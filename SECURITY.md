# Security Policy

This document describes how to report vulnerabilities, the scope of security issues we handle, and our response process.

## Supported Versions

| Version | Supported |
|---------|-----------|
| 0.3.x   | Yes       |
| 0.2.x   | Security fixes only |
| < 0.2   | No        |

## Scope

### Critical (fix within 14 days)

- Signature verification bypasses that allow tampered statements to pass.
- DSSE envelope forgery or payload substitution.
- Private key leakage through logs, error messages, or bundle metadata.
- Encrypted payload decryption without the correct age recipient key.

### High (fix within 30 days)

- Digest mismatch bypasses that allow modified artifacts to verify.
- Policy gate bypasses in YAML or Rego engines.
- Provenance chain validation bypasses (DAG cycle injection, missing link acceptance).
- OCI registry bundle substitution or tag confusion attacks.

### Medium (fix within 60 days)

- Schema validation bypasses that accept malformed statements.
- Information disclosure through verbose error messages.
- Denial of service through crafted attestation payloads.
- Race conditions in concurrent verification.

### Out of Scope

- Vulnerabilities in upstream dependencies (report to the upstream project).
- Social engineering attacks against maintainers.
- Denial of service through excessive API calls to OCI registries.
- Issues in example configurations or demo scripts.

## Reporting a Vulnerability

**Do not open a public GitHub issue for security vulnerabilities.**

1. Go to the [Security Advisories](https://github.com/ogulcanaydogan/LLM-Supply-Chain-Attestation/security/advisories) tab.
2. Click **"Report a vulnerability"**.
3. Provide a clear description including:
   - Affected component (signing, verification, policy, store).
   - Steps to reproduce.
   - Impact assessment.
   - Suggested fix if available.

Alternatively, email security reports to the maintainer listed in [GOVERNANCE.md](GOVERNANCE.md).

## Response Timeline

| Stage | Timeline |
|-------|----------|
| Acknowledgement | Within 48 hours |
| Triage and severity assessment | Within 7 days |
| Fix for Critical issues | Within 14 days |
| Fix for High issues | Within 30 days |
| Fix for Medium issues | Within 60 days |
| Public disclosure | After fix is released |

We follow coordinated disclosure. Security fixes are released as patch versions with a security advisory published on GitHub.

## Security Best Practices for Users

- **Rotate signing keys regularly.** Generate new PEM keys and re-attest artifacts on a schedule.
- **Use Sigstore keyless signing in CI/CD.** Binds attestations to verifiable OIDC identities instead of long-lived keys.
- **Pin OCI digests, not tags.** Tags are mutable; always reference bundles by `sha256:` digest.
- **Enable encrypted payloads** (`--privacy encrypted_payload`) for attestations containing sensitive prompts or evaluation data.
- **Run the tamper detection suite** (`scripts/tamper-tests.sh`) after any upgrade to verify verification integrity.
- **Apply policy gates** using `llmsa policy eval` to enforce organisational requirements before deployment.
- **Review provenance chains** with `llmsa verify chain` to ensure all dependencies are attested.
