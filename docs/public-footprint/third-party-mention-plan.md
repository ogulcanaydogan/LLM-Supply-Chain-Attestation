# Third-Party Mention Plan

## Objective

Secure at least one external technical mention that links to `llmsa` evidence artifacts (release/workflow/docs), not just repository homepage.

## Candidate Channels

1. Security engineering newsletters.
2. Platform engineering community roundups.
3. Meetup recap posts or conference community notes.

## Publication Package (English)

1. Technical write-up (primary):
   - problem statement (LLM artifact integrity gap),
   - method (typed attestations + chain + policy + admission),
   - measured outputs (tamper detection, verify latency, gate coverage),
   - known limitations.
2. Community post (secondary) linking to primary write-up.

## Evidence Rules

1. Every metric must link to a source URL (workflow artifact, docs, release note, PR).
2. No superlative claims without external benchmark evidence.
3. Include a dedicated "What we do not claim" section.

## Acceptance Criteria

1. Mention is hosted outside this repository domain.
2. Mention includes at least one deep link to technical evidence.
3. Mention accurately reflects scope and limitations.
4. Mention date is captured in evidence pack.

## Current Status (2026-02-20) — PUBLISHED + NORMALIZED

- **LLMSA technical article published on Dev.to** (2026-02-19):
  - https://dev.to/ogulcanaydogan/i-spent-3-months-solving-a-security-gap-nobody-talks-about-llm-artifact-integrity-6co
  - Article ID: 3269328
  - Tags: security, golang, kubernetes, ai
  - Includes deep links to GitHub release, upstream PRs, and repo
  - Includes honest limitations section
- **AI Detection article also published on Dev.to** (2026-02-19):
  - https://dev.to/ogulcanaydogan/how-i-built-an-ai-content-detection-system-from-scratch-oe4
  - Article ID: 3269248
  - Tags: ai, python, machinelearning, opensource
- Original Gist mirror remains at:
  - https://gist.github.com/ogulcanaydogan/7cffe48a760a77cb42cb1f87644909bb
- Canonical mention source-of-truth is fixed to:
  - `docs/public-footprint/third-party-mention-canonical-url.txt`
- Additional publication links for Evidence Summary are tracked in:
  - `docs/public-footprint/third-party-mention-publications.tsv`
- Evidence pack uses URL-only source for machine-consumed mention metrics:
  - `docs/public-footprint/evidence-pack-2026-02-18.md`

## Minimal Mention Outline

1. Why generic supply-chain controls miss LLM-specific change risk.
2. How typed attestations and chain verification close that gap.
3. What measured outcomes look like in CI and admission paths.
4. Where the system still has blind spots.
