# llmsa Full Completion Roadmap (as of 2026-02-23)

## Current Status

The project is operational and released (`v1.0.1`) with attestation, verification, policy, CI workflows, and footprint automation in place.

Public-footprint gates currently stand at:

- Upstream merged PRs: met (`>=1`)
- Canonical third-party mention: met
- Rolling 30-day CI pass rate: not yet met (`94.12%`, target `>=95%`)
- Post-hardening CI pass rate: met (`98.57%`)

## Infrastructure Usage Report (V100/A100/Spark)

Current repository state shows no usage of:

- NVIDIA V100
- NVIDIA A100
- Apache Spark / Databricks Spark runtime

The project process is currently CI/CD and CLI driven (Go + GitHub Actions), not GPU-cluster or Spark-pipeline driven.

## Completion Definition

Project completion is treated as done when all below are true:

1. Rolling 30-day CI pass rate reaches `>=95%` in automated snapshot output.
2. Public-footprint evidence files point to current snapshots and remain internally consistent.
3. Open upstream validation lane has either:
   - one additional merged PR, or
   - documented maintainer closure with reduced-scope follow-up attempts.
4. Release and verification evidence remain green and reproducible from scripts.

## Execution Plan

### Phase 1: Strict CI Closure (Highest Priority)

Goal: move rolling CI from `94.12%` to `>=95%`.

Process:

1. Run non-overlapping `ci-attest-verify` manual dispatches (`workflow_dispatch`) while monitoring each run to completion.
2. Stop immediately on first non-success run; classify failure and fix before any further dispatch.
3. After each successful batch, refresh:
   - `scripts/public-footprint-snapshot.sh`
   - `scripts/ci-health-snapshot.sh`
   - `scripts/ci-passrate-forecast.sh`
   - `scripts/generate-evidence-pack.sh`
   - `scripts/roadmap-completion-check.sh`
4. Continue until latest roadmap completion output reports rolling `>=95%`.

Exit criteria:

- `strict_complete=true` in latest roadmap completion JSON.

### Phase 2: External Validation Depth

Goal: strengthen merged external signal beyond minimum bar.

Process:

1. Continue 48-hour maintainer follow-up cadence on remaining open upstream PRs.
2. If no response after 5 business days, submit reduced-scope fallback PR with explicit link-back.
3. Keep contribution log and evidence pack synchronized after every state change.

Exit criteria:

- At least one additional external merge OR documented closure path with reproducible follow-up record.

### Phase 3: Evidence Freeze

Goal: publish a stable, auditable closure set.

Process:

1. Freeze latest artifact set under `.llmsa/public-footprint/<timestamp>/`.
2. Ensure these docs reference the same current sources:
   - `docs/public-footprint/measurement-dashboard.md`
   - `docs/public-footprint/evidence-pack-2026-02-18.md`
   - `docs/public-footprint/roadmap-status-2026-02.md`
   - `docs/public-footprint/day-30-outcomes-2026-02.md`
3. Verify all metric rows contain URL/path pointers.

Exit criteria:

- No stale snapshot paths and no conflicting values across footprint docs.

### Phase 4: Release Closure (Optional Patch if Needed)

Goal: keep release lane trustworthy after closure.

Process:

1. If fixes were required, cut `v1.0.2` as hardening-only patch.
2. Ensure release artifacts include binaries, signatures, SBOM, and verify links.
3. Re-run footprint generation and pin release links in evidence pack.

Exit criteria:

- Latest release evidence is complete and reproducible.

## Operating Cadence

Daily:

1. Check workflow health (`ci-attest-verify`, `release`, `release-verify`, `public-footprint-weekly`).
2. Trigger controlled green CI runs only when current tracked workflows are healthy.
3. Refresh footprint outputs if numerator/denominator changes.

Weekly:

1. Run full footprint regeneration pipeline.
2. Validate source traceability in evidence pack.
3. Update roadmap status notes.

## Risk Controls

1. Do not game methodology: keep workflow set and CI window definition unchanged.
2. Do not backfill or rewrite historical failures.
3. Treat network/API transient issues as operational blockers; retry transparently and log attempts.
4. Keep claims strict: every metric must map to a script-generated artifact.
