# Contributing to LLM Supply-Chain Attestation

Thank you for considering contributing. This document provides guidelines for development, testing, and submitting contributions.

## Prerequisites

- Go 1.25 or later
- Git
- Make
- [cosign](https://docs.sigstore.dev/cosign/installation/) (optional, for Sigstore keyless signing)
- [age](https://age-encryption.org/) (optional, for encrypted payload tests)

## Development Setup

```bash
git clone https://github.com/ogulcanaydogan/LLM-Supply-Chain-Attestation.git
cd LLM-Supply-Chain-Attestation
go mod download
go build -o llmsa ./cmd/llmsa
./llmsa init
```

Run the full demo to verify your environment:

```bash
make demo
```

## Running Tests

### Unit Tests

```bash
make test                          # Run all tests
go test -v ./internal/attest/      # Verbose output for a specific package
go test -cover ./...               # Coverage report per package
```

### Tamper Detection Suite

The project includes a 20-case tamper detection suite that validates signature, digest, schema, and chain integrity:

```bash
./scripts/tamper-tests.sh
```

### Benchmark Suite

Performance benchmarks for signing, verification, and policy evaluation:

```bash
./scripts/benchmark.sh
```

### E2E Integration Tests

Full pipeline tests covering attestation creation, signing, verification, and webhook admission:

```bash
go test -tags=e2e -v ./test/e2e/
```

### Coverage Targets

Current package coverage targets:

| Package | Target |
|---------|--------|
| `internal/verify` | >= 85% |
| `internal/sign` | >= 75% |
| `internal/attest` | >= 80% |
| `internal/store` | >= 80% |
| `internal/policy/yaml` | >= 90% |
| `internal/policy/rego` | >= 90% |
| `pkg/schema` | 100% |
| `pkg/types` | 100% |

## Code Style

- Format code with `gofmt` before committing.
- Wrap errors with context: `fmt.Errorf("description: %w", err)`.
- Follow Go naming conventions: exported identifiers in CamelCase, JSON/YAML fields in snake_case.
- Keep functions focused and under 60 lines where practical.
- Add comments for all exported functions, types, and constants.

### Package Organisation

```
cmd/llmsa/       CLI entry point and command definitions
internal/        Non-exported implementation packages
  attest/        Typed collectors for each attestation type
  sign/          DSSE bundle creation and signing providers
  verify/        Multi-stage verification engine
  policy/        YAML and Rego policy engines
  hash/          SHA-256 digest and canonical JSON utilities
  store/         Local filesystem and OCI registry storage
  report/        JSON and Markdown report generation
pkg/             Exported packages (types, schema)
```

## Adding a New Attestation Type

To add a new attestation type (for example, `deployment_attestation`):

1. **Define the predicate type** in `pkg/types/predicate_deployment.go`.
2. **Create the collector** in `internal/attest/collectors_deployment.go` following the pattern in `collectors_prompt.go`.
3. **Add a JSON Schema** in `schemas/v1/deployment_attestation.schema.json`.
4. **Register the type** in `cmd/llmsa/main.go` under the `attest create` command switch.
5. **Add the constant** `AttestationDeployment` in `pkg/types/statement.go` and update `PredicateURI`.
6. **Write tests** in `internal/attest/collectors_deployment_test.go` covering happy path, missing required fields, and optional fields.
7. **Add an example config** in `examples/tiny-rag/configs/`.
8. **Update documentation**: README attestation type table, CHANGELOG.

## Architecture Decision Records

Key design decisions are documented as ADRs in [`docs/adr/`](docs/adr/):

| ADR | Title |
|-----|-------|
| [0001](docs/adr/0001-llm-specific-attestation-taxonomy.md) | LLM-Specific Attestation Taxonomy |
| [0002](docs/adr/0002-dsse-envelope-format.md) | DSSE Envelope Format |
| [0003](docs/adr/0003-dual-policy-engine.md) | Dual Policy Engine (YAML + Rego) |
| [0004](docs/adr/0004-privacy-modes.md) | Privacy Modes |
| [0005](docs/adr/0005-provenance-chain-dag.md) | Provenance Chain DAG |
| [0006](docs/adr/0006-sigstore-keyless-signing.md) | Sigstore Keyless Signing |
| [0007](docs/adr/0007-semantic-exit-codes.md) | Semantic Exit Codes |

When proposing architectural changes, create a new ADR following the template:
- **Status**: Proposed → Accepted/Rejected/Superseded
- **Context**: What problem are we solving?
- **Decision**: What did we decide and why?
- **Consequences**: What are the trade-offs?

## Submitting Pull Requests

### Before Opening a PR

1. Run `make test` and confirm all tests pass.
2. Run `go vet ./...` and fix any warnings.
3. Ensure your changes maintain or increase test coverage.
4. Update CHANGELOG.md if adding features or fixing bugs.

### PR Guidelines

- One logical change per PR.
- Use conventional commit prefixes in the PR title: `feat:`, `fix:`, `chore:`, `docs:`, `refactor:`, `test:`.
- Reference related issues in the description.
- Keep schema changes backward-compatible for `1.x` releases.
- Add fixture coverage for new collectors or policy rules.

### Review Process

PRs require approval from a core maintainer and passing CI checks before merge.

## Release Process

Releases are automated via GoReleaser. To create a release:

1. Update CHANGELOG.md with the new version section.
2. Tag the version: `git tag -a v0.4.0 -m "Release v0.4.0"`
3. Push the tag: `git push origin v0.4.0`
4. GitHub Actions builds multi-platform binaries and publishes the release.
