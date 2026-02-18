# Policy Guide

Policy file: `policy/examples/mvp-gates.yaml`

## Gate model
- `trigger_paths`: file path globs that activate a gate.
- `required_attestations`: attestation types that must be present.
- `message`: CI failure text.

## Run locally
```bash
go run ./cmd/llmsa gate \
  --policy policy/examples/mvp-gates.yaml \
  --attestations .llmsa/attestations \
  --git-ref HEAD~1
```

## Privacy guard
Statements with `privacy.mode=plaintext_explicit` are blocked unless `statement_id` is in `plaintext_allowlist`.
