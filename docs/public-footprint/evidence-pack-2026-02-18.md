# Evidence Pack (2026-02-18)

## Project

- Name: `LLM-Supply-Chain-Attestation (llmsa)`
- Repository: https://github.com/ogulcanaydogan/LLM-Supply-Chain-Attestation
- Reporting window: 2026-02-18 to 2026-03-19 (UTC, rolling 30-day execution window)
- Last evidence refresh: **2026-02-19 UTC (post v1.0.1 release)**

## Evidence Summary

| Claim | Evidence Type | Date (UTC) | Public URL |
|---|---|---|---|
| Release shipped with signed artifacts | Release | 2026-02-19 | https://github.com/ogulcanaydogan/LLM-Supply-Chain-Attestation/releases/tag/v1.0.1 |
| Release workflow for `v1.0.1` completed successfully | Workflow | 2026-02-19 | https://github.com/ogulcanaydogan/LLM-Supply-Chain-Attestation/actions/runs/22189290319 |
| Release verification for `v1.0.1` succeeded (checksum + cosign) | Workflow | 2026-02-19 | https://github.com/ogulcanaydogan/LLM-Supply-Chain-Attestation/actions/runs/22189499214 |
| CI attestation gate enforced | Workflow | 2026-02-19 | https://github.com/ogulcanaydogan/LLM-Supply-Chain-Attestation/actions/workflows/ci-attest-verify.yml |
| Public-footprint weekly workflow active (latest run succeeded) | Workflow | 2026-02-19 | https://github.com/ogulcanaydogan/LLM-Supply-Chain-Attestation/actions/runs/22189686718 |
| Upstream follow-up cadence workflow active (48h) | Workflow | 2026-02-19 | https://github.com/ogulcanaydogan/LLM-Supply-Chain-Attestation/actions/workflows/upstream-pr-followup.yml |
| Tamper test suite executed (20 cases) | Benchmark/Security | 2026-02-19 | repository artifact path: `.llmsa/tamper/results.md` |
| Upstream contribution opened (Sigstore) | External PR | 2026-02-18 | https://github.com/sigstore/cosign/pull/4710 |
| Upstream contribution opened (OPA) | External PR | 2026-02-18 | https://github.com/open-policy-agent/opa/pull/8343 |
| Upstream fallback contribution opened (OPA reduced scope) | External PR | 2026-02-19 | https://github.com/open-policy-agent/opa/pull/8346 |
| Upstream contribution opened (OpenSSF Scorecard) | External PR | 2026-02-18 | https://github.com/ossf/scorecard/pull/4942 |
| Maintainer follow-up posted (Sigstore) | External PR Comment | 2026-02-19 | https://github.com/sigstore/cosign/pull/4710#issuecomment-3927950828 |
| Maintainer follow-up posted (OPA) | External PR Comment | 2026-02-19 | https://github.com/open-policy-agent/opa/pull/8343#issuecomment-3927950855 |
| Maintainer follow-up posted (Scorecard) | External PR Comment | 2026-02-19 | https://github.com/ossf/scorecard/pull/4942#issuecomment-3927950830 |
| OPA upstream PR closed (unmerged) | External PR Status | 2026-02-19 | https://github.com/open-policy-agent/opa/pull/8343 |
| OPA fallback PR status (open) | External PR Status | 2026-02-19 | https://github.com/open-policy-agent/opa/pull/8346 |
| Anonymous pilot case study published | Adoption | 2026-02-18 | `docs/public-footprint/case-study-anonymous-pilot-2026-02.md` |
| Third-party technical mention published | Mention | 2026-02-18 | https://gist.github.com/ogulcanaydogan/7cffe48a760a77cb42cb1f87644909bb |

## Metrics Snapshot

| Metric | Value | Source |
|---|---:|---|
| Upstream PRs opened | 4 | `docs/public-footprint/external-contribution-log.md` |
| Upstream PRs merged | 0 | https://github.com/sigstore/cosign/pull/4710, https://github.com/open-policy-agent/opa/pull/8346, https://github.com/ossf/scorecard/pull/4942 |
| Upstream PRs in review | 3 | `docs/public-footprint/external-contribution-log.md` |
| Upstream PRs closed (unmerged) | 1 | https://github.com/open-policy-agent/opa/pull/8343 |
| Third-party mentions | 1 | https://gist.github.com/ogulcanaydogan/7cffe48a760a77cb42cb1f87644909bb |
| Anonymous case studies | 1 | `docs/public-footprint/case-study-anonymous-pilot-2026-02.md` |
| Stars / forks / watchers | 0 / 0 / 0 | `.llmsa/public-footprint/20260218T205404Z/snapshot.json` |
| Release downloads (cumulative) | 184 | `.llmsa/public-footprint/20260218T205404Z/snapshot.json` |
| CI pass rate (last 30 days) | 83% (34/41) | https://github.com/ogulcanaydogan/LLM-Supply-Chain-Attestation/actions/runs/22189686718 |
| Tamper detection success rate | 100% (20/20) | `.llmsa/tamper/results.md` |
| Verify p95 (100 statements) | 304 ms | `.llmsa/benchmarks/20260219T070954Z/summary.md` |

## Reproducibility Notes

1. Commands:
   - `go test ./...`
   - `./scripts/benchmark.sh`
   - `./scripts/tamper-tests.sh`
   - `./scripts/public-footprint-snapshot.sh`
   - `./scripts/ci-health-snapshot.sh`
   - `./scripts/generate-evidence-pack.sh`
2. Current limitations:
   - external merged PR evidence is still pending maintainer approval.
   - one upstream PR (OPA #8343) is closed, while reduced-scope replacement PR #8346 is open and awaiting review.
   - current mention is hosted on Gist; canonical non-GitHub publication is still pending.
3. Reliability hardening in progress:
   - OCI verification and release-asset download steps now include retries and preflight checks in workflow YAML.

## Non-Claims Statement

Refer to `docs/public-footprint/what-we-do-not-claim.md`.
