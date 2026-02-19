# ADR-0005: Provenance Chain as a Directed Acyclic Graph

## Status

Accepted

## Date

2025-05-15

## Context

Individual attestations prove that specific files are intact and signed. But LLM systems have sequential dependencies: evaluation results are only meaningful if they reference the exact prompt and corpus versions that were tested. We needed a mechanism to verify that downstream attestations reference upstream ones.

## Decision

We defined a provenance chain as a directed acyclic graph (DAG) with the following edges:

```
eval → prompt (eval must reference the attested prompt version)
eval → corpus (eval must reference the attested corpus version)
route → eval   (route must reference the attested eval results)
slo   → route  (SLO must reference the attested routing config)
```

Dependencies are expressed via `depends_on` annotations in each attestation statement's `annotations` map. The verification engine (`internal/verify/chain_verify.go`) validates:

1. **Referential integrity** — Every `depends_on` reference points to a statement ID that exists in the attestation set.
2. **Temporal ordering** — Downstream statements have `generated_at` timestamps equal to or later than their upstream dependencies.
3. **Type constraints** — Only valid dependency edges are permitted (e.g., route cannot depend directly on corpus).
4. **Graph acyclicity** — The dependency graph contains no cycles.

## Rationale

- **DAG, not linear chain**: A linear chain (prompt → corpus → eval → route → slo) would require artificial ordering between prompt and corpus, which are independent. The DAG allows eval to depend on both prompt and corpus in parallel.
- **Annotation-based references**: Using `depends_on` in the annotations map rather than a dedicated field keeps the statement schema extensible and backward-compatible. Statements without `depends_on` are valid (standalone attestations).
- **Lenient single-bundle mode**: When only one attestation exists, the chain is trivially valid. This supports incremental adoption where teams start with a single attestation type.

## Consequences

- The full chain requires all 5 attestation types to be present. Missing types produce warnings but not failures in lenient mode.
- Statement IDs must be stable and unique within an attestation set. The current implementation uses `<type>-<subject-digest-prefix>`.
- Adding a new attestation type requires defining its position in the DAG. A type with no upstream dependencies is a root node; a type with no downstream dependents is a leaf.
