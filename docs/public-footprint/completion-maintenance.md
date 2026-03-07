# Completion Maintenance Runbook

This runbook defines the operational checks required to keep the roadmap in strict completion state.

## Baseline (2026-03-07 UTC)

- Roadmap completion baseline: `.llmsa/public-footprint/20260307T125957Z/roadmap-completion.json`
- Consistency baseline: `.llmsa/public-footprint/20260307T130043Z/consistency-check.json`
- Expected machine state:
  - `verdict.strict_complete = true`
  - `verdict.practical_complete = true`
  - `blockers = []`
  - `consistent = true`

## Daily Guard

Run:

```bash
./scripts/completion-daily-check.sh
```

What it checks:

1. Latest run health for `ci-attest-verify`, `release`, `release-verify`, `public-footprint-weekly`.
2. Latest roadmap completion still satisfies strict gate (`FAIL_ON_INCOMPLETE=true`).
3. Docs and machine artifacts remain consistent (`CONSISTENCY_SCOPE=core`, `FAIL_ON_INCONSISTENT=true`).

## Weekly Refresh

Run:

```bash
./scripts/completion-weekly-refresh.sh
```

What it refreshes:

1. `scripts/public-footprint-snapshot.sh`
2. `scripts/ci-health-snapshot.sh`
3. `scripts/generate-evidence-pack.sh`
4. `scripts/roadmap-completion-check.sh`
5. `scripts/check-footprint-consistency.sh`

## Acceptance Threshold

After daily or weekly checks, the newest artifacts under `.llmsa/public-footprint/<timestamp>/` must satisfy:

- `roadmap-completion.json`:
  - `verdict.strict_complete == true`
  - `blockers` is empty
- `consistency-check.json`:
  - `consistent == true`

## CI Automation

- Daily automation workflow: `.github/workflows/completion-daily-health.yml`
- Weekly refresh workflow: `.github/workflows/public-footprint-weekly.yml`
