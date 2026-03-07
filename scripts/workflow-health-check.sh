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

if [[ -n "${WORKFLOWS:-}" ]]; then
  IFS=',' read -r -a WORKFLOW_LIST <<<"${WORKFLOWS}"
else
  WORKFLOW_LIST=(
    "ci-attest-verify"
    "release"
    "release-verify"
    "public-footprint-weekly"
  )
fi

if [[ -n "${MISSING_ALLOWED_WORKFLOWS:-}" ]]; then
  IFS=',' read -r -a MISSING_ALLOWED_LIST <<<"${MISSING_ALLOWED_WORKFLOWS}"
else
  MISSING_ALLOWED_LIST=(
    "release"
    "release-verify"
  )
fi

MAX_RUNS="${MAX_RUNS:-100}"
FAIL_ON_UNHEALTHY="${FAIL_ON_UNHEALTHY:-true}"
TS="$(date -u +"%Y%m%dT%H%M%SZ")"
GENERATED_AT="$(date -u +"%Y-%m-%dT%H:%M:%SZ")"
OUT_DIR="${FOOTPRINT_OUT_DIR:-.llmsa/public-footprint/${TS}}"
mkdir -p "${OUT_DIR}"

RUNS_JSON="${OUT_DIR}/workflow-health-runs.json"
OUT_JSON="${OUT_DIR}/workflow-health.json"
OUT_MD="${OUT_DIR}/workflow-health.md"
DETAILS_NDJSON="${OUT_DIR}/workflow-health.ndjson"
BLOCKERS_TXT="${OUT_DIR}/workflow-health.blockers"
WARNINGS_TXT="${OUT_DIR}/workflow-health.warnings"

: > "${DETAILS_NDJSON}"
: > "${BLOCKERS_TXT}"
: > "${WARNINGS_TXT}"

gh_api_retry "repos/${REPO}/actions/runs?per_page=${MAX_RUNS}" > "${RUNS_JSON}"

healthy="true"

is_missing_allowed() {
  local workflow="$1"
  local allowed
  for allowed in "${MISSING_ALLOWED_LIST[@]}"; do
    if [[ "${workflow}" == "${allowed}" ]]; then
      return 0
    fi
  done
  return 1
}

for wf in "${WORKFLOW_LIST[@]}"; do
  latest_run="$(jq -c --arg wf "${wf}" '
    [.workflow_runs[] | select(.name == $wf)] | sort_by(.created_at) | last // {}
  ' "${RUNS_JSON}")"

  run_id="$(echo "${latest_run}" | jq -r '.id // empty')"
  status="$(echo "${latest_run}" | jq -r '.status // "missing"')"
  conclusion="$(echo "${latest_run}" | jq -r '.conclusion // "missing"')"
  created_at="$(echo "${latest_run}" | jq -r '.created_at // empty')"
  run_url="$(echo "${latest_run}" | jq -r '.html_url // empty')"

  jq -cn \
    --arg workflow "${wf}" \
    --argjson run_id "${run_id:-null}" \
    --arg status "${status}" \
    --arg conclusion "${conclusion}" \
    --arg created_at "${created_at}" \
    --arg run_url "${run_url}" \
    '{
      workflow: $workflow,
      run_id: $run_id,
      status: $status,
      conclusion: $conclusion,
      created_at: $created_at,
      run_url: $run_url
    }' >> "${DETAILS_NDJSON}"

  if [[ -z "${run_id}" ]]; then
    if is_missing_allowed "${wf}"; then
      echo "No workflow run found for optional workflow ${wf} in the latest ${MAX_RUNS} runs." >> "${WARNINGS_TXT}"
    else
      healthy="false"
      echo "No workflow run found for ${wf} in the latest ${MAX_RUNS} runs." >> "${BLOCKERS_TXT}"
    fi
    continue
  fi

  if [[ "${status}" != "completed" ]]; then
    healthy="false"
    echo "Latest ${wf} run is not completed (status=${status}, run=${run_url:-n/a})." >> "${BLOCKERS_TXT}"
    continue
  fi

  case "${conclusion}" in
    success|neutral|skipped)
      ;;
    *)
      healthy="false"
      echo "Latest ${wf} completed with conclusion=${conclusion} (run=${run_url:-n/a})." >> "${BLOCKERS_TXT}"
      ;;
  esac
