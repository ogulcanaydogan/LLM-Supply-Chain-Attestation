# OpenSSF Best Practices Badge Guide

## Application for llmsa: LLM Supply Chain Attestation

| Field | Detail |
|-------|--------|
| **Programme** | OpenSSF (formerly CII) Best Practices Badge |
| **URL** | <https://www.bestpractices.dev/en> |
| **Badge levels** | Passing, Silver, Gold |
| **Target level** | Passing (initial), Silver (stretch) |
| **Project** | llmsa: LLM Supply Chain Attestation |
| **Repository** | <https://github.com/ogulcanaydogan/LLM-Supply-Chain-Attestation> |
| **License** | Apache-2.0 |

---

## 1. Pre-Submission Checklist

This checklist maps the OpenSSF Best Practices criteria to what the llmsa repository already satisfies and what may need attention. Criteria are grouped by the badge programme's categories.

### 1.1 Basics

| Criterion | Status | Evidence |
|-----------|--------|----------|
| Project website or README | SATISFIED | `README.md` with architecture diagrams, quick start, CLI reference, and documentation links |
| Project description | SATISFIED | One-paragraph description at top of README; "One-Message Positioning" section |
| Interaction mechanisms (issue tracker, discussions) | SATISFIED | GitHub Issues enabled; GitHub Discussions referenced in GOVERNANCE.md |
| Contribution guide | SATISFIED | `CONTRIBUTING.md` with prerequisites, setup, test commands, code style, PR guidelines, release process |
| License (OSI-approved, in standard location) | SATISFIED | Apache-2.0 in `LICENSE` file at repository root; badge in README |
| License in each source file (recommended) | CHECK | Go source files may not have individual license headers; add SPDX headers if required |
| Documentation of project governance | SATISFIED | `GOVERNANCE.md` with maintainer list, decision-making process, release cadence, code of conduct |

### 1.2 Change Control

| Criterion | Status | Evidence |
|-----------|--------|----------|
| Public version control (e.g., git) | SATISFIED | GitHub repository, public |
| Unique version numbering | SATISFIED | Semantic versioning; 7 tagged releases (v0.1.0 through v1.0.1) |
| Release notes | SATISFIED | `CHANGELOG.md` following Keep a Changelog format with Added/Changed/Fixed sections per release |
| Version-controlled build/install instructions | SATISFIED | `Makefile`, `go.mod`, `Dockerfile`, GoReleaser config |

### 1.3 Reporting

| Criterion | Status | Evidence |
|-----------|--------|----------|
| Bug reporting process | SATISFIED | GitHub Issues for bugs; CONTRIBUTING.md describes the process |
| Vulnerability reporting process | SATISFIED | `SECURITY.md` with private reporting via GitHub Security Advisories, scope definitions (Critical/High/Medium/Out-of-scope), response timelines |
| Vulnerability response timeline | SATISFIED | SECURITY.md: Critical 14 days, High 30 days, Medium 60 days; acknowledgement within 48 hours |

### 1.4 Quality

| Criterion | Status | Evidence |
|-----------|--------|----------|
| Working build system | SATISFIED | `go build ./cmd/llmsa`, `make build`, Dockerfile, GoReleaser |
| Automated test suite | SATISFIED | `go test ./...`; unit tests across all packages; E2E tests in `test/e2e/` |
| Test coverage | SATISFIED | verify 86.9%, attest 83.6%, store 84.5%, sign 77.3%, rego 94.3%, schema 100%, types 100% |
| New functionality has tests | SATISFIED | CONTRIBUTING.md requires tests for new features; coverage targets per package |
| Test policy documented | SATISFIED | CONTRIBUTING.md documents test commands, coverage targets, tamper suite, benchmark suite, E2E instructions |
| CI/CD passes on each commit | SATISFIED | 9 GitHub Actions workflows; `ci-attest-verify.yml` runs on every push/PR |

### 1.5 Security

