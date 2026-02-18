# External Technical Write-up Draft (Ready to Publish)

## Title

From Scripts to Evidence: Enforcing LLM Artifact Integrity with OPA, Sigstore, and Admission Controls

## Short Abstract

Most AI delivery pipelines still rely on loosely coupled checks for prompts, evals, and runtime controls.  
This write-up presents a practical approach where typed attestations, policy decisions, and deployment admission are treated as one auditable control path.

## Draft Body

LLM systems introduce change surfaces that generic software supply-chain controls usually do not model directly: prompts, corpus snapshots, eval packs, model-routing policy, and SLO constraints.  
When these are validated independently, incident response and audits become difficult because evidence is fragmented.

Our approach in `llmsa` is intentionally operational:

1. Build typed attestations for prompt/corpus/eval/route/SLO artifacts.
2. Sign and verify statements using DSSE-compatible bundles.
3. Enforce policy gates in CI with explicit non-zero exits.
4. Validate deployment-time requirements through a webhook path.

Measured outputs from one pilot run:

- Tamper suite detection: 20/20 seeded attacks detected.
- Verify p95 (100 statements): 27 ms on the benchmark harness.
- Determinism stability: 1.0 in repeated-generation checks.

Key evidence URLs:

- Release: https://github.com/ogulcanaydogan/LLM-Supply-Chain-Attestation/releases/tag/v1.0.0
- CI + artifact snapshot run: https://github.com/ogulcanaydogan/LLM-Supply-Chain-Attestation/actions/runs/22157220499
- Upstream contributions:
  - https://github.com/sigstore/cosign/pull/4710
  - https://github.com/open-policy-agent/opa/pull/8343

What this does not claim:

- It is not a runtime prompt-injection defense by itself.
- It is not a universal performance guarantee.
- It is not a replacement for independent security review.

## Suggested Channels

1. CNCF/security newsletters.
2. OPA or Sigstore community recap posts.
3. Platform engineering meetup recap with links to evidence artifacts.

## Publish Checklist

1. Keep all numeric claims linked to source URLs.
2. Include non-claims paragraph.
3. Add publication URL to `docs/public-footprint/evidence-pack-2026-02-18.md`.
