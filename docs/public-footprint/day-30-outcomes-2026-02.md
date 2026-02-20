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
| CI pass-rate (rolling 30-day) | >=95% | 92.22% (83/90) | not met | `.llmsa/public-footprint/20260220T104215Z/ci-health.json` |
| CI pass-rate (post-hardening baseline) | >=95% | 98.11% (52/53) | achieved | `.llmsa/public-footprint/20260220T104215Z/ci-health.json` |

## What Worked

1. `v1.0.1` release and verification path closed cleanly with signed assets.
2. Weekly footprint automation now produces snapshot + CI health + generated evidence docs.
3. Upstream cadence process became explicit and traceable (48h follow-up workflow), and maintainer-requested revisions were pushed to `cosign#4710`.
4. Reliability trend after hardening is now explicit and measurable (`post_hardening_pass_rate_percent` in CI health snapshots).
5. Canonical third-party mention gate is now met with a non-GitHub source URL.

## What Did Not Convert

1. One upstream PR remains open (`scorecard#4942`) and not yet converted to merged evidence.
2. OPA docs contribution line was closed twice (`#8343`, `#8346`) with maintainer guidance to keep this pattern in project docs.
3. Strict closure is still blocked by rolling 30-day CI pass-rate remaining below `>=95%`.

## Next Cycle Priorities

1. Convert a second merge from remaining open PR (`scorecard#4942`) if maintainers accept scope.
2. Keep CI/workflows green so pass-rate rises as older failed runs age out of the 30-day denominator.
3. Refresh evidence pack and roadmap completion output after each successful CI cycle.
