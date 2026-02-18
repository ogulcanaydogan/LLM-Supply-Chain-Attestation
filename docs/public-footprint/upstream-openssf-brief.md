# Upstream Contribution Brief: OpenSSF-Aligned Hardening

## Target Repositories / Programs

- https://github.com/ossf/scorecard
- OpenSSF-aligned project security checklists and hardening guidance repos

## Problem Statement

AI delivery pipelines add artifact classes (prompt, corpus, eval, route, SLO) that are not represented in many standard supply-chain checklists.  
A practical mapping from existing hardening controls to LLM-specific artifacts is missing in most public guidance.

## Proposed Contribution

Submit a documentation-oriented mapping that connects:

1. Existing software supply-chain controls (signing, provenance, policy gates).
2. LLM-specific evidence objects (typed attestations and provenance chains).
3. Deployment-time enforcement pattern (admission verification).

## Acceptance Criteria

1. Mapping remains control-focused, not product-promotional.
2. Every recommendation has a verifiable implementation example.
3. Scope boundaries are explicit (what remains out of scope).
4. Reviewers can validate examples independently.

## Reproducible Steps (Draft)

1. Publish a minimal reproducibility bundle:
   - signed attestation set,
   - gate policy,
   - verify/gate command outputs.
2. Attach mapping table: control -> expected evidence -> validation command.
3. Include failure injection case with observed denial path.

## Non-Goals

- No claim of standards-body endorsement.
- No claim that checklist conformance equals full security.

## Evidence to Capture

- PR/discussion URL.
- Reviewer feedback and revisions.
- Published mapping URL with date.
