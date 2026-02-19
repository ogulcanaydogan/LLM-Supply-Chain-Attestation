# Measurement Dashboard (Day 0 -> Day 30)

This dashboard tracks public-footprint metrics only.  
Day-0 values are captured on **2026-02-18 UTC**.

## Scoreboard

| Metric | Day 0 | Day 30 Target | Current Delta | Source |
|---|---:|---:|---:|---|
| Upstream PRs opened | 0 | >=2 | +3 | https://github.com/sigstore/cosign/pull/4710, https://github.com/open-policy-agent/opa/pull/8343, https://github.com/ossf/scorecard/pull/4942 |
| Upstream PRs merged | 0 | >=1 | +0 | https://github.com/sigstore/cosign/pull/4710, https://github.com/open-policy-agent/opa/pull/8343, https://github.com/ossf/scorecard/pull/4942 |
| Upstream PRs in review | 0 | <=2 | +3 | `docs/public-footprint/external-contribution-log.md` |
| Third-party mentions | 0 | >=1 | +1 | https://gist.github.com/ogulcanaydogan/7cffe48a760a77cb42cb1f87644909bb |
| Anonymous pilot case studies | 0 | >=1 | +1 | `docs/public-footprint/case-study-anonymous-pilot-2026-02.md` |
| GitHub stars | 0 | >=25 | +0 | `.llmsa/public-footprint/20260218T205404Z/snapshot.json` |
| GitHub forks | 0 | >=5 | +0 | `.llmsa/public-footprint/20260218T205404Z/snapshot.json` |
| GitHub watchers | 0 | >=5 | +0 | `.llmsa/public-footprint/20260218T205404Z/snapshot.json` |
| Release downloads (cumulative) | 184 | >=400 | +0 | `.llmsa/public-footprint/20260218T205404Z/snapshot.json` |
| CI pass rate (last 30 days) | 93.3% | >=95% | 89.47% | https://github.com/ogulcanaydogan/LLM-Supply-Chain-Attestation/actions/runs/22172837380 |

## Reliability Focus (Current)

- Primary fail cluster is `ci-attest-verify` in OCI verification path.
- Root-cause classes in current window:
  - deterministic config/verification failures,
  - credential/token failure previously seen in `release-verify` asset download.
- Mitigations now implemented in workflows:
  - preflight checks for required files/tokens,
  - retries on publish/download/OCI verification steps,
  - explicit job timeouts and concurrency control.

## Automation Commands

```bash
./scripts/public-footprint-snapshot.sh
./scripts/ci-health-snapshot.sh
./scripts/generate-evidence-pack.sh
./scripts/upstream-pr-followup.sh
```

## Current Snapshot (2026-02-19 UTC)

- Snapshot artifact (local): `.llmsa/public-footprint/20260218T205404Z/snapshot.md`
- CI pass rate source window: 17/19 successful runs (`89.47%`).
- External mention URL (current primary): https://gist.github.com/ogulcanaydogan/7cffe48a760a77cb42cb1f87644909bb
- Upstream PR review stage:
  - `in-review`: 3
  - `merged`: 0
