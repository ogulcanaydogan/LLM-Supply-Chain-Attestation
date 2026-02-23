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

require_cmd jq

SEARCH_BIN=""
if command -v rg >/dev/null 2>&1; then
  SEARCH_BIN="rg"
elif command -v grep >/dev/null 2>&1; then
  SEARCH_BIN="grep"
else
  echo "error: required command not found: rg or grep" >&2
  exit 1
fi

SNAPSHOT_JSON="${1:-$(latest_file ".llmsa/public-footprint/*/snapshot.json")}"
CI_HEALTH_JSON="${2:-$(latest_file ".llmsa/public-footprint/*/ci-health.json")}"
COMPLETION_JSON="${3:-$(latest_file ".llmsa/public-footprint/*/roadmap-completion.json")}"

if [[ -z "${SNAPSHOT_JSON}" || ! -f "${SNAPSHOT_JSON}" ]]; then
  echo "error: snapshot json not found (expected .llmsa/public-footprint/*/snapshot.json)" >&2
  exit 1
fi
if [[ -z "${CI_HEALTH_JSON}" || ! -f "${CI_HEALTH_JSON}" ]]; then
  echo "error: ci-health json not found (expected .llmsa/public-footprint/*/ci-health.json)" >&2
  exit 1
fi
if [[ -z "${COMPLETION_JSON}" || ! -f "${COMPLETION_JSON}" ]]; then
  echo "error: roadmap completion json not found (expected .llmsa/public-footprint/*/roadmap-completion.json)" >&2
  exit 1
fi

DOCS=(
  "docs/public-footprint/measurement-dashboard.md"
  "docs/public-footprint/evidence-pack-2026-02-18.md"
)

DOCS_FULL=(
  "docs/public-footprint/measurement-dashboard.md"
  "docs/public-footprint/evidence-pack-2026-02-18.md"
  "docs/public-footprint/roadmap-status-2026-02.md"
  "docs/public-footprint/day-30-outcomes-2026-02.md"
  "docs/public-footprint/v1.0.1-hardening-closure.md"
)

CONSISTENCY_SCOPE="${CONSISTENCY_SCOPE:-full}"
if [[ "${CONSISTENCY_SCOPE}" == "full" ]]; then
  DOCS=("${DOCS_FULL[@]}")
elif [[ "${CONSISTENCY_SCOPE}" == "core" ]]; then
  DOCS=(
    "docs/public-footprint/measurement-dashboard.md"
    "docs/public-footprint/evidence-pack-2026-02-18.md"
  )
else
  echo "error: CONSISTENCY_SCOPE must be 'full' or 'core'" >&2
  exit 1
fi

for doc in "${DOCS[@]}"; do
  if [[ ! -f "${doc}" ]]; then
    echo "error: required doc missing: ${doc}" >&2
    exit 1
  fi
done

STRICT_MACHINE="$(jq -r '.verdict.strict_complete // false' "${COMPLETION_JSON}")"
PRACTICAL_MACHINE="$(jq -r '.verdict.practical_complete // false' "${COMPLETION_JSON}")"

TS="$(date -u +"%Y%m%dT%H%M%SZ")"
OUT_DIR="${FOOTPRINT_OUT_DIR:-.llmsa/public-footprint/${TS}}"
mkdir -p "${OUT_DIR}"
OUT_JSON="${OUT_DIR}/consistency-check.json"
OUT_MD="${OUT_DIR}/consistency-check.md"
GENERATED_AT="$(date -u +"%Y-%m-%dT%H:%M:%SZ")"

BLOCKERS_JSON="$(mktemp)"
trap 'rm -f "${BLOCKERS_JSON}"' EXIT
echo "[]" > "${BLOCKERS_JSON}"

append_blocker() {
  local message="$1"
  jq --arg msg "${message}" '. += [$msg]' "${BLOCKERS_JSON}" > "${BLOCKERS_JSON}.tmp"
  mv "${BLOCKERS_JSON}.tmp" "${BLOCKERS_JSON}"
}

mapfile -t SNAPSHOT_REFS < <(
  if [[ "${SEARCH_BIN}" == "rg" ]]; then
    rg -o --no-filename '\.llmsa/public-footprint/[0-9TZ]+/snapshot\.json' "${DOCS[@]}"
  else
    grep -Eho '\.llmsa/public-footprint/[0-9TZ]+/snapshot\.json' "${DOCS[@]}"
  fi | sort -u
)
mapfile -t CI_REFS < <(
  if [[ "${SEARCH_BIN}" == "rg" ]]; then
    rg -o --no-filename '\.llmsa/public-footprint/[0-9TZ]+/ci-health\.json' "${DOCS[@]}"
  else
    grep -Eho '\.llmsa/public-footprint/[0-9TZ]+/ci-health\.json' "${DOCS[@]}"
  fi | sort -u
)

if [[ "${#SNAPSHOT_REFS[@]}" -eq 0 ]]; then
  append_blocker "No snapshot artifact path found in footprint docs."
fi
if [[ "${#CI_REFS[@]}" -eq 0 ]]; then
  append_blocker "No ci-health artifact path found in footprint docs."
