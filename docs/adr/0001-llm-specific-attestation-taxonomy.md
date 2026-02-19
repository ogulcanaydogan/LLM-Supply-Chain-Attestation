# ADR-0001: LLM-Specific Attestation Taxonomy

## Status

Accepted

## Date

2025-05-15

## Context

Existing software supply-chain frameworks (SLSA, in-toto, SBOM) secure traditional build artifacts — container images, binaries, packages — but provide no coverage for AI-specific components that define LLM system behaviour. We needed to decide whether to extend an existing framework or define a new taxonomy purpose-built for LLM artifacts.

The key question was: what categories of LLM artifacts require independent attestation statements?

## Decision

We defined five attestation types that cover the complete LLM artifact lifecycle:

1. **prompt_attestation** — System prompts, templates, tool schemas, safety policies
2. **corpus_attestation** — Training data, RAG documents, embeddings, vector indices
3. **eval_attestation** — Test suites, benchmark results, scoring configs, baselines
4. **route_attestation** — Routing tables, fallback graphs, canary configs, budget policies
5. **slo_attestation** — Latency targets, cost budgets, accuracy thresholds, query profiles

Each type has a dedicated JSON Schema (`schemas/v1/<type>.schema.json`), a predicate type definition (`pkg/types/predicate_<type>.go`), and a collector implementation (`internal/attest/collectors_<type>.go`).

## Rationale

- **Five types, not fewer**: Collapsing prompt and corpus into a single type would lose the ability to enforce that evaluations reference both independently. The provenance chain requires typed nodes.
- **Five types, not more**: We considered separate types for safety policies, tool schemas, and embedding configurations. These are better modeled as subjects within existing types than as standalone attestation categories.
- **Predicate URI namespacing**: Each type uses `https://llmsa.dev/attestation/<type>/v1` as its predicate URI, following in-toto's extensible predicate model while establishing a distinct namespace.
- **Config-driven collectors**: Each attestation type is created from a YAML config file that declares subjects, metadata, and type-specific fields. This keeps the CLI generic while allowing type-specific validation via JSON Schema.

## Consequences

- Adding a sixth attestation type requires: predicate type, collector, JSON Schema, CLI registration, and provenance chain updates. The process is documented in CONTRIBUTING.md.
- The five-type taxonomy is encoded in the provenance chain DAG (eval depends on prompt + corpus, route depends on eval, slo depends on route). Changing the dependency structure requires updating `internal/verify/chain_verify.go`.
- External integrations must understand all five types to provide complete coverage. Partial adoption (e.g., only prompt + eval) is supported but produces warnings in verification.
