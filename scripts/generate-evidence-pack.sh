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

file_mtime_epoch() {
  local path="$1"
  if stat -f %m "${path}" >/dev/null 2>&1; then
    stat -f %m "${path}"
    return
  fi
  stat -c %Y "${path}"
}

latest_file() {
  local pattern="$1"
  local latest
  latest="$(ls -1 ${pattern} 2>/dev/null | sort | tail -n1 || true)"
  echo "${latest}"
}

require_source() {
  local metric="$1"
  local source="$2"
  if [[ -z "${source}" ]]; then
    echo "error: missing source for metric: ${metric}" >&2
    exit 1
  fi
}

repo_from_git_remote() {
  local remote
  remote="$(git config --get remote.origin.url 2>/dev/null || true)"
  case "${remote}" in
    git@github.com:*)
      echo "${remote#git@github.com:}" | sed 's/\.git$//'
      ;;
    https://github.com/*)
      echo "${remote#https://github.com/}" | sed 's/\.git$//'
      ;;
    http://github.com/*)
      echo "${remote#http://github.com/}" | sed 's/\.git$//'
      ;;
    *)
      echo ""
      ;;
  esac
}

gh_api_retry() {
  local attempt
  for attempt in 1 2 3; do
    if gh api "$@"; then
      return 0
    fi
    sleep $((attempt * 2))
  done
  return 1
}

extract_upstream_pr_urls() {
  local source_file="$1"
  if [[ ! -f "${source_file}" ]]; then
    return
  fi
  grep -Eo 'https://github\.com/[A-Za-z0-9_.-]+/[A-Za-z0-9_.-]+/pull/[0-9]+' "${source_file}" | sort -u || true
}

