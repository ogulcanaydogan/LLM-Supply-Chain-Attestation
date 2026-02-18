# Evidence Baseline (Day 0)

As of **2026-02-18 (UTC)**, this is the public evidence baseline for `llmsa`.

## Public Inventory

| Evidence Type | Current State (Day 0) | Public URL / Artifact |
|---|---|---|
| Repository | Public, default branch `main` | https://github.com/ogulcanaydogan/LLM-Supply-Chain-Attestation |
| Stars / Forks / Watchers | `0 / 0 / 0` | https://github.com/ogulcanaydogan/LLM-Supply-Chain-Attestation |
| Latest release | `v1.0.0` | https://github.com/ogulcanaydogan/LLM-Supply-Chain-Attestation/releases/tag/v1.0.0 |
| Total release asset downloads | `184` | https://github.com/ogulcanaydogan/LLM-Supply-Chain-Attestation/releases |
| Latest CI pass | `ci-attest-verify` success | https://github.com/ogulcanaydogan/LLM-Supply-Chain-Attestation/actions/runs/22149416407 |
| Release verification signal | manual verification success (latest), one earlier failed run visible | https://github.com/ogulcanaydogan/LLM-Supply-Chain-Attestation/actions/workflows/release-verify.yml |
| Security/tamper benchmark harness | present in repo (`scripts/tamper-tests.sh`, `scripts/benchmark.sh`) | `/Users/ogulcanaydogan/Desktop/Projects/YaPAY/ LLM Supply-Chain Attestation/scripts/` |
| Coverage snapshot (local) | `55.5%` statement coverage from `coverage.out` | `/Users/ogulcanaydogan/Desktop/Projects/YaPAY/ LLM Supply-Chain Attestation/coverage.out` |
| Upstream external PRs | `2` opened on 2026-02-18 | https://github.com/sigstore/cosign/pull/4710 ; https://github.com/open-policy-agent/opa/pull/8343 |
| Third-party mentions | none linked yet | `docs/public-footprint/third-party-mention-plan.md` |
| Anonymous pilot case study | published (repo-hosted) | `docs/public-footprint/case-study-anonymous-pilot-2026-02.md` |

## Core Gap Statement

Current signal quality is strong on implementation depth and release hygiene, but weak on third-party validation.  
The primary gap is external proof: upstream acceptance, independent mention, and adoption evidence.

## Day-30 Outcomes Required

1. At least `2` upstream PRs opened and `>=1` merged.
2. At least `1` third-party technical mention with a link.
3. At least `1` anonymous metrics-backed pilot case study.
4. `v1.0.1` hardening outputs visible through release/docs/workflow evidence.
5. A copy-ready evidence pack with claim-to-URL traceability.

## Week-2 Execution Update (2026-02-18)

1. Upstream Wave 1 complete for opening stage:
   - Sigstore/cosign PR opened.
   - OPA PR opened.
2. Public snapshot workflow manually triggered and completed:
   - https://github.com/ogulcanaydogan/LLM-Supply-Chain-Attestation/actions/runs/22157220499
3. Evidence pack draft with live URLs is now available:
   - `docs/public-footprint/evidence-pack-2026-02-18.md`

## Update Procedure

1. Run `scripts/public-footprint-snapshot.sh`.
2. Append delta values to `measurement-dashboard.md`.
3. Add net-new URLs into `evidence-pack-template.md` (or derived evidence pack document).
4. Keep all entries date-stamped in UTC.
