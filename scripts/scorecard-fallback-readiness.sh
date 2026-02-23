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
  for attempt in 1 2 3; do
    if "$@"; then
      return 0
    fi
    sleep $((attempt * 2))
  done
  return 1
}

iso_epoch() {
  local value="$1"
  if date -u -d "${value}" +%s >/dev/null 2>&1; then
    date -u -d "${value}" +%s
    return
  fi
  local normalized
  normalized="$(echo "${value}" | sed -E 's/Z$/+0000/; s/([+-][0-9]{2}):([0-9]{2})$/\1\2/')"
  date -u -j -f "%Y-%m-%dT%H:%M:%S%z" "${normalized}" +%s
}

latest_file() {
  local pattern="$1"
  local latest
  latest="$(ls -1 ${pattern} 2>/dev/null | sort | tail -n1 || true)"
  echo "${latest}"
}

usage() {
  cat <<'EOF'
Usage: scripts/scorecard-fallback-readiness.sh [options]

Options:
  --repo <owner/repo>            Target repository (default: ossf/scorecard)
  --pr-number <n>                Pull request number (default: 4942)
  --fallback-not-before <utc>    UTC cutoff for fallback recommendation (default: 2026-03-02T00:00:00Z)
  --out-dir <path>               Output directory (default: .llmsa/public-footprint/<timestamp>)
  -h, --help                     Show help
EOF
}

require_cmd gh
require_cmd jq

REPO="ossf/scorecard"
PR_NUMBER="4942"
FALLBACK_NOT_BEFORE_UTC="${FALLBACK_NOT_BEFORE_UTC:-2026-03-02T00:00:00Z}"
OUT_DIR=""

while [[ $# -gt 0 ]]; do
  case "$1" in
    --repo)
      REPO="${2:-}"
      shift 2
      ;;
    --pr-number)
      PR_NUMBER="${2:-}"
      shift 2
      ;;
    --fallback-not-before)
      FALLBACK_NOT_BEFORE_UTC="${2:-}"
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

if [[ ! "${PR_NUMBER}" =~ ^[0-9]+$ ]]; then
  echo "error: --pr-number must be numeric" >&2
  exit 1
fi
if [[ -z "${REPO}" ]]; then
  echo "error: --repo must not be empty" >&2
  exit 1
fi

if [[ -z "${OUT_DIR}" ]]; then
  TS="$(date -u +"%Y%m%dT%H%M%SZ")"
  OUT_DIR=".llmsa/public-footprint/${TS}"
fi
mkdir -p "${OUT_DIR}"

PR_JSON="$(gh_retry gh pr view "${PR_NUMBER}" --repo "${REPO}" --json state,isDraft,mergeable,reviewDecision,updatedAt,url,comments)"
COMMENTS_JSON="$(gh_retry gh api "repos/${REPO}/issues/${PR_NUMBER}/comments?per_page=100")"

NOW_UTC="$(date -u +"%Y-%m-%dT%H:%M:%SZ")"
NOW_EPOCH="$(iso_epoch "${NOW_UTC}")"
FALLBACK_NOT_BEFORE_EPOCH="$(iso_epoch "${FALLBACK_NOT_BEFORE_UTC}")"

STATE="$(echo "${PR_JSON}" | jq -r '.state')"
MERGEABLE="$(echo "${PR_JSON}" | jq -r '.mergeable')"
REVIEW_DECISION="$(echo "${PR_JSON}" | jq -r '.reviewDecision')"
UPDATED_AT="$(echo "${PR_JSON}" | jq -r '.updatedAt')"
URL="$(echo "${PR_JSON}" | jq -r '.url')"
IS_DRAFT="$(echo "${PR_JSON}" | jq -r '.isDraft')"
COMMENTS_COUNT="$(echo "${PR_JSON}" | jq -r '.comments | length')"
LATEST_COMMENT_AUTHOR="$(echo "${COMMENTS_JSON}" | jq -r 'if length == 0 then "" else .[-1].user.login // "" end')"
LATEST_COMMENT_AT="$(echo "${COMMENTS_JSON}" | jq -r 'if length == 0 then "" else .[-1].created_at // "" end')"

UPDATED_EPOCH="$(iso_epoch "${UPDATED_AT}")"
HOURS_SINCE_UPDATE=$(((NOW_EPOCH - UPDATED_EPOCH) / 3600))

FALLBACK_WINDOW_OPEN=false
if (( NOW_EPOCH >= FALLBACK_NOT_BEFORE_EPOCH )); then
  FALLBACK_WINDOW_OPEN=true
fi

FALLBACK_RECOMMENDED=false
if [[ "${STATE}" == "OPEN" && "${FALLBACK_WINDOW_OPEN}" == "true" ]]; then
  FALLBACK_RECOMMENDED=true