| Criterion | Status | Evidence |
|-----------|--------|----------|
| Secure development knowledge | SATISFIED | Threat model in `docs/threat-model.md` with STRIDE analysis; SECURITY.md; 7 ADRs documenting security-relevant design decisions |
| Use good cryptographic practices | SATISFIED | SHA-256 for digests, Ed25519 for PEM signing, Sigstore for keyless signing, age/X25519 for encryption; no custom crypto |
| No unpatched known vulnerabilities | CHECK | Run `govulncheck ./...` and confirm no known vulnerabilities in dependencies |
| Secure defaults | SATISFIED | `hash_only` privacy mode default (no payload leakage); fail-closed webhook default; OIDC issuer verification enabled by default |
| Hardened delivery mechanism | SATISFIED | OCI digest-pinned bundles; content-addressable storage; GoReleaser multi-platform builds with checksums |
| Memory-safe language | SATISFIED | Written in Go (memory-safe, garbage-collected) |

### 1.6 Analysis

| Criterion | Status | Evidence |
|-----------|--------|----------|
| Static analysis | CHECK | Add `go vet ./...` and `staticcheck ./...` to CI if not already present |
| Dynamic analysis or fuzzing | CHECK | Consider adding Go fuzzing tests for JSON parsing, schema validation, and DSSE deserialization |
| Compiler warnings addressed | SATISFIED | Go compiler warnings are errors by default; `go vet` in CI |

### 1.7 Summary

| Category | Satisfied | Needs Work |
|----------|-----------|------------|
| Basics | 6/7 | License headers in source files (recommended, not required) |
| Change Control | 4/4 | None |
| Reporting | 3/3 | None |
| Quality | 6/6 | None |
| Security | 5/6 | Run govulncheck to confirm no known vulnerabilities |
| Analysis | 1/3 | Add static analysis to CI; consider fuzzing |
| **Total** | **25/29** | **4 items to address** |

**Assessment:** The project already satisfies the vast majority of Passing-level criteria. The remaining items (license headers, govulncheck, static analysis in CI, fuzzing) are straightforward additions that can be completed in 1 to 2 days.

---

## 2. Step-by-Step Form Filling Guide

### Step 1: Create an Account

1. Navigate to <https://www.bestpractices.dev/en>.
2. Click "Get Your Badge Now!" or "Log in" in the top navigation.
3. Sign in with your GitHub account (`ogulcanaydogan`).
4. Authorise the OpenSSF Best Practices app to access your public repositories.

### Step 2: Add Your Project

1. After logging in, click "Get Your Badge Now!" or navigate to <https://www.bestpractices.dev/en/projects/new>.
2. Enter the repository URL: `https://github.com/ogulcanaydogan/LLM-Supply-Chain-Attestation`
3. The form will auto-populate some fields from the GitHub API.
4. Click "Submit URL" to create the project entry.

### Step 3: Fill in Basic Information

| Form Field | Value |
|------------|-------|
| **Project name** | llmsa: LLM Supply Chain Attestation |
| **Project description** | A cryptographic attestation framework that brings software supply-chain security to large language model lifecycles with typed LLM artefact provenance, policy enforcement, and deployment-time admission checks. |
| **Project home page URL** | `https://github.com/ogulcanaydogan/LLM-Supply-Chain-Attestation` |
| **Repository URL** | `https://github.com/ogulcanaydogan/LLM-Supply-Chain-Attestation` |
| **License** | Apache-2.0 |
| **Primary programming language** | Go |
| **Project governance documentation URL** | `https://github.com/ogulcanaydogan/LLM-Supply-Chain-Attestation/blob/main/GOVERNANCE.md` |

### Step 4: Fill in Each Criteria Section

The form is divided into sections matching the badge categories. Below is how to answer each question with references to specific repository files.

#### 4.1 Basics

**Q: Does the project have a website or README?**
- Answer: Met
- Justification: `README.md` in repository root contains project description, architecture diagrams, quick start guide, CLI reference, and links to all documentation.
- URL: `https://github.com/ogulcanaydogan/LLM-Supply-Chain-Attestation/blob/main/README.md`

**Q: Does the project have a description of what it does?**
- Answer: Met
- Justification: README opens with a one-sentence description and includes a "One-Message Positioning" section, "The Problem" section, and "What llmsa Does" section with a feature comparison table.

