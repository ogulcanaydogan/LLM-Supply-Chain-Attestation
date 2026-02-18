# Upstream Contribution Brief: OPA / Rego

## Target Repository

- https://github.com/open-policy-agent/opa
- Adjacent policy example collections that accept Rego CI patterns

## Problem Statement

Security gates for AI artifact supply chains often start with path-based triggers, then need richer Rego policies.  
Teams need a practical migration path from simple gate rules to Rego without changing enforcement semantics.

## Proposed Contribution

Contribute a policy-input example and guardrail pattern showing:

1. Stable policy input schema for attestation metadata.
2. Equivalent baseline gate checks in Rego for common CI controls.
3. Privacy mode guardrail (`hash_only` default, controlled exceptions).

## Acceptance Criteria

1. Rego examples are runnable with `opa eval` or standard test harness.
2. Sample input and expected decisions are included.
3. Failure decisions are as explicit as allow decisions.
4. Docs include uncertainty and edge-case notes.

## Reproducible Steps (Draft)

1. Define sample JSON policy input with statement set metadata.
2. Evaluate allow/deny rules against:
   - complete attestation set,
   - missing required type,
   - disallowed plaintext mode.
3. Validate output messages are deterministic and actionable.

## Non-Goals

- No claim that Rego replaces cryptographic verification.
- No claim that one policy pack fits all regulated environments.

## Evidence to Capture

- PR URL and commit hash.
- Maintainer validation feedback.
- Final merged artifact URL and short diff summary.
