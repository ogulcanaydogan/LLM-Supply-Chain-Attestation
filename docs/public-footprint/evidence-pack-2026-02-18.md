# Evidence Pack (2026-02-18)

## Project

- Name: `LLM-Supply-Chain-Attestation (llmsa)`
- Repository: https://github.com/ogulcanaydogan/LLM-Supply-Chain-Attestation
- Reporting window: 2026-02-18 to 2026-03-19 (UTC, rolling 30-day execution window)
- Generated at (UTC): `2026-02-19T21:26:46Z`

## Evidence Summary

| Claim | Evidence Type | Date (UTC) | Public URL |
|---|---|---|---|
| Release shipped with signed artifacts | Release | 2026-02-19 | https://github.com/ogulcanaydogan/LLM-Supply-Chain-Attestation/releases/tag/v1.0.1 |
| Release workflow completed successfully | Workflow | 2026-02-19 | https://github.com/ogulcanaydogan/LLM-Supply-Chain-Attestation/actions/runs/22189290319 |
| Release verification completed successfully | Workflow | 2026-02-19 | https://github.com/ogulcanaydogan/LLM-Supply-Chain-Attestation/actions/runs/22189499214 |
| CI attestation gate enforced and passing | Workflow | 2026-02-19 | https://github.com/ogulcanaydogan/LLM-Supply-Chain-Attestation/actions/runs/22200750261 |
| Public-footprint snapshot workflow executed | Workflow | 2026-02-19 | https://github.com/ogulcanaydogan/LLM-Supply-Chain-Attestation/actions/runs/22198998266 |
| Tamper test suite executed (20 cases) | Benchmark/Security | 2026-02-19 | repository artifact path: `.llmsa/tamper/results.json` |
| Upstream contribution closed (unmerged) | External PR | 2026-02-19 | https://github.com/open-policy-agent/opa/pull/8343 |
| Upstream contribution closed (unmerged) | External PR | 2026-02-19 | https://github.com/open-policy-agent/opa/pull/8346 |
| Upstream contribution in review | External PR | 2026-02-19 | https://github.com/ossf/scorecard/pull/4942 |
| Upstream contribution in review | External PR | 2026-02-19 | https://github.com/sigstore/cosign/pull/4710 |
| Anonymous pilot case study published | Adoption | 2026-02-18 | `docs/public-footprint/case-study-anonymous-pilot-2026-02.md` |
| Third-party technical mention published | Mention | 2026-02-18 | https://gist.github.com/ogulcanaydogan/7cffe48a760a77cb42cb1f87644909bb |

## Metrics Snapshot

| Metric | Value | Source |
|---|---:|---|
| Upstream PRs opened | 4 | https://github.com/open-policy-agent/opa/pull/8343, https://github.com/open-policy-agent/opa/pull/8346, https://github.com/ossf/scorecard/pull/4942, https://github.com/sigstore/cosign/pull/4710 |
| Upstream PRs merged | 0 | https://github.com/open-policy-agent/opa/pull/8343, https://github.com/open-policy-agent/opa/pull/8346, https://github.com/ossf/scorecard/pull/4942, https://github.com/sigstore/cosign/pull/4710 |
| Upstream PRs in review | 2 | https://github.com/open-policy-agent/opa/pull/8343, https://github.com/open-policy-agent/opa/pull/8346, https://github.com/ossf/scorecard/pull/4942, https://github.com/sigstore/cosign/pull/4710 |
| Upstream PRs closed (unmerged) | 2 | https://github.com/open-policy-agent/opa/pull/8343, https://github.com/open-policy-agent/opa/pull/8346, https://github.com/ossf/scorecard/pull/4942, https://github.com/sigstore/cosign/pull/4710 |
| Third-party mentions | 1 | https://gist.github.com/ogulcanaydogan/7cffe48a760a77cb42cb1f87644909bb |
| Anonymous case studies | 1 | `docs/public-footprint/case-study-anonymous-pilot-2026-02.md` |
| Stars / forks / watchers | 0 / 0 / 0 | `.llmsa/public-footprint/20260219T212635Z/snapshot.json` |
| Release downloads (cumulative) | 282 | `.llmsa/public-footprint/20260219T212635Z/snapshot.json` |
| CI pass rate (last 30 days) | 86.54% (45/52) | `.llmsa/public-footprint/20260219T212641Z/ci-health.json` |
| CI pass rate (post-hardening window) | 100% (12/12) | `.llmsa/public-footprint/20260219T212641Z/ci-health.json` |
| Tamper detection success rate | 100.00% (20/20) | `.llmsa/tamper/results.json` |
| Verify p95 (100 statements) | 304.0 ms | `.llmsa/benchmarks/20260219T070954Z/summary.md` |

## Upstream PR Status Details

| URL | State | Merged | Merged At | Updated At |
|---|---|---|---|---|
| https://github.com/open-policy-agent/opa/pull/8343 | closed | false | n/a | 2026-02-19T16:27:58Z |
| https://github.com/open-policy-agent/opa/pull/8346 | closed | false | n/a | 2026-02-19T16:32:37Z |
| https://github.com/ossf/scorecard/pull/4942 | open | false | n/a | 2026-02-19T21:04:34Z |
| https://github.com/sigstore/cosign/pull/4710 | open | false | n/a | 2026-02-19T21:04:34Z |

## Reproducibility Notes

1. Commands:
   - `go test ./...`
   - `./scripts/benchmark.sh`
   - `./scripts/tamper-tests.sh`
   - `./scripts/public-footprint-snapshot.sh`
   - `./scripts/ci-health-snapshot.sh`
   - `./scripts/generate-evidence-pack.sh`
2. Environment notes:
   - GitHub Actions + local benchmark/tamper outputs.
3. Limitations:
   - merged-status external validation is still pending maintainer approval on open upstream PRs.
   - 2 upstream PR(s) are closed-unmerged and count as non-converted evidence.

## Non-Claims Statement

Refer to `docs/public-footprint/what-we-do-not-claim.md`.
