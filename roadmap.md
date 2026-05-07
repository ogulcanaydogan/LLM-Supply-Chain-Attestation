# Roadmap

## Current: v1.0.2 (GA)

CLI and admission webhook that produce signed DSSE attestations for prompts, corpora, evaluations, routes, and SLOs — enforced at the Kubernetes API boundary via OPA.

---

## v1.1.0 — in-toto v2 + GitHub Action (Q2 2026)

- [ ] in-toto v2 statement support alongside DSSE (backwards-compatible, negotiated by flag)
- [ ] GitHub Action `ogulcanaydogan/llmsa-attest@v1` for one-line CI attestation
- [ ] E2E tamper-detection test matrix: corrupt payload, clock skew, revoked key, replay
- [ ] Helm chart updates for Kubernetes 1.32

**Target branch**: `feature/v1.1.0`

---

## v1.2.0 — Policy Authoring UX (Q3 2026)

- [ ] Rego policy library: 10 pre-built policies for common LLM supply-chain risks
- [ ] `llmsa policy lint` — static analysis for Rego attestation policies
- [ ] Attestation diff tool: compare two attestation bundles, surface divergences
- [ ] Multi-cluster federation: single webhook, multiple cluster admission targets

---

## v2.0.0 — Cross-Org Verification (Q4 2026)

- [ ] Cross-organisation attestation exchange (publish/subscribe via Sigstore transparency log)
- [ ] Attestation expiry + rotation lifecycle management
- [ ] SBOM attestation type (CycloneDX + SPDX)
- [ ] Dashboard: real-time attestation health for all enrolled pipelines

---

## Known issues / backlog

See [open issues](https://github.com/ogulcanaydogan/LLM-Supply-Chain-Attestation/issues).
