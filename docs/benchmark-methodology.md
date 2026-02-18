# Benchmark Methodology

This document describes the benchmark methodology used to validate the correctness, determinism, and performance characteristics of the LLM Supply-Chain Attestation framework. All benchmarks are designed to be fully reproducible and produce machine-readable artifacts.

## Benchmark Categories

### 1. Determinism Validation

Attestation generation must be deterministic — running the same command with the same inputs must produce identical statement hashes. Non-determinism would undermine the integrity of the verification pipeline, as re-generated attestations would not match previously signed bundles.

**Methodology**: The `--determinism-check N` flag runs attestation generation N times and compares the canonical JSON hashes of each output. If any pair differs, the check fails with a descriptive error identifying the divergent fields.

```bash
go run ./cmd/llmsa attest create \
  --type prompt_attestation \
  --config examples/tiny-rag/configs/prompt.yaml \
  --out /tmp/determinism \
  --determinism-check 5
```

**Expected result**: All 5 runs produce identical SHA-256 hashes. Known sources of non-determinism (timestamps, UUIDs) are excluded from the comparison by using canonical JSON encoding.

### 2. Tamper Detection

The verification engine must detect modifications to any subject file referenced in an attestation statement. This validates the core security property of the framework.

**Methodology**: The tamper detection suite (`scripts/tamper-tests.sh`) runs 20 test cases across all five attestation types:

1. Generate a valid attestation and sign it.
2. Modify a subject file (append bytes, truncate, replace content, zero-fill).
3. Run verification and assert exit code 12 (digest mismatch).
4. Restore the original file.

```bash
./scripts/tamper-tests.sh
```

**Expected result**: All 20 cases produce exit code 12. Zero false negatives (tampered files passing verification) and zero false positives (unmodified files failing verification).

### 3. Signature Verification

Signature spoofing attempts must be detected. The benchmark validates both positive cases (valid signatures pass) and negative cases (corrupted, truncated, or replaced signatures fail).

**Methodology**:
1. Sign a bundle with a known PEM key.
2. Verify the bundle passes (exit code 0).
3. Corrupt the `sig` field in the bundle JSON.
4. Verify the corrupted bundle fails (exit code 11).
5. Replace the public key with a different key.
6. Verify the mismatched key fails (exit code 11).

### 4. Policy Gate Enforcement

Policy gates must correctly block deployments when required attestation types are missing for changed files.

**Methodology**:
1. Configure a policy with gates for all five attestation types.
2. Simulate file changes that trigger each gate.
3. Run the gate engine with complete attestations — assert zero violations.
4. Remove one attestation type — assert the corresponding gate produces a violation.
5. Test privacy mode enforcement — assert `plaintext_explicit` is blocked without allowlisting.

### 5. Provenance Chain Integrity

The dependency graph between attestation types must be validated correctly.

**Methodology**:
1. Generate all five attestation types with correct temporal ordering.
2. Verify the chain report shows all four edges satisfied (eval→prompt, eval→corpus, route→eval, slo→route).
3. Remove a predecessor attestation and verify the chain reports a violation.
4. Alter timestamps to violate temporal ordering and verify detection.

### 6. Performance Benchmarks

Pipeline performance is measured to ensure the framework adds minimal overhead to CI/CD pipelines.

**Methodology**: The benchmark suite (`scripts/benchmark.sh`) measures wall-clock time for each pipeline stage:

| Stage | Operation | Target |
|-------|-----------|--------|
| Attest | Generate all 5 attestation types | < 500ms |
| Sign | PEM-sign all 5 bundles | < 200ms |
| Verify | Four-stage verification pipeline | < 1s |
| Gate | Policy evaluation (YAML engine) | < 100ms |
| Gate | Policy evaluation (Rego engine) | < 200ms |
| Report | Markdown report generation | < 100ms |
| Publish | OCI registry push (per bundle) | Network-dependent |

```bash
./scripts/benchmark.sh
```

## Reproducibility Requirements

All benchmark runs must capture:

1. **Environment metadata**: Go version, OS, architecture, CPU cores, available memory.
2. **Git provenance**: Commit SHA, branch, and dirty state.
3. **Raw results**: JSON or CSV output with individual timing data.
4. **Markdown summary**: Human-readable report with key metrics and observations.

These artifacts are stored in the CI pipeline as build artifacts for auditability.

## E2E Integration Tests

The Go-based E2E test suite (`test/e2e/`) serves as an automated benchmark that runs in CI:

| Test | What It Validates |
|------|-------------------|
| `TestFullPipeline_AttestSignVerify` | Complete 5-type pipeline produces exit code 0 |
| `TestFullPipeline_TamperDetection` | Modified subject file produces exit code 12 |
| `TestFullPipeline_ChainVerification` | All 4 provenance chain edges are satisfied |
| `TestFullPipeline_MissingAttestation` | Empty directory produces exit code 10 |
| `TestFullPipeline_SignatureCorruption` | Corrupted signature produces exit code 11 |
| `TestFullPipeline_DeterminismCheck` | Determinism validation runs without panic |

Run the E2E suite:

```bash
go test -tags=e2e -v -timeout 120s ./test/e2e/
```

## CI/CD Benchmark Integration

The nightly benchmark workflow (`.github/workflows/nightly-benchmark.yml`) runs the full benchmark suite on a schedule, tracking performance trends over time. Results are uploaded as build artifacts and can be compared across commits to detect performance regressions.

## Interpreting Results

A healthy benchmark run produces:
- **Zero failed tamper detection cases** (20/20 pass).
- **Zero false positives** in signature verification.
- **All provenance chain edges satisfied** when all five types are present.
- **Sub-second total pipeline time** for the standard five-type attestation set.
- **Deterministic hashes** across repeated runs.

Any deviation from these baselines indicates a regression that should be investigated before release.
