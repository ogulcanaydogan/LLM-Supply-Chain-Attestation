# External Contribution Log

Track only contributions to repositories outside this project.

## Entries

| Date (UTC) | Target Project | PR URL | Type | Status | Last Follow-up (UTC) | Technical Summary |
|---|---|---|---|---|---|---|
| 2026-02-18 | Sigstore / cosign | https://github.com/sigstore/cosign/pull/4710 | Docs contribution | merged | 2026-02-19 (merged: https://github.com/sigstore/cosign/pull/4710) | Added keyless `verify-blob` identity/issuer-constrained verification example; applied maintainer-requested placement/heading updates in commit `14ed255c`; PR merged on 2026-02-19. |
| 2026-02-18 | Open Policy Agent / OPA | https://github.com/open-policy-agent/opa/pull/8343 | Docs contribution | closed | 2026-02-19 (https://github.com/open-policy-agent/opa/pull/8343#issuecomment-3927950855) | Added CI/CD guardrail pattern using stable input contract + deny-reasons policy output and `--fail-defined` behavior; DCO status passing. PR was closed on 2026-02-19 without merge. |
| 2026-02-19 | Open Policy Agent / OPA | https://github.com/open-policy-agent/opa/pull/8346 | Docs contribution (fallback PR) | closed | 2026-02-19 (https://github.com/open-policy-agent/opa/pull/8346#issuecomment-3928358057) | Reduced-scope replacement PR opened after closure of #8343; maintainer closed PR on 2026-02-19 with guidance to keep this pattern in project-specific docs. |
| 2026-02-18 | OpenSSF / Scorecard | https://github.com/ossf/scorecard/pull/4942 | Docs contribution | closed | 2026-03-06 (https://github.com/ossf/scorecard/pull/4942#issuecomment-4011028727) | Original wider-scope PR. Closed by author in favor of reduced-scope fallback `#4960` to keep a single active review path for maintainers. |
| 2026-03-06 | OpenSSF / Scorecard | https://github.com/ossf/scorecard/pull/4960 | Docs contribution (reduced-scope fallback PR) | closed | 2026-03-06 (https://github.com/ossf/scorecard/pull/4960#issuecomment-4012254642) | Reduced the change to a 2-line docs-only insertion in `docs/beginner-checks.md` and cross-linked `#4942`; closed as superseded after `cosign#4740` merged. |
| 2026-03-06 | SLSA / slsa-github-generator | https://github.com/slsa-framework/slsa-github-generator/pull/4468 | Docs contribution (parallel backup PR) | closed | 2026-03-06 (https://github.com/slsa-framework/slsa-github-generator/pull/4468#issuecomment-4012254970) | Opened narrow docs-only PR fixing a broken README anchor (`#verification-of-provenance` -> `#verify-provenance`) with DCO sign-off; closed as superseded after `cosign#4740` merged. |
| 2026-03-06 | Sigstore / cosign | https://github.com/sigstore/cosign/pull/4740 | Docs contribution (primary merge lane) | merged | 2026-03-06 (merged: https://github.com/sigstore/cosign/pull/4740) | One-line docs-only fix for stale URL scheme in `CHANGELOG.md` (`[sigstore](sigstore.dev)` -> `[sigstore](https://sigstore.dev)`); merged on 2026-03-06 and used as second upstream conversion proof. |

## Cadence Policy

- Follow-up cadence: every `12h` while a PR remains open (auto-post best-effort when due).
- Fallback rule (historical): no maintainer response after `5 days` triggers reduced-scope fallback PR + cross-link.
- Automation workflow: `.github/workflows/upstream-pr-followup.yml`.
- Current execution note: OPA maintainer requested this content remain project-specific; no further OPA docs PRs planned for this topic.
- Operational note (2026-03-06 UTC): primary merge lane `sigstore/cosign#4740` is merged; secondary/backup lanes (`#4960`, `#4468`) are closed as superseded.
- Reviewer-request limitation: direct reviewer request API to `ossf/scorecard` from external contributor context returned `404` (`requested_reviewers` endpoint), so cadence remains comment-based plus fallback-date policy.

## Follow-up Attempt Ledger (`ossf/scorecard#4942`)

| Attempted At (UTC) | Channel | Outcome | Notes |
|---|---|---|---|
| 2026-02-23T19:59:41Z | `gh pr comment` (manual retries) | failed | Blocked by transient `api.github.com` connectivity failures during retry loop. |
| 2026-02-23T20:00:08Z | `scripts/upstream-pr-followup.sh` (`POST_FOLLOWUPS=true`) | posted | Follow-up comment posted: https://github.com/ossf/scorecard/pull/4942#issuecomment-3946999971 |
| 2026-02-23T20:14:29Z | `scripts/upstream-pr-followup.sh` (`POST_FOLLOWUPS=true`) | skipped | Automation ran on cadence and skipped duplicate posting because latest comment author was current actor. |
| 2026-02-23T21:32:42Z | `scripts/upstream-pr-followup.sh` (`POST_FOLLOWUPS=true`) | skipped | Automation cycle recorded `followup_due=false`; latest PR comment author remained current actor, so no duplicate nudge was posted. |
| 2026-02-23T21:33:00Z | `gh pr edit --add-reviewer ossf/scorecard-doc-maintainers` | failed | GitHub API returned `404` on `requested_reviewers` endpoint for external contributor context; no reviewer assignment was made. |
| 2026-03-06T09:19:59Z | `gh pr edit` + `gh pr update-branch` + maintainer comment | posted | Title changed to include `:book:` indicator, branch synced to latest `main`, and explicit workflow-approval/final-review request posted (https://github.com/ossf/scorecard/pull/4942#issuecomment-4010565820). |
| 2026-03-06T09:23:33Z | signed empty commit push | posted | Added `chore: retrigger PR checks after title+sync update` to retrigger PR checks under current policy gates. |
| 2026-03-06T10:08:22Z | fallback PR creation + cross-link comment | posted | Opened reduced-scope fallback PR `#4960` and cross-linked from `#4942` (https://github.com/ossf/scorecard/pull/4942#issuecomment-4010735248). |
| 2026-03-06T10:44:00Z | maintainer follow-up on `#4960` + superseded note on `#4942` | posted | Requested workflow approval/final review on fallback PR (https://github.com/ossf/scorecard/pull/4960#issuecomment-4010886725) and marked `#4942` as supersedable by `#4960` pending maintainer preference (https://github.com/ossf/scorecard/pull/4942#issuecomment-4010887592). |
| 2026-03-06T10:54:24Z | `gh pr close` on `#4942` | posted | Closed `#4942` in favor of `#4960` to keep one active scorecard review path (https://github.com/ossf/scorecard/pull/4942#issuecomment-4011028727). |
| 2026-03-06T15:00:53Z | superseded close on `#4960` | posted | Closed fallback scorecard lane as superseded by merged `cosign#4740` (https://github.com/ossf/scorecard/pull/4960#issuecomment-4012254642). |
| 2026-03-06T15:00:56Z | superseded close on `#4468` | posted | Closed backup slsa lane as superseded by merged `cosign#4740` (https://github.com/slsa-framework/slsa-github-generator/pull/4468#issuecomment-4012254970). |

## Status Definitions

- `opened`: PR submitted.
- `in-review`: maintainer feedback received, revision pending/applied, awaiting maintainer decision.
- `merged`: PR accepted and merged.
- `closed`: PR closed without merge.