done

DETAILS_JSON="${OUT_DIR}/workflow-health-details.json"
if [[ -s "${DETAILS_NDJSON}" ]]; then
  jq -s '.' "${DETAILS_NDJSON}" > "${DETAILS_JSON}"
else
  echo '[]' > "${DETAILS_JSON}"
fi

BLOCKERS_JSON="${OUT_DIR}/workflow-health-blockers.json"
if [[ -s "${BLOCKERS_TXT}" ]]; then
  jq -R -s 'split("\n") | map(select(length > 0))' "${BLOCKERS_TXT}" > "${BLOCKERS_JSON}"
else
  echo '[]' > "${BLOCKERS_JSON}"
fi

WARNINGS_JSON="${OUT_DIR}/workflow-health-warnings.json"
if [[ -s "${WARNINGS_TXT}" ]]; then
  jq -R -s 'split("\n") | map(select(length > 0))' "${WARNINGS_TXT}" > "${WARNINGS_JSON}"
else
  echo '[]' > "${WARNINGS_JSON}"
fi

jq -n \
  --arg generated_at "${GENERATED_AT}" \
  --arg repo "${REPO}" \
  --argjson workflows "$(printf '%s\n' "${WORKFLOW_LIST[@]}" | jq -R . | jq -s .)" \
  --argjson allowed_missing_workflows "$(printf '%s\n' "${MISSING_ALLOWED_LIST[@]}" | jq -R . | jq -s .)" \
  --argjson healthy "${healthy}" \
  --slurpfile details "${DETAILS_JSON}" \
  --slurpfile blockers "${BLOCKERS_JSON}" \
  --slurpfile warnings "${WARNINGS_JSON}" \
  '{
    generated_at_utc: $generated_at,
    repo: $repo,
    workflows_checked: $workflows,
    allowed_missing_workflows: $allowed_missing_workflows,
    healthy: $healthy,
    latest_runs: ($details[0] // []),
    blockers: ($blockers[0] // []),
    warnings: ($warnings[0] // [])
  }' > "${OUT_JSON}"

{
  echo "# Workflow Health Check"
  echo
  echo "- Generated (UTC): \`${GENERATED_AT}\`"
  echo "- Repository: \`${REPO}\`"
  echo "- Healthy: \`${healthy}\`"
  echo
  echo "## Latest Runs"
  echo
  echo "| Workflow | Status | Conclusion | Created At | Run URL |"
  echo "|---|---|---|---|---|"
  jq -r '.[] | "| \(.workflow) | \(.status) | \(.conclusion) | \(.created_at // "n/a") | \(.run_url // "n/a") |"' "${DETAILS_JSON}"
  echo
  echo "## Blockers"
  echo
  if [[ "${healthy}" == "true" ]]; then
    echo "- none"
  else
    jq -r '.[] | "- " + .' "${BLOCKERS_JSON}"
  fi
  echo
  echo "## Warnings"
  echo
  if [[ "$(jq -r 'length' "${WARNINGS_JSON}")" == "0" ]]; then
    echo "- none"
  else
    jq -r '.[] | "- " + .' "${WARNINGS_JSON}"
  fi
} > "${OUT_MD}"

echo "workflow health json: ${OUT_JSON}"
echo "workflow health markdown: ${OUT_MD}"

if [[ "${FAIL_ON_UNHEALTHY}" == "true" && "${healthy}" != "true" ]]; then
  exit 2
fi
