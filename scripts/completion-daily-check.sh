#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "${ROOT}"

REPO="${1:-}"

chmod +x scripts/workflow-health-check.sh scripts/roadmap-completion-check.sh scripts/check-footprint-consistency.sh

./scripts/workflow-health-check.sh "${REPO}"
FAIL_ON_INCOMPLETE=true ./scripts/roadmap-completion-check.sh
CONSISTENCY_SCOPE=core FAIL_ON_INCONSISTENT=true ./scripts/check-footprint-consistency.sh

echo "daily completion checks: healthy=true strict_complete=true consistency=true"
