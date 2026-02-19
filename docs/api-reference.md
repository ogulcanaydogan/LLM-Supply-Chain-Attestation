# API Reference

This document describes the exported Go API surface of the `llmsa` framework. All internal packages are under `internal/` and are not importable by external modules. The public API is exposed via `pkg/types` and `pkg/schema`.

## Public Packages

### `pkg/types`

Core type definitions for attestation statements and predicates.

#### Types

| Type | Description |
|------|-------------|
| `Statement` | Top-level attestation statement containing metadata, subjects, predicate, and privacy config |
| `Generator` | Tool metadata: name, version, git SHA |
| `Subject` | Artifact reference: name, URI, digest, size |
| `Digest` | SHA-256 digest wrapper |
| `Privacy` | Privacy mode config: mode, encrypted blob digest, recipient fingerprint |
| `PromptPredicate` | Predicate for prompt attestations: template digests, tool schemas, safety policies |
| `CorpusPredicate` | Predicate for corpus attestations: connector configs, chunking, embedding model, vector index |
| `EvalPredicate` | Predicate for eval attestations: test sets, scoring, metrics, thresholds, regression flag |
| `RoutePredicate` | Predicate for route attestations: provider set, budget policy, fallback graph, routing strategy |
| `SLOPredicate` | Predicate for SLO attestations: latency targets, cost caps, error budgets, time windows |
| `NamedDigest` | Name-digest pair used in corpus connector configs |
| `ProviderModel` | Provider-model pair used in route provider sets |
| `TimeWindow` | Start-end time range for SLO measurement windows |

#### Constants

| Constant | Value |
|----------|-------|
| `AttestationPrompt` | `"prompt_attestation"` |
| `AttestationCorpus` | `"corpus_attestation"` |
| `AttestationEval` | `"eval_attestation"` |
| `AttestationRoute` | `"route_attestation"` |
| `AttestationSLO` | `"slo_attestation"` |

#### Functions

| Function | Signature | Description |
|----------|-----------|-------------|
| `PredicateURI` | `(attestationType string) string` | Returns the predicate URI for a given attestation type |

### `pkg/schema`

JSON Schema validation for attestation statements.

| Function | Signature | Description |
|----------|-----------|-------------|
| `Validate` | `(schemaPath string, doc any) ([]string, error)` | Validates a document against a JSON Schema file, returns validation errors |

---

## Internal Packages

These packages are not importable externally but define the framework's core logic.

### `internal/attest`

Attestation creation and collection.

| Function | Signature | Description |
|----------|-----------|-------------|
| `CreateByType` | `(opts CreateOptions) ([]string, error)` | Creates attestation statement(s) for a given type and config, returns output file paths |

| Type | Description |
|------|-------------|
| `CreateOptions` | Options for attestation creation: Type, ConfigPath, OutDir, ChangedOnly, DeterminismCheck, Ref |

### `internal/sign`

DSSE bundle creation and cryptographic signing.

#### Signer Interface

```go
type Signer interface {
    Sign(payload []byte) (SignMaterial, error)
}
```

#### Implementations

| Type | Description |
|------|-------------|
| `PEMSigner` | Ed25519 signing with local PEM key files |
| `SigstoreSigner` | Sigstore keyless signing via cosign with OIDC, falls back to PEM |
| `KMSSigner` | Placeholder for cloud KMS integration (not yet implemented) |

#### Bundle Types

| Type | Description |
|------|-------------|
| `Bundle` | DSSE envelope with metadata: envelope + metadata |
| `Envelope` | Payload type, base64 payload, signatures array |
| `Signature` | Key ID, signature, provider, public key PEM, certificate PEM, OIDC claims |
| `Metadata` | Bundle version, creation timestamp, statement hash |
| `SignMaterial` | Output of signing: key ID, signature base64, provider, public key, OIDC claims |

#### Functions

| Function | Signature | Description |
|----------|-----------|-------------|
| `NewPEMSigner` | `(keyPath string) (*PEMSigner, error)` | Creates a PEM signer from an Ed25519 private key file |
| `GeneratePEMPrivateKey` | `(path string) error` | Generates a new Ed25519 key pair and writes the private key as PEM |
| `CreateBundle` | `(statement any, material SignMaterial) (Bundle, error)` | Wraps a statement in a DSSE envelope with signing material |
| `DecodePayload` | `(bundle Bundle, out any) error` | Decodes the base64 payload from a bundle into a target struct |
| `WriteBundle` | `(path string, b Bundle) error` | Writes a bundle to a JSON file |
| `ReadBundle` | `(path string) (Bundle, error)` | Reads a bundle from a JSON file |

### `internal/verify`

Multi-stage verification engine.