fi

NEXT_ACTION="continue_12h_followup_cadence"
if [[ "${STATE}" == "MERGED" ]]; then
  NEXT_ACTION="freeze_evidence_and_mark_second_upstream_merge_complete"
elif [[ "${STATE}" == "CLOSED" ]]; then
  NEXT_ACTION="open_reduced_scope_fallback_pr_if_not_already_merged"
elif [[ "${FALLBACK_RECOMMENDED}" == "true" ]]; then
  NEXT_ACTION="open_reduced_scope_fallback_pr_and_cross_link_original"
fi

OUT_JSON="${OUT_DIR}/scorecard-fallback-readiness.json"
OUT_MD="${OUT_DIR}/scorecard-fallback-readiness.md"

jq -n \
  --arg generated_at_utc "${NOW_UTC}" \
  --arg repo "${REPO}" \
  --argjson pr_number "${PR_NUMBER}" \
  --arg url "${URL}" \
  --arg state "${STATE}" \
  --arg mergeable "${MERGEABLE}" \
  --arg review_decision "${REVIEW_DECISION}" \
  --argjson is_draft "${IS_DRAFT}" \
  --arg updated_at "${UPDATED_AT}" \
  --argjson hours_since_update "${HOURS_SINCE_UPDATE}" \
  --arg fallback_not_before_utc "${FALLBACK_NOT_BEFORE_UTC}" \
  --argjson fallback_window_open "${FALLBACK_WINDOW_OPEN}" \
  --argjson fallback_recommended "${FALLBACK_RECOMMENDED}" \
  --argjson comments_count "${COMMENTS_COUNT}" \
  --arg latest_comment_author "${LATEST_COMMENT_AUTHOR}" \
  --arg latest_comment_at "${LATEST_COMMENT_AT}" \
  --arg next_action "${NEXT_ACTION}" \
  '{
    generated_at_utc: $generated_at_utc,
    target: {
      repo: $repo,
      pr_number: $pr_number,
      url: $url
    },
    pr_status: {
      state: $state,
      is_draft: $is_draft,
      mergeable: $mergeable,
      review_decision: $review_decision,
      updated_at: $updated_at,
      hours_since_update: $hours_since_update
    },
    fallback_policy: {
      fallback_not_before_utc: $fallback_not_before_utc,
      fallback_window_open: $fallback_window_open,
      fallback_recommended: $fallback_recommended
    },
    activity: {
      comments_count: $comments_count,
      latest_comment_author: $latest_comment_author,
      latest_comment_at: $latest_comment_at
    },
    next_action: $next_action
  }' > "${OUT_JSON}"

{
  echo "# Scorecard Fallback Readiness"
  echo
  echo "- Generated (UTC): \`${NOW_UTC}\`"
  echo "- Target PR: \`${URL}\`"
  echo
  echo "## Status"
  echo
  echo "| Field | Value |"
  echo "|---|---|"
  echo "| State | $(jq -r '.pr_status.state' "${OUT_JSON}") |"
  echo "| Mergeable | $(jq -r '.pr_status.mergeable' "${OUT_JSON}") |"
  echo "| Review decision | $(jq -r '.pr_status.review_decision' "${OUT_JSON}") |"
  echo "| Updated at | $(jq -r '.pr_status.updated_at' "${OUT_JSON}") |"
  echo "| Hours since update | $(jq -r '.pr_status.hours_since_update' "${OUT_JSON}") |"
  echo
  echo "## Fallback Policy"
  echo
  echo "| Field | Value |"
  echo "|---|---|"
  echo "| Not before (UTC) | $(jq -r '.fallback_policy.fallback_not_before_utc' "${OUT_JSON}") |"
  echo "| Window open | $(jq -r '.fallback_policy.fallback_window_open' "${OUT_JSON}") |"
  echo "| Fallback recommended | $(jq -r '.fallback_policy.fallback_recommended' "${OUT_JSON}") |"
  echo
  echo "## Activity"
  echo
  echo "| Field | Value |"
  echo "|---|---|"
  echo "| Comment count | $(jq -r '.activity.comments_count' "${OUT_JSON}") |"
  echo "| Latest comment author | $(jq -r '.activity.latest_comment_author // "n/a"' "${OUT_JSON}") |"
  echo "| Latest comment at | $(jq -r '.activity.latest_comment_at // "n/a"' "${OUT_JSON}") |"
  echo
  echo "## Next Action"
  echo
  jq -r '.next_action' "${OUT_JSON}"
} > "${OUT_MD}"

echo "scorecard fallback readiness json: ${OUT_JSON}"
echo "scorecard fallback readiness markdown: ${OUT_MD}"
