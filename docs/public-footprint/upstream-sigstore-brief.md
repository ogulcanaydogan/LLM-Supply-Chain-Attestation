# Upstream Contribution Brief: Sigstore

## Target Repository

- https://github.com/sigstore/cosign
- https://github.com/sigstore/sigstore

## Problem Statement

`llmsa` uses keyless signing and DSSE bundles, but many engineering teams still struggle to map CI identity controls to practical verification commands.  
A concise, reproducible example for issuer/identity-constrained verification would reduce misuse.

## Proposed Contribution

Provide a docs/example contribution focused on:

1. DSSE/attestation verification flow with strict issuer + identity checks.
2. Before/after examples showing permissive verification vs constrained verification.
3. Failure cases:
   - wrong OIDC issuer,
   - wrong workflow identity,
   - stale or mismatched bundle subject.

## Acceptance Criteria

1. Example is executable from a clean checkout.
2. Includes expected pass and fail outputs.
3. Maintainers can validate behavior without project-specific dependencies.
4. Security trade-offs are explicit (what is verified and what is not).

## Reproducible Steps (Draft)

1. Generate a signed DSSE test artifact in CI context.
2. Verify with constrained identity (`issuer`, `subject/identity regex`).
3. Replay with intentionally mismatched identity to prove failure mode.
4. Document minimal command set and interpretation.

## Non-Goals

- No claim that this solves runtime compromise.
- No claim of full transparency-log-based revocation lifecycle.

## Evidence to Capture

- PR URL.
- Maintainer review comments.
- Final merged doc/example URL.
- Date and short outcome summary.
