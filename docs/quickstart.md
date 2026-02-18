# Quickstart

This guide walks you through the complete LLM Supply-Chain Attestation pipeline — from bootstrapping a project to deploying attestation enforcement in Kubernetes. By the end, you will have cryptographically signed attestation bundles for all five LLM artifact types, verified their integrity, and enforced policy gates.

## Prerequisites

- **Go 1.25+** installed and available on your `$PATH`.
- **Git** repository (the tool uses git for change detection and provenance).
- **cosign** (optional) for Sigstore keyless signing. Install via:
  ```bash
  go install github.com/sigstore/cosign/v2/cmd/cosign@latest
  ```

## 1. Bootstrap the Project

Initialise the local store, default configuration, policy template, and a development signing key:

```bash
go run ./cmd/llmsa init
```

This creates:
- `.llmsa/attestations/` — local attestation store directory.
- `llmsa.yaml` — project configuration file.
- `policy/examples/mvp-gates.yaml` — default YAML policy gates.
- `.llmsa/dev_ed25519.pem` — Ed25519 private key for local development signing.

## 2. Generate Attestations

Create attestation statements for each of the five LLM artifact types. Each statement captures cryptographic digests of the referenced artifacts:

```bash
# Prompt attestation — system prompts, templates, tool schemas
go run ./cmd/llmsa attest create \
  --type prompt_attestation \
  --config examples/tiny-rag/configs/prompt.yaml \
  --out .llmsa/attestations

# Corpus attestation — training/retrieval data
go run ./cmd/llmsa attest create \
  --type corpus_attestation \
  --config examples/tiny-rag/configs/corpus.yaml \
  --out .llmsa/attestations

# Eval attestation — evaluation benchmarks and metrics
go run ./cmd/llmsa attest create \
  --type eval_attestation \
  --config examples/tiny-rag/configs/eval.yaml \
  --out .llmsa/attestations

# Route attestation — model routing configuration
go run ./cmd/llmsa attest create \
  --type route_attestation \
  --config examples/tiny-rag/configs/route.yaml \
  --out .llmsa/attestations

# SLO attestation — service-level objectives
go run ./cmd/llmsa attest create \
  --type slo_attestation \
  --config examples/tiny-rag/configs/slo.yaml \
  --out .llmsa/attestations
```

Each command outputs a `statement_*.json` file containing the attestation statement with subject digests, predicate data, and generator metadata.

### Changed-Only Mode

For CI pipelines, generate attestations only for artifact types whose source files have changed since the last commit:

```bash
go run ./cmd/llmsa attest create --changed-only --git-ref origin/main
```

### Determinism Validation

Verify that attestation generation is deterministic by running it multiple times and comparing hashes:

```bash
go run ./cmd/llmsa attest create \
  --type prompt_attestation \
  --config examples/tiny-rag/configs/prompt.yaml \
  --out .llmsa/attestations \
  --determinism-check 3
```

## 3. Sign Bundles

Wrap each statement into a DSSE (Dead Simple Signing Envelope) bundle with a cryptographic signature:

### Local PEM Signing (Development)

```bash
for s in .llmsa/attestations/statement_*.json; do
  go run ./cmd/llmsa sign \
    --in "$s" \
    --provider pem \
    --key .llmsa/dev_ed25519.pem \
    --out .llmsa/attestations
done
```

### Sigstore Keyless Signing (CI/CD)

In GitHub Actions, use OIDC-based keyless signing:

```bash
for s in .llmsa/attestations/statement_*.json; do
  go run ./cmd/llmsa sign \
    --in "$s" \
    --provider sigstore \
    --out .llmsa/attestations
done
```

Each signed bundle produces an `*.bundle.json` file containing the DSSE envelope, signature, public key material, and metadata.

## 4. Verify Attestations

Run the four-stage verification pipeline (signature → schema → digest → chain):

```bash
go run ./cmd/llmsa verify \
  --source local \
  --attestations .llmsa/attestations \
  --format json \
  --out verify.json
```

The verification engine checks:
1. **Signature** — DSSE envelope signature is valid against the embedded public key.
2. **Schema** — Statement conforms to the JSON Schema for its attestation type.
3. **Digest** — Subject file digests match the values recorded in the statement.
4. **Chain** — Provenance dependency graph is satisfied (eval→prompt+corpus, route→eval, slo→route).

Semantic exit codes:
| Code | Meaning |
|------|---------|
| `0` | All checks passed |
| `10` | Missing attestation bundles |
| `11` | Signature verification failed |
| `12` | Digest mismatch (tampering detected) |
| `13` | Policy gate violation |
| `14` | Schema validation failed |

## 5. Enforce Policy Gates

Evaluate attestations against policy rules to enforce governance requirements:

```bash
go run ./cmd/llmsa gate \
  --policy policy/examples/mvp-gates.yaml \
  --attestations .llmsa/attestations \
  --git-ref origin/main
```

### Using Rego Policies (OPA)

For more expressive rules, use the Open Policy Agent engine:

```bash
go run ./cmd/llmsa gate \
  --engine rego \
  --rego-policy policy/examples/rego-gates.rego \
  --policy policy/examples/mvp-gates.yaml \
  --attestations .llmsa/attestations
```

## 6. Generate Audit Report

Create a human-readable Markdown report from verification results:

```bash
go run ./cmd/llmsa report --in verify.json --out verify.md
```

## 7. Publish to OCI Registry

Distribute attestation bundles via OCI-compliant container registries:

```bash
go run ./cmd/llmsa publish \
  --in .llmsa/attestations/prompt.bundle.json \
  --oci ghcr.io/your-org/attestations:sha256-abc123
```

## 8. Run the Full Demo

Execute the complete pipeline end-to-end with the bundled `tiny-rag` example:

```bash
go run ./cmd/llmsa demo run
```

This runs: init → attest (all 5 types) → sign → verify → gate → report in a single command.

## Next Steps

- **Kubernetes Enforcement**: Deploy the validating admission webhook — see [K8s Admission Guide](k8s-admission.md).
- **Policy Customisation**: Write custom gates — see [Policy Guide](policy-guide.md).
- **CI/CD Integration**: Add to your GitHub Actions workflow — see `.github/workflows/ci-attest-verify.yml`.
- **Threat Model**: Understand the security properties — see [Threat Model](threat-model.md).
