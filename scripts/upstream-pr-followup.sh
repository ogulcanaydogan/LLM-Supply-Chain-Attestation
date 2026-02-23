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

TS="$(date -u +"%Y%m%dT%H%M%SZ")"
OUT_DIR="${FOOTPRINT_OUT_DIR:-.llmsa/public-footprint/${TS}}"
mkdir -p "${OUT_DIR}"

FOLLOWUP_MD="${OUT_DIR}/upstream-followup.md"
FOLLOWUP_JSON="${OUT_DIR}/upstream-followup.json"
NOW_EPOCH="$(date -u +%s)"
FOLLOWUP_INTERVAL_HOURS="${FOLLOWUP_INTERVAL_HOURS:-12}"
if ! [[ "${FOLLOWUP_INTERVAL_HOURS}" =~ ^[0-9]+$ ]]; then
  echo "error: FOLLOWUP_INTERVAL_HOURS must be a positive integer" >&2
  exit 1
fi
FOLLOWUP_INTERVAL_SECONDS=$((FOLLOWUP_INTERVAL_HOURS * 60 * 60))
FALLBACK_THRESHOLD_SECONDS=$((5 * 24 * 60 * 60))
FALLBACK_NOT_BEFORE_UTC="${FALLBACK_NOT_BEFORE_UTC:-}"
FALLBACK_NOT_BEFORE_EPOCH=""
if [[ -n "${FALLBACK_NOT_BEFORE_UTC}" ]]; then
  FALLBACK_NOT_BEFORE_EPOCH="$(iso_epoch "${FALLBACK_NOT_BEFORE_UTC}")"
fi
EXTERNAL_LOG="docs/public-footprint/external-contribution-log.md"
POST_FOLLOWUPS="${POST_FOLLOWUPS:-false}"
FOLLOWUP_COMMENT_BODY="${FOLLOWUP_COMMENT_BODY:-Maintainer follow-up on current HEAD:

- docs-only diff remains intentionally narrow (no check behavior changes)
- beginner evidence progression still maps SBOM generation to Signed-Releases verification
- DCO/sign-off is passing

If scope looks acceptable, could a maintainer take final review for merge?}"
FOLLOWUP_ATTEMPTS_JSON="${OUT_DIR}/upstream-followup-attempts.json"
FOLLOWUP_ATTEMPTS_MD="${OUT_DIR}/upstream-followup-attempts.md"
CURRENT_ACTOR="$(gh api user --jq '.login' 2>/dev/null || true)"
export FOLLOWUP_INTERVAL_HOURS
export POST_FOLLOWUPS
export FALLBACK_NOT_BEFORE_UTC
mapfile -t PRS < <(extract_upstream_pr_urls "${EXTERNAL_LOG}" | while IFS= read -r pr_url; do parse_pr_repo_number "${pr_url}"; done)
if [[ "${#PRS[@]}" -eq 0 ]]; then
  PRS=(
    "sigstore/cosign:4710"
    "open-policy-agent/opa:8343"
    "ossf/scorecard:4942"
  )
fi

TMP_NDJSON="$(mktemp)"
TMP_ATTEMPTS_NDJSON="$(mktemp)"
trap 'rm -f "${TMP_NDJSON}" "${TMP_ATTEMPTS_NDJSON}"' EXIT

