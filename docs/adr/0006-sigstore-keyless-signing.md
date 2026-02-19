# ADR-0006: Sigstore Keyless Signing with OIDC Identity Binding

## Status

Accepted

## Date

2025-06-01

## Context

Attestation bundles must be cryptographically signed. Traditional approaches require managing long-lived private keys, which creates key distribution, rotation, and revocation challenges. We needed a signing strategy that works in CI/CD environments without persistent key management.

The candidates were:

1. **PEM key signing** — Simple Ed25519 keys stored as PEM files. Easy to implement, requires key management.
2. **Sigstore keyless signing** — Ephemeral keys with OIDC identity binding via Fulcio certificate authority. No persistent keys, but requires cosign binary and network access.
3. **KMS-backed signing** — Cloud provider key management (AWS KMS, GCP KMS). Strong security but vendor-locked.
4. **All of the above** — Provider interface supporting multiple backends.

## Decision

We implemented a `Signer` interface (`internal/sign/signer.go`) with three providers:

- **`PEMSigner`** — Ed25519 signing with local PEM key files. Used for local development, air-gapped environments, and testing. Key generation via `llmsa init --generate-key`.
- **`SigstoreSigner`** — Keyless signing via cosign with Fulcio OIDC certificates. In CI (GitHub Actions), uses the ambient OIDC token. Falls back to PEM signing when cosign is unavailable.
- **`KMSSigner`** — Placeholder for future cloud KMS integration. Currently returns "not implemented".

The `SigstoreSigner` captures OIDC claims (issuer and identity) in the bundle's signature metadata, enabling identity-based verification policies.

## Rationale

- **Keyless as the CI default**: In GitHub Actions, the workflow's OIDC identity token is available without any secret configuration. Sigstore keyless signing binds the signature to the workflow identity (`https://github.com/org/repo/.github/workflows/ci.yml@refs/heads/main`), providing non-repudiation without key management.
- **PEM fallback for offline use**: Not all environments have network access or cosign installed. PEM signing ensures the tool works in air-gapped, local, and testing contexts.
- **OIDC claims in bundle metadata**: Storing the issuer and identity in the signature block enables verification policies like "only accept signatures from the release workflow in the main repository."
- **Provider abstraction**: The `Signer` interface allows adding new providers (KMS, hardware tokens) without changing the bundle creation or verification logic.

## Consequences

- Sigstore keyless signing requires the `cosign` binary and network access to Fulcio and Rekor. This is a runtime dependency, not a build dependency.
- OIDC-based verification policies are only meaningful for Sigstore-signed bundles. PEM-signed bundles skip identity policy checks.
- The PEM fallback in `SigstoreSigner` means the provider field is always "sigstore" regardless of whether keyless or PEM signing was used. The presence or absence of `certificate_pem` and `oidc_issuer` fields distinguishes the two paths.