**Q: Does the project have an interaction mechanism (issue tracker)?**
- Answer: Met
- Justification: GitHub Issues are enabled. GOVERNANCE.md references GitHub Issues, Discussions, Pull Requests, and Security Advisories as communication channels.
- URL: `https://github.com/ogulcanaydogan/LLM-Supply-Chain-Attestation/issues`

**Q: Is there a contribution guide?**
- Answer: Met
- Justification: `CONTRIBUTING.md` covers prerequisites, development setup, running tests (unit, tamper, benchmark, E2E), code style, package organisation, adding new attestation types, ADR process, PR guidelines, and release process.
- URL: `https://github.com/ogulcanaydogan/LLM-Supply-Chain-Attestation/blob/main/CONTRIBUTING.md`

**Q: Does the project have an OSI-approved license?**
- Answer: Met
- Justification: Apache-2.0 license file at repository root (`LICENSE`). Badge displayed in README.
- URL: `https://github.com/ogulcanaydogan/LLM-Supply-Chain-Attestation/blob/main/LICENSE`

**Q: Is the project governance documented?**
- Answer: Met
- Justification: `GOVERNANCE.md` documents maintainer list, becoming a maintainer, decision-making process (day-to-day and major decisions), release cadence, code of conduct, and contribution model.
- URL: `https://github.com/ogulcanaydogan/LLM-Supply-Chain-Attestation/blob/main/GOVERNANCE.md`

#### 4.2 Change Control

**Q: Does the project use public version control?**
- Answer: Met
- Justification: Git repository hosted on GitHub, public.

**Q: Does the project have unique version numbering?**
- Answer: Met
- Justification: Semantic versioning (v0.1.0, v0.2.0, v0.3.0, v1.0.0, v1.0.1). Each release has a corresponding git tag and CHANGELOG entry.

**Q: Does the project have release notes?**
- Answer: Met
- Justification: `CHANGELOG.md` follows Keep a Changelog format with version sections, dates, and Added/Changed/Fixed subsections.
- URL: `https://github.com/ogulcanaydogan/LLM-Supply-Chain-Attestation/blob/main/CHANGELOG.md`

#### 4.3 Reporting

**Q: Does the project have a process for reporting bugs?**
- Answer: Met
- Justification: GitHub Issues for bug reports. CONTRIBUTING.md describes the process.

**Q: Does the project have a vulnerability reporting process?**
- Answer: Met
- Justification: `SECURITY.md` provides detailed vulnerability reporting instructions (GitHub Security Advisories), scope definitions with severity levels (Critical/High/Medium/Out-of-scope), and response timelines.
- URL: `https://github.com/ogulcanaydogan/LLM-Supply-Chain-Attestation/blob/main/SECURITY.md`

**Q: Is there a vulnerability response timeline?**
- Answer: Met
- Justification: SECURITY.md defines: Acknowledgement within 48 hours, triage within 7 days, Critical fix within 14 days, High fix within 30 days, Medium fix within 60 days.

#### 4.4 Quality

**Q: Does the project have a working build system?**
- Answer: Met
- Justification: `go build -o llmsa ./cmd/llmsa` builds the binary. `Makefile` provides `make build`, `make test`, `make demo` targets. `Dockerfile` provides containerised build. GoReleaser config for multi-platform release builds.

**Q: Does the project have an automated test suite?**
- Answer: Met
- Justification: `go test ./...` runs unit tests across all packages. `test/e2e/` contains integration tests. `scripts/tamper-tests.sh` provides 20-case security validation. `scripts/benchmark.sh` provides performance benchmarks.

**Q: Are test results publicly available?**
- Answer: Met
- Justification: CI workflow `ci-attest-verify.yml` runs tests on every push and pull request. Results are publicly visible on the GitHub Actions tab.
- URL: `https://github.com/ogulcanaydogan/LLM-Supply-Chain-Attestation/actions/workflows/ci-attest-verify.yml`

**Q: Is new functionality tested?**
- Answer: Met
- Justification: CONTRIBUTING.md requires tests for all new features and collectors. Coverage targets are defined per package (verify >= 85%, sign >= 75%, attest >= 80%, store >= 80%, policy >= 90%, schema 100%, types 100%).

**Q: Is the test policy documented?**
- Answer: Met
- Justification: CONTRIBUTING.md documents test commands (`make test`, `go test`, E2E, tamper suite, benchmarks) and coverage targets.

