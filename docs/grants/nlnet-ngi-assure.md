# NLnet NGI Assure Grant Proposal

## Supply Chain Security for AI Models

| Field | Detail |
|-------|--------|
| **Call** | NGI Assure (NLnet Foundation) |
| **URL** | <https://nlnet.nl/propose/> |
| **Requested amount** | EUR 40,000 |
| **Eligible range** | EUR 5,000 -- 50,000 |
| **Deadline** | 2026-04-01 |
| **Applicant** | Ogulcan Aydogan |
| **Project** | llmsa -- LLM Supply Chain Attestation |
| **Repository** | <https://github.com/ogulcanaydogan/LLM-Supply-Chain-Attestation> |
| **License** | Apache-2.0 |

---

## 1. Abstract

Large language models are increasingly deployed in safety-critical applications, yet the artefacts that define their behaviour -- system prompts, training corpora, evaluation benchmarks, routing tables, and latency budgets -- flow through CI/CD pipelines with no integrity verification, no provenance tracking, and no policy enforcement. Existing supply-chain frameworks (SLSA, in-toto, SBOM) address traditional build artefacts but leave AI-specific components entirely unprotected.

**llmsa** is an open-source Go toolchain that closes this gap. It generates, signs, distributes, verifies, and enforces typed cryptographic attestations for five categories of LLM artefacts: prompt, corpus, eval, route, and SLO. Attestations are wrapped in DSSE envelopes, signed via Sigstore keyless OIDC binding (eliminating key management), published to OCI-compliant registries, and enforced at deployment time through a Kubernetes validating admission webhook. A dual policy engine (declarative YAML gates plus Open Policy Agent/Rego) evaluates provenance chain integrity before any workload reaches a cluster.

The project is production-ready with 7 tagged releases, 9 CI workflows, a 20-case tamper-detection test suite, privacy-preserving attestation modes (hash-only, encrypted payload via age/X25519), and comprehensive documentation including a formal threat model. This grant would fund an independent security audit, SLSA Level 3 compliance hardening, OCI distribution resilience improvements, and expanded documentation to enable adoption by EU organisations preparing for the Cyber Resilience Act.

---

## 2. Description of Work

### 2.1 Background and Motivation

The EU Cyber Resilience Act (CRA), expected to enter full enforcement in 2027, will require manufacturers and deployers of software products -- including AI systems -- to demonstrate supply-chain integrity throughout the product lifecycle. Current attestation frameworks were designed for container images, binaries, and packages. They have no understanding of the artefact types that determine LLM behaviour: the prompts that steer outputs, the data that shaped model weights, the benchmarks that validated quality, the routing logic that selects models, or the operational budgets that bound latency and cost.

This creates a systemic blind spot. A system prompt can be silently modified in production, altering model behaviour without any audit trail. Training data can be poisoned between preparation and deployment with no tamper detection. Evaluation results can be fabricated or replayed from stale runs to bypass quality gates. Routing logic can be changed to redirect traffic to cheaper, less capable models without accountability.

llmsa was designed from the ground up to address these threats with a domain-specific attestation taxonomy, cryptographic signing, and policy-enforced verification at every stage of the LLM delivery pipeline.

### 2.2 What llmsa Does

llmsa is a local-first CLI and CI toolchain written in Go that implements a complete attestation lifecycle for LLM artefacts:

**Typed Attestation Generation.** Five domain-specific collectors generate attestation statements for prompts (system prompts, templates, tool schemas, safety policies), corpora (training data, RAG documents, embeddings, vector indices), evaluations (test suites, benchmark results, scoring configs), routes (routing tables, fallback graphs, canary configs), and SLOs (latency targets, cost budgets, accuracy thresholds). Each collector understands the semantic structure of its artefact type, extracting the correct digests and metadata rather than treating everything as opaque blobs.

**Provenance Chain Verification.** A directed acyclic dependency graph enforces logical ordering between attestation types: evaluations must reference both prompt and corpus attestations; routing attestations must reference evaluations; SLO attestations must reference routes. The chain verifier validates type-based dependencies, explicit ID references, temporal ordering, and dangling reference detection.

