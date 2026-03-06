# Measurement Dashboard (Day 0 -> Day 30)

This dashboard tracks public-footprint metrics only.
Day-0 values are captured on **2026-02-18 UTC**.

## Scoreboard

| Metric | Day 0 | Day 30 Target | Current Delta | Source |
|---|---:|---:|---:|---|
| Upstream PRs opened | 0 | >=2 | +5 | https://github.com/open-policy-agent/opa/pull/8343, https://github.com/open-policy-agent/opa/pull/8346, https://github.com/ossf/scorecard/pull/4942, https://github.com/sigstore/cosign/pull/4710, https://github.com/slsa-framework/slsa-github-generator/pull/4468 |
| Upstream PRs merged | 0 | >=1 | +1 | https://github.com/open-policy-agent/opa/pull/8343, https://github.com/open-policy-agent/opa/pull/8346, https://github.com/ossf/scorecard/pull/4942, https://github.com/sigstore/cosign/pull/4710, https://github.com/slsa-framework/slsa-github-generator/pull/4468 |
| Upstream PRs in review | 0 | <=2 | +2 | https://github.com/open-policy-agent/opa/pull/8343, https://github.com/open-policy-agent/opa/pull/8346, https://github.com/ossf/scorecard/pull/4942, https://github.com/sigstore/cosign/pull/4710, https://github.com/slsa-framework/slsa-github-generator/pull/4468 |
| Upstream PRs closed (unmerged) | 0 | <=1 | +2 | https://github.com/open-policy-agent/opa/pull/8343, https://github.com/open-policy-agent/opa/pull/8346, https://github.com/ossf/scorecard/pull/4942, https://github.com/sigstore/cosign/pull/4710, https://github.com/slsa-framework/slsa-github-generator/pull/4468 |
| Third-party mentions | 0 | >=1 | +5 | https://dev.to/ogulcanaydogan/i-spent-3-months-solving-a-security-gap-nobody-talks-about-llm-artifact-integrity-6co |
| Anonymous pilot case studies | 0 | >=1 | +1 | `docs/public-footprint/case-study-anonymous-pilot-2026-02.md` |
| GitHub stars | 0 | >=25 | 0 | `.llmsa/public-footprint/20260306T092913Z/snapshot.json` |
| GitHub forks | 0 | >=5 | 1 | `.llmsa/public-footprint/20260306T092913Z/snapshot.json` |
| GitHub watchers | 0 | >=5 | 0 | `.llmsa/public-footprint/20260306T092913Z/snapshot.json` |
| Release downloads (cumulative) | 184 | >=400 | 224 | `.llmsa/public-footprint/20260306T092913Z/snapshot.json` |
| CI pass rate (last 30 days) | 93.3% | >=95% | 95% | `.llmsa/public-footprint/20260306T095052Z/ci-health.json` |
| CI pass rate (post-hardening window) | n/a | >=95% | 95% | `.llmsa/public-footprint/20260306T095052Z/ci-health.json` |

## Current Snapshot (2026-03-06 UTC)

- Snapshot artifact (local): `.llmsa/public-footprint/20260306T092913Z/snapshot.json`
- CI health artifact (local): `.llmsa/public-footprint/20260306T095052Z/ci-health.json`
- CI pass rate source window: 57/60 successful runs (`95%`).
- CI pass rate post-hardening (`2026-02-19T16:08:22Z`): 57/60 successful runs (`95%`, meets target=`true`).
- External write-up URL: https://dev.to/ogulcanaydogan/i-spent-3-months-solving-a-security-gap-nobody-talks-about-llm-artifact-integrity-6co
- Upstream PR review stage:
  - `in-review`: 2
  - `closed`: 2
  - `merged`: 1
- Dashboard generated at: `2026-03-06T09:50:56Z`