#### 4.5 Security

**Q: Does at least one developer know how to design secure software?**
- Answer: Met
- Justification: The project includes a formal threat model (`docs/threat-model.md`) with STRIDE analysis covering 8 threat scenarios and mitigations. 7 Architecture Decision Records document security-relevant design decisions. The 20-case tamper detection suite demonstrates adversarial testing methodology.

**Q: Does the project use good cryptographic practices?**
- Answer: Met
- Justification: Uses established cryptographic libraries and standards: SHA-256 (NIST FIPS 180-4) for digests, Ed25519 for PEM signing, Sigstore for keyless signing with OIDC identity binding, age (X25519 + ChaCha20-Poly1305) for encryption. No custom cryptographic implementations.

**Q: Does the project have secure defaults?**
- Answer: Met
- Justification: Default privacy mode is `hash_only` (no payload stored). Kubernetes webhook defaults to `failurePolicy: Fail` (fail-closed). Sigstore verification checks OIDC issuer by default. Plaintext payload mode requires explicit policy allowlisting.

**Q: Does the project address known vulnerabilities?**
- Answer: Met (verify with govulncheck)
- Justification: Dependencies are maintained at current versions (see `go.mod`). Run `govulncheck ./...` to confirm no known vulnerabilities. To add as a CI step:
  ```yaml
  - name: Check for known vulnerabilities
    run: |
      go install golang.org/x/vuln/cmd/govulncheck@latest
      govulncheck ./...
  ```

**Q: Does the project use a memory-safe language?**
- Answer: Met
- Justification: Written entirely in Go, a memory-safe garbage-collected language.

#### 4.6 Analysis

**Q: Does the project use static analysis?**
- Answer: Met (ensure in CI)
- Justification: CONTRIBUTING.md requires `go vet ./...` before PR submission. To satisfy this criterion fully, add static analysis as an explicit CI step:
  ```yaml
  - name: Static analysis
    run: |
      go vet ./...
      go install honnef.co/go/tools/cmd/staticcheck@latest
      staticcheck ./...
  ```

**Q: Does the project use dynamic analysis or fuzzing?**
- Answer: Partially met
- Justification: The 20-case tamper detection suite provides adversarial dynamic testing. E2E tests exercise the full pipeline. For full compliance, consider adding Go fuzz tests:
  ```go
  func FuzzVerifyBundle(f *testing.F) {
      f.Add([]byte(`{"payloadType":"...","payload":"...","signatures":[...]}`))
      f.Fuzz(func(t *testing.T, data []byte) {
          // Attempt to verify malformed bundles
      })
  }
  ```

### Step 5: Review and Submit

1. After filling in all sections, the form will show a percentage score and highlight any unmet criteria.
2. Review each "Not met" or "Unknown" item and add justification or mark as "N/A" where applicable.
3. For criteria that require repository changes (govulncheck, static analysis CI step), make those changes first, then update the form.
4. Click "Submit" when all Passing-level criteria show "Met."
5. The badge URL will be generated immediately and can be added to the README.

### Step 6: Add the Badge to README

After the badge is granted, add it to the README badge row:

```markdown
[![OpenSSF Best Practices](https://www.bestpractices.dev/projects/XXXXX/badge)](https://www.bestpractices.dev/projects/XXXXX)
```

Replace `XXXXX` with the project number assigned by the badge programme.

---

## 3. Recommended Pre-Submission Changes

Before submitting the badge application, make these changes to ensure all Passing criteria are met on first attempt:

### 3.1 Add govulncheck to CI

Add a vulnerability check step to `ci-attest-verify.yml`:

```yaml
- name: Check for known vulnerabilities
  run: |
    go install golang.org/x/vuln/cmd/govulncheck@latest
    govulncheck ./...
```

### 3.2 Add Static Analysis to CI

Add explicit static analysis steps to `ci-attest-verify.yml`:

```yaml
- name: Static analysis
  run: |
    go vet ./...
    go install honnef.co/go/tools/cmd/staticcheck@latest
    staticcheck ./...
```

### 3.3 Add SPDX License Headers (Recommended)

