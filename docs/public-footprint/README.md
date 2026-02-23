# Public Footprint Playbook (30 Days)

This folder tracks public evidence for `/Users/ogulcanaydogan/Desktop/Projects/YaPAY/ LLM Supply-Chain Attestation` over a 30-day execution window.

The goal is to convert strong internal engineering quality into externally verifiable signals:

- upstream technical contributions,
- third-party technical mentions,
- reproducible anonymous adoption evidence.

## Scope

- Project: `llmsa` only.
- Language: English only.
- Audience: security engineering, platform engineering, and SRE leaders.
- Constraint: keep repository language strictly technical and product-focused.

## Contents

- `evidence-baseline.md`: Day-0 inventory of current public evidence and gaps.
- `measurement-dashboard.md`: Day-0 to Day-30 metric scoreboard.
- `upstream-sigstore-brief.md`: one-page contribution proposal for Sigstore.
- `upstream-opa-brief.md`: one-page contribution proposal for OPA/Rego ecosystem.
- `upstream-openssf-brief.md`: one-page contribution proposal for OpenSSF-aligned docs/checklists.
- `third-party-mention-plan.md`: external write-up and community distribution plan.
- `third-party-mention-draft-2026-02.md`: publish-ready technical write-up draft.
- `case-study-template.md`: anonymous pilot case study template with reproducibility sections.
- `case-study-anonymous-pilot-2026-02.md`: first anonymized pilot case study using benchmark and tamper artifacts.
- `evidence-pack-template.md`: copy-ready evidence format for submission packets.
- `evidence-pack-2026-02-18.md`: filled evidence pack with live URLs and metric snapshot.
- `what-we-do-not-claim.md`: explicit non-claims and known limitations language.
- `day0-metrics-2026-02-18.md`: initial baseline snapshot.
- `roadmap-status-2026-02.md`: current roadmap-completion status and endgame gates.
- `external-contribution-log.md`: tracker for opened/merged external PRs.
- `v1.0.1-hardening-closure.md`: hardening completion checklist and linked evidence.
- `../project-completion-roadmap-2026-02.md`: full project completion roadmap with current closure process and operating cadence.
- `third-party-mention-canonical-url.txt`: one-line canonical mention URL (non-GitHub) used as machine source-of-truth.
- `third-party-mention-publications.tsv`: optional additional publication links rendered in evidence summary.
- `scripts/check-footprint-consistency.sh`: guard script that validates machine verdict vs narrative docs and snapshot/ci-health artifact consistency.
- `scripts/scorecard-fallback-readiness.sh`: readiness snapshot for `ossf/scorecard#4942` fallback path (`not-before` date + recommended next action).

## Operating Rhythm

1. Update metrics weekly via `scripts/public-footprint-snapshot.sh`.
2. Update `external-contribution-log.md` on every external PR event.
3. Keep all public claims linked to public URLs (release, workflow, PR, post, talk).
4. Keep non-claims updated when scope changes.
5. Publish canonical third-party mention via `.github/workflows/publish-third-party-mention.yml` once `DEVTO_API_KEY` secret is configured.
6. Run `scripts/roadmap-completion-check.sh` to determine strict/practical roadmap completion and current blockers.
7. Run `scripts/ci-passrate-forecast.sh` to estimate success-only runs needed for rolling `>=95%`.
8. Run `scripts/check-footprint-consistency.sh` before freezing docs to block stale or contradictory footprint narratives.
9. Run `scripts/scorecard-fallback-readiness.sh` to confirm whether fallback conversion should be executed or cadence should continue.