fi
if [[ "${#SNAPSHOT_REFS[@]}" -gt 1 ]]; then
  append_blocker "Mixed snapshot artifact timestamps detected across docs."
fi
if [[ "${#CI_REFS[@]}" -gt 1 ]]; then
  append_blocker "Mixed ci-health artifact timestamps detected across docs."
fi

if [[ "${#SNAPSHOT_REFS[@]}" -eq 1 && "${SNAPSHOT_REFS[0]}" != "${SNAPSHOT_JSON}" ]]; then
  append_blocker "Snapshot artifact reference is stale. expected=${SNAPSHOT_JSON} found=${SNAPSHOT_REFS[0]}"
fi
if [[ "${#CI_REFS[@]}" -eq 1 && "${CI_REFS[0]}" != "${CI_HEALTH_JSON}" ]]; then
  append_blocker "CI-health artifact reference is stale. expected=${CI_HEALTH_JSON} found=${CI_REFS[0]}"
fi

while IFS= read -r line; do
  if [[ "${line}" =~ Strict[[:space:]]+complete:[[:space:]]*\`?true\`? ]] && [[ "${STRICT_MACHINE}" != "true" ]]; then
    append_blocker "Narrative claims strict complete=true while machine verdict is false."
  fi
  if [[ "${line}" =~ Strict[[:space:]]+complete:[[:space:]]*\`?false\`? ]] && [[ "${STRICT_MACHINE}" == "true" ]]; then
    append_blocker "Narrative claims strict complete=false while machine verdict is true."
  fi
  if [[ "${line}" =~ Practical[[:space:]]+complete:[[:space:]]*\`?true\`? ]] && [[ "${PRACTICAL_MACHINE}" != "true" ]]; then
    append_blocker "Narrative claims practical complete=true while machine verdict is false."
  fi
  if [[ "${line}" =~ Practical[[:space:]]+complete:[[:space:]]*\`?false\`? ]] && [[ "${PRACTICAL_MACHINE}" == "true" ]]; then
    append_blocker "Narrative claims practical complete=false while machine verdict is true."
  fi
done < <(cat "${DOCS[@]}")

CONSISTENT="true"
if [[ "$(jq -r 'length' "${BLOCKERS_JSON}")" != "0" ]]; then
  CONSISTENT="false"
fi

jq -n \
  --arg generated_at "${GENERATED_AT}" \
  --arg snapshot_json "${SNAPSHOT_JSON}" \
  --arg ci_health_json "${CI_HEALTH_JSON}" \
  --arg completion_json "${COMPLETION_JSON}" \
  --arg strict_machine "${STRICT_MACHINE}" \
  --arg practical_machine "${PRACTICAL_MACHINE}" \
  --arg consistency_scope "${CONSISTENCY_SCOPE}" \
  --argjson consistent "${CONSISTENT}" \
  --argjson docs "$(printf '%s\n' "${DOCS[@]}" | jq -R . | jq -s .)" \
  --argjson snapshot_refs "$(printf '%s\n' "${SNAPSHOT_REFS[@]-}" | sed '/^$/d' | jq -R . | jq -s .)" \
  --argjson ci_refs "$(printf '%s\n' "${CI_REFS[@]-}" | sed '/^$/d' | jq -R . | jq -s .)" \
  --slurpfile blockers "${BLOCKERS_JSON}" \
  '{
    generated_at_utc: $generated_at,
    sources: {
      snapshot_json: $snapshot_json,
      ci_health_json: $ci_health_json,
      completion_json: $completion_json
    },
    machine_verdict: {
      strict_complete: ($strict_machine == "true"),
      practical_complete: ($practical_machine == "true")
    },
    consistency_scope: $consistency_scope,
    docs_scanned: $docs,
    doc_references: {
      snapshot_json_refs: $snapshot_refs,
      ci_health_json_refs: $ci_refs
    },
    consistent: $consistent,
    blockers: ($blockers[0] // [])
  }' > "${OUT_JSON}"

{
  echo "# Public Footprint Consistency Check"
  echo
  echo "- Generated (UTC): \`${GENERATED_AT}\`"
  echo "- Snapshot source: \`${SNAPSHOT_JSON}\`"
  echo "- CI health source: \`${CI_HEALTH_JSON}\`"
  echo "- Completion source: \`${COMPLETION_JSON}\`"
  echo
  echo "## Machine Verdict"
  echo
  echo "- Strict complete: \`${STRICT_MACHINE}\`"
  echo "- Practical complete: \`${PRACTICAL_MACHINE}\`"
  echo
  echo "## Consistency"
  echo
  echo "- Consistent: \`${CONSISTENT}\`"
  echo
  echo "## Blockers"
  echo
  if [[ "${CONSISTENT}" == "true" ]]; then
    echo "- none"
  else
    jq -r '.[] | "- " + .' "${BLOCKERS_JSON}"
  fi
} > "${OUT_MD}"

echo "consistency json: ${OUT_JSON}"
echo "consistency markdown: ${OUT_MD}"

if [[ "${FAIL_ON_INCONSISTENT:-true}" == "true" && "${CONSISTENT}" != "true" ]]; then
  exit 2
fi