**Sigstore Keyless Signing.** Production signing uses Sigstore with OIDC tokens from CI providers (GitHub Actions, GitLab CI), binding attestations to the specific workflow identity that produced them. No private keys to manage, rotate, or protect. Ed25519 PEM signing is available for local development and air-gapped environments.

**OCI-Native Distribution.** Signed DSSE bundles are published to any OCI-compliant container registry (GHCR, ECR, ACR, Docker Hub) as first-class artefacts with content-addressable digest pinning.

**Kubernetes Admission Enforcement.** A validating admission webhook intercepts Pod, Deployment, ReplicaSet, StatefulSet, DaemonSet, and Job creation. For each container image, it pulls the attestation bundle from the OCI registry and runs the full four-stage verification pipeline (signature, schema, digest recomputation, chain integrity) before allowing or denying admission.

**Dual Policy Engine.** Declarative YAML gates cover the majority of CI/CD use cases. OPA/Rego policies enable advanced cross-statement analysis, privacy guards, and conditional gates based on attestation metadata.

**Privacy Modes.** Three modes control payload handling: hash-only (default, no content stored), encrypted payload (age/X25519 encryption with deterministic digest binding), and plaintext explicit (policy-gated, for audit scenarios).

### 2.3 Current Status

| Metric | Value |
|--------|-------|
| Tagged releases | 7 (v0.1.0 through v1.0.1) |
| CI workflows | 9 (attest-verify, nightly benchmark, release, release-verify, scorecard, public footprint, completion health, upstream PR followup, third-party mention publish) |
| Tamper test cases | 20 |
| Test coverage | verify 86.9%, attest 83.6%, store 84.5%, sign 77.3%, policy/rego 94.3%, schema 100%, types 100% |
| Architecture Decision Records | 7 |
| Documentation pages | 15+ (threat model, policy guide, quickstart, benchmark methodology, K8s admission, API reference, ADRs, public footprint playbook) |
| License | Apache-2.0 |
| Language | Go 1.25 |

The project has reached feature completeness for its initial scope: all five attestation types, full signing and verification pipelines, OCI distribution, Kubernetes admission enforcement, and dual policy engines are implemented, tested, and documented.

### 2.4 What the Grant Would Fund

The grant would fund four workstreams that move llmsa from a feature-complete open-source project to a hardened, audited, and adoption-ready security tool suitable for EU CRA compliance:

1. **Independent Security Audit of the Signing and Verification Pipeline** (EUR 28,000). Commission a third-party security firm to perform a comprehensive audit of the DSSE signing, Sigstore integration, Ed25519 PEM signing, signature verification, digest recomputation, chain verification, and OCI bundle integrity checks. This is the most critical workstream: the project's security guarantees rest entirely on the correctness of these codepaths. The audit would produce a public report with findings, severity ratings, and remediation status.

2. **SLSA Level 3 Compliance** (EUR 5,000). Implement the remaining requirements for SLSA Build Level 3: non-forgeable provenance generated by the build platform, isolated build environment, and parameterless build definition. This includes integrating SLSA provenance generation into the GoReleaser pipeline, adding Rekor transparency log verification, and documenting the compliance posture.

3. **OCI Distribution Hardening** (EUR 4,000). Add retry logic, registry failover, content-type validation, and bundle integrity verification during pull operations. Implement signature verification of OCI manifest metadata before extracting attestation payloads. Add support for OCI 1.1 referrers API for attestation discovery.

4. **Documentation and Adoption Enablement** (EUR 3,000). Write a CRA compliance mapping document showing how llmsa attestations satisfy specific CRA requirements. Produce deployment guides for major cloud providers (AWS EKS, GCP GKE, Azure AKS). Create a contributor onboarding tutorial and expand the existing quickstart guide with production deployment scenarios.

---

## 3. Budget

| Workstream | Amount (EUR) | Effort | Deliverable |
|------------|-------------|--------|-------------|
| Security audit of signing/verification pipeline | 28,000 | External audit firm engagement | Public audit report with findings and remediations |
| SLSA Level 3 compliance | 5,000 | ~4 weeks part-time | SLSA provenance generation, Rekor verification, compliance documentation |
| OCI distribution hardening | 4,000 | ~3 weeks part-time | Retry/failover logic, OCI 1.1 referrers, bundle integrity checks |
| Documentation and adoption enablement | 3,000 | ~3 weeks part-time | CRA mapping, cloud deployment guides, contributor tutorial |
| **Total** | **40,000** | **16 weeks** | |

