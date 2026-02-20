#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "${ROOT}"

require_cmd() {
  if ! command -v "$1" >/dev/null 2>&1; then
    echo "error: required command not found: $1" >&2
    exit 1
  fi
}

latest_file() {
  local pattern="$1"
  local latest
  latest="$(ls -1 ${pattern} 2>/dev/null | sort | tail -n1 || true)"
  echo "${latest}"
}

usage() {
  cat <<'EOF'
Usage: scripts/ci-passrate-forecast.sh [options]

Options:
  --ci-health <path>         Path to ci-health.json (default: latest under .llmsa/public-footprint/*/ci-health.json)
  --target <percent>         Target pass rate percent (default: 95)
  --assume-success-runs <n>  Hypothetical additional all-success runs (default: 0)
  --out-dir <path>           Output directory (default: .llmsa/public-footprint/<timestamp>)
  -h, --help                 Show help
EOF
}

require_cmd jq
require_cmd python3

CI_HEALTH_JSON=""
TARGET_PERCENT="95"
ASSUME_SUCCESS_RUNS="0"
OUT_DIR=""

while [[ $# -gt 0 ]]; do
  case "$1" in
    --ci-health)
      CI_HEALTH_JSON="${2:-}"
      shift 2
      ;;
    --target)
      TARGET_PERCENT="${2:-}"
      shift 2
      ;;
    --assume-success-runs)
      ASSUME_SUCCESS_RUNS="${2:-}"
      shift 2
      ;;
    --out-dir)
      OUT_DIR="${2:-}"
      shift 2
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      echo "error: unknown option: $1" >&2
      usage
      exit 1
      ;;
  esac
done

if [[ -z "${CI_HEALTH_JSON}" ]]; then
  CI_HEALTH_JSON="$(latest_file ".llmsa/public-footprint/*/ci-health.json")"
fi
if [[ -z "${CI_HEALTH_JSON}" || ! -f "${CI_HEALTH_JSON}" ]]; then
  echo "error: ci-health json not found" >&2
  exit 1
fi

if [[ ! "${TARGET_PERCENT}" =~ ^[0-9]+([.][0-9]+)?$ ]]; then
  echo "error: --target must be numeric" >&2
  exit 1
fi
if [[ ! "${ASSUME_SUCCESS_RUNS}" =~ ^[0-9]+$ ]]; then
  echo "error: --assume-success-runs must be a non-negative integer" >&2
  exit 1
fi

if [[ -z "${OUT_DIR}" ]]; then
  TS="$(date -u +"%Y%m%dT%H%M%SZ")"
  OUT_DIR=".llmsa/public-footprint/${TS}"
fi
mkdir -p "${OUT_DIR}"

CURRENT_SUCCESS="$(jq -r '.totals.successful_runs // 0' "${CI_HEALTH_JSON}")"
CURRENT_TOTAL="$(jq -r '.totals.completed_runs // 0' "${CI_HEALTH_JSON}")"
CURRENT_FAIL="$(jq -r '.totals.failed_runs // 0' "${CI_HEALTH_JSON}")"
GENERATED_AT="$(date -u +"%Y-%m-%dT%H:%M:%SZ")"

FORECAST_JSON="${OUT_DIR}/ci-passrate-forecast.json"
FORECAST_MD="${OUT_DIR}/ci-passrate-forecast.md"

python3 - "${CURRENT_SUCCESS}" "${CURRENT_TOTAL}" "${CURRENT_FAIL}" "${TARGET_PERCENT}" "${ASSUME_SUCCESS_RUNS}" "${CI_HEALTH_JSON}" "${FORECAST_JSON}" "${GENERATED_AT}" <<'PY'
import json
import math
import sys

current_success = int(sys.argv[1])
current_total = int(sys.argv[2])
current_fail = int(sys.argv[3])
target_percent = float(sys.argv[4])
assume_success_runs = int(sys.argv[5])
source_path = sys.argv[6]
out_path = sys.argv[7]
generated_at = sys.argv[8]

target_ratio = target_percent / 100.0
current_ratio = (current_success / current_total) if current_total > 0 else 0.0

def min_additional_successes(p, t, target):
    if t == 0:
        return 0
    if p / t >= target:
        return 0
    # Solve (p + n) / (t + n) >= target for integer n.
    n = (target * t - p) / (1 - target)
    return max(0, math.ceil(n))

required_success_runs = min_additional_successes(current_success, current_total, target_ratio)

projected_success = current_success + assume_success_runs
projected_total = current_total + assume_success_runs
projected_ratio = (projected_success / projected_total) if projected_total > 0 else 0.0
remaining_after_assumption = min_additional_successes(projected_success, projected_total, target_ratio)

if required_success_runs == 0:
    confidence_note = "Target already met for the current rolling window."
elif remaining_after_assumption == 0:
    confidence_note = (
        "If all assumed runs succeed and no new failures occur, the target is reachable in this step. "
        "This is conservative; aging out old failures can reduce required runs."
    )
else:
    confidence_note = (
        "Requires additional all-success runs with no new failures. "
        "Estimate is conservative because old failures naturally aging out can reduce required runs."
    )

result = {
    "generated_at_utc": generated_at,
    "source_ci_health_json": source_path,
    "target_percent": target_percent,
    "current": {
      "successful_runs": current_success,
      "completed_runs": current_total,
      "failed_runs": current_fail,
      "pass_rate_percent": round(current_ratio * 100.0, 2)
    },
    "forecast": {
      "assume_success_runs": assume_success_runs,
      "required_additional_success_runs_no_new_failures": required_success_runs,
      "projected_pass_rate_percent_after_assumption": round(projected_ratio * 100.0, 2),
      "remaining_success_runs_after_assumption": remaining_after_assumption
    },
    "confidence_note": confidence_note
}

with open(out_path, "w", encoding="utf-8") as f:
    json.dump(result, f, indent=2)
PY

{
  echo "# CI Pass-Rate Forecast"
  echo
  echo "- Generated (UTC): \`${GENERATED_AT}\`"
  echo "- Source: \`${CI_HEALTH_JSON}\`"
  echo
  echo "## Current"
  echo
  echo "| Metric | Value |"
  echo "|---|---:|"
  echo "| Pass rate | $(jq -r '.current.pass_rate_percent' "${FORECAST_JSON}")% |"
  echo "| Successful runs | $(jq -r '.current.successful_runs' "${FORECAST_JSON}") |"
  echo "| Completed runs | $(jq -r '.current.completed_runs' "${FORECAST_JSON}") |"
  echo "| Failed runs | $(jq -r '.current.failed_runs' "${FORECAST_JSON}") |"
  echo
  echo "## Forecast"
  echo
  echo "| Metric | Value |"
  echo "|---|---:|"
  echo "| Target pass rate | $(jq -r '.target_percent' "${FORECAST_JSON}")% |"
  echo "| Assumed extra success runs | $(jq -r '.forecast.assume_success_runs' "${FORECAST_JSON}") |"
  echo "| Required additional successes (no new failures) | $(jq -r '.forecast.required_additional_success_runs_no_new_failures' "${FORECAST_JSON}") |"
  echo "| Projected pass rate after assumption | $(jq -r '.forecast.projected_pass_rate_percent_after_assumption' "${FORECAST_JSON}")% |"
  echo "| Remaining successes after assumption | $(jq -r '.forecast.remaining_success_runs_after_assumption' "${FORECAST_JSON}") |"
  echo
  echo "## Confidence Note"
  echo
  jq -r '.confidence_note' "${FORECAST_JSON}"
} > "${FORECAST_MD}"

echo "ci forecast json: ${FORECAST_JSON}"
echo "ci forecast markdown: ${FORECAST_MD}"