Add SPDX license headers to Go source files. This is recommended but not strictly required for the Passing level:

```go
// SPDX-License-Identifier: Apache-2.0
// Copyright 2025-2026 Ogulcan Aydogan
```

A script can automate this across all `.go` files:

```bash
for f in $(find . -name '*.go' -not -path './vendor/*'); do
  if ! head -1 "$f" | grep -q 'SPDX'; then
    sed -i '' '1i\
// SPDX-License-Identifier: Apache-2.0\
// Copyright 2025-2026 Ogulcan Aydogan\
' "$f"
  fi
done
```

### 3.4 Add Go Fuzz Tests (Recommended for Silver)

Add fuzz tests for security-critical parsing paths:

- `internal/verify/signature_verify_test.go`: fuzz DSSE envelope parsing
- `internal/verify/schema_verify_test.go`: fuzz JSON Schema validation with malformed input
- `pkg/schema/loader_test.go`: fuzz schema loading with corrupted files
- `internal/sign/dsse_bundle_test.go`: fuzz bundle deserialization

### 3.5 Add OpenSSF Scorecard Workflow

The repository already has `scorecard.yml` in `.github/workflows/`. Verify it's running and producing results visible at <https://securityscorecards.dev/>. The Scorecard score can be referenced in the badge application as additional evidence.

---

## 4. Path to Silver Badge

After achieving the Passing badge, the following additional work would be needed for Silver:

| Silver Criterion | Current Status | Work Needed |
|-----------------|----------------|-------------|
| Code review policy (all changes reviewed) | Branch protection likely configured | Verify branch protection requires PR review |
| Signed releases | GoReleaser produces checksums | Add Sigstore signing of release artefacts |
| Reproducible builds | Go builds are largely reproducible | Verify with `go build -trimpath` and document |
| Dynamic analysis / fuzzing | Tamper suite exists | Add Go fuzz tests for parsing paths |
| Cryptographic agility | Uses SHA-256 and Ed25519 | Document algorithm migration path |

---

## 5. Quick Reference: Repository Files for Badge Evidence

| Badge Criterion | Repository File |
|----------------|-----------------|
| Project description | `README.md` |
| License | `LICENSE` |
| Contribution guide | `CONTRIBUTING.md` |
| Governance | `GOVERNANCE.md` |
| Security policy | `SECURITY.md` |
| Changelog | `CHANGELOG.md` |
| Threat model | `docs/threat-model.md` |
| Architecture decisions | `docs/adr/` (7 ADRs) |
| CI pipeline | `.github/workflows/ci-attest-verify.yml` |
| Scorecard | `.github/workflows/scorecard.yml` |
| Tamper tests | `scripts/tamper-tests.sh` |
| E2E tests | `test/e2e/` |
| Benchmarks | `scripts/benchmark.sh` |
| Example project | `examples/tiny-rag/` |
| K8s deployment | `deploy/webhook/`, `deploy/helm/` |
| Dockerfile | `Dockerfile` |

---

## 6. Estimated Timeline

| Task | Effort | Dependency |
|------|--------|------------|
| Add govulncheck to CI | 30 minutes | None |
| Add static analysis to CI | 30 minutes | None |
| Add SPDX headers to source files | 1 hour | None |
| Fill in badge form | 2 hours | Above changes merged |
| Verify all criteria show "Met" | 30 minutes | Form filled |
| Add badge to README | 5 minutes | Badge granted |
| **Total** | **~4.5 hours** | |

---

## 7. Links

| Resource | URL |
|----------|-----|
| OpenSSF Best Practices Badge Programme | <https://www.bestpractices.dev/en> |
| Badge criteria (Passing) | <https://www.bestpractices.dev/en/criteria/0> |
| Badge criteria (Silver) | <https://www.bestpractices.dev/en/criteria/1> |
| Badge criteria (Gold) | <https://www.bestpractices.dev/en/criteria/2> |
| OpenSSF Scorecard | <https://securityscorecards.dev/> |
| llmsa repository | <https://github.com/ogulcanaydogan/LLM-Supply-Chain-Attestation> |
| llmsa CI status | <https://github.com/ogulcanaydogan/LLM-Supply-Chain-Attestation/actions> |
