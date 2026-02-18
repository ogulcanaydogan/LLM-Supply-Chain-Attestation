# Positioning Message (Single Technical Narrative)

`llmsa` exists to enforce integrity controls for LLM-specific change surfaces that generic software supply-chain tooling does not model directly.

The core message:

1. **Typed LLM attestation taxonomy**  
   Prompt, corpus, eval, route, and SLO artifacts are measured and signed as first-class objects.
2. **Provenance chain verification**  
   Downstream deployment decisions are cryptographically bound to upstream eval and artifact evidence.
3. **Policy enforcement in CI and admission**  
   Evidence is not only generated; it is enforced with deny semantics when controls are missing or invalid.
4. **Operationally practical defaults**  
   Local-first CLI, OCI distribution, and explicit privacy modes support real deployment constraints.

This message should be used consistently in README, release notes, external write-ups, and case studies.
