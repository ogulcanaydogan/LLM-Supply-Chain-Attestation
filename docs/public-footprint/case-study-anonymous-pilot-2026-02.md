# Anonymous Pilot Case Study (2026-02)

## 1. Context

- Industry profile: regulated software delivery (anonymized).
- Team size: small platform/security function.
- Environment: GitHub Actions + OCI publishing + Kubernetes admission path.
- Audit pressure: need for verifiable change evidence before deployment.

## 2. Problem Before `llmsa`

- LLM-critical artifacts (prompt, corpus, eval, route, SLO) were not cryptographically linked in one enforcement path.
- Policy checks were fragmented across scripts without consistent provenance output.
- Deployment admission had no attestation-aware control boundary.

## 3. Integration Steps

1. Enabled all five attestation types via `llmsa attest create`.
2. Added signing (`llmsa sign`), verification (`llmsa verify`), and gate enforcement (`llmsa gate`) in CI.
3. Added OCI publish path for bundle distribution and pull-based verification.
4. Added webhook flow for deployment-time validation testing.

## 4. Measurement Method

- Date window (UTC): 2026-02-18.
- Baseline method: script-driven benchmark and tamper harness comparison.
- Tooling and commands:
  - `./scripts/benchmark.sh`
  - `./scripts/tamper-tests.sh`
  - `./scripts/public-footprint-snapshot.sh`
- Sources:
  - `.llmsa/benchmarks/20260218T165828Z/summary.md`
  - `.llmsa/tamper/results.md`
  - workflow run: https://github.com/ogulcanaydogan/LLM-Supply-Chain-Attestation/actions/runs/22157220499

## 5. Metrics

| Metric | Before | After | Collection Method |
|---|---:|---:|---|
| Tamper detection success rate | n/a | 100% (20/20 seeded cases) | `.llmsa/tamper/results.md` |
| Verify p95 latency (100 statements) | n/a | 27 ms | `.llmsa/benchmarks/20260218T165828Z/summary.md` |
| Gate violation detection rate | n/a | 100% on seeded missing/invalid evidence scenarios | tamper + policy fixtures |
| Release assurance evidence completeness | fragmented | release + CI + benchmark + policy artifacts linkable in one pack | `docs/public-footprint/evidence-pack-2026-02-18.md` |

## 6. Reproducibility Commands

```bash
go test ./...
./scripts/benchmark.sh
./scripts/tamper-tests.sh
./scripts/public-footprint-snapshot.sh
```

## 7. Failure Modes Observed

- CI pass-rate trend dipped due an earlier failed release verification run in the 30-day window.
- Single-run performance values are not treated as universal; trend tracking is required.
- External validation (merged upstream PRs, independent mention) remains in progress.

## 8. What This Case Study Does Not Prove

- It does not prove runtime compromise prevention after admission.
- It does not prove universal latency behavior across all infrastructure profiles.
- It does not replace independent security review or formal compliance assessment.

## 9. Public Evidence Links

- Release: https://github.com/ogulcanaydogan/LLM-Supply-Chain-Attestation/releases/tag/v1.0.0
- CI workflow: https://github.com/ogulcanaydogan/LLM-Supply-Chain-Attestation/actions/runs/22157220499
- Upstream PRs:
  - https://github.com/sigstore/cosign/pull/4710
  - https://github.com/open-policy-agent/opa/pull/8343