for entry in "${PRS[@]}"; do
  repo="${entry%%:*}"
  number="${entry##*:}"
  pr_json="$(gh_api_retry "repos/${repo}/pulls/${number}")"
  comments_json="$(gh_api_retry "repos/${repo}/issues/${number}/comments?per_page=100")"
  url="$(echo "${pr_json}" | jq -r '.html_url')"
  title="$(echo "${pr_json}" | jq -r '.title')"
  state="$(echo "${pr_json}" | jq -r '.state')"
  merged="$(echo "${pr_json}" | jq -r '.merged')"
  updated_at="$(echo "${pr_json}" | jq -r '.updated_at')"
  latest_comment_author="$(echo "${comments_json}" | jq -r 'if length == 0 then "" else .[-1].user.login // "" end')"
  latest_comment_at="$(echo "${comments_json}" | jq -r 'if length == 0 then "" else .[-1].created_at // "" end')"

  updated_epoch="$(iso_epoch "${updated_at}")"
  age_seconds=$((NOW_EPOCH - updated_epoch))
  hours_since_update=$((age_seconds / 3600))

  followup_due=false
  fallback_pr_recommended=false
  followup_posted=false
  followup_attempted=false
  followup_attempt_status="skipped"
  followup_attempt_message=""
  followup_attempted_at=""
  if [[ "${state}" == "open" && "${merged}" == "false" ]]; then
    if (( age_seconds >= FOLLOWUP_INTERVAL_SECONDS )); then
      followup_due=true
    fi
    if (( age_seconds >= FALLBACK_THRESHOLD_SECONDS )); then
      fallback_pr_recommended=true
    fi
    if [[ -n "${FALLBACK_NOT_BEFORE_EPOCH}" && "${NOW_EPOCH}" -lt "${FALLBACK_NOT_BEFORE_EPOCH}" ]]; then
      fallback_pr_recommended=false
    fi

    should_attempt_post=false
    if [[ "${followup_due}" == "true" && "${POST_FOLLOWUPS}" == "true" ]]; then
      # Post follow-up only when the latest comment is not from the current actor.
      # This prevents repeating the same nudge every interval without maintainer response.
      if [[ -n "${CURRENT_ACTOR}" && "${latest_comment_author}" != "${CURRENT_ACTOR}" ]]; then
        should_attempt_post=true
      fi
      if [[ -z "${latest_comment_author}" ]]; then
        should_attempt_post=true
      fi
    fi

    if [[ "${should_attempt_post}" == "true" ]]; then
      followup_attempted=true
      followup_attempted_at="$(date -u +"%Y-%m-%dT%H:%M:%SZ")"
      if gh pr comment "${number}" --repo "${repo}" --body "${FOLLOWUP_COMMENT_BODY}" >/dev/null 2>&1; then
        followup_posted=true
        followup_attempt_status="posted"
        followup_attempt_message="follow-up comment posted"
      else
        followup_attempt_status="failed"
        followup_attempt_message="failed to post comment (API/auth/transient issue)"
      fi
    fi
  fi

  jq -cn \
    --arg repo "${repo}" \
    --argjson number "${number}" \
    --arg url "${url}" \
    --arg title "${title}" \
    --arg state "${state}" \
    --arg merged "${merged}" \
    --arg updated_at "${updated_at}" \
    --argjson hours_since_update "${hours_since_update}" \
    --argjson followup_due "${followup_due}" \
    --argjson fallback_pr_recommended "${fallback_pr_recommended}" \
    --argjson followup_attempted "${followup_attempted}" \
    --argjson followup_posted "${followup_posted}" \
    --arg followup_attempt_status "${followup_attempt_status}" \
    --arg followup_attempt_message "${followup_attempt_message}" \
    --arg followup_attempted_at "${followup_attempted_at}" \
    --arg latest_comment_author "${latest_comment_author}" \
    --arg latest_comment_at "${latest_comment_at}" \
    '{
      repo: $repo,
      number: $number,
      url: $url,
      title: $title,
      state: $state,
      merged: ($merged == "true"),
      updated_at: $updated_at,
      hours_since_update: $hours_since_update,
      followup_due: $followup_due,
      fallback_pr_recommended: $fallback_pr_recommended,
      followup_attempted: $followup_attempted,
      followup_posted: $followup_posted,
      followup_attempt_status: $followup_attempt_status,
      followup_attempt_message: $followup_attempt_message,
      followup_attempted_at: $followup_attempted_at,
      latest_comment_author: $latest_comment_author,
      latest_comment_at: $latest_comment_at
    }' >> "${TMP_NDJSON}"

  if [[ "${followup_attempted}" == "true" ]]; then
    jq -cn \
      --arg repo "${repo}" \
      --argjson number "${number}" \
      --arg url "${url}" \
      --arg attempted_at "${followup_attempted_at}" \
      --arg status "${followup_attempt_status}" \
      --arg message "${followup_attempt_message}" \
      --arg actor "${CURRENT_ACTOR}" \
      '{
        repo: $repo,
        number: $number,
        url: $url,
        actor: $actor,
        attempted_at_utc: $attempted_at,
        status: $status,
        message: $message
      }' >> "${TMP_ATTEMPTS_NDJSON}"
  fi
