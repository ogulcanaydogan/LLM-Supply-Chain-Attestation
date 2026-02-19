# ADR-0004: Privacy Modes for Attestation Statements

## Status

Accepted

## Date

2025-06-01

## Context

Attestation statements contain metadata about LLM artifacts — file paths, configuration values, model names, and evaluation results. In regulated industries (healthcare, finance) and multi-tenant environments, this metadata may be sensitive. We needed a mechanism to control what information appears in plaintext within attestation bundles.

## Decision

We implemented three privacy modes, configurable per attestation type via the YAML config:

1. **`hash_only`** (default) — Subject file contents are never included in the statement. Only SHA-256 digests and file paths appear. All metadata fields are present but values that could contain sensitive content are hashed.

2. **`plaintext_explicit`** — Statement includes full metadata values in plaintext. Used when transparency is preferred over confidentiality (e.g., open-source projects, public model evaluations).

3. **`encrypted_payload`** — The full statement is encrypted using age (X25519) before being placed in the DSSE envelope. The bundle contains the encrypted ciphertext, the recipient's public key fingerprint, and the path to the encrypted payload file. Only holders of the corresponding age private key can decrypt the statement.

Privacy configuration is applied in `internal/attest/service.go` via `ApplyPrivacyConfig()`.

## Rationale

- **Default to `hash_only`**: The principle of least privilege applies to attestation metadata. Most verification use cases (signature validation, digest comparison, chain checking) work with digests alone.
- **`encrypted_payload` uses age, not GPG**: age is simpler, has no configuration complexity, and produces deterministic ciphertext sizes. The X25519 key exchange is well-suited to CI/CD environments where recipient public keys can be committed to the repository.
- **Per-type configuration**: Different attestation types have different sensitivity profiles. Prompt attestations may use `hash_only` while eval attestations use `plaintext_explicit` to share benchmark results.

## Consequences

- Encrypted payloads cannot be verified without the decryption key. This is by design — the signature is over the encrypted ciphertext, not the plaintext.
- The `encrypted_payload` mode requires the `age` binary at attestation creation time and the corresponding private key at decryption time.
- Schema validation operates on the decrypted statement. Encrypted bundles pass schema checks during creation but cannot be re-validated without the key.
