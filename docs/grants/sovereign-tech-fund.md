# Sovereign Tech Fund Application

## Sovereign AI Supply Chain Security Infrastructure

| Field | Detail |
|-------|--------|
| **Programme** | Sovereign Tech Fund: Fund |
| **URL** | <https://www.sovereign.tech/programs/fund> |
| **Requested amount** | EUR 50,000 |
| **Eligible range** | EUR 50,000+ |
| **Deadline** | 2026-03-25 |
| **Applicant** | Ogulcan Aydogan |
| **Project** | llmsa: LLM Supply Chain Attestation |
| **Repository** | <https://github.com/ogulcanaydogan/LLM-Supply-Chain-Attestation> |
| **License** | Apache-2.0 |

---

## 1. Executive Summary

Governments and critical infrastructure operators across Europe are deploying large language models for citizen services, document processing, legal analysis, and decision support. These deployments depend on artefacts that have no supply-chain protection: the system prompts that steer model behaviour, the training data that shaped model capabilities, the evaluation benchmarks that validated model quality, the routing logic that selects which model handles which request, and the latency budgets that bound operational cost.

If any of these artefacts is tampered with (silently modified in transit, replayed from a stale version, or substituted without authorisation), the behaviour of the AI system changes with no audit trail, no detection, and no accountability. For sovereign deployments where AI systems process citizen data and inform government decisions, this represents an unacceptable security gap.

**llmsa** (LLM Supply Chain Attestation) is an open-source Go toolchain that provides cryptographic attestation, signing, distribution, verification, and enforcement for all artefacts in the LLM delivery pipeline. It is fully self-hosted, requires no external cloud services for core operation, builds exclusively on open standards (DSSE, Sigstore, OCI, OPA, SLSA), and enforces zero-trust verification at every stage from development through Kubernetes admission.

This application requests EUR 50,000 to fund an independent security audit of the signing and verification pipeline, SLSA Level 3 compliance hardening, Rekor transparency log integration, OCI distribution resilience improvements, and sovereign deployment documentation. The result will be an audited, standards-compliant attestation framework that European governments and operators can deploy with confidence to secure their AI supply chains.

---

## 2. Relevance to Digital Sovereignty

### 2.1 Self-Hosted, Zero External Dependencies for Core Operation

llmsa runs entirely on-premise. The core attestation lifecycle (generation, signing with local PEM keys, local storage, verification, and policy enforcement) requires no internet connectivity and no external service. This is a deliberate design choice for sovereign deployments:

- **No cloud signing services required.** Ed25519 PEM signing works in fully air-gapped environments. Sigstore keyless signing is available for CI/CD pipelines with internet access but isn't mandatory.
- **No external policy servers.** The dual policy engine (YAML gates + OPA/Rego) evaluates policies locally. OPA policies are shipped as static files, not fetched from external endpoints.
- **No vendor lock-in.** OCI distribution works with any compliant registry: self-hosted Harbor, GitLab Container Registry, or government-operated registries. There's no dependency on GHCR, ECR, or any commercial registry.
- **Single binary distribution.** The Go toolchain compiles to a single statically-linked binary with no runtime dependencies, simplifying deployment in restricted environments.

### 2.2 Zero-Trust Verification Model

llmsa implements a zero-trust verification model where every artefact is verified at every boundary:

- **At build time:** Attestation statements are generated with SHA-256 digests of all referenced artefacts.
- **At signing time:** Statements are wrapped in DSSE envelopes and cryptographically signed.
- During distribution, bundles are published to OCI registries with content-addressable digest pinning.
- **At verification time:** A four-stage pipeline validates signature integrity, schema conformance, digest recomputation (re-hashing source files to detect tampering), and provenance chain integrity (dependency graph validation).
- Before deployment reaches the cluster, a Kubernetes validating admission webhook runs the full verification pipeline. Default mode is fail-closed: if verification can't be completed, the workload is rejected.

