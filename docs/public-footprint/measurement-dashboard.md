# Measurement Dashboard (Day 0 -> Day 30)

This dashboard tracks public-footprint metrics only.
Day-0 values are captured on **2026-02-18 UTC**.

## Scoreboard

| Metric | Day 0 | Day 30 Target | Current Delta | Source |
|---|---:|---:|---:|---|
| Upstream PRs opened | 0 | >=2 | +4 | https://github.com/open-policy-agent/opa/pull/8343, https://github.com/open-policy-agent/opa/pull/8346, https://github.com/ossf/scorecard/pull/4942, https://github.com/sigstore/cosign/pull/4710 |
| Upstream PRs merged | 0 | >=1 | +1 | https://github.com/open-policy-agent/opa/pull/8343, https://github.com/open-policy-agent/opa/pull/8346, https://github.com/ossf/scorecard/pull/4942, https://github.com/sigstore/cosign/pull/4710 |
| Upstream PRs in review | 0 | <=2 | +1 | https://github.com/open-policy-agent/opa/pull/8343, https://github.com/open-policy-agent/opa/pull/8346, https://github.com/ossf/scorecard/pull/4942, https://github.com/sigstore/cosign/pull/4710 |
| Upstream PRs closed (unmerged) | 0 | <=1 | +2 | https://github.com/open-policy-agent/opa/pull/8343, https://github.com/open-policy-agent/opa/pull/8346, https://github.com/ossf/scorecard/pull/4942, https://github.com/sigstore/cosign/pull/4710 |
| Third-party mentions | 0 | >=1 | +5 | https://dev.to/ogulcanaydogan/i-spent-3-months-solving-a-security-gap-nobody-talks-about-llm-artifact-integrity-6co |
| Anonymous pilot case studies | 0 | >=1 | +1 | `docs/public-footprint/case-study-anonymous-pilot-2026-02.md` |
| GitHub stars | 0 | >=25 | 0 | `.llmsa/public-footprint/20260223T201649Z/snapshot.json` |
| GitHub forks | 0 | >=5 | 0 | `.llmsa/public-footprint/20260223T201649Z/snapshot.json` |
| GitHub watchers | 0 | >=5 | 0 | `.llmsa/public-footprint/20260223T201649Z/snapshot.json` |
| Release downloads (cumulative) | 184 | >=400 | 140 | `.llmsa/public-footprint/20260223T201649Z/snapshot.json` |
| CI pass rate (last 30 days) | 93.3% | >=95% | 95.4% | `.llmsa/public-footprint/20260223T201700Z/ci-health.json` |
| CI pass rate (post-hardening window) | n/a | >=95% | 98.77% | `.llmsa/public-footprint/20260223T201700Z/ci-health.json` |

## Current Snapshot (2026-02-23 UTC)

- Snapshot artifact (local): `.llmsa/public-footprint/20260223T201649Z/snapshot.json`
- CI health artifact (local): `.llmsa/public-footprint/20260223T201700Z/ci-health.json`
- CI pass rate source window: 83/87 successful runs (`95.4%`).
- CI pass rate post-hardening (`2026-02-19T16:08:22Z`): 80/81 successful runs (`98.77%`, meets target=`true`).
- External write-up URL: https://dev.to/ogulcanaydogan/i-spent-3-months-solving-a-security-gap-nobody-talks-about-llm-artifact-integrity-6co
- Upstream PR review stage:
  - `in-review`: 1
  - `closed`: 2
  - `merged`: 1
- Dashboard generated at: `2026-02-23T20:17:07Z`
