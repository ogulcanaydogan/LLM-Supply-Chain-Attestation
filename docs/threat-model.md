# Threat Model

This document describes the threat model for the LLM Supply-Chain Attestation framework, covering the adversarial scenarios addressed, the mitigations implemented, and the known limitations of the current design.

## Scope

The framework protects the integrity and provenance of five categories of LLM artifacts throughout the software supply chain: system prompts and templates, training/retrieval corpora, evaluation benchmarks, model routing configurations, and service-level objectives. The threat model considers attacks that could occur between artifact creation and deployment-time enforcement.

## Trust Boundaries

The system defines three trust boundaries:

1. **Developer Workstation → CI/CD Pipeline** — Artifacts are committed to a git repository and flow into automated pipelines. Attestations are created and signed at this boundary.
2. **CI/CD Pipeline → OCI Registry** — Signed bundles are published to OCI-compliant registries with digest-pinned references. Bundle integrity is preserved by content-addressable storage.
3. **OCI Registry → Kubernetes Cluster** — The validating admission webhook pulls and verifies attestation bundles at pod admission time, enforcing the verification boundary at deployment.

## Threats Addressed

### T1: Artifact Tampering After Signing

An attacker modifies a subject file (e.g., a system prompt or evaluation dataset) after the attestation statement has been signed. The tampered artifact would produce different behaviour at runtime without detection.

**Mitigation**: The verification engine recomputes SHA-256 digests of all subject files referenced in each attestation statement and compares them against the recorded values. Any mismatch produces exit code 12 (tamper detected). The 20-case tamper detection test suite validates this across all five attestation types.

### T2: Signature Spoofing

An attacker crafts a valid-looking DSSE envelope with a forged signature, attempting to pass verification with an untrusted key.

**Mitigation**: Signature verification extracts the public key from the bundle's signature metadata and verifies the Ed25519/ECDSA signature against the canonical JSON payload. For Sigstore bundles, the framework additionally validates OIDC issuer and identity claims against the configured policy, binding signatures to specific CI/CD workflow identities.

### T3: Replay of Stale Attestations

An attacker replays an old but legitimately signed attestation bundle after the underlying artifacts have been updated, bypassing checks for the new version.

**Mitigation**: The provenance chain verification enforces temporal ordering — predecessor attestations must have a `generated_at` timestamp that precedes or equals the dependent attestation's timestamp. Combined with the `--changed-only` mode and git-based change detection, stale attestations are detected when source files have been modified.

### T4: Sensitive Payload Leakage

Attestation statements may inadvertently include sensitive intellectual property (model weights, proprietary prompts, training data samples) in plaintext form, exposing them to anyone with registry access.

**Mitigation**: Three privacy modes control payload handling:
- **`hash_only`** (default): Only cryptographic digests are stored. No recoverable content.
- **`plaintext_explicit`**: Raw content is included but gated by policy — the statement ID must be explicitly allowlisted in the policy's `plaintext_allowlist`.
- **`encrypted_payload`**: Content is encrypted using age (X25519) before inclusion, allowing authorised parties to decrypt while preventing casual exposure.

### T5: Missing Attestation Types

A deployment proceeds without the required attestation types, creating gaps in the provenance chain. For example, deploying without an evaluation attestation means model quality was never validated.

**Mitigation**: YAML policy gates define trigger paths and required attestation types. The gate engine checks that all required types are present when files matching the trigger paths have changed. The Kubernetes admission webhook enforces this at deployment time by requiring valid attestation bundles before admitting pods.

### T6: Provenance Chain Breaks

An attestation references dependencies that do not exist or are invalid, creating a broken provenance chain where downstream artifacts appear verified but lack upstream attestation.

**Mitigation**: The provenance chain verifier constructs a dependency DAG from the five attestation types (eval→prompt+corpus, route→eval, slo→route) and validates that all required predecessors are present with valid signatures and correct temporal ordering. Missing edges produce chain violations.

### T7: Registry Compromise

An attacker gains write access to the OCI registry and replaces attestation bundles with crafted ones containing valid-looking but incorrect digests.

**Mitigation**: OCI references are digest-pinned (`@sha256:...`) after publish, ensuring content integrity via content-addressable storage. The verification pipeline validates the cryptographic signature of the bundle payload, so replaced bundles with invalid signatures are detected at verification time.

### T8: Webhook Bypass

An attacker attempts to deploy workloads in namespaces where the admission webhook is not enforced, or the webhook is unavailable.

**Mitigation**: The webhook uses namespace opt-in via the `llmsa-attestation: enabled` label, allowing gradual rollout. In production, the webhook is configured with `failurePolicy: Fail` (fail-closed), meaning that if the webhook is unavailable, pod admission is denied. The `--fail-open` mode is available only for non-production gradual rollout.

## STRIDE Analysis Summary

| Threat Category | Threat | Mitigation |
|----------------|--------|------------|
| **Spoofing** | Forged signatures | DSSE signature verification, OIDC identity binding |
| **Tampering** | Modified artifacts | Subject digest recomputation, content-addressable OCI storage |
| **Repudiation** | Denied provenance | Signed attestations with generator metadata and timestamps |
| **Information Disclosure** | Payload leakage | Privacy modes (hash_only, encrypted_payload), policy gating |
| **Denial of Service** | Webhook unavailability | Fail-closed default, health probes, replica scaling |
| **Elevation of Privilege** | Namespace bypass | Namespace-scoped labels, RBAC on webhook configuration |

## Known Limitations

1. **No Transparency Log Verification**: The framework signs bundles using Sigstore but does not currently verify Rekor transparency log inclusion proofs. This means signed bundles are cryptographically valid but lack public auditability.
2. **Single-Cluster Scope**: The admission webhook operates within a single Kubernetes cluster. Multi-cluster federation requires deploying the webhook to each cluster independently.
3. **No Runtime Attestation**: The framework verifies artifacts at deployment time, not runtime. If artifacts are modified after pod admission (e.g., via mounted volumes), the change is not detected.
4. **Trust-on-First-Use for PEM Keys**: Local PEM signing trusts the generated key without a certificate chain. Production deployments should use Sigstore keyless signing for identity-bound verification.
5. **No Revocation**: There is no mechanism to revoke a previously signed attestation bundle. Revocation would require integration with a transparency log or a separate revocation list.

## Future Mitigations

- **Rekor integration** for transparency log proof validation.
- **Runtime attestation hooks** for continuous verification via eBPF or admission controller mutation.
- **KMS provider** for hardware-backed key management.
- **Multi-cluster federation** via federated webhook configuration.
