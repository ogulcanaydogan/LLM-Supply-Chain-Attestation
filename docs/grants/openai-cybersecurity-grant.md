# OpenAI Cybersecurity Grant Application

| Field | Value |
|-------|-------|
| **URL** | https://openai.com/form/cybersecurity-grant-program |
| **Amount** | $1M grants pool + $10M API credits pool |
| **Deadline** | Rolling |
| **Project** | LLM-Supply-Chain-Attestation |
| **Applicant** | Ogulcan Aydogan |

---

## Project Title

LLM Supply Chain Attestation: Sigstore/SLSA Security for AI Model Artifacts

## Project Description

LLM-Supply-Chain-Attestation is an open-source Go toolkit that applies software supply chain security practices (Sigstore, in-toto, SLSA) to LLM artifacts. It generates signed attestations for model weights, training data, fine-tuning runs, and inference outputs, creating a verifiable chain of custody from training through deployment.

The project has 7 releases, 8 CI workflows, and supports keyless signing via Sigstore Fulcio, SLSA provenance generation, in-toto layout verification, and SBOM creation for ML pipelines. It integrates with container registries (OCI), model hubs, and CI/CD systems.

## How does this project advance AI cybersecurity?

1. AI model supply chains have no standard integrity verification. Organizations download model weights from HuggingFace, fine-tune them, and deploy to production with no cryptographic proof that the weights haven't been tampered with. This toolkit brings the same supply chain security that Sigstore provides for software packages to AI model artifacts.

2. SLSA provenance attestations create verifiable records of how a model was built: which base model, which training data, which fine-tuning script, which hardware. If a model produces unexpected outputs, teams can trace back to the exact training pipeline and verify nothing was modified.

3. The toolkit integrates with OpenAI's fine-tuning workflow. When you fine-tune an OpenAI model, the attestation toolkit can generate signed records of the training job parameters, dataset hash, and resulting model ID, creating an audit trail for your custom models.

4. In-toto layout verification ensures that multi-step ML pipelines (data preparation, training, evaluation, deployment) followed the prescribed workflow. Any skipped step or unauthorized modification breaks the verification chain.

5. SBOM (Software Bill of Materials) generation for ML pipelines documents every dependency, from the base model version to the training framework version to the evaluation metrics library. This is becoming a regulatory requirement (US Executive Order 14028, EU Cyber Resilience Act).

## Specific use of funding

**OpenAI fine-tuning integration ($15,000-20,000):**
- Build native attestation support for OpenAI fine-tuning jobs
- Sign and verify fine-tuning parameters, dataset hashes, and model IDs
- Create a transparent log of all fine-tuning operations per organization
- Document the complete attestation workflow for OpenAI API users

**Attestation verification tooling ($10,000-15,000):**
- CLI and SDK tools for verifying model attestations before deployment
- Policy engine that enforces attestation requirements (e.g., "don't deploy models without signed provenance")
- Integration with Kubernetes admission controllers for model deployment gates

**Security audit ($10,000):**
- Independent review of the signing pipeline (Sigstore keyless, DSSE envelope format)
- Verification of SLSA provenance generation against the specification
- Audit of the in-toto layout verification logic

**API credits ($10,000):**
- Test attestation generation across OpenAI's fine-tuning API
- Benchmark signing overhead for high-volume inference logging
- Generate test datasets for the public benchmark

## Team

Ogulcan Aydogan, sole developer and maintainer. Background in software engineering, supply chain security, and AI infrastructure. Built the complete toolkit including Sigstore integration, SLSA provenance generation, in-toto verification, and CI pipeline. Based in the United Kingdom.

## Repository

https://github.com/ogulcanaydogan/LLM-Supply-Chain-Attestation

## Requested Amount

$30,000 grant + $10,000 API credits

---

## Submission Steps

1. Go to https://openai.com/form/cybersecurity-grant-program
2. Fill in form fields using responses above
3. Submit (rolling deadline)
4. Record confirmation and update tracker

---

*Last updated: 2026-03-14*
