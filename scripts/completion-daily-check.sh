#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "${ROOT}"

REPO="${1:-}"

chmod +x scripts/workflow-health-check.sh scripts/roadmap-completion-check.sh scripts/check-footprint-consistency.sh

DOCS=(
  "docs/public-footprint/measurement-dashboard.md"
  "docs/public-footprint/evidence-pack-2026-02-18.md"
)

search_refs() {
  local pattern="$1"
  if command -v rg >/dev/null 2>&1; then
    rg -o --no-filename "${pattern}" "${DOCS[@]}" | sort -u
    return
  fi
  grep -Eho "${pattern}" "${DOCS[@]}" | sort -u
}

ROADMAP_ARGS=()
CONSISTENCY_ARGS=()

if [[ "${BOOTSTRAP_FRESH_ARTIFACTS:-false}" == "true" ]]; then
  chmod +x scripts/public-footprint-snapshot.sh scripts/ci-health-snapshot.sh

  SNAPSHOT_REF="$(search_refs '\.llmsa/public-footprint/[0-9TZ]+/snapshot\.json' | head -n1 || true)"
  CI_REF="$(search_refs '\.llmsa/public-footprint/[0-9TZ]+/ci-health\.json' | head -n1 || true)"

  if [[ -z "${SNAPSHOT_REF}" || -z "${CI_REF}" ]]; then
    echo "error: could not resolve snapshot/ci-health references from core footprint docs" >&2
    exit 1
  fi

  FOOTPRINT_OUT_DIR="$(dirname "${SNAPSHOT_REF}")" ./scripts/public-footprint-snapshot.sh "${REPO}"
  FOOTPRINT_OUT_DIR="$(dirname "${CI_REF}")" ./scripts/ci-health-snapshot.sh "${REPO}"

  ROADMAP_ARGS=("${SNAPSHOT_REF}" "${CI_REF}")
  CONSISTENCY_ARGS=("${SNAPSHOT_REF}" "${CI_REF}")
fi

./scripts/workflow-health-check.sh "${REPO}"
FAIL_ON_INCOMPLETE=true ./scripts/roadmap-completion-check.sh "${ROADMAP_ARGS[@]}"
CONSISTENCY_SCOPE=core FAIL_ON_INCONSISTENT=true ./scripts/check-footprint-consistency.sh "${CONSISTENCY_ARGS[@]}"

echo "daily completion checks: healthy=true strict_complete=true consistency=true"
