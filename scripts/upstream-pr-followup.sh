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

require_cmd gh
require_cmd jq

if ! gh auth status >/dev/null 2>&1; then
  if [[ -z "${GH_TOKEN:-}" && -z "${GITHUB_TOKEN:-}" ]]; then
    echo "error: gh is not authenticated. run: gh auth login" >&2
    exit 1
  fi
fi

TS="$(date -u +"%Y%m%dT%H%M%SZ")"
OUT_DIR="${FOOTPRINT_OUT_DIR:-.llmsa/public-footprint/${TS}}"
mkdir -p "${OUT_DIR}"

FOLLOWUP_MD="${OUT_DIR}/upstream-followup.md"
FOLLOWUP_JSON="${OUT_DIR}/upstream-followup.json"
NOW_EPOCH="$(date -u +%s)"
FOLLOWUP_INTERVAL_SECONDS=$((48 * 60 * 60))
FALLBACK_THRESHOLD_SECONDS=$((5 * 24 * 60 * 60))

PRS=(
  "sigstore/cosign:4710"
  "open-policy-agent/opa:8343"
  "ossf/scorecard:4942"
)

TMP_NDJSON="$(mktemp)"
trap 'rm -f "${TMP_NDJSON}"' EXIT

for entry in "${PRS[@]}"; do
  repo="${entry%%:*}"
  number="${entry##*:}"
  pr_json="$(gh api "repos/${repo}/pulls/${number}")"
  url="$(echo "${pr_json}" | jq -r '.html_url')"
  title="$(echo "${pr_json}" | jq -r '.title')"
  state="$(echo "${pr_json}" | jq -r '.state')"
  merged="$(echo "${pr_json}" | jq -r '.merged')"
  updated_at="$(echo "${pr_json}" | jq -r '.updated_at')"

  updated_epoch="$(iso_epoch "${updated_at}")"
  age_seconds=$((NOW_EPOCH - updated_epoch))
  hours_since_update=$((age_seconds / 3600))

  followup_due=false
  fallback_pr_recommended=false
  if [[ "${state}" == "open" && "${merged}" == "false" ]]; then
    if (( age_seconds >= FOLLOWUP_INTERVAL_SECONDS )); then
      followup_due=true
    fi
    if (( age_seconds >= FALLBACK_THRESHOLD_SECONDS )); then
      fallback_pr_recommended=true
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
      fallback_pr_recommended: $fallback_pr_recommended
    }' >> "${TMP_NDJSON}"
done

jq -s '
  {
    generated_at_utc: now | todateiso8601,
    followup_interval_hours: 48,
    fallback_threshold_days: 5,
    prs: .
  }
' "${TMP_NDJSON}" > "${FOLLOWUP_JSON}"

{
  echo "# Upstream PR Follow-up Snapshot"
  echo
  echo "- Generated (UTC): \`$(date -u +"%Y-%m-%dT%H:%M:%SZ")\`"
  echo "- Follow-up cadence: every \`48h\` while PR is open"
  echo "- Fallback trigger: no maintainer response for \`5 days\`"
  echo
  echo "| PR | State | Merged | Hours Since Update | Follow-up Due | Fallback PR Recommended |"
  echo "|---|---|---|---:|---|---|"
  jq -r '.prs[] | "| \(.url) | \(.state) | \(.merged) | \(.hours_since_update) | \(.followup_due) | \(.fallback_pr_recommended) |"' "${FOLLOWUP_JSON}"
} > "${FOLLOWUP_MD}"

echo "upstream followup json: ${FOLLOWUP_JSON}"
echo "upstream followup markdown: ${FOLLOWUP_MD}"
