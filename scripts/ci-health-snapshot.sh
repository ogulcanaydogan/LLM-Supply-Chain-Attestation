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

date_days_ago() {
  local days="$1"
  if date -u -d "${days} days ago" +"%Y-%m-%d" >/dev/null 2>&1; then
    date -u -d "${days} days ago" +"%Y-%m-%d"
    return
  fi
  date -u -v-"${days}"d +"%Y-%m-%d"
}

classify_failure() {
  local workflow="$1"
  local failed_steps="$2"
  local haystack
  haystack="$(printf "%s %s" "${workflow}" "${failed_steps}" | tr '[:upper:]' '[:lower:]')"

  if [[ "${haystack}" =~ bad\ credentials|unauthorized|forbidden|permission|missing\ signature|missing\ certificate|token|download\ release\ assets|login\ to\ ghcr|authentication ]]; then
    echo "credential_token_permission"
    return
  fi
  if [[ "${haystack}" =~ timeout|timed\ out|connection\ reset|service\ unavailable|temporary|network|tls|rate\ limit|502|503|download|fetch ]]; then
    echo "flaky_transient_infrastructure"
    return
  fi
  echo "deterministic_config_error"
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

WINDOW_DAYS="${WINDOW_DAYS:-30}"
SINCE_DATE="$(date_days_ago "${WINDOW_DAYS}")"
GENERATED_AT="$(date -u +"%Y-%m-%dT%H:%M:%SZ")"
TS="$(date -u +"%Y%m%dT%H%M%SZ")"
OUT_DIR="${FOOTPRINT_OUT_DIR:-.llmsa/public-footprint/${TS}}"
TMP_DIR="${OUT_DIR}/.ci-health-tmp"
mkdir -p "${TMP_DIR}"

if [[ -n "${CI_WORKFLOWS:-}" ]]; then
  IFS=',' read -r -a WORKFLOWS <<<"${CI_WORKFLOWS}"
else
  WORKFLOWS=(
    "ci-attest-verify"
    "release"
    "release-verify"
    "public-footprint-weekly"
  )
fi

ALL_RUNS_JSON="${TMP_DIR}/all-runs.json"
ALL_RUNS_REST_JSON="${TMP_DIR}/all-runs-rest.json"
gh_api_retry "repos/${REPO}/actions/runs?per_page=100" > "${ALL_RUNS_REST_JSON}"

workflows_json="$(printf '%s\n' "${WORKFLOWS[@]}" | jq -R . | jq -s .)"
since_ts="${SINCE_DATE}T00:00:00Z"

jq \
  --arg since "${since_ts}" \
  --argjson workflows "${workflows_json}" \
  '
    [
      .workflow_runs[]
      | select(.created_at >= $since)
      | select(($workflows | index(.name)) != null)
      | {
          databaseId: .id,
          name: .name,
          workflowName: .name,
          conclusion: .conclusion,
          status: .status,
          createdAt: .created_at,
          updatedAt: .updated_at,
          url: .html_url,
          event: .event,
          headSha: .head_sha
        }
    ]
  ' "${ALL_RUNS_REST_JSON}" > "${ALL_RUNS_JSON}"

FAILURE_DETAILS_NDJSON="${TMP_DIR}/failure-details.ndjson"
: > "${FAILURE_DETAILS_NDJSON}"

mapfile -t FAILED_IDS < <(
  jq -r '
    .[]
    | select(.status == "completed")
    | select(.conclusion != "success" and .conclusion != "skipped" and .conclusion != "neutral")
    | .databaseId
  ' "${ALL_RUNS_JSON}" | sort -u
)

for run_id in "${FAILED_IDS[@]}"; do
  run_file="${TMP_DIR}/run-${run_id}.json"
  jobs_file="${TMP_DIR}/run-${run_id}-jobs.json"
  jq --argjson rid "${run_id}" '.[] | select(.databaseId == $rid)' "${ALL_RUNS_JSON}" > "${run_file}" || continue
  gh_api_retry "repos/${REPO}/actions/runs/${run_id}/jobs?per_page=100" > "${jobs_file}" || continue

  workflow_name="$(jq -r '.workflowName // .name // "unknown"' "${run_file}")"
  failed_steps="$(jq -r '[.jobs[]?.steps[]? | select(.conclusion == "failure") | .name] | unique | join("; ")' "${jobs_file}")"
  classification="$(classify_failure "${workflow_name}" "${failed_steps}")"

  jq -cn \
    --argjson run_id "${run_id}" \
    --arg workflow "${workflow_name}" \
    --arg conclusion "$(jq -r '.conclusion // ""' "${run_file}")" \
    --arg created_at "$(jq -r '.createdAt // ""' "${run_file}")" \
    --arg url "$(jq -r '.url // ""' "${run_file}")" \
    --arg failed_steps "${failed_steps}" \
    --arg classification "${classification}" \
    '{
      run_id: $run_id,
      workflow: $workflow,
      conclusion: $conclusion,
      created_at: $created_at,
      url: $url,
      failed_steps: $failed_steps,
      classification: $classification
    }' >> "${FAILURE_DETAILS_NDJSON}"
done

FAILURES_JSON="${TMP_DIR}/failures.json"
if [[ -s "${FAILURE_DETAILS_NDJSON}" ]]; then
  jq -s '.' "${FAILURE_DETAILS_NDJSON}" > "${FAILURES_JSON}"
else
  echo '[]' > "${FAILURES_JSON}"
fi

CI_HEALTH_JSON="${OUT_DIR}/ci-health.json"
CI_HEALTH_MD="${OUT_DIR}/ci-health.md"

