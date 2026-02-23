# Evidence Pack (2026-02-18)

## Project

- Name: `LLM-Supply-Chain-Attestation (llmsa)`
- Repository: https://github.com/ogulcanaydogan/LLM-Supply-Chain-Attestation
- Reporting window: 2026-02-18 to 2026-03-19 (UTC, rolling 30-day execution window)
- Generated at (UTC): `2026-02-23T15:38:17Z`

## Evidence Summary

| Claim | Evidence Type | Date (UTC) | Public URL |
|---|---|---|---|
| Release shipped with signed artifacts | Release | 2026-02-19 | https://github.com/ogulcanaydogan/LLM-Supply-Chain-Attestation/releases/tag/v1.0.1 |
| Release workflow completed successfully | Workflow | 2026-02-19 | https://github.com/ogulcanaydogan/LLM-Supply-Chain-Attestation/actions/runs/22189290319 |
| Release verification completed successfully | Workflow | 2026-02-19 | https://github.com/ogulcanaydogan/LLM-Supply-Chain-Attestation/actions/runs/22201007496 |
| CI attestation gate enforced and passing | Workflow | 2026-02-20 | https://github.com/ogulcanaydogan/LLM-Supply-Chain-Attestation/actions/runs/22229255552 |
| Public-footprint snapshot workflow executed | Workflow | 2026-02-23 | https://github.com/ogulcanaydogan/LLM-Supply-Chain-Attestation/actions/runs/22292215111 |
| Tamper test suite executed (20 cases) | Benchmark/Security | 2026-02-23 | repository artifact path: `.llmsa/tamper/results.json` |
| Upstream contribution closed (unmerged) | External PR | 2026-02-19 | https://github.com/open-policy-agent/opa/pull/8343 |
| Upstream contribution closed (unmerged) | External PR | 2026-02-19 | https://github.com/open-policy-agent/opa/pull/8346 |
| Upstream contribution in review | External PR | 2026-02-19 | https://github.com/ossf/scorecard/pull/4942 |
| Upstream contribution merged | External PR | 2026-02-19 | https://github.com/sigstore/cosign/pull/4710 |
| Anonymous pilot case study published | Adoption | 2026-02-18 | `docs/public-footprint/case-study-anonymous-pilot-2026-02.md` |
| Third-party technical mention published (canonical) | Mention | 2026-02-18 | https://dev.to/ogulcanaydogan/i-spent-3-months-solving-a-security-gap-nobody-talks-about-llm-artifact-integrity-6co |
| Technical article published on Dev.to (LLMSA) | Article | 2026-02-19 | https://dev.to/ogulcanaydogan/i-spent-3-months-solving-a-security-gap-nobody-talks-about-llm-artifact-integrity-6co |
| Technical article published on Dev.to (AI Detection) | Article | 2026-02-19 | https://dev.to/ogulcanaydogan/how-i-built-an-ai-content-detection-system-from-scratch-oe4 |
| Technical article cross-posted to Medium (LLMSA) | Article | 2026-02-20 | https://medium.com/@ogulcanaydogan/i-spent-3-months-solving-a-security-gap-nobody-talks-about-llm-artifact-integrity-2d127da150d4 |
| Technical article cross-posted to Medium (AI Detection) | Article | 2026-02-20 | https://medium.com/@ogulcanaydogan/how-i-built-an-ai-content-detection-system-from-scratch-6db14942d844 |

## Metrics Snapshot

| Metric | Value | Source |
|---|---:|---|
| Upstream PRs opened | 4 | https://github.com/open-policy-agent/opa/pull/8343, https://github.com/open-policy-agent/opa/pull/8346, https://github.com/ossf/scorecard/pull/4942, https://github.com/sigstore/cosign/pull/4710 |
| Upstream PRs merged | 1 | https://github.com/open-policy-agent/opa/pull/8343, https://github.com/open-policy-agent/opa/pull/8346, https://github.com/ossf/scorecard/pull/4942, https://github.com/sigstore/cosign/pull/4710 |
| Upstream PRs in review | 1 | https://github.com/open-policy-agent/opa/pull/8343, https://github.com/open-policy-agent/opa/pull/8346, https://github.com/ossf/scorecard/pull/4942, https://github.com/sigstore/cosign/pull/4710 |
| Upstream PRs closed (unmerged) | 2 | https://github.com/open-policy-agent/opa/pull/8343, https://github.com/open-policy-agent/opa/pull/8346, https://github.com/ossf/scorecard/pull/4942, https://github.com/sigstore/cosign/pull/4710 |
| Third-party mentions | 5 | https://dev.to/ogulcanaydogan/i-spent-3-months-solving-a-security-gap-nobody-talks-about-llm-artifact-integrity-6co |
| Anonymous case studies | 1 | `docs/public-footprint/case-study-anonymous-pilot-2026-02.md` |
| Stars / forks / watchers | 0 / 0 / 0 | `.llmsa/public-footprint/20260223T153759Z/snapshot.json` |
| Release downloads (cumulative) | 324 | `.llmsa/public-footprint/20260223T153759Z/snapshot.json` |
| CI pass rate (last 30 days) | 94.12% (80/85) | `.llmsa/public-footprint/20260223T153809Z/ci-health.json` |
| CI pass rate (post-hardening window) | 98.57% (69/70) | `.llmsa/public-footprint/20260223T153809Z/ci-health.json` |
| Tamper detection success rate | 100.00% (20/20) | `.llmsa/tamper/results.json` |
| Verify p95 (100 statements) | 27.0 ms | `.llmsa/benchmarks/20260219T214353Z/summary.md` |

## Upstream PR Status Details

| URL | State | Merged | Merged At | Updated At |
|---|---|---|---|---|
| https://github.com/open-policy-agent/opa/pull/8343 | closed | false | n/a | 2026-02-19T16:27:58Z |
| https://github.com/open-policy-agent/opa/pull/8346 | closed | false | n/a | 2026-02-19T16:32:37Z |
| https://github.com/ossf/scorecard/pull/4942 | open | false | n/a | 2026-02-19T21:04:34Z |
| https://github.com/sigstore/cosign/pull/4710 | closed | true | 2026-02-19T21:49:44Z | 2026-02-19T21:59:28Z |

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
   - 1 upstream PR(s) are merged; 1 PR(s) remain open and pending maintainer decision.
   - 2 upstream PR(s) are closed-unmerged and count as non-converted evidence.

## Non-Claims Statement

Refer to `docs/public-footprint/what-we-do-not-claim.md`.
