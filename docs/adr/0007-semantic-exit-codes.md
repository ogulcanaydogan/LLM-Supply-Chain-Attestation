# ADR-0007: Semantic Exit Codes for Verification Results

## Status

Accepted

## Date

2025-05-15

## Context

The `llmsa verify` and `llmsa gate` commands need to communicate verification results to CI/CD pipelines and shell scripts. A simple pass/fail (0/1) exit code loses information about what went wrong. We needed a structured exit code scheme that enables automated response logic.

## Decision

We defined six semantic exit codes:

| Code | Meaning | Typical Response |
|------|---------|-----------------|
| 0 | All checks passed | Proceed with deployment |
| 10 | Missing attestations | Block: required attestation types not found |
| 11 | Signature verification failed | Block: investigate key compromise or signing failure |
| 12 | Tamper detected (digest mismatch) | Block: artifact modified after signing |
| 13 | Policy violation | Block: attestation set doesn't meet policy requirements |
| 14 | Schema validation failed | Block: statement structure is invalid |

Exit codes escalate — if multiple failure types occur, the highest-severity code is returned. The priority order is: tamper (12) > signature (11) > schema (14) > policy (13) > missing (10).

## Rationale

- **Distinct codes for distinct failures**: A CI pipeline that receives exit code 12 (tamper) should trigger a security incident investigation, while exit code 10 (missing) might just need a pipeline reconfiguration. Collapsing these into a single non-zero code loses this distinction.
- **Range 10-14**: Exit codes 1-9 are reserved for generic errors (file not found, invalid arguments). Starting at 10 avoids conflicts with standard shell conventions.
- **Escalation semantics**: When a bundle fails both signature verification and schema validation, the signature failure is more severe. Returning the highest-severity code ensures automated systems respond to the worst problem.

## Consequences

- CI/CD pipelines can use `if [ $? -eq 12 ]; then` to detect specific failure modes and trigger appropriate responses.
- The 20-case tamper detection test suite (`scripts/tamper-tests.sh`) validates that each corruption type produces the expected exit code.
- Adding new exit codes requires updating the escalation logic in `internal/verify/engine.go` (`addFailure`) and documenting the new code in the CLI help text.
