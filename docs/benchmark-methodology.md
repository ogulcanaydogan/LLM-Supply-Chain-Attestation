# Benchmark Methodology

This document describes what the current benchmark and tamper harnesses actually measure in this repository.

## Scripts

- Performance/reproducibility: `/Users/ogulcanaydogan/Desktop/Projects/YaPAY/ LLM Supply-Chain Attestation/scripts/benchmark.sh`
- Tamper suite (20 seeded cases): `/Users/ogulcanaydogan/Desktop/Projects/YaPAY/ LLM Supply-Chain Attestation/scripts/tamper-tests.sh`

## Measured Benchmark Dimensions

### 1. Signing overhead

`scripts/benchmark.sh` measures total wall-clock signing duration for workload sizes:

- 1 statement
- 10 statements
- 100 statements

Metric emitted: `sign_total` duration in milliseconds.

### 2. Verification overhead

The same workloads are verified with `llmsa verify` against local bundles.

Metric emitted: `verify_total` duration in milliseconds.

### 3. Policy evaluation cost at scale

`llmsa gate` is executed 100 times per iteration.

Metric emitted: `policy_total` duration in milliseconds for 100 evaluations.

### 4. Reproducibility stability

Prompt attestation generation is repeated and normalized hashes are compared (excluding runtime nonce fields).

Metrics emitted:

- `stable` (boolean)
- `stability_rate` (0.0/1.0)

## Tamper Suite Scope (20 Cases)

`scripts/tamper-tests.sh` executes 20 deterministic mutation cases:

- 10 subject byte mutations (expected `exit 12`)
- 4 signature/bundle mutations (expected `exit 11`)
- 3 schema mutations with valid re-signing (expected `exit 14`)
- 3 provenance-chain integrity mutations (expected `exit 14`)

Each case records expected vs. actual exit code and fails the run if any mismatch occurs.

## Output Artifacts

### Benchmark artifacts

Per run, outputs are written under:

`/Users/ogulcanaydogan/Desktop/Projects/YaPAY/ LLM Supply-Chain Attestation/.llmsa/benchmarks/<timestamp>/`

Key files:

- `raw/timings.csv`
- `raw/reproducibility.json`
- `raw/run-manifest.json`
- `summary.md`

A monthly summary is also written to:

`/Users/ogulcanaydogan/Desktop/Projects/YaPAY/ LLM Supply-Chain Attestation/docs/benchmarks/YYYY-MM.md`

### Tamper artifacts

Outputs are written under:

`/Users/ogulcanaydogan/Desktop/Projects/YaPAY/ LLM Supply-Chain Attestation/.llmsa/tamper/`

Key files:

- `results.csv`
- `results.json`
- `results.md`

## Reproducibility Manifest Requirements

`run-manifest.json` includes:

- generation timestamp (UTC)
- git SHA
- working tree status
- Go version
- platform and CPU architecture
- Python version

## CI Integration

Nightly workflow:

`/Users/ogulcanaydogan/Desktop/Projects/YaPAY/ LLM Supply-Chain Attestation/.github/workflows/nightly-benchmark.yml`

It executes both scripts and uploads:

- `.llmsa/benchmarks/**`
- `.llmsa/tamper/**`
- `docs/benchmarks/*.md`

## Interpretation Guidance

Use trend comparison across nightly runs, not single-run score claims.

Red flags:

- any tamper case not matching expected exit code
- reproducibility `stable=false`
- abrupt timing regressions in `sign_total`, `verify_total`, or `policy_total`