---

## 4. Milestones and Timeline

### Milestone 1: Security Audit Preparation (Weeks 1--2)

- Compile audit scope document covering all signing and verification codepaths
- Identify and engage qualified security audit firm
- Prepare isolated test environment with full attestation pipeline
- Document all trust boundaries and threat model assumptions for auditors

**Deliverable:** Audit scope document, signed engagement letter

### Milestone 2: Security Audit Execution (Weeks 3--8)

- External auditors perform code review of `internal/sign/`, `internal/verify/`, `internal/store/`, and `internal/webhook/`
- Auditors execute the 20-case tamper detection suite and attempt additional bypass scenarios
- Weekly sync calls to clarify architecture and triage findings
- Maintainer begins remediating findings as they are reported

**Deliverable:** Draft audit report with findings categorised by severity

### Milestone 3: Audit Remediation and SLSA Level 3 (Weeks 9--12)

- Remediate all Critical and High severity audit findings
- Implement SLSA Build Level 3 requirements:
  - Non-forgeable provenance via GitHub Actions OIDC
  - Isolated build environment with GoReleaser in ephemeral runners
  - Parameterless build definition
- Integrate Rekor transparency log proof verification into `llmsa verify`
- Document SLSA compliance posture

**Deliverable:** Remediated codebase, SLSA provenance generation, Rekor integration, final audit report (public)

### Milestone 4: OCI Hardening and Documentation (Weeks 13--16)

- Implement OCI distribution hardening: retry with exponential backoff, registry failover, content-type validation, manifest signature verification
- Add OCI 1.1 referrers API support for attestation discovery
- Write CRA compliance mapping document
- Produce cloud deployment guides (EKS, GKE, AKS)
- Write contributor onboarding tutorial
- Update quickstart guide with production deployment scenarios
- Final integration testing across all changes

**Deliverable:** Hardened OCI distribution, comprehensive documentation set, tagged release incorporating all grant-funded work

---

## 5. Relevance to the NGI Mission

### 5.1 Trustworthiness

llmsa directly addresses the trustworthiness dimension of the Next Generation Internet by making AI system behaviour auditable and tamper-evident. Every artefact that influences model output -- from the system prompt to the latency budget -- is cryptographically attested, signed, and verified before deployment. This creates an unbroken chain of evidence from development through production, enabling organisations and individuals to trust that the AI system they interact with operates on verified components.

The independent security audit funded by this grant would provide external validation of these trust guarantees, producing a public report that any deployer can review before relying on llmsa for their own attestation infrastructure.

### 5.2 Openness

The project is Apache-2.0 licensed, developed entirely in the open on GitHub, and built exclusively on open standards and open-source components:

- **DSSE** (Dead Simple Signing Envelope) from the in-toto project
- **Sigstore** for keyless signing and transparency
- **OCI** (Open Container Initiative) for artefact distribution
- **OPA** (Open Policy Agent) for policy evaluation
- **SLSA** (Supply-chain Levels for Software Artifacts) for provenance

There are no proprietary dependencies, no vendor lock-in, and no cloud-specific requirements. The tool runs entirely self-hosted with local PEM signing for air-gapped environments.

### 5.3 Resilience

By distributing attestation bundles through OCI registries with content-addressable digest pinning, llmsa ensures that attestation evidence survives independently of any single infrastructure provider. The fail-closed Kubernetes admission webhook prevents unverified workloads from entering clusters even when external services are degraded. The dual policy engine (YAML + Rego) allows organisations to define and enforce their own verification requirements without depending on centralised policy servers.

### 5.4 EU Cyber Resilience Act Alignment

The CRA requires manufacturers to ensure that products with digital elements are delivered with documented security properties throughout the supply chain. llmsa provides the technical infrastructure for demonstrating this compliance for AI systems: cryptographic provenance for all AI-specific artefacts, policy enforcement at build and deployment time, and auditable verification reports. The CRA compliance mapping document funded by this grant would make this alignment explicit and actionable for EU organisations.