| Function | Signature | Description |
|----------|-----------|-------------|
| `Run` | `(opts Options) (Result, error)` | Executes the full verification pipeline: signatures, subjects, schemas, chain |
| `VerifySignature` | `(bundle Bundle, policy SignerPolicy) error` | Verifies the cryptographic signature on a bundle |
| `VerifySubjects` | `(statement Statement, sourceDir string) error` | Recomputes subject digests and compares against recorded values |
| `VerifySchemas` | `(statement Statement, schemaDir string) error` | Validates statement against its JSON Schema |
| `VerifyProvenanceChain` | `(statements []Statement) (*ChainResult, error)` | Validates the provenance DAG: references, temporal ordering, type constraints |
| `WriteJSON` | `(path string, result Result) error` | Writes verification results as JSON |

| Type | Description |
|------|-------------|
| `Options` | Verification options: BundleDir, SourceDir, SchemaDir, SignerPolicy |
| `Result` | Verification outcome: Passed, ExitCode, BundleCount, Failures, Chain |
| `SignerPolicy` | Policy for identity verification: required OIDC issuer, identity pattern (regex) |
| `ChainResult` | Provenance chain outcome: Valid, Edges, Violations |

#### Exit Codes

| Code | Constant | Meaning |
|------|----------|---------|
| 0 | — | All checks passed |
| 10 | `ExitMissing` | Required attestations missing |
| 11 | `ExitSignature` | Signature verification failed |
| 12 | `ExitTamper` | Subject digest mismatch (tamper detected) |
| 13 | `ExitPolicy` | Policy violation |
| 14 | `ExitSchema` | Schema validation failed |

### `internal/store`

Local filesystem and OCI registry storage.

| Function | Signature | Description |
|----------|-----------|-------------|
| `SaveLocal` | `(srcPath, dstDir string) (string, error)` | Copies a bundle file to a local directory, returns destination path |
| `PublishOCI` | `(bundlePath, ref string) (string, error)` | Publishes a bundle to an OCI registry, returns digest-pinned reference |
| `PullOCI` | `(ref, outputPath string) error` | Pulls a bundle from an OCI registry to a local file |
| `EnsureDefaultAttestationDir` | `() (string, error)` | Creates `.llmsa/attestations/` directory, returns relative path |

### `internal/hash`

SHA-256 digest and canonical JSON utilities.

| Function | Signature | Description |
|----------|-----------|-------------|
| `DigestFile` | `(path string) (string, error)` | Computes SHA-256 digest of a file, returns `sha256:<hex>` |
| `DigestBytes` | `(data []byte) string` | Computes SHA-256 digest of bytes, returns `sha256:<hex>` |
| `DigestDir` | `(dirPath string) (string, error)` | Computes a deterministic tree digest of a directory |
| `CanonicalJSON` | `(v any) ([]byte, error)` | Produces canonical JSON with sorted keys for deterministic hashing |

### `internal/policy/yaml`

Declarative YAML policy gate engine.

| Function | Signature | Description |
|----------|-----------|-------------|
| `Evaluate` | `(policyPath string, input Input) ([]Violation, error)` | Evaluates attestation results against a YAML policy file |

| Type | Description |
|------|-------------|
| `Input` | Policy evaluation input: attestation results, bundle metadata |
| `Violation` | Policy violation: severity, message, rule reference, affected bundle |

### `internal/policy/rego`

OPA Rego policy engine.

| Function | Signature | Description |
|----------|-----------|-------------|
| `Evaluate` | `(policyDir string, input any) ([]Violation, error)` | Evaluates attestation results against Rego policies in a directory |
| `BuildInput` | `(result Result) map[string]any` | Constructs the input document for Rego evaluation from verification results |

### `internal/report`

Audit report generation.

| Function | Signature | Description |
|----------|-----------|-------------|
| `GenerateJSON` | `(result Result, path string) error` | Writes verification results as a JSON audit report |
| `GenerateMarkdown` | `(result Result, path string) error` | Writes verification results as a Markdown audit report |

### `internal/webhook`

Kubernetes validating admission webhook.

| Type | Description |
|------|-------------|
| `Handler` | HTTP handler for admission review requests |
| `Config` | Webhook configuration: registry prefix, fail-open, policy path, cache TTL |
| `ImageRef` | Container image reference extracted from Pod spec |

| Function | Signature | Description |
|----------|-----------|-------------|
| `NewHandler` | `(cfg Config) *Handler` | Creates a new admission webhook handler |
| `ExtractImageRefs` | `(spec PodSpec) []ImageRef` | Extracts all container image references from a Pod spec |
| `AttestationRef` | `(registryPrefix, imageRef string) (string, error)` | Constructs the OCI reference for an image's attestation bundle |
