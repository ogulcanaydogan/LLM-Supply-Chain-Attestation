# ADR-0003: Dual Policy Engine (YAML Gates + OPA Rego)

## Status

Accepted

## Date

2025-06-01

## Context

The verification pipeline needs a policy enforcement layer that decides whether a set of attestation results constitutes a passing or failing gate. We needed to support both simple declarative policies for common use cases and expressive programmatic policies for advanced requirements.

The candidates were:

1. **YAML-only** — Simple, no runtime dependency, but limited expressiveness
2. **OPA Rego-only** — Powerful, industry-standard, but steep learning curve for simple cases
3. **Both** — YAML for simple gates, Rego for advanced policies, with a unified interface

## Decision

We implemented a dual policy engine with a common `Evaluator` interface:

- **YAML gate engine** (`internal/policy/yaml/`) — Declarative rules that specify required attestation types, minimum bundle counts, allowed signature providers, and path-based triggers. Evaluated by iterating over rules and checking conditions.
- **Rego policy engine** (`internal/policy/rego/`) — OPA-based evaluation that receives structured input (attestation results, bundle metadata, verification outcomes) and returns typed violation objects. Supports arbitrary policy logic.

Both engines implement the same interface and return `[]Violation` results. The CLI `gate` command accepts either a YAML policy file or a Rego policy directory.

## Rationale

- **YAML handles 80% of use cases**: Most teams need simple rules like "all 5 attestation types must be present and signed." YAML gates are readable, version-controllable, and require no policy language knowledge.
- **Rego handles the remaining 20%**: Advanced policies like "eval attestations must reference corpus versions from the last 30 days" or "route changes require two independent signatures" need programmatic logic. OPA Rego is the industry standard for infrastructure policy.
- **Unified violation model**: Both engines produce the same `Violation` struct with severity, message, rule reference, and affected bundle. The gate command doesn't need to know which engine produced a violation.

## Consequences

- Two policy formats to document and maintain. The policy guide (`docs/policy-guide.md`) covers both.
- OPA adds a dependency (`github.com/open-policy-agent/opa`). The binary size increases by approximately 8MB.
- Users must choose between YAML and Rego per invocation. Combining both in a single gate evaluation is not currently supported.