No stage trusts the output of any previous stage. Every stage independently verifies the evidence it receives.

### 2.3 Standards-Based Architecture

Every component of llmsa builds on established open standards:

| Component | Standard | Governance |
|-----------|----------|------------|
| Signing envelope | DSSE (Dead Simple Signing Envelope) | in-toto project, CNCF |
| Keyless signing | Sigstore (Fulcio + Rekor) | OpenSSF, Linux Foundation |
| Artefact distribution | OCI Distribution Spec 1.1 | Open Container Initiative, Linux Foundation |
| Policy evaluation | OPA (Open Policy Agent) / Rego | CNCF Graduated Project |
| Build provenance | SLSA (Supply-chain Levels for Software Artifacts) | OpenSSF, Linux Foundation |
| Admission control | Kubernetes Admission API v1 | CNCF Graduated Project |
| Hashing | SHA-256 | NIST FIPS 180-4 |
| Encryption | age (X25519 + ChaCha20-Poly1305) | Open specification |

This standards alignment ensures interoperability with the broader cloud-native security ecosystem and avoids dependence on any single vendor's implementation.

---

## 3. Technical Description

### 3.1 Architecture Overview

llmsa is structured as a modular Go application with clearly separated concerns:

```
cmd/llmsa/           CLI entry point (Cobra-based command routing)
internal/
  attest/            Typed collectors for 5 attestation types + privacy engine
  sign/              DSSE bundle creation (Sigstore, Ed25519 PEM, KMS providers)
  verify/            Four-stage verification engine + provenance chain validator
  policy/
    yaml/            Declarative gate engine (path-based triggers)
    rego/            OPA integration engine (Rego policy evaluation)
  store/             OCI registry publish/pull with digest pinning + local storage
  hash/              SHA-256 digest computation, canonical JSON, Merkle tree hashing
  report/            JSON and Markdown audit report generation
  webhook/           Kubernetes validating admission webhook handler
pkg/
  types/             Shared type definitions (Statement, Predicate types, Privacy)
  schema/            JSON Schema loader and validator
```

### 3.2 Attestation Types

llmsa defines five domain-specific attestation types that capture the full lifecycle of LLM artefacts:

| Type | Artefacts Covered | Security Guarantee |
|------|-------------------|-------------------|
| **Prompt** | System prompts, templates, tool schemas, safety policies | The exact prompts deployed match what was reviewed and approved |
| **Corpus** | Training data, RAG documents, embeddings, vector indices | Data lineage is intact from preparation through indexing |
| **Eval** | Test suites, benchmark results, scoring configs, baselines | Model quality was validated against specific prompt and corpus versions |
| **Route** | Routing tables, fallback graphs, canary configs, budget policies | Traffic routing matches the tested and approved configuration |
| **SLO** | Latency targets, cost budgets, accuracy thresholds, query profiles | Operational constraints were defined against the verified routing setup |

Each attestation type has a dedicated collector that understands the semantic structure of its artefact class, a JSON Schema for validation, and a predicate type registered in the in-toto predicate URI namespace.

### 3.3 Provenance Chain

Attestations form a directed acyclic graph (DAG) with enforced dependencies:

```
Prompt ──┐
         ├──> Eval ──> Route ──> SLO
Corpus ──┘
```

The chain verifier validates:
- **Type-based dependencies:** An eval attestation must reference both prompt and corpus attestations.
- **ID-based references:** Explicit `depends_on` fields link specific statement IDs.
- **Temporal ordering:** Predecessors must predate successors.
- **Dangling reference detection:** References to non-existent attestations are flagged.

### 3.4 Signing Providers

| Provider | Use Case | Key Management |
|----------|----------|----------------|
| **Sigstore Keyless** | CI/CD pipelines with OIDC (GitHub Actions, GitLab CI) | No keys; attestations bound to workflow identity via OIDC token |
| **Ed25519 PEM** | Local development, air-gapped environments | User-managed PEM files |
| **KMS** (planned) | Production with hardware-backed keys | AWS KMS, GCP Cloud KMS, Azure Key Vault |

