# Roadmap Status (2026-02)

This page tracks execution status against the Day-30 public-footprint roadmap.

## Snapshot

- Generated: 2026-02-19 UTC
- Scope: `llmsa` public footprint only
- Source of truth:
  - `docs/public-footprint/evidence-pack-2026-02-18.md`
  - `docs/public-footprint/measurement-dashboard.md`
  - `docs/public-footprint/external-contribution-log.md`

## Workstream Status

| Workstream | Target | Current | Status |
|---|---|---|---|
| WS1: Upstream conversion | `>=1` merged external PR | `0` merged, `2` open (`cosign#4710`, `scorecard#4942`) | in progress (external dependency) |
| WS2: CI pass-rate | rolling `>=95%` | rolling `87.5% (49/56)`; post-hardening `100% (16/16)` | partially complete |
| WS3: `v1.0.1` hardening evidence | release + verification artifacts complete | complete with release and verification runs linked | complete |
| WS4: Evidence automation | one-command refresh and source-traceable docs | complete (`public-footprint-snapshot`, `ci-health-snapshot`, `generate-evidence-pack`) | complete |
| WS5: Third-party mention | canonical non-GitHub publication URL | still gist mirror; Dev.to workflow blocked by missing `DEVTO_API_KEY` | blocked (credential dependency) |

## Completion Gates Remaining

1. Canonical publication:
   - Set repository secret `DEVTO_API_KEY`.
   - Run `.github/workflows/publish-third-party-mention.yml`.
   - Replace gist-primary mention with canonical URL in evidence docs.
2. External merge conversion:
   - Convert at least one open upstream PR to `merged`.
   - Update `external-contribution-log.md` and evidence pack with merge URL/date.
3. Rolling-window CI target:
   - Keep workflows green so historical failures age out of the 30-day denominator.
   - Recompute rolling pass-rate in each snapshot cycle.

## Practical Endgame

When the roadmap is considered complete:

1. `>=1` upstream PR is merged and reflected in docs.
2. Canonical third-party mention URL is published and linked from evidence pack.
3. Rolling CI pass-rate is `>=95%` or explicitly justified with post-hardening trend and an updated acceptance note.
