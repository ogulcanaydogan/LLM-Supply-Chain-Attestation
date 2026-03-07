#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "${ROOT}"

REPO="${1:-}"

chmod +x \
  scripts/public-footprint-snapshot.sh \
  scripts/ci-health-snapshot.sh \
  scripts/generate-evidence-pack.sh \
  scripts/roadmap-completion-check.sh \
  scripts/check-footprint-consistency.sh

./scripts/public-footprint-snapshot.sh "${REPO}"
./scripts/ci-health-snapshot.sh "${REPO}"
./scripts/generate-evidence-pack.sh "${REPO}"
FAIL_ON_INCOMPLETE=true ./scripts/roadmap-completion-check.sh
CONSISTENCY_SCOPE=core FAIL_ON_INCONSISTENT=true ./scripts/check-footprint-consistency.sh

echo "weekly completion refresh: snapshot+ci-health+evidence+completion+consistency done"