### 3.5 Verification Pipeline

Four independent verification stages, each with a distinct exit code:

| Stage | Validates | Failure Code |
|-------|-----------|-------------|
| 1. Signature | DSSE envelope signature against public key or Sigstore certificate | Exit 11 |
| 2. Schema | Statement structure against JSON Schema for the attestation type | Exit 14 |
| 3. Digest | Recomputed SHA-256 of referenced files vs. statement subjects | Exit 12 |
| 4. Chain | Provenance DAG satisfies dependency, ordering, and reference constraints | Exit 14 |

### 3.6 Policy Enforcement

**YAML Gates** (declarative): Path-based triggers that require specific attestation types when files in specified directories change. Covers the majority of CI/CD use cases with zero learning curve.

**OPA/Rego Engine** (programmatic): Full Open Policy Agent integration for cross-statement analysis, privacy guards, custom predicates, and conditional gates. Policies are evaluated locally with no external service dependency.

### 3.7 Kubernetes Admission Webhook

The `llmsa webhook serve` command runs a validating admission webhook that:

1. Intercepts Pod, Deployment, ReplicaSet, StatefulSet, DaemonSet, and Job creation
2. Extracts container image references from the pod spec
3. For each image, pulls the corresponding attestation bundle from the OCI registry
4. Runs the full four-stage verification pipeline
5. Returns an AdmissionReview response allowing or denying the resource

Configuration: namespace opt-in via `llmsa-attestation: enabled` label, configurable fail-open/fail-closed mode, TLS termination, verifier result caching.

### 3.8 Privacy Modes

| Mode | Behaviour | Use Case |
|------|-----------|----------|
| `hash_only` | Only SHA-256 digests stored; no payload in statement | Default; proves integrity without exposing content |
| `encrypted_payload` | age (X25519) encrypted blob with deterministic digest binding | Compliance workflows requiring content recovery by authorised parties |
| `plaintext_explicit` | Full payload embedded; blocked unless policy-allowlisted | Audit scenarios requiring direct content inspection |

---

## 4. Security Audit Need

The core value proposition of llmsa is the integrity of its signing and verification pipeline. If there is a bug in signature verification, digest recomputation, chain validation, or OCI bundle handling, the entire security model collapses. The project has extensive unit tests (86.9% coverage for the verify package) and a 20-case tamper detection suite, but these are developer-written tests that may share blind spots with the implementation.

An independent security audit by a qualified firm would:

1. **Review all signing codepaths** (`internal/sign/`): DSSE envelope construction, Sigstore OIDC token handling, Ed25519 signature generation, bundle serialisation.
2. **Review all verification codepaths** (`internal/verify/`): Signature verification, schema validation, subject digest recomputation, provenance chain validation.
3. **Review OCI distribution** (`internal/store/`): Bundle publish/pull, digest pinning, manifest handling, content-type validation.
4. **Review the admission webhook** (`internal/webhook/`): Image reference extraction, cache bypass scenarios, TLS configuration, fail-closed enforcement.
5. **Attempt bypass scenarios** beyond the existing 20-case tamper suite: signature confusion attacks, type confusion, chain manipulation, timing attacks on the verifier cache.
6. **Produce a public audit report** with findings, severity ratings, and remediation status, establishing trust for downstream adopters.

---

## 5. Budget

| Workstream | Amount (EUR) | Weeks | Deliverable |
|------------|-------------|-------|-------------|
| Independent security audit (external firm) | 28,000 | 6 | Public audit report with findings and remediations |
| SLSA Level 3 compliance + Rekor integration | 7,000 | 3 | Non-forgeable provenance, transparency log verification, compliance docs |
| OCI distribution hardening | 5,000 | 2 | Retry/failover, OCI 1.1 referrers, manifest signature verification |
| Sovereign deployment documentation | 5,000 | 2 | Air-gapped deployment guide, CRA mapping, government deployment playbook |
| Project management and reporting | 5,000 | 1 | STF progress reports, milestone documentation |
| **Total** | **50,000** | **14** | |

