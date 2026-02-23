# Day-30 Outcomes (2026-02)

## Scope

This outcome summary covers the 30-day public-footprint plan execution for `llmsa`.

## Targets vs Actuals

| Target | Expected | Actual | Status | Evidence |
|---|---:|---:|---|---|
| Upstream PRs opened | >=2 | 4 | achieved | `docs/public-footprint/external-contribution-log.md` |
| Upstream PRs merged | >=1 | 1 | achieved | `docs/public-footprint/external-contribution-log.md` |
| Third-party mention count | >=1 | 5 (canonical Dev.to + additional publications) | achieved | `docs/public-footprint/evidence-pack-2026-02-18.md` |
| Anonymous pilot case study | >=1 | 1 | achieved | `docs/public-footprint/case-study-anonymous-pilot-2026-02.md` |
| Hardening release closure | v1.0.1 complete evidence | complete | achieved | https://github.com/ogulcanaydogan/LLM-Supply-Chain-Attestation/actions/runs/22189290319 |
| CI pass-rate (rolling 30-day) | >=95% | 97.65% (83/85) | achieved | `.llmsa/public-footprint/20260223T213311Z/ci-health.json` |
| CI pass-rate (post-hardening baseline) | >=95% | 98.78% (81/82) | achieved | `.llmsa/public-footprint/20260223T213311Z/ci-health.json` |

## What Worked

1. `v1.0.1` release and verification path closed cleanly with signed assets.
2. Weekly footprint automation now produces snapshot + CI health + generated evidence docs.
3. Upstream cadence process became explicit and traceable (48h follow-up workflow), and maintainer-requested revisions were pushed to `cosign#4710`.
4. Reliability trend after hardening is now explicit and measurable (`post_hardening_pass_rate_percent` in CI health snapshots).
5. Canonical third-party mention gate is now met with a non-GitHub source URL.

## What Did Not Convert

1. One upstream PR remains open (`scorecard#4942`) and not yet converted to merged evidence.
2. OPA docs contribution line was closed twice (`#8343`, `#8346`) with maintainer guidance to keep this pattern in project docs.
3. One upstream PR remains open (`scorecard#4942`), so external-validation depth can still be improved beyond the minimum merged target.

## Next Cycle Priorities

1. Convert a second merge from remaining open PR (`scorecard#4942`) if maintainers accept scope.
2. Maintain CI/workflow stability above strict threshold and rerun consistency checks on each evidence refresh.
3. Refresh and freeze evidence pack + roadmap outputs after each upstream PR status change.

## Runtime/Infra Profile

- Current execution model: GitHub Actions workflows plus local Go CLI tooling.
- Not used in this repository runtime/process: NVIDIA `V100`, NVIDIA `A100`, Apache Spark.
- Update trigger: add/update this section only if GPU/Spark benchmark execution is introduced.
