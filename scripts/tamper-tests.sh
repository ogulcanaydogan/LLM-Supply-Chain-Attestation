#!/usr/bin/env bash
set -euo pipefail

make attest
make sign

go build -o .llmsa/llmsa ./cmd/llmsa

cp examples/tiny-rag/app/system_prompt.txt examples/tiny-rag/app/system_prompt.txt.bak
echo "tampered" >> examples/tiny-rag/app/system_prompt.txt

set +e
.llmsa/llmsa verify --source local --attestations .llmsa/attestations --policy policy/examples/mvp-gates.yaml --format json --out verify.json
CODE=$?
set -e

mv examples/tiny-rag/app/system_prompt.txt.bak examples/tiny-rag/app/system_prompt.txt

if [[ "$CODE" -ne 12 ]]; then
  echo "expected exit code 12, got $CODE"
  exit 1
fi

echo "tamper test passed"
