# External Contribution Log

Track only contributions to repositories outside this project.

## Entries

| Date (UTC) | Target Project | PR URL | Type | Status | Last Follow-up (UTC) | Technical Summary |
|---|---|---|---|---|---|---|
| 2026-02-18 | Sigstore / cosign | https://github.com/sigstore/cosign/pull/4710 | Docs contribution | merged | 2026-02-19 (merged: https://github.com/sigstore/cosign/pull/4710) | Added keyless `verify-blob` identity/issuer-constrained verification example; applied maintainer-requested placement/heading updates in commit `14ed255c`; PR merged on 2026-02-19. |
| 2026-02-18 | Open Policy Agent / OPA | https://github.com/open-policy-agent/opa/pull/8343 | Docs contribution | closed | 2026-02-19 (https://github.com/open-policy-agent/opa/pull/8343#issuecomment-3927950855) | Added CI/CD guardrail pattern using stable input contract + deny-reasons policy output and `--fail-defined` behavior; DCO status passing. PR was closed on 2026-02-19 without merge. |
| 2026-02-19 | Open Policy Agent / OPA | https://github.com/open-policy-agent/opa/pull/8346 | Docs contribution (fallback PR) | closed | 2026-02-19 (https://github.com/open-policy-agent/opa/pull/8346#issuecomment-3928358057) | Reduced-scope replacement PR opened after closure of #8343; maintainer closed PR on 2026-02-19 with guidance to keep this pattern in project-specific docs. |
| 2026-02-18 | OpenSSF / Scorecard | https://github.com/ossf/scorecard/pull/4942 | Docs contribution | in-review | 2026-02-23 (https://github.com/ossf/scorecard/pull/4942#issuecomment-3946999971) | Added beginner-focused release evidence progression connecting SBOM and Signed-Releases checks; docs scope kept narrow after follow-up; DCO status passing. |

## Cadence Policy

- Follow-up cadence: every `12h` while a PR remains open (auto-post best-effort when due).
- No maintainer response after `5 days`: open a reduced-scope fallback PR and cross-link both PRs.
- Automation workflow: `.github/workflows/upstream-pr-followup.yml`.
- Current execution note: OPA maintainer requested this content remain project-specific; no further OPA docs PRs planned for this topic.
- Operational note (2026-02-23 UTC): automated follow-up posting succeeded for `ossf/scorecard#4942` at `2026-02-23T20:00:08Z`; current state remains `in-review` and awaiting maintainer decision.

## Follow-up Attempt Ledger (`ossf/scorecard#4942`)

| Attempted At (UTC) | Channel | Outcome | Notes |
|---|---|---|---|
| 2026-02-23T19:59:41Z | `gh pr comment` (manual retries) | failed | Blocked by transient `api.github.com` connectivity failures during retry loop. |
| 2026-02-23T20:00:08Z | `scripts/upstream-pr-followup.sh` (`POST_FOLLOWUPS=true`) | posted | Follow-up comment posted: https://github.com/ossf/scorecard/pull/4942#issuecomment-3946999971 |

## Status Definitions

- `opened`: PR submitted.
- `in-review`: maintainer feedback received, revision pending/applied, awaiting maintainer decision.
- `merged`: PR accepted and merged.
- `closed`: PR closed without merge.