jq -n \
  --arg generated_at "${GENERATED_AT}" \
  --arg repo "${REPO}" \
  --arg since_date "${SINCE_DATE}" \
  --argjson window_days "${WINDOW_DAYS}" \
  --argjson workflows "${workflows_json}" \
  --slurpfile runs "${ALL_RUNS_JSON}" \
  --slurpfile failures "${FAILURES_JSON}" '
  def pct($ok; $total):
    if $total == 0 then 0
    else ((($ok * 10000.0) / $total) | round / 100)
    end;
  ($runs[0] // []) as $all_runs
  | ($all_runs | map(select(.status == "completed"))) as $completed_runs
  | ($completed_runs | length) as $completed_total
  | ($completed_runs | map(select(.conclusion == "success")) | length) as $success_total
  | ($completed_total - $success_total) as $failure_total
  | ($failures[0] // []) as $failure_details
  | {
      generated_at_utc: $generated_at,
      repo: $repo,
      window_days: $window_days,
      window_start_utc: ($since_date + "T00:00:00Z"),
      sources: {
        gh_runs_api: "gh api repos/<owner>/<repo>/actions/runs?per_page=100",
        gh_run_jobs_api: "gh api repos/<owner>/<repo>/actions/runs/<run_id>/jobs?per_page=100"
      },
      totals: {
        completed_runs: $completed_total,
        successful_runs: $success_total,
        failed_runs: $failure_total,
        pass_rate_percent: pct($success_total; $completed_total),
        pass_rate_target_percent: 95.0,
        meets_pass_rate_target: (pct($success_total; $completed_total) >= 95.0)
      },
      by_workflow: [
        $workflows[] as $wf
        | ($completed_runs | map(select((.workflowName // .name) == $wf or .name == $wf))) as $wf_runs
        | ($wf_runs | length) as $wf_total
        | ($wf_runs | map(select(.conclusion == "success")) | length) as $wf_success
        | {
            workflow: $wf,
            completed_runs: $wf_total,
            successful_runs: $wf_success,
            failed_runs: ($wf_total - $wf_success),
            pass_rate_percent: pct($wf_success; $wf_total)
          }
      ],
      failure_classification: {
        deterministic_config_error: ($failure_details | map(select(.classification == "deterministic_config_error")) | length),
        flaky_transient_infrastructure: ($failure_details | map(select(.classification == "flaky_transient_infrastructure")) | length),
        credential_token_permission: ($failure_details | map(select(.classification == "credential_token_permission")) | length)
      },
      failures: $failure_details
    }
  ' > "${CI_HEALTH_JSON}"

{
  echo "# CI Health Snapshot"
  echo
  echo "- Generated (UTC): \`${GENERATED_AT}\`"
  echo "- Repository: \`${REPO}\`"
  echo "- Window: last \`${WINDOW_DAYS}\` days (from \`${SINCE_DATE}\`)"
  echo
  echo "## Totals"
  echo
  echo "| Metric | Value | Source |"
  echo "|---|---:|---|"
  echo "| Completed runs | $(jq -r '.totals.completed_runs' "${CI_HEALTH_JSON}") | \`gh api repos/${REPO}/actions/runs\` |"
  echo "| Successful runs | $(jq -r '.totals.successful_runs' "${CI_HEALTH_JSON}") | \`gh api repos/${REPO}/actions/runs\` |"
  echo "| Failed runs | $(jq -r '.totals.failed_runs' "${CI_HEALTH_JSON}") | \`gh api repos/${REPO}/actions/runs\` |"
  echo "| Pass rate | $(jq -r '.totals.pass_rate_percent' "${CI_HEALTH_JSON}")% | \`gh api repos/${REPO}/actions/runs\` |"
  echo "| Meets >=95% target | $(jq -r '.totals.meets_pass_rate_target' "${CI_HEALTH_JSON}") | \`${CI_HEALTH_JSON}\` |"
  echo
  echo "## Workflow Breakdown"
  echo
  echo "| Workflow | Completed | Success | Failed | Pass rate |"
  echo "|---|---:|---:|---:|---:|"
  jq -r '.by_workflow[] | "| \(.workflow) | \(.completed_runs) | \(.successful_runs) | \(.failed_runs) | \(.pass_rate_percent)% |"' "${CI_HEALTH_JSON}"
  echo
  echo "## Failure Classification"
  echo
  echo "| Classification | Count |"
  echo "|---|---:|"
  echo "| deterministic_config_error | $(jq -r '.failure_classification.deterministic_config_error' "${CI_HEALTH_JSON}") |"
  echo "| flaky_transient_infrastructure | $(jq -r '.failure_classification.flaky_transient_infrastructure' "${CI_HEALTH_JSON}") |"
  echo "| credential_token_permission | $(jq -r '.failure_classification.credential_token_permission' "${CI_HEALTH_JSON}") |"
  echo
  echo "## Failing Run Details"
  echo
  if [[ "$(jq -r '.failures | length' "${CI_HEALTH_JSON}")" == "0" ]]; then
    echo "No failed runs in the selected window."
  else
    echo "| Run ID | Workflow | Classification | Failed Steps | URL |"
    echo "|---:|---|---|---|---|"
    jq -r '.failures[] | "| \(.run_id) | \(.workflow) | \(.classification) | \(.failed_steps // "n/a") | \(.url) |"' "${CI_HEALTH_JSON}"
  fi
} > "${CI_HEALTH_MD}"

echo "ci health json: ${CI_HEALTH_JSON}"
echo "ci health markdown: ${CI_HEALTH_MD}"
