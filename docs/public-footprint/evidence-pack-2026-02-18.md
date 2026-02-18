# Evidence Pack (2026-02-18)

## Project

- Name: `LLM-Supply-Chain-Attestation (llmsa)`
- Repository: https://github.com/ogulcanaydogan/LLM-Supply-Chain-Attestation
- Reporting window: 2026-02-18 to 2026-03-19 (UTC, rolling 30-day execution window)

## Evidence Summary

| Claim | Evidence Type | Date (UTC) | Public URL |
|---|---|---|---|
| Release shipped with signed artifacts | Release | 2026-02-18 | https://github.com/ogulcanaydogan/LLM-Supply-Chain-Attestation/releases/tag/v1.0.0 |
| CI attestation gate enforced and passing | Workflow | 2026-02-18 | https://github.com/ogulcanaydogan/LLM-Supply-Chain-Attestation/actions/runs/22149416407 |
| Public-footprint snapshot workflow executed | Workflow | 2026-02-18 | https://github.com/ogulcanaydogan/LLM-Supply-Chain-Attestation/actions/runs/22157220499 |
| Tamper test suite executed (20 cases) | Benchmark/Security | 2026-02-18 | repository artifact path: `.llmsa/tamper/results.md` |
| Upstream contribution opened (Sigstore) | External PR | 2026-02-18 | https://github.com/sigstore/cosign/pull/4710 |
| Upstream contribution opened (OPA) | External PR | 2026-02-18 | https://github.com/open-policy-agent/opa/pull/8343 |
| Upstream contribution opened (OpenSSF Scorecard) | External PR | 2026-02-18 | https://github.com/ossf/scorecard/pull/4942 |
| Anonymous pilot case study published | Adoption | 2026-02-18 | `docs/public-footprint/case-study-anonymous-pilot-2026-02.md` |
| Third-party technical mention published | Mention | 2026-02-18 | https://gist.github.com/ogulcanaydogan/7cffe48a760a77cb42cb1f87644909bb |

## Metrics Snapshot

| Metric | Value | Source |
|---|---:|---|
| Upstream PRs opened | 3 | `docs/public-footprint/external-contribution-log.md` |
| Upstream PRs merged | 0 | `docs/public-footprint/external-contribution-log.md` |
| Third-party mentions | 1 | gist URL above |
| Anonymous case studies | 1 | `docs/public-footprint/case-study-anonymous-pilot-2026-02.md` |
| Stars / forks / watchers | 0 / 0 / 0 | `.llmsa/public-footprint/20260218T205404Z/snapshot.json` |
| Release downloads (cumulative) | 184 | `.llmsa/public-footprint/20260218T205404Z/snapshot.json` |
| CI pass rate (last 30 days) | 89.47% (17/19) | `.llmsa/public-footprint/20260218T205404Z/snapshot.json` |
| Tamper detection success rate | 100% (20/20) | `.llmsa/tamper/results.md` |
| Verify p95 (100 statements) | 27 ms | `.llmsa/benchmarks/20260218T165828Z/summary.md` |

## Reproducibility Notes

1. Commands:
   - `go test ./...`
   - `./scripts/benchmark.sh`
   - `./scripts/tamper-tests.sh`
   - `./scripts/public-footprint-snapshot.sh`
2. `llmsa` commit for footprint toolkit:
   - `d3b5248`
3. Environment notes:
   - GitHub Actions + local benchmark harness outputs.
4. Limitations:
   - merged-status external validation is pending maintainer review on opened PRs.

## Non-Claims Statement

Refer to `docs/public-footprint/what-we-do-not-claim.md`.
