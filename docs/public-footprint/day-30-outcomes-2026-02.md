# Day-30 Outcomes (2026-02)

## Scope

This outcome summary covers the 30-day public-footprint plan execution for `llmsa`.

## Targets vs Actuals

| Target | Expected | Actual | Status | Evidence |
|---|---:|---:|---|---|
| Upstream PRs opened | >=2 | 4 | achieved | `docs/public-footprint/external-contribution-log.md` |
| Upstream PRs merged | >=1 | 1 | achieved | `docs/public-footprint/external-contribution-log.md` |
| Third-party mention count | >=1 | 1 (gist mirror) | partially met | `docs/public-footprint/evidence-pack-2026-02-18.md` |
| Anonymous pilot case study | >=1 | 1 | achieved | `docs/public-footprint/case-study-anonymous-pilot-2026-02.md` |
| Hardening release closure | v1.0.1 complete evidence | complete | achieved | https://github.com/ogulcanaydogan/LLM-Supply-Chain-Attestation/actions/runs/22189290319 |
| CI pass-rate (rolling 30-day) | >=95% | 87.72% (50/57) | not met | `.llmsa/public-footprint/20260219T215312Z/ci-health.json` |
| CI pass-rate (post-hardening baseline) | >=95% | 100% (17/17) | achieved | `.llmsa/public-footprint/20260219T215312Z/ci-health.json` |

## What Worked

1. `v1.0.1` release and verification path closed cleanly with signed assets.
2. Weekly footprint automation now produces snapshot + CI health + generated evidence docs.
3. Upstream cadence process became explicit and traceable (48h follow-up workflow), and maintainer-requested revisions were pushed to `cosign#4710`.
4. Reliability trend after hardening is now explicit and measurable (`post_hardening_pass_rate_percent` in CI health snapshots).

## What Did Not Convert

1. One upstream PR remains open (`scorecard#4942`) and not yet converted to merged evidence.
2. OPA docs contribution line was closed twice (`#8343`, `#8346`) with maintainer guidance to keep this pattern in project docs.
3. Third-party mention is still gist-based; canonical non-GitHub publication is pending because `DEVTO_API_KEY` is not yet configured (latest failed secret preflight run: https://github.com/ogulcanaydogan/LLM-Supply-Chain-Attestation/actions/runs/22200978995).

## Next Cycle Priorities

1. Convert a second merge from remaining open PR (`scorecard#4942`) if maintainers accept scope.
2. Publish canonical external post (Dev.to/Medium/newsletter) and mirror back into evidence pack.
3. Keep CI/workflows green so pass-rate rises as older failed runs age out of the 30-day denominator.