---

## 6. Comparable Projects

| Project | Scope | Relation to llmsa |
|---------|-------|-------------------|
| [in-toto](https://in-toto.io/) | Generic software supply chain layout and verification | llmsa builds on in-toto's DSSE format but adds LLM-specific attestation types, provenance chain semantics, and K8s admission enforcement |
| [Sigstore/cosign](https://www.sigstore.dev/) | Container image signing and verification | llmsa uses Sigstore as a signing provider but extends attestation beyond container images to LLM artefacts (prompts, corpora, evals, routes, SLOs) |
| [SLSA](https://slsa.dev/) | Build provenance framework | llmsa generates SLSA-compatible provenance and aims for Level 3 compliance; extends the model to AI-specific artefact types |
| [ModelScan](https://github.com/protectai/modelscan) | Static analysis of serialised ML models for malicious payloads | Complementary: ModelScan detects malicious model files; llmsa attests the integrity of the broader LLM pipeline |
| [ML BOM](https://cyclonedx.org/capabilities/mlbom/) | Machine learning bill of materials | Complementary: ML BOM inventories components; llmsa cryptographically attests their integrity and enforces policy |

None of these projects provide typed attestations for LLM-specific artefacts with signed provenance chains and deployment-time admission enforcement. llmsa occupies a unique position at the intersection of software supply-chain security and AI operations.

---

## 7. Supporting Materials Checklist

- [ ] GitHub repository: <https://github.com/ogulcanaydogan/LLM-Supply-Chain-Attestation>
- [ ] Apache-2.0 LICENSE file in repository root
- [ ] README with architecture diagrams, quick start, and CLI reference
- [ ] SECURITY.md with vulnerability reporting policy and response timelines
- [ ] CONTRIBUTING.md with development setup, testing, and PR guidelines
- [ ] GOVERNANCE.md with maintainer list and decision-making process
- [ ] Threat model document: `docs/threat-model.md`
- [ ] 7 Architecture Decision Records in `docs/adr/`
- [ ] API reference: `docs/api-reference.md`
- [ ] Policy guide: `docs/policy-guide.md`
- [ ] Kubernetes admission guide: `docs/k8s-admission.md`
- [ ] Benchmark methodology: `docs/benchmark-methodology.md`
- [ ] CHANGELOG.md with full release history
- [ ] 9 CI/CD workflows in `.github/workflows/`
- [ ] 20-case tamper detection test suite: `scripts/tamper-tests.sh`
- [ ] End-to-end integration tests: `test/e2e/`
- [ ] Working example: `examples/tiny-rag/`
- [ ] Helm chart: `deploy/helm/`
- [ ] Kubernetes manifests: `deploy/webhook/`
- [ ] Dockerfile with multi-stage distroless build

---

## 8. Submission Steps

1. **Navigate** to <https://nlnet.nl/propose/>.
2. **Select** the NGI Assure fund from the available calls.
3. **Fill in** applicant details (name, email, country, organisation if applicable).
4. **Project name:** llmsa -- LLM Supply Chain Attestation
5. **Project URL:** `https://github.com/ogulcanaydogan/LLM-Supply-Chain-Attestation`
6. **Requested amount:** EUR 40,000
7. **Abstract:** Paste Section 1 above (200 words).
8. **Description:** Paste Section 2 above, adapted to the form's field length limits. If the form has separate fields for "what exists" and "what the grant funds," split Section 2.3 and 2.4 accordingly.
9. **Budget and milestones:** Paste Sections 3 and 4.
10. **Relevance to NGI:** Paste Section 5 (trustworthiness, openness, resilience).
11. **Comparable efforts:** Paste Section 6.
12. **Attach or link** supporting materials from the checklist in Section 7.
13. **Review** the full application for consistency and completeness.
14. **Submit** before 2026-04-01.

---

## 9. Post-Submission

- Monitor email for NLnet review questions (typically 2--4 weeks after submission)
- Be prepared to provide a short video demo of the attestation pipeline
- NLnet may request a revised budget or milestone plan during review
- If accepted, NLnet assigns a mentor and funds are disbursed per milestone completion