---

## 6. Timeline

### Phase 1: Audit Preparation and Scope (Weeks 1 to 2)

| Task | Output |
|------|--------|
| Compile audit scope covering all signing, verification, store, and webhook codepaths | Audit scope document |
| Identify and engage qualified security audit firm | Signed engagement letter |
| Prepare isolated test environment with full attestation pipeline | Test environment with documented setup |
| Document trust boundaries, threat model, and architecture for auditors | Auditor briefing package |

### Phase 2: Security Audit Execution (Weeks 3 to 8)

| Task | Output |
|------|--------|
| External auditors perform code review of critical packages | Running findings log |
| Auditors execute tamper detection suite and attempt additional bypasses | Bypass attempt report |
| Weekly sync calls for architecture clarification and finding triage | Meeting notes |
| Begin remediating Critical/High findings as they arrive | Patches for early findings |

### Phase 3: Remediation and SLSA Compliance (Weeks 9 to 12)

| Task | Output |
|------|--------|
| Remediate all Critical and High severity audit findings | Verified patches with tests |
| Implement SLSA Build Level 3: non-forgeable provenance, isolated builds, parameterless definition | SLSA provenance in release pipeline |
| Integrate Rekor transparency log proof verification into `llmsa verify` | `--rekor` flag with log verification |
| Document SLSA compliance posture | Compliance matrix document |
| Publish final audit report | Public audit report on GitHub |

### Phase 4: OCI Hardening and Sovereign Documentation (Weeks 13 to 14)

| Task | Output |
|------|--------|
| Implement OCI retry with exponential backoff and registry failover | Hardened `internal/store/` |
| Add OCI 1.1 referrers API support for attestation discovery | Referrers-based bundle lookup |
| Add manifest signature verification before payload extraction | Pre-extraction integrity check |
| Write air-gapped deployment guide (PEM signing, local registry, offline verification) | `docs/sovereign-deployment.md` |
| Write EU Cyber Resilience Act compliance mapping | `docs/cra-mapping.md` |
| Write government deployment playbook (RBAC, network policy, audit logging) | `docs/government-deployment.md` |
| Final integration testing and tagged release | Release incorporating all funded work |

---

## 7. Impact

### 7.1 Immediate Impact

- **First audited attestation framework for LLM supply chains.** The public audit report establishes a baseline of trust that no comparable project currently offers. Organisations deploying LLMs in regulated sectors (healthcare, finance, government) can point to an independent assessment of the security guarantees.
- **SLSA Level 3 provenance for AI artefact attestations.** This is the first project to apply SLSA provenance standards to AI-specific artefacts, extending the SLSA framework beyond its traditional scope of container images and binaries.
- **Sovereign deployment readiness.** Air-gapped deployment documentation and CRA compliance mapping enable European governments to adopt the tool without dependency on external cloud services.

### 7.2 Community Impact

- **Standards contribution.** The LLM-specific attestation taxonomy (prompt, corpus, eval, route, SLO) is a novel contribution to the supply-chain security space. By publishing the predicate types as open specifications, llmsa enables other tools and platforms to generate and verify compatible attestations, creating an interoperable attestation network.
- **Reference implementation.** The Kubernetes admission webhook provides a reference implementation for deployment-time attestation enforcement that can be adopted or adapted by other projects and platforms.
- **Upstream contributions.** The project has identified and reported issues in Sigstore libraries related to certificate parse variability, contributing back to the broader signing community.

### 7.3 Long-Term Impact

