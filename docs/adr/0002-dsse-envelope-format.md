# ADR-0002: DSSE Envelope Format for Bundle Signing

## Status

Accepted

## Date

2025-05-15

## Context

We needed a signing envelope format that wraps attestation statements with cryptographic signatures. The main candidates were:

1. **DSSE (Dead Simple Signing Envelope)** — Lightweight, widely adopted by Sigstore and in-toto ecosystems
2. **JWS (JSON Web Signature)** — Mature standard but complex header negotiation
3. **Custom envelope** — Full control but no ecosystem compatibility
4. **CMS/PKCS#7** — Enterprise-grade but heavyweight for CI/CD use cases

## Decision

We adopted DSSE as the envelope format. Each bundle contains:

- `envelope.payloadType`: MIME type identifying the statement format
- `envelope.payload`: Base64-encoded canonical JSON of the attestation statement
- `envelope.signatures[]`: Array of signatures, each with key ID, signature bytes, provider metadata, public key PEM, and optional OIDC identity claims
- `metadata`: Bundle version, creation timestamp, and statement content hash

The bundle structure is defined in `internal/sign/dsse_bundle.go`.

## Rationale

- **Ecosystem alignment**: DSSE is the signing format used by Sigstore's cosign, in-toto attestations, and SLSA provenance. Using DSSE means our bundles can be processed by existing tooling.
- **Payload-type agnosticism**: DSSE separates the signing mechanism from the payload format. We use `application/vnd.llmsa.statement.v1+json` as the payload type, allowing future statement format evolution without envelope changes.
- **Multi-signature support**: The `signatures[]` array natively supports multiple signers, enabling use cases like developer + CI co-signing.
- **Canonical JSON payload**: We base64-encode the canonical JSON (sorted keys, no trailing whitespace) of the statement, ensuring deterministic signature verification regardless of JSON serialization differences.

## Consequences

- Bundle files are larger than raw statements due to base64 encoding and signature metadata. A typical bundle is 2-3x the size of its statement.
- Verification requires base64 decoding before JSON parsing, adding a step compared to plaintext envelopes.
- The canonical JSON implementation (`internal/hash/canonical.go`) must handle all JSON types consistently. Edge cases (unicode escaping, number formatting) are covered by tests.
