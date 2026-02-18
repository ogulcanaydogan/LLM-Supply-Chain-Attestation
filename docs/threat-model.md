# Threat Model (MVP)

## Threats covered
- Artifact tampering after signing.
- Signature spoofing with untrusted key material.
- Replay of stale attestations.
- Sensitive payload leakage in statements.

## MVP mitigations
- Subject digest recomputation in verify.
- DSSE envelope signature verification.
- OIDC issuer/identity checks for sigstore provider metadata.
- Policy block for plaintext payload mode unless allowlisted.

## Known limits
- No production Rekor proof validation in MVP.
- No OCI verifier implementation in MVP.
- Rego policy engine not enabled until v0.2.
