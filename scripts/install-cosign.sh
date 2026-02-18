#!/usr/bin/env bash
set -euo pipefail

if command -v cosign >/dev/null 2>&1; then
  echo "cosign already installed"
  exit 0
fi

echo "Install cosign from https://github.com/sigstore/cosign/releases"
