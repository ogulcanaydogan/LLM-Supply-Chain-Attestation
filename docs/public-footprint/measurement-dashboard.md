# Measurement Dashboard (Day 0 -> Day 30)

This dashboard tracks public-footprint metrics only.  
Day-0 values are captured on **2026-02-18 UTC**.

## Scoreboard

| Metric | Day 0 | Day 30 Target | Current Delta |
|---|---:|---:|---:|
| Upstream PRs opened | 0 | >=2 | +3 |
| Upstream PRs merged | 0 | >=1 | 0 |
| Third-party mentions | 0 | >=1 | +1 |
| Anonymous pilot case studies | 0 | >=1 | +1 |
| GitHub stars | 0 | >=25 | 0 |
| GitHub forks | 0 | >=5 | 0 |
| GitHub watchers | 0 | >=5 | 0 |
| Release downloads (cumulative) | 184 | >=400 | 0 |
| CI pass rate (last 30 days) | 93.3% | >=95% | -3.83 |

## Notes

- Targets are directional and can be revised based on real conversion rates.
- Upstream PR metrics count only PRs to repositories outside this project.
- Mention metrics require linkable, third-party-hosted content.

## Update Commands

```bash
./scripts/public-footprint-snapshot.sh
```

Then copy metrics from the latest output into this table.

## Current Snapshot (2026-02-18 UTC)

- Snapshot artifact (local): `.llmsa/public-footprint/20260218T205404Z/snapshot.md`
- Workflow run (public artifact): https://github.com/ogulcanaydogan/LLM-Supply-Chain-Attestation/actions/runs/22157220499
- CI pass rate source window: 17/19 successful runs (`89.47%`).
- External write-up URL: https://gist.github.com/ogulcanaydogan/7cffe48a760a77cb42cb1f87644909bb