done

jq -s '
  {
    generated_at_utc: now | todateiso8601,
    followup_interval_hours: env.FOLLOWUP_INTERVAL_HOURS | tonumber,
    fallback_threshold_days: 5,
    fallback_not_before_utc: (if (env.FALLBACK_NOT_BEFORE_UTC // "") == "" then null else env.FALLBACK_NOT_BEFORE_UTC end),
    post_followups_enabled: (env.POST_FOLLOWUPS == "true"),
    prs: .
  }
' "${TMP_NDJSON}" > "${FOLLOWUP_JSON}"

if [[ -s "${TMP_ATTEMPTS_NDJSON}" ]]; then
  jq -s '.' "${TMP_ATTEMPTS_NDJSON}" > "${FOLLOWUP_ATTEMPTS_JSON}"
else
  echo '[]' > "${FOLLOWUP_ATTEMPTS_JSON}"
fi

{
  echo "# Upstream PR Follow-up Snapshot"
  echo
  echo "- Generated (UTC): \`$(date -u +"%Y-%m-%dT%H:%M:%SZ")\`"
  echo "- Follow-up cadence: every \`${FOLLOWUP_INTERVAL_HOURS}h\` while PR is open"
  echo "- Fallback trigger: no maintainer response for \`5 days\`"
  echo "- Auto-post follow-ups: \`${POST_FOLLOWUPS}\`"
  echo
  echo "| PR | State | Merged | Hours Since Update | Follow-up Due | Latest Comment Author | Attempt Status | Fallback PR Recommended |"
  echo "|---|---|---|---:|---|---|---|---|"
  jq -r '.prs[] | "| \(.url) | \(.state) | \(.merged) | \(.hours_since_update) | \(.followup_due) | \(.latest_comment_author // "n/a") | \(.followup_attempt_status) | \(.fallback_pr_recommended) |"' "${FOLLOWUP_JSON}"
} > "${FOLLOWUP_MD}"

{
  echo "# Upstream Follow-up Attempts"
  echo
  echo "- Generated (UTC): \`$(date -u +"%Y-%m-%dT%H:%M:%SZ")\`"
  echo "- Actor: \`${CURRENT_ACTOR:-unknown}\`"
  echo
  echo "| PR | Attempted At (UTC) | Status | Message |"
  echo "|---|---|---|---|"
  if [[ "$(jq -r 'length' "${FOLLOWUP_ATTEMPTS_JSON}")" == "0" ]]; then
    echo "| n/a | n/a | skipped | no follow-up attempts in this run |"
  else
    jq -r '.[] | "| \(.url) | \(.attempted_at_utc) | \(.status) | \(.message) |"' "${FOLLOWUP_ATTEMPTS_JSON}"
  fi
} > "${FOLLOWUP_ATTEMPTS_MD}"

echo "upstream followup json: ${FOLLOWUP_JSON}"
echo "upstream followup markdown: ${FOLLOWUP_MD}"
echo "upstream followup attempts json: ${FOLLOWUP_ATTEMPTS_JSON}"
echo "upstream followup attempts markdown: ${FOLLOWUP_ATTEMPTS_MD}"
