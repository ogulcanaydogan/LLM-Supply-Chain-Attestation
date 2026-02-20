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

gh_retry() {
  local attempt
  for attempt in 1 2 3 4 5; do
    if "$@"; then
      return 0
    fi
    sleep $((attempt * 2))
  done
  return 1
}

require_cmd gh
require_cmd jq

RUNS="${RUNS:-${1:-10}}"
REF="${REF:-main}"
WORKFLOW="${WORKFLOW:-ci-attest-verify.yml}"
POLL_SECONDS="${POLL_SECONDS:-8}"
MAX_POLLS="${MAX_POLLS:-120}"
STOP_ON_FAILURE="${STOP_ON_FAILURE:-1}"

if ! [[ "${RUNS}" =~ ^[0-9]+$ ]] || [[ "${RUNS}" -lt 1 ]]; then
  echo "error: RUNS must be a positive integer" >&2
  exit 1
fi

TS="$(date -u +"%Y%m%dT%H%M%SZ")"
OUT_DIR=".llmsa/public-footprint/${TS}"
mkdir -p "${OUT_DIR}"

RESULTS_JSON="${OUT_DIR}/ci-burst-results.json"
RESULTS_MD="${OUT_DIR}/ci-burst-results.md"
echo '[]' > "${RESULTS_JSON}"

append_result() {
  local tmp
  tmp="$(mktemp)"
  jq \
    --argjson idx "${1}" \
    --arg run_id "${2}" \
    --arg status "${3}" \
    --arg conclusion "${4}" \
    --arg url "${5}" \
    --arg note "${6}" \
    '. + [{
      index: $idx,
      run_id: $run_id,
      status: $status,
      conclusion: $conclusion,
      url: $url,
      note: $note
    }]' "${RESULTS_JSON}" > "${tmp}"
  mv "${tmp}" "${RESULTS_JSON}"
}

echo "starting ci burst: runs=${RUNS} workflow=${WORKFLOW} ref=${REF}"

for ((i=1; i<=RUNS; i++)); do
  echo "=== burst run ${i}/${RUNS} ==="

  if ! gh_retry gh workflow run "${WORKFLOW}" --ref "${REF}" >/dev/null 2>&1; then
    append_result "${i}" "" "dispatch_failed" "" "" "failed to dispatch workflow run"
    if [[ "${STOP_ON_FAILURE}" == "1" ]]; then
      break
    fi
    continue
  fi

  run_id=""
  for _ in {1..25}; do
    run_id="$(gh_retry gh run list --workflow "${WORKFLOW}" --limit 1 --json databaseId -q '.[0].databaseId' 2>/dev/null || true)"
    if [[ -n "${run_id}" && "${run_id}" != "null" ]]; then
      break
    fi
    sleep 2
  done

  if [[ -z "${run_id}" || "${run_id}" == "null" ]]; then
    append_result "${i}" "" "id_resolution_failed" "" "" "could not resolve workflow run id"
    if [[ "${STOP_ON_FAILURE}" == "1" ]]; then
      break
    fi
    continue
  fi

  status=""
  conclusion=""
  url="$(gh_retry gh run view "${run_id}" --json url -q '.url' 2>/dev/null || true)"

  for ((p=1; p<=MAX_POLLS; p++)); do
    status="$(gh_retry gh run view "${run_id}" --json status -q '.status' 2>/dev/null || true)"
    conclusion="$(gh_retry gh run view "${run_id}" --json conclusion -q '.conclusion // ""' 2>/dev/null || true)"
    if [[ "${status}" == "completed" ]]; then
      break
    fi
    sleep "${POLL_SECONDS}"
  done

  if [[ "${status}" != "completed" ]]; then
    append_result "${i}" "${run_id}" "${status:-unknown}" "${conclusion}" "${url}" "poll timeout before completion"
    if [[ "${STOP_ON_FAILURE}" == "1" ]]; then
      break
    fi
    continue
  fi

  append_result "${i}" "${run_id}" "${status}" "${conclusion}" "${url}" ""
  echo "run_id=${run_id} conclusion=${conclusion}"

  if [[ "${STOP_ON_FAILURE}" == "1" && "${conclusion}" != "success" ]]; then
    break
  fi
done

jq -r '
  def pct($ok; $total):
    if $total == 0 then 0 else ((($ok * 10000.0) / $total) | round / 100) end;
  . as $runs
  | ($runs | length) as $total
  | ($runs | map(select(.conclusion == "success")) | length) as $success
  | ($total - $success) as $non_success
  | {
      total_runs_observed: $total,
      successful_runs: $success,
      non_success_runs: $non_success,
      success_rate_percent: pct($success; $total)
    }
' "${RESULTS_JSON}" > "${OUT_DIR}/ci-burst-summary.json"

{
  echo "# CI Burst Summary"
  echo
  echo "- Generated (UTC): \`$(date -u +"%Y-%m-%dT%H:%M:%SZ")\`"
  echo "- Workflow: \`${WORKFLOW}\`"
  echo "- Ref: \`${REF}\`"
  echo
  echo "## Aggregate"
  echo
  jq -r '
    "| Metric | Value |",
    "|---|---:|",
    "| Total observed runs | \(.total_runs_observed) |",
    "| Successful runs | \(.successful_runs) |",
    "| Non-success runs | \(.non_success_runs) |",
    "| Success rate | \(.success_rate_percent)% |"
  ' "${OUT_DIR}/ci-burst-summary.json"
  echo
  echo "## Run Results"
  echo
  echo "| # | Run ID | Status | Conclusion | URL | Note |"
  echo "|---|---:|---|---|---|---|"
  jq -r '
    .[]
    | "| \(.index) | \(.run_id // "") | \(.status // "") | \(.conclusion // "") | \(.url // "") | \(.note // "") |"
  ' "${RESULTS_JSON}"
} > "${RESULTS_MD}"

echo "ci burst summary json: ${OUT_DIR}/ci-burst-summary.json"
echo "ci burst summary markdown: ${OUT_DIR}/ci-burst-results.md"

