# External Contribution Log

Track only contributions to repositories outside this project.

## Entries

| Date (UTC) | Target Project | PR URL | Type | Status | Last Follow-up (UTC) | Technical Summary |
|---|---|---|---|---|---|---|
| 2026-02-18 | Sigstore / cosign | https://github.com/sigstore/cosign/pull/4710 | Docs contribution | in-review | 2026-02-19 (https://github.com/sigstore/cosign/pull/4710#issuecomment-3930021517) | Added keyless `verify-blob` identity/issuer-constrained verification example; applied maintainer-requested placement/heading updates in commit `14ed255c`; DCO status passing. |
| 2026-02-18 | Open Policy Agent / OPA | https://github.com/open-policy-agent/opa/pull/8343 | Docs contribution | closed | 2026-02-19 (https://github.com/open-policy-agent/opa/pull/8343#issuecomment-3927950855) | Added CI/CD guardrail pattern using stable input contract + deny-reasons policy output and `--fail-defined` behavior; DCO status passing. PR was closed on 2026-02-19 without merge. |
| 2026-02-19 | Open Policy Agent / OPA | https://github.com/open-policy-agent/opa/pull/8346 | Docs contribution (fallback PR) | closed | 2026-02-19 (https://github.com/open-policy-agent/opa/pull/8346#issuecomment-3928358057) | Reduced-scope replacement PR opened after closure of #8343; maintainer closed PR on 2026-02-19 with guidance to keep this pattern in project-specific docs. |
| 2026-02-18 | OpenSSF / Scorecard | https://github.com/ossf/scorecard/pull/4942 | Docs contribution | in-review | 2026-02-19 (https://github.com/ossf/scorecard/pull/4942#issuecomment-3930021503) | Added beginner-focused release evidence progression connecting SBOM and Signed-Releases checks; docs scope kept narrow after follow-up; DCO status passing. |

## Cadence Policy

- Follow-up cadence: every `48h` while a PR remains open.
- No maintainer response after `5 days`: open a reduced-scope fallback PR and cross-link both PRs.
- Automation workflow: `.github/workflows/upstream-pr-followup.yml`.
- Current execution note: OPA maintainer requested this content remain project-specific; no further OPA docs PRs planned for this topic.

## Status Definitions

- `opened`: PR submitted.
- `in-review`: maintainer feedback received, revision pending/applied, awaiting maintainer decision.
- `merged`: PR accepted and merged.
- `closed`: PR closed without merge.
