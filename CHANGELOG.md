# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).

## [Unreleased]

### Added
- Public-footprint documentation pack under `docs/public-footprint/` covering evidence baseline, upstream contribution briefs, mention plan, anonymous case study template, non-claims, and evidence-pack template.
- `scripts/public-footprint-snapshot.sh` for reproducible public-metrics snapshots (stars/forks/releases/CI pass rate/PR activity).
- Weekly metrics artifact workflow `.github/workflows/public-footprint-weekly.yml`.
- Kubernetes validating admission webhook (`internal/webhook/`) for deployment-time attestation enforcement.
- `llmsa webhook serve` CLI command with TLS, fail-open, and policy configuration flags.
- Kubernetes manifests (`deploy/webhook/`) and Helm chart (`deploy/helm/`) for webhook deployment.
- E2E integration test suite (`test/e2e/`) covering full pipeline, tamper detection, chain verification, and webhook admit/deny flows.
- Test coverage for all collector types (corpus, eval, route, slo), hash, sign, and verify packages.
- Test coverage for `internal/policy/yaml`, `internal/report`, `internal/store`, and `pkg/types` packages.
- Apache 2.0 LICENSE file.
- Dockerfile with multi-stage distroless build.
- GoReleaser configuration for multi-platform binary releases and container images.
- Privacy-aware attestation modes: `hash_only`, `plaintext_explicit`, `encrypted_payload` (age X25519).
- Rego (OPA) policy engine alongside existing YAML gate engine.
- Provenance chain verification with dependency graph, temporal ordering, and reference validation.

### Changed
- `docs/benchmark-methodology.md` rewritten with governance-yield-vs-overhead framing, strict raw artifact contract, and publication rules.
- `README.md` tightened with single-message positioning, explicit non-claims section, and public-footprint documentation links.
- README overhaul with 6 Mermaid diagrams (architecture, provenance chain, verification pipeline, privacy modes, K8s admission flow, CI/CD pipeline).
- Expanded CONTRIBUTING.md, SECURITY.md, and GOVERNANCE.md with comprehensive community documentation.
- Expanded docs/k8s-admission.md from placeholder to full deployment guide.

## [0.3.0] - 2025-06-15

### Added
- Release verification workflow with automated post-release attestation validation.
- Signer regression tests for Sigstore certificate parse variability.
- Auto-triggered release verification after successful release workflow.

## [0.2.0] - 2025-06-01

### Added
- OCI registry publish and pull with digest-pinned references.
- Sigstore keyless signing with OIDC identity binding (GitHub Actions).
- OCI round-trip verification in CI pipeline.
- Identity policy alignment for Sigstore provider metadata.
- Hardened bundle naming conventions.

### Fixed
- Tolerate Sigstore certificate parse variability across provider versions.

## [0.1.0] - 2025-05-15

### Added
- Initial MVP CLI: `init`, `attest create`, `sign`, `verify`, `gate`, `report`, `demo run`.
- Five attestation types: prompt, corpus, eval, route, SLO.
- DSSE bundle signing with Sigstore and PEM providers.
- YAML policy gate engine with path-based triggers.
- JSON Schema validation for all statement types.
- Subject digest recomputation during verification.
- Semantic exit codes: 0 (pass), 10 (missing), 11 (signature), 12 (tamper), 13 (policy), 14 (schema).
- Determinism validation with `--determinism-check` flag.
- Git-aware `--changed-only` attestation generation.
- Markdown and JSON audit report generation.
- `examples/tiny-rag/` end-to-end working demo.
- 20-case tamper detection test suite (`scripts/tamper-tests.sh`).
- Performance benchmark suite (`scripts/benchmark.sh`).
- GitHub Actions CI workflow with full attestation pipeline.
- Threat model, policy guide, quickstart, and benchmark methodology documentation.
