# Policy Guide

This guide explains how to write and configure policy gates for the LLM Supply-Chain Attestation framework. Policies define which attestation types are required for specific file paths, controlling when the CI/CD pipeline blocks on missing or non-compliant attestations.

## Policy Engines

The framework supports two policy engines:

1. **YAML Gate Engine** (default) — Declarative trigger-path rules with required attestation types.
2. **Rego (OPA) Engine** — Expressive Open Policy Agent rules for advanced logic.

## YAML Policy Structure

Policy files use YAML format with the following top-level fields:

```yaml
version: "1"
oidc_issuer: https://token.actions.githubusercontent.com
identity_regex: '^https://github\.com/your-org/.+/.github/workflows/.+@refs/.+$'
plaintext_allowlist:
  - stmt-abc-123
gates:
  - id: G001
    trigger_paths:
      - app/prompts/**
    required_attestations:
      - prompt_attestation
    message: "Prompt attestation required for prompt changes"
```

### Top-Level Fields

| Field | Required | Description |
|-------|----------|-------------|
| `version` | Yes | Policy schema version (currently `"1"`) |
| `oidc_issuer` | No | Expected OIDC issuer for Sigstore signatures |
| `identity_regex` | No | Regex pattern for allowed signing identities |
| `plaintext_allowlist` | No | List of statement IDs allowed to use `plaintext_explicit` privacy mode |
| `gates` | Yes | Array of gate rules |

### Gate Fields

| Field | Required | Description |
|-------|----------|-------------|
| `id` | Yes | Unique gate identifier (e.g., `G001`) |
| `trigger_paths` | Yes | File path patterns that activate this gate |
| `required_attestations` | Yes | Attestation types that must be present when gate triggers |
| `message` | No | Custom error message (defaults to `"<id> missing attestations: <types>"`) |

### Path Pattern Matching

Trigger paths support two pattern formats:

- **Double-star suffix** (`app/**`): Matches any file under the directory, including nested subdirectories. The pattern `app/**` matches `app/main.go`, `app/sub/deep.go`, and the directory `app` itself.
- **Standard glob** (`*.yaml`, `config.json`): Matches files using Go's `filepath.Match` rules.

## Writing Effective Gates

### Example: Full LLM Pipeline

```yaml
version: "1"
oidc_issuer: https://token.actions.githubusercontent.com
identity_regex: '^https://github\.com/acme-corp/.+'
plaintext_allowlist: []
gates:
  - id: G001-prompts
    trigger_paths:
      - app/prompts/**
      - app/templates/**
    required_attestations:
      - prompt_attestation
    message: "Prompt changes require a prompt attestation"

  - id: G002-data
    trigger_paths:
      - data/**
      - embeddings/**
    required_attestations:
      - corpus_attestation
    message: "Data changes require a corpus attestation"

  - id: G003-eval
    trigger_paths:
      - benchmarks/**
      - eval/**
    required_attestations:
      - eval_attestation
    message: "Evaluation changes require an eval attestation"

  - id: G004-routing
    trigger_paths:
      - config/routes/**
      - app/router/**
    required_attestations:
      - route_attestation
    message: "Routing changes require a route attestation"

  - id: G005-slo
    trigger_paths:
      - config/slo/**
      - monitoring/**
    required_attestations:
      - slo_attestation
    message: "SLO changes require an SLO attestation"
```

### Example: Strict Mode (All Types Required)

```yaml
version: "1"
gates:
  - id: G-STRICT
    trigger_paths:
      - "**"
    required_attestations:
      - prompt_attestation
      - corpus_attestation
      - eval_attestation
      - route_attestation
      - slo_attestation
    message: "All five attestation types are required for any change"
```

## Privacy Policy

The `plaintext_allowlist` field controls which attestation statements are permitted to use `plaintext_explicit` privacy mode. This prevents accidental exposure of sensitive IP (model weights, proprietary prompts) in attestation bundles.

Statements with `privacy.mode: plaintext_explicit` that are not in the allowlist will trigger a policy violation: `"Sensitive payload exposure blocked by policy."`

To allow specific statements:
```yaml
plaintext_allowlist:
  - stmt-public-prompt-v1
  - stmt-open-eval-benchmark
```

## Running the YAML Gate Engine

```bash
go run ./cmd/llmsa gate \
  --policy policy/examples/mvp-gates.yaml \
  --attestations .llmsa/attestations \
  --git-ref origin/main
```

The engine:
1. Runs `git diff --name-only <ref>...HEAD` to identify changed files.
2. For each gate, checks if any changed file matches the trigger paths.
3. For triggered gates, verifies all required attestation types are present.
4. Returns violations for any missing attestation types.

## Rego (OPA) Policy Engine

For advanced policy logic, use the Rego engine. Rego policies can express conditions beyond simple path matching, including cross-attestation constraints, time-based rules, and custom validation logic.

### Example Rego Policy

```rego
package llmsa.gate

import rego.v1

default allow := false

violations contains msg if {
    some st in input.statements
    st.privacy_mode == "plaintext_explicit"
    not st.statement_id in input.plaintext_allowlist
    msg := "Sensitive payload exposure blocked by Rego policy"
}

violations contains msg if {
    required := {"prompt_attestation", "corpus_attestation", "eval_attestation"}
    present := {st.attestation_type | some st in input.statements}
    missing := required - present
    count(missing) > 0
    msg := sprintf("Missing required attestations: %v", [missing])
}
```

### Running Rego Policies

```bash
go run ./cmd/llmsa gate \
  --engine rego \
  --rego-policy policy/examples/rego-gates.rego \
  --policy policy/examples/mvp-gates.yaml \
  --attestations .llmsa/attestations
```

## CI/CD Integration

Add policy enforcement to your GitHub Actions workflow:

```yaml
- name: Gate
  run: |
    go run ./cmd/llmsa gate \
      --policy policy/examples/mvp-gates.yaml \
      --attestations .llmsa/attestations \
      --git-ref origin/main
```

The gate command exits with code 13 when policy violations are detected, which fails the CI step.

## Troubleshooting

**Gate not triggering**: Verify that changed files match the `trigger_paths` patterns. Use `git diff --name-only origin/main...HEAD` to see what files the engine detects.

**Unexpected violations**: Check that attestation types in `required_attestations` exactly match the `attestation_type` field in your statement files (e.g., `prompt_attestation`, not `prompt`).

**Plaintext blocked**: If you intentionally use `plaintext_explicit` mode, add the statement ID to `plaintext_allowlist` in the policy file.