- **CRA compliance infrastructure.** As the EU Cyber Resilience Act enters enforcement, organisations will need technical tools to demonstrate supply-chain integrity for their AI systems. llmsa provides the foundational infrastructure for this compliance posture.
- **Trust foundation for AI deployment.** By making AI system behaviour auditable and tamper-evident, llmsa contributes to the broader goal of trustworthy AI deployment in European digital infrastructure.

---

## 8. Open Source Commitment

- **License:** Apache-2.0, permitting commercial and government use with no copyleft obligations
- **Development:** All development is conducted in the open on GitHub
- Built exclusively on open standards (DSSE, Sigstore, OCI, OPA, SLSA, age) with no proprietary dependencies or cloud-specific requirements
- **Governance:** Published GOVERNANCE.md with clear decision-making process and path to maintainership
- SECURITY.md defines vulnerability reporting policy, scope definitions, and response timelines
- **Contributions:** Fork-and-pull-request model documented in CONTRIBUTING.md

---

## 9. Supporting Materials Checklist

- [ ] GitHub repository: <https://github.com/ogulcanaydogan/LLM-Supply-Chain-Attestation>
- [ ] Apache-2.0 LICENSE file
- [ ] README with architecture diagrams, Mermaid sequence diagrams, and CLI reference
- [ ] SECURITY.md with vulnerability scope, reporting process, and response timelines
- [ ] CONTRIBUTING.md with development setup, test coverage targets, and ADR template
- [ ] GOVERNANCE.md with maintainer list and decision-making process
- [ ] Threat model: `docs/threat-model.md` (STRIDE analysis, 8 threat scenarios, known limitations)
- [ ] 7 Architecture Decision Records in `docs/adr/`
- [ ] API reference: `docs/api-reference.md`
- [ ] Policy guide: `docs/policy-guide.md`
- [ ] Kubernetes admission guide: `docs/k8s-admission.md`
- [ ] Benchmark methodology: `docs/benchmark-methodology.md`
- [ ] CHANGELOG.md with full release history (v0.1.0 through v1.0.1)
- [ ] 9 CI/CD workflows in `.github/workflows/`
- [ ] 20-case tamper detection test suite
- [ ] End-to-end integration test suite
- [ ] Working example with all 5 attestation types: `examples/tiny-rag/`
- [ ] Helm chart and Kubernetes manifests for webhook deployment
- [ ] Dockerfile with multi-stage distroless build

---

## 10. Submission Steps

1. **Navigate** to <https://www.sovereign.tech/programs/fund>.
2. **Review** the current call guidelines and eligibility criteria.
3. **Prepare** an application using the STF submission form or email template.
4. **Project name:** llmsa: LLM Supply Chain Attestation
5. **Project URL:** `https://github.com/ogulcanaydogan/LLM-Supply-Chain-Attestation`
6. **Requested amount:** EUR 50,000
7. **Executive Summary:** Paste Section 1 above.
8. **Relevance to Digital Sovereignty:** Paste Section 2 (self-hosted, zero-trust, standards-based).
9. **Technical Description:** Paste Section 3 (architecture, attestation types, signing, verification, K8s webhook, privacy).
10. **Security Audit Need:** Paste Section 4.
11. **Budget:** Paste Section 5.
12. **Timeline:** Paste Section 6.
13. **Impact:** Paste Section 7.
14. **Open Source Commitment:** Paste Section 8.
15. **Attach or link** supporting materials from the checklist in Section 9.
16. **Review** the full application for consistency, verify all links resolve, and confirm the total matches the requested amount.
17. **Submit** before 2026-03-25.

---

## 11. Post-Submission

- The Sovereign Tech Fund review process typically takes 4 to 8 weeks
- Be prepared for a technical interview or demo session with the STF evaluation team
- STF may request a revised scope, budget, or timeline during evaluation
- If accepted, funding is disbursed based on milestone completion with regular progress reporting
- All funded work must remain open-source under the existing Apache-2.0 license
- STF requires public acknowledgement of funding in the project's documentation and release notes