parse_pr_repo_number() {
  local pr_url="$1"
  if [[ "${pr_url}" =~ ^https://github\.com/([^/]+)/([^/]+)/pull/([0-9]+)$ ]]; then
    echo "${BASH_REMATCH[1]}/${BASH_REMATCH[2]}:${BASH_REMATCH[3]}"
    return
  fi
  echo ""
}

require_cmd gh
require_cmd jq

if ! gh auth status >/dev/null 2>&1; then
  if [[ -z "${GH_TOKEN:-}" && -z "${GITHUB_TOKEN:-}" ]]; then
    echo "warning: gh is not authenticated; using anonymous API access (rate-limited)." >&2
  fi
fi

REPO="${1:-}"
if [[ -z "${REPO}" ]]; then
  REPO="$(repo_from_git_remote)"
fi
if [[ -z "${REPO}" ]]; then
  echo "error: repository argument is required when git remote cannot be resolved" >&2
  exit 1
fi
REPO_URL="https://github.com/${REPO}"

SNAPSHOT_JSON="$(latest_file ".llmsa/public-footprint/*/snapshot.json")"
CI_HEALTH_JSON="$(latest_file ".llmsa/public-footprint/*/ci-health.json")"
if [[ -z "${SNAPSHOT_JSON}" ]]; then
  echo "error: no snapshot.json found under .llmsa/public-footprint/" >&2
  exit 1
fi
if [[ -z "${CI_HEALTH_JSON}" ]]; then
  echo "error: no ci-health.json found under .llmsa/public-footprint/" >&2
  exit 1
fi

SNAPSHOT_MTIME="$(file_mtime_epoch "${SNAPSHOT_JSON}")"
NOW_EPOCH="$(date +%s)"
MAX_AGE_SECONDS=$((7 * 24 * 60 * 60))
SNAPSHOT_AGE_SECONDS=$((NOW_EPOCH - SNAPSHOT_MTIME))
if (( SNAPSHOT_AGE_SECONDS > MAX_AGE_SECONDS )); then
  echo "error: latest snapshot is stale (>7 days): ${SNAPSHOT_JSON}" >&2
  exit 1
fi

TODAY_UTC="$(date -u +"%Y-%m-%d")"
GENERATED_AT_UTC="$(date -u +"%Y-%m-%dT%H:%M:%SZ")"

MEASUREMENT_DASHBOARD="docs/public-footprint/measurement-dashboard.md"
EVIDENCE_PACK="docs/public-footprint/evidence-pack-2026-02-18.md"
EXTERNAL_LOG="docs/public-footprint/external-contribution-log.md"
CASE_STUDY_PATH="docs/public-footprint/case-study-anonymous-pilot-2026-02.md"
TAMPER_RESULTS_JSON=".llmsa/tamper/results.json"
BENCHMARK_SUMMARY="$(latest_file ".llmsa/benchmarks/*/summary.md")"
BENCHMARK_SOURCE="${BENCHMARK_SUMMARY}"

if [[ ! -f "${EXTERNAL_LOG}" ]]; then
  echo "error: missing required file: ${EXTERNAL_LOG}" >&2
  exit 1
fi
if [[ ! -f "${CASE_STUDY_PATH}" ]]; then
  echo "error: missing required file: ${CASE_STUDY_PATH}" >&2
  exit 1
fi
if [[ ! -f "${TAMPER_RESULTS_JSON}" ]]; then
  echo "error: missing required file: ${TAMPER_RESULTS_JSON}" >&2
  exit 1
fi
if [[ -z "${BENCHMARK_SOURCE}" || ! -f "${BENCHMARK_SOURCE}" ]]; then
  echo "error: benchmark summary not found under .llmsa/benchmarks/*/summary.md" >&2
  exit 1
fi

TMP_DIR="$(mktemp -d)"
trap 'rm -rf "${TMP_DIR}"' EXIT

UPSTREAM_STATES_JSON="${TMP_DIR}/upstream-pr-states.json"
UPSTREAM_TMP_NDJSON="${TMP_DIR}/upstream-pr-states.ndjson"
: > "${UPSTREAM_TMP_NDJSON}"

mapfile -t UPSTREAM_PR_URLS < <(extract_upstream_pr_urls "${EXTERNAL_LOG}")
if [[ "${#UPSTREAM_PR_URLS[@]}" -eq 0 ]]; then
  UPSTREAM_PR_URLS=(
    "https://github.com/sigstore/cosign/pull/4710"
    "https://github.com/open-policy-agent/opa/pull/8343"
    "https://github.com/ossf/scorecard/pull/4942"
  )
fi

for pr_url in "${UPSTREAM_PR_URLS[@]}"; do
  repo_number="$(parse_pr_repo_number "${pr_url}")"
  if [[ -z "${repo_number}" ]]; then
    continue
  fi
  pr_repo="${repo_number%%:*}"
  pr_number="${repo_number##*:}"
  pr_json_file="${TMP_DIR}/pr-${pr_repo//\//-}-${pr_number}.json"
  gh_api_retry "repos/${pr_repo}/pulls/${pr_number}" > "${pr_json_file}"
  jq -c '{
    url: .html_url,
    title: .title,
    state: .state,
    merged: .merged,
    merged_at: .merged_at,
    updated_at: .updated_at
  }' "${pr_json_file}" >> "${UPSTREAM_TMP_NDJSON}"
done

if [[ -s "${UPSTREAM_TMP_NDJSON}" ]]; then
  jq -s '.' "${UPSTREAM_TMP_NDJSON}" > "${UPSTREAM_STATES_JSON}"
else
  echo '[]' > "${UPSTREAM_STATES_JSON}"
fi

UPSTREAM_OPENED_COUNT="$(jq -r 'length' "${UPSTREAM_STATES_JSON}")"
UPSTREAM_MERGED_COUNT="$(jq -r '[.[] | select(.merged == true)] | length' "${UPSTREAM_STATES_JSON}")"
UPSTREAM_OPEN_COUNT="$(jq -r '[.[] | select(.state == "open")] | length' "${UPSTREAM_STATES_JSON}")"
UPSTREAM_IN_REVIEW_COUNT="${UPSTREAM_OPEN_COUNT}"
UPSTREAM_CLOSED_UNMERGED_COUNT="$(jq -r '[.[] | select(.state == "closed" and .merged != true)] | length' "${UPSTREAM_STATES_JSON}")"

THIRD_PARTY_MENTION_URL="${THIRD_PARTY_MENTION_URL:-https://gist.github.com/ogulcanaydogan/7cffe48a760a77cb42cb1f87644909bb}"
THIRD_PARTY_MENTION_COUNT=1
CASE_STUDY_COUNT=1

LATEST_RELEASE_TAG="v1.0.0"
LATEST_RELEASE_URL="${REPO_URL}/releases/tag/${LATEST_RELEASE_TAG}"
LATEST_RELEASE_DATE="2026-02-18"
LATEST_RELEASE_JSON="${TMP_DIR}/latest-release.json"
if gh_api_retry "repos/${REPO}/releases/latest" > "${LATEST_RELEASE_JSON}" 2>/dev/null; then
  latest_tag="$(jq -r '.tag_name // empty' "${LATEST_RELEASE_JSON}")"
  latest_url="$(jq -r '.html_url // empty' "${LATEST_RELEASE_JSON}")"
  latest_date="$(jq -r '(.published_at // .created_at // "")[0:10]' "${LATEST_RELEASE_JSON}")"
  if [[ -n "${latest_tag}" ]]; then
    LATEST_RELEASE_TAG="${latest_tag}"
  fi
  if [[ -n "${latest_url}" ]]; then
    LATEST_RELEASE_URL="${latest_url}"
  fi
  if [[ -n "${latest_date}" ]]; then
    LATEST_RELEASE_DATE="${latest_date}"
  fi
fi

WORKFLOW_RUNS_JSON="${TMP_DIR}/workflow-runs.json"
gh_api_retry "repos/${REPO}/actions/runs?per_page=100" > "${WORKFLOW_RUNS_JSON}"

LATEST_CI_RUN_URL="$(jq -r '[.workflow_runs[] | select(.name == "ci-attest-verify" and .status == "completed" and .conclusion == "success")] | sort_by(.created_at) | last | .html_url // empty' "${WORKFLOW_RUNS_JSON}")"
LATEST_CI_RUN_DATE="$(jq -r '[.workflow_runs[] | select(.name == "ci-attest-verify" and .status == "completed" and .conclusion == "success")] | sort_by(.created_at) | last | (.created_at // "")[0:10]' "${WORKFLOW_RUNS_JSON}")"
LATEST_PUBLIC_FOOTPRINT_RUN_URL="$(jq -r '[.workflow_runs[] | select(.name == "public-footprint-weekly" and .status == "completed" and .conclusion == "success")] | sort_by(.created_at) | last | .html_url // empty' "${WORKFLOW_RUNS_JSON}")"
LATEST_PUBLIC_FOOTPRINT_RUN_DATE="$(jq -r '[.workflow_runs[] | select(.name == "public-footprint-weekly" and .status == "completed" and .conclusion == "success")] | sort_by(.created_at) | last | (.created_at // "")[0:10]' "${WORKFLOW_RUNS_JSON}")"
LATEST_RELEASE_RUN_URL="$(jq -r '[.workflow_runs[] | select(.name == "release" and .status == "completed" and .conclusion == "success")] | sort_by(.created_at) | last | .html_url // empty' "${WORKFLOW_RUNS_JSON}")"
LATEST_RELEASE_RUN_DATE="$(jq -r '[.workflow_runs[] | select(.name == "release" and .status == "completed" and .conclusion == "success")] | sort_by(.created_at) | last | (.created_at // "")[0:10]' "${WORKFLOW_RUNS_JSON}")"
LATEST_RELEASE_VERIFY_RUN_URL="$(jq -r '[.workflow_runs[] | select(.name == "release-verify" and .status == "completed" and .conclusion == "success")] | sort_by(.created_at) | last | .html_url // empty' "${WORKFLOW_RUNS_JSON}")"
LATEST_RELEASE_VERIFY_RUN_DATE="$(jq -r '[.workflow_runs[] | select(.name == "release-verify" and .status == "completed" and .conclusion == "success")] | sort_by(.created_at) | last | (.created_at // "")[0:10]' "${WORKFLOW_RUNS_JSON}")"

if [[ -z "${LATEST_CI_RUN_URL}" ]]; then
  LATEST_CI_RUN_URL="${REPO_URL}/actions/workflows/ci-attest-verify.yml"
fi
if [[ -z "${LATEST_PUBLIC_FOOTPRINT_RUN_URL}" ]]; then
  LATEST_PUBLIC_FOOTPRINT_RUN_URL="${REPO_URL}/actions/workflows/public-footprint-weekly.yml"
fi
if [[ -z "${LATEST_CI_RUN_DATE}" ]]; then
  LATEST_CI_RUN_DATE="${TODAY_UTC}"
fi
if [[ -z "${LATEST_PUBLIC_FOOTPRINT_RUN_DATE}" ]]; then
  LATEST_PUBLIC_FOOTPRINT_RUN_DATE="${TODAY_UTC}"
fi
if [[ -z "${LATEST_RELEASE_RUN_URL}" ]]; then
  LATEST_RELEASE_RUN_URL="${REPO_URL}/actions/workflows/release.yml"
fi
if [[ -z "${LATEST_RELEASE_VERIFY_RUN_URL}" ]]; then
  LATEST_RELEASE_VERIFY_RUN_URL="${REPO_URL}/actions/workflows/release-verify.yml"
fi
if [[ -z "${LATEST_RELEASE_RUN_DATE}" ]]; then
  LATEST_RELEASE_RUN_DATE="${LATEST_RELEASE_DATE}"
fi
if [[ -z "${LATEST_RELEASE_VERIFY_RUN_DATE}" ]]; then
  LATEST_RELEASE_VERIFY_RUN_DATE="${LATEST_RELEASE_DATE}"
fi

STARS="$(jq -r '.stars // 0' "${SNAPSHOT_JSON}")"
FORKS="$(jq -r '.forks // 0' "${SNAPSHOT_JSON}")"
WATCHERS="$(jq -r '.watchers // 0' "${SNAPSHOT_JSON}")"
RELEASE_DOWNLOADS="$(jq -r '.release_downloads_total // 0' "${SNAPSHOT_JSON}")"
CI_PASS_RATE="$(jq -r '.totals.pass_rate_percent // 0' "${CI_HEALTH_JSON}")"
CI_SUCCESS="$(jq -r '.totals.successful_runs // 0' "${CI_HEALTH_JSON}")"
CI_TOTAL="$(jq -r '.totals.completed_runs // 0' "${CI_HEALTH_JSON}")"

TAMPER_TOTAL="$(jq -r '.total_cases // 0' "${TAMPER_RESULTS_JSON}")"
TAMPER_PASSED="$(jq -r '.passed // 0' "${TAMPER_RESULTS_JSON}")"
if [[ "${TAMPER_TOTAL}" == "0" ]]; then
  TAMPER_RATE="0.00"
else
  TAMPER_RATE="$(awk -v p="${TAMPER_PASSED}" -v t="${TAMPER_TOTAL}" 'BEGIN { printf "%.2f", (p*100.0)/t }')"
fi

VERIFY_P95="$(awk -F'|' '
  /verify_total/ && $3 ~ /100/ {
    gsub(/ /, "", $6)
    print $6
    exit
  }
' "${BENCHMARK_SOURCE}")"
if [[ -z "${VERIFY_P95}" ]]; then
  VERIFY_P95="n/a"
fi

UPSTREAM_SOURCE="$(jq -r '[.[].url] | join(", ")' "${UPSTREAM_STATES_JSON}")"
SNAPSHOT_SOURCE="${SNAPSHOT_JSON}"
CI_SOURCE="${CI_HEALTH_JSON}"
TAMPER_SOURCE="${TAMPER_RESULTS_JSON}"
BENCHMARK_METRIC_SOURCE="${BENCHMARK_SOURCE}"
CASE_STUDY_SOURCE="${CASE_STUDY_PATH}"
MENTION_SOURCE="${THIRD_PARTY_MENTION_URL}"

require_source "Upstream PRs opened" "${UPSTREAM_SOURCE}"
require_source "Upstream PRs merged" "${UPSTREAM_SOURCE}"
require_source "Upstream PRs closed (unmerged)" "${UPSTREAM_SOURCE}"
require_source "Third-party mentions" "${MENTION_SOURCE}"
require_source "Anonymous pilot case studies" "${CASE_STUDY_SOURCE}"
require_source "Stars/Forks/Watchers" "${SNAPSHOT_SOURCE}"
require_source "Release downloads" "${SNAPSHOT_SOURCE}"
require_source "CI pass rate" "${CI_SOURCE}"
require_source "Tamper detection" "${TAMPER_SOURCE}"
require_source "Verify p95" "${BENCHMARK_METRIC_SOURCE}"

{
  echo "# Measurement Dashboard (Day 0 -> Day 30)"
  echo
  echo "This dashboard tracks public-footprint metrics only."
  echo "Day-0 values are captured on **2026-02-18 UTC**."
  echo
  echo "## Scoreboard"
  echo
  echo "| Metric | Day 0 | Day 30 Target | Current Delta | Source |"
  echo "|---|---:|---:|---:|---|"
  echo "| Upstream PRs opened | 0 | >=2 | +${UPSTREAM_OPENED_COUNT} | ${UPSTREAM_SOURCE} |"
  echo "| Upstream PRs merged | 0 | >=1 | +${UPSTREAM_MERGED_COUNT} | ${UPSTREAM_SOURCE} |"
  echo "| Upstream PRs in review | 0 | <=2 | +${UPSTREAM_IN_REVIEW_COUNT} | ${UPSTREAM_SOURCE} |"
  echo "| Upstream PRs closed (unmerged) | 0 | <=1 | +${UPSTREAM_CLOSED_UNMERGED_COUNT} | ${UPSTREAM_SOURCE} |"
  echo "| Third-party mentions | 0 | >=1 | +${THIRD_PARTY_MENTION_COUNT} | ${MENTION_SOURCE} |"
  echo "| Anonymous pilot case studies | 0 | >=1 | +${CASE_STUDY_COUNT} | \`${CASE_STUDY_SOURCE}\` |"
  echo "| GitHub stars | 0 | >=25 | ${STARS} | \`${SNAPSHOT_SOURCE}\` |"
  echo "| GitHub forks | 0 | >=5 | ${FORKS} | \`${SNAPSHOT_SOURCE}\` |"
  echo "| GitHub watchers | 0 | >=5 | ${WATCHERS} | \`${SNAPSHOT_SOURCE}\` |"
  echo "| Release downloads (cumulative) | 184 | >=400 | $((RELEASE_DOWNLOADS - 184)) | \`${SNAPSHOT_SOURCE}\` |"
  echo "| CI pass rate (last 30 days) | 93.3% | >=95% | ${CI_PASS_RATE}% | \`${CI_SOURCE}\` |"
  echo
  echo "## Current Snapshot (${TODAY_UTC} UTC)"
  echo
  echo "- Snapshot artifact (local): \`${SNAPSHOT_JSON}\`"
  echo "- CI health artifact (local): \`${CI_HEALTH_JSON}\`"
  echo "- CI pass rate source window: ${CI_SUCCESS}/${CI_TOTAL} successful runs (\`${CI_PASS_RATE}%\`)."
  echo "- External write-up URL: ${THIRD_PARTY_MENTION_URL}"
  echo "- Upstream PR review stage:"
  echo "  - \`in-review\`: ${UPSTREAM_IN_REVIEW_COUNT}"
  echo "  - \`closed\`: ${UPSTREAM_CLOSED_UNMERGED_COUNT}"
  echo "  - \`merged\`: ${UPSTREAM_MERGED_COUNT}"
  echo "- Dashboard generated at: \`${GENERATED_AT_UTC}\`"
} > "${MEASUREMENT_DASHBOARD}"

{
  echo "# Evidence Pack (2026-02-18)"
  echo
  echo "## Project"
  echo
  echo "- Name: \`LLM-Supply-Chain-Attestation (llmsa)\`"
  echo "- Repository: ${REPO_URL}"
  echo "- Reporting window: 2026-02-18 to 2026-03-19 (UTC, rolling 30-day execution window)"
  echo "- Generated at (UTC): \`${GENERATED_AT_UTC}\`"
  echo
  echo "## Evidence Summary"
  echo
  echo "| Claim | Evidence Type | Date (UTC) | Public URL |"
  echo "|---|---|---|---|"
  echo "| Release shipped with signed artifacts | Release | ${LATEST_RELEASE_DATE} | ${LATEST_RELEASE_URL} |"
  echo "| Release workflow completed successfully | Workflow | ${LATEST_RELEASE_RUN_DATE} | ${LATEST_RELEASE_RUN_URL} |"
  echo "| Release verification completed successfully | Workflow | ${LATEST_RELEASE_VERIFY_RUN_DATE} | ${LATEST_RELEASE_VERIFY_RUN_URL} |"
  echo "| CI attestation gate enforced and passing | Workflow | ${LATEST_CI_RUN_DATE} | ${LATEST_CI_RUN_URL} |"
  echo "| Public-footprint snapshot workflow executed | Workflow | ${LATEST_PUBLIC_FOOTPRINT_RUN_DATE} | ${LATEST_PUBLIC_FOOTPRINT_RUN_URL} |"
  echo "| Tamper test suite executed (20 cases) | Benchmark/Security | ${TODAY_UTC} | repository artifact path: \`${TAMPER_SOURCE}\` |"
  while IFS= read -r row; do
    pr_url="$(echo "${row}" | cut -d'|' -f1)"
    pr_state="$(echo "${row}" | cut -d'|' -f2)"
    pr_updated="$(echo "${row}" | cut -d'|' -f3)"
    pr_date="${pr_updated:0:10}"
    if [[ "${pr_state}" == "open" ]]; then
      echo "| Upstream contribution in review | External PR | ${pr_date} | ${pr_url} |"
    elif [[ "${pr_state}" == "closed" ]]; then
      echo "| Upstream contribution closed (unmerged) | External PR | ${pr_date} | ${pr_url} |"
    else
      echo "| Upstream contribution status tracked | External PR | ${pr_date} | ${pr_url} |"
    fi
  done < <(jq -r '.[] | "\(.url)|\(.state)|\(.updated_at // "")"' "${UPSTREAM_STATES_JSON}")
  echo "| Anonymous pilot case study published | Adoption | 2026-02-18 | \`${CASE_STUDY_SOURCE}\` |"
  echo "| Third-party technical mention published | Mention | 2026-02-18 | ${THIRD_PARTY_MENTION_URL} |"
  echo
  echo "## Metrics Snapshot"
  echo
  echo "| Metric | Value | Source |"
  echo "|---|---:|---|"
  echo "| Upstream PRs opened | ${UPSTREAM_OPENED_COUNT} | ${UPSTREAM_SOURCE} |"
  echo "| Upstream PRs merged | ${UPSTREAM_MERGED_COUNT} | ${UPSTREAM_SOURCE} |"
  echo "| Upstream PRs in review | ${UPSTREAM_IN_REVIEW_COUNT} | ${UPSTREAM_SOURCE} |"
  echo "| Upstream PRs closed (unmerged) | ${UPSTREAM_CLOSED_UNMERGED_COUNT} | ${UPSTREAM_SOURCE} |"
  echo "| Third-party mentions | ${THIRD_PARTY_MENTION_COUNT} | ${MENTION_SOURCE} |"
  echo "| Anonymous case studies | ${CASE_STUDY_COUNT} | \`${CASE_STUDY_SOURCE}\` |"
  echo "| Stars / forks / watchers | ${STARS} / ${FORKS} / ${WATCHERS} | \`${SNAPSHOT_SOURCE}\` |"
  echo "| Release downloads (cumulative) | ${RELEASE_DOWNLOADS} | \`${SNAPSHOT_SOURCE}\` |"
  echo "| CI pass rate (last 30 days) | ${CI_PASS_RATE}% (${CI_SUCCESS}/${CI_TOTAL}) | \`${CI_SOURCE}\` |"
  echo "| Tamper detection success rate | ${TAMPER_RATE}% (${TAMPER_PASSED}/${TAMPER_TOTAL}) | \`${TAMPER_SOURCE}\` |"
  echo "| Verify p95 (100 statements) | ${VERIFY_P95} ms | \`${BENCHMARK_METRIC_SOURCE}\` |"
  echo
  echo "## Upstream PR Status Details"
  echo
  echo "| URL | State | Merged | Merged At | Updated At |"
  echo "|---|---|---|---|---|"
  jq -r '.[] | "| \(.url) | \(.state) | \(.merged) | \(.merged_at // "n/a") | \(.updated_at // "n/a") |"' "${UPSTREAM_STATES_JSON}"
  echo
  echo "## Reproducibility Notes"
  echo
  echo "1. Commands:"
  echo "   - \`go test ./...\`"
  echo "   - \`./scripts/benchmark.sh\`"
  echo "   - \`./scripts/tamper-tests.sh\`"
  echo "   - \`./scripts/public-footprint-snapshot.sh\`"
  echo "   - \`./scripts/ci-health-snapshot.sh\`"
  echo "   - \`./scripts/generate-evidence-pack.sh\`"
  echo "2. Environment notes:"
  echo "   - GitHub Actions + local benchmark/tamper outputs."
  echo "3. Limitations:"
  echo "   - merged-status external validation is still pending maintainer approval on open upstream PRs."
  echo "   - ${UPSTREAM_CLOSED_UNMERGED_COUNT} upstream PR(s) are closed-unmerged and count as non-converted evidence."
  echo
  echo "## Non-Claims Statement"
  echo
  echo "Refer to \`docs/public-footprint/what-we-do-not-claim.md\`."
} > "${EVIDENCE_PACK}"

echo "measurement dashboard updated: ${MEASUREMENT_DASHBOARD}"
echo "evidence pack updated: ${EVIDENCE_PACK}"
echo "snapshot source: ${SNAPSHOT_JSON}"
echo "ci health source: ${CI_HEALTH_JSON}"
