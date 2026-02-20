#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "${ROOT}"

latest_file() {
  local pattern="$1"
  local latest
  latest="$(ls -1 ${pattern} 2>/dev/null | sort | tail -n1 || true)"
  echo "${latest}"
}

trim() {
  echo "$1" | sed -E 's/^[[:space:]]+//; s/[[:space:]]+$//'
}

normalize_url() {
  local value="$1"
  value="$(trim "${value}")"
  value="${value//\`/}"
  value="$(echo "${value}" | cut -d',' -f1)"
  value="$(trim "${value}")"
  echo "${value}"
}

is_canonical_mention() {
  local url="$1"
  local lower
  if [[ -z "${url}" ]]; then
    echo "false"
    return
  fi
  lower="$(echo "${url}" | tr '[:upper:]' '[:lower:]')"
  if [[ ! "${lower}" =~ ^https?:// ]]; then
    echo "false"
    return
  fi
  if [[ "${lower}" =~ ^https?://([a-z0-9-]+\.)?gist\.github\.com(/|$) ]]; then
    echo "false"
    return
  fi
  if [[ "${lower}" =~ ^https?://([a-z0-9-]+\.)?github\.com(/|$) ]]; then
    echo "false"
    return
  fi
  echo "true"
}

SNAPSHOT_JSON="${1:-$(latest_file ".llmsa/public-footprint/*/snapshot.json")}"
CI_HEALTH_JSON="${2:-$(latest_file ".llmsa/public-footprint/*/ci-health.json")}"

if [[ -z "${SNAPSHOT_JSON}" || ! -f "${SNAPSHOT_JSON}" ]]; then
  echo "error: snapshot json not found (expected .llmsa/public-footprint/*/snapshot.json)" >&2
  exit 1
fi
if [[ -z "${CI_HEALTH_JSON}" || ! -f "${CI_HEALTH_JSON}" ]]; then
  echo "error: ci-health json not found (expected .llmsa/public-footprint/*/ci-health.json)" >&2
  exit 1
fi

CANONICAL_URL_FILE="docs/public-footprint/third-party-mention-canonical-url.txt"
EVIDENCE_PACK_FILE="docs/public-footprint/evidence-pack-2026-02-18.md"
MENTION_URL="${THIRD_PARTY_MENTION_URL:-}"

if [[ -f "${CANONICAL_URL_FILE}" ]]; then
  MENTION_URL="$(head -n1 "${CANONICAL_URL_FILE}" || true)"
fi
if [[ -z "${MENTION_URL}" ]]; then
  MENTION_URL="${THIRD_PARTY_MENTION_URL:-}"
fi
if [[ -z "${MENTION_URL}" && -f "${EVIDENCE_PACK_FILE}" ]]; then
  MENTION_URL="$(awk -F'|' '/^\| Third-party mentions \|/ {print $4; exit}' "${EVIDENCE_PACK_FILE}" || true)"
fi
if [[ -z "${MENTION_URL}" ]]; then
  MENTION_URL="https://gist.github.com/ogulcanaydogan/7cffe48a760a77cb42cb1f87644909bb"
fi
MENTION_URL="$(normalize_url "${MENTION_URL}")"

UPSTREAM_MERGED_COUNT="$(jq -r '.upstream_pr_merged // 0' "${SNAPSHOT_JSON}")"
UPSTREAM_MERGED_MET="false"
if (( UPSTREAM_MERGED_COUNT >= 1 )); then
  UPSTREAM_MERGED_MET="true"
fi

ROLLING_PASS_RATE="$(jq -r '.totals.pass_rate_percent // 0' "${CI_HEALTH_JSON}")"
ROLLING_PASS_TARGET="$(jq -r '.totals.pass_rate_target_percent // 95' "${CI_HEALTH_JSON}")"
ROLLING_PASS_MET="$(jq -r '.totals.meets_pass_rate_target // false' "${CI_HEALTH_JSON}")"

POST_HARDENING_PASS_RATE="$(jq -r '.totals.post_hardening_pass_rate_percent // 0' "${CI_HEALTH_JSON}")"
POST_HARDENING_PASS_TARGET="$(jq -r '.totals.post_hardening_pass_rate_target_percent // 95' "${CI_HEALTH_JSON}")"
POST_HARDENING_PASS_MET="$(jq -r '.totals.meets_post_hardening_pass_rate_target // false' "${CI_HEALTH_JSON}")"

CANONICAL_MENTION_MET="$(is_canonical_mention "${MENTION_URL}")"

STRICT_COMPLETE="false"
if [[ "${UPSTREAM_MERGED_MET}" == "true" && "${CANONICAL_MENTION_MET}" == "true" && "${ROLLING_PASS_MET}" == "true" ]]; then
  STRICT_COMPLETE="true"
fi

PRACTICAL_COMPLETE="false"
if [[ "${UPSTREAM_MERGED_MET}" == "true" && "${CANONICAL_MENTION_MET}" == "true" && "${POST_HARDENING_PASS_MET}" == "true" ]]; then
  PRACTICAL_COMPLETE="true"
fi

BLOCKERS_JSON="$(mktemp)"
trap 'rm -f "${BLOCKERS_JSON}"' EXIT
echo "[]" > "${BLOCKERS_JSON}"

if [[ "${UPSTREAM_MERGED_MET}" != "true" ]]; then
  jq '. += ["At least one upstream PR must be merged."]' "${BLOCKERS_JSON}" > "${BLOCKERS_JSON}.tmp" && mv "${BLOCKERS_JSON}.tmp" "${BLOCKERS_JSON}"
fi
if [[ "${CANONICAL_MENTION_MET}" != "true" ]]; then
  jq '. += ["A canonical third-party mention URL (non-GitHub/Gist) is required."]' "${BLOCKERS_JSON}" > "${BLOCKERS_JSON}.tmp" && mv "${BLOCKERS_JSON}.tmp" "${BLOCKERS_JSON}"
fi
if [[ "${ROLLING_PASS_MET}" != "true" ]]; then
  jq '. += ["Rolling 30-day CI pass rate is below target (>=95%)."]' "${BLOCKERS_JSON}" > "${BLOCKERS_JSON}.tmp" && mv "${BLOCKERS_JSON}.tmp" "${BLOCKERS_JSON}"
fi

TS="$(date -u +"%Y%m%dT%H%M%SZ")"
OUT_DIR="${FOOTPRINT_OUT_DIR:-.llmsa/public-footprint/${TS}}"
mkdir -p "${OUT_DIR}"
OUT_JSON="${OUT_DIR}/roadmap-completion.json"
OUT_MD="${OUT_DIR}/roadmap-completion.md"
GENERATED_AT="$(date -u +"%Y-%m-%dT%H:%M:%SZ")"

jq -n \
  --arg generated_at "${GENERATED_AT}" \
  --arg snapshot_source "${SNAPSHOT_JSON}" \
  --arg ci_health_source "${CI_HEALTH_JSON}" \
  --arg mention_url "${MENTION_URL}" \
  --argjson upstream_merged_count "${UPSTREAM_MERGED_COUNT}" \
  --argjson rolling_pass_rate "${ROLLING_PASS_RATE}" \
  --argjson rolling_pass_target "${ROLLING_PASS_TARGET}" \
  --argjson post_hardening_pass_rate "${POST_HARDENING_PASS_RATE}" \
  --argjson post_hardening_pass_target "${POST_HARDENING_PASS_TARGET}" \
  --argjson upstream_merged_met "${UPSTREAM_MERGED_MET}" \
  --argjson canonical_mention_met "${CANONICAL_MENTION_MET}" \
  --argjson rolling_pass_met "${ROLLING_PASS_MET}" \
  --argjson post_hardening_pass_met "${POST_HARDENING_PASS_MET}" \
  --argjson strict_complete "${STRICT_COMPLETE}" \
  --argjson practical_complete "${PRACTICAL_COMPLETE}" \
  --slurpfile blockers "${BLOCKERS_JSON}" \
  '{
    generated_at_utc: $generated_at,
    gates: {
      upstream_merged_count: $upstream_merged_count,
      upstream_merged_met: $upstream_merged_met,
      canonical_mention_url: $mention_url,
      canonical_mention_met: $canonical_mention_met,
      rolling_ci_pass_rate_percent: $rolling_pass_rate,
      rolling_ci_pass_rate_target_percent: $rolling_pass_target,
      rolling_ci_pass_met: $rolling_pass_met,
      post_hardening_ci_pass_rate_percent: $post_hardening_pass_rate,
      post_hardening_ci_pass_rate_target_percent: $post_hardening_pass_target,
      post_hardening_ci_pass_met: $post_hardening_pass_met
    },
    verdict: {
      strict_complete: $strict_complete,
      practical_complete: $practical_complete
    },
    blockers: ($blockers[0] // []),
    sources: {
      snapshot_json: $snapshot_source,
      ci_health_json: $ci_health_source,
      mention_url: $mention_url
    }
  }' > "${OUT_JSON}"

{
  echo "# Roadmap Completion Check"
  echo
  echo "- Generated (UTC): \`${GENERATED_AT}\`"
  echo "- Snapshot Source: \`${SNAPSHOT_JSON}\`"
  echo "- CI Health Source: \`${CI_HEALTH_JSON}\`"
  echo
  echo "## Gate Status"
  echo
  echo "| Gate | Current | Target | Met |"
  echo "|---|---|---|---|"
  echo "| Upstream merged PRs | ${UPSTREAM_MERGED_COUNT} | >=1 | ${UPSTREAM_MERGED_MET} |"
  echo "| Canonical third-party mention | ${MENTION_URL} | non-GitHub/Gist URL | ${CANONICAL_MENTION_MET} |"
  echo "| Rolling CI pass rate | ${ROLLING_PASS_RATE}% | >=${ROLLING_PASS_TARGET}% | ${ROLLING_PASS_MET} |"
  echo "| Post-hardening CI pass rate | ${POST_HARDENING_PASS_RATE}% | >=${POST_HARDENING_PASS_TARGET}% | ${POST_HARDENING_PASS_MET} |"
  echo
  echo "## Verdict"
  echo
  echo "- Strict complete: \`${STRICT_COMPLETE}\`"
  echo "- Practical complete: \`${PRACTICAL_COMPLETE}\`"
  echo
  echo "## Blockers"
  echo
  if [[ "$(jq -r 'length' "${BLOCKERS_JSON}")" == "0" ]]; then
    echo "- none"
  else
    jq -r '.[] | "- " + .' "${BLOCKERS_JSON}"
  fi
} > "${OUT_MD}"

echo "roadmap completion json: ${OUT_JSON}"
echo "roadmap completion markdown: ${OUT_MD}"

if [[ "${FAIL_ON_INCOMPLETE:-false}" == "true" && "${STRICT_COMPLETE}" != "true" ]]; then
  exit 2
fi
