# LLM Supply-Chain Attestation

Local-first CLI + CI toolchain for tamper-evident LLM artifact attestations.

## What it does
- Creates typed attestations for prompt/corpus/eval/route/slo artifacts.
- Signs statements into DSSE bundles (Sigstore keyless with cosign, or PEM fallback).
- Verifies signatures, schemas, and subject digests.
- Enforces YAML policy gates in CI.
- Generates JSON and Markdown audit reports.

## Quick start
```bash
make init
make demo
```

## CLI commands
- `llmsa init`
- `llmsa attest create --type <prompt_attestation|corpus_attestation|eval_attestation|route_attestation|slo_attestation> --config <path> --out <dir>`
- `llmsa attest create --changed-only --git-ref <ref> --out <dir>`
- `llmsa sign --in <statement.json> --provider <sigstore|pem|kms> [--out <bundle.json|dir>]`
- `llmsa publish --in <bundle.json> --oci <registry/repo:tag>`
- `llmsa verify --source <local|oci> --attestations <dir|oci-ref[,oci-ref...]> --policy <file> --format <json|md> --out <file>`
- `llmsa gate --source <local|oci> --policy <file> --attestations <dir|oci-ref[,oci-ref...]>`
- `llmsa report --in <verify.json> --out <verify.md>`
- `llmsa demo run`

## Exit codes
- `0`: pass
- `10`: missing attestation/bundle
- `11`: signature failure
- `12`: digest mismatch/tamper
- `13`: policy violation
- `14`: schema/version incompatibility
