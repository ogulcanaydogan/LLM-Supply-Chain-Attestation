# Benchmark Methodology

This document defines a publication-ready benchmark method for `llmsa` that is both practical and auditable.

The primary benchmark question is:

**How much governance yield do we get per unit of overhead when attestation verification is enforced for LLM delivery pipelines?**

## 1) Core Benchmark Angle

`llmsa` should be evaluated on two dimensions together, not separately:

1. **Governance yield**
   - Can the system detect and block high-value failure modes (tamper, missing evidence, invalid signatures, broken provenance)?
2. **Performance overhead**
   - What latency/cost is introduced by signing, verifying, and policy evaluation?

A benchmark report is incomplete if it reports only speed without detection quality, or detection quality without overhead.

## 2) Workloads and Scenarios

## Standard workloads

- `1` statement
- `10` statements
- `100` statements

These are executed by `/Users/ogulcanaydogan/Desktop/Projects/YaPAY/ LLM Supply-Chain Attestation/scripts/benchmark.sh`.

## Security/fault scenarios

Use `/Users/ogulcanaydogan/Desktop/Projects/YaPAY/ LLM Supply-Chain Attestation/scripts/tamper-tests.sh`, which runs 20 seeded attacks:

- subject/material byte mutation cases,
- signature/bundle corruption cases,
- schema and provenance-chain integrity cases.

## 3) Metric Definitions

## Governance yield metrics

1. **Tamper detection recall**
   - Formula: `detected_tamper_cases / total_seeded_tamper_cases`
   - Target: `1.0` on seeded suite.
2. **Gate violation catch rate**
   - Formula: `blocked_invalid_changes / total_invalid_changes`
3. **Provenance integrity catch rate**
   - Formula: `blocked_chain_break_cases / total_chain_break_cases`
4. **Evidence completeness rate**
   - Formula: `runs_with_all_required_artifacts / total_runs`

## Overhead metrics

1. **Signing overhead**
   - `sign_total` duration in ms per workload.
2. **Verification overhead**
   - `verify_total` duration in ms per workload.
3. **Policy overhead**
   - `policy_total` duration in ms for 100 evaluations.
4. **Admission-path latency (when webhook benchmark is included)**
   - p50/p95 decision latency.

## Stability metric

1. **Determinism stability rate**
   - Formula: `stable_runs / total_repeated_runs`
   - Uses normalized comparison that excludes runtime nonce fields.

## 4) Required Output Contract

Every benchmark run must produce raw artifacts under:

`/Users/ogulcanaydogan/Desktop/Projects/YaPAY/ LLM Supply-Chain Attestation/.llmsa/benchmarks/<timestamp>/`

### Required files

1. `raw/timings.csv`
2. `raw/reproducibility.json`
3. `raw/run-manifest.json`
4. `summary.md`

Tamper run outputs must exist under:

`/Users/ogulcanaydogan/Desktop/Projects/YaPAY/ LLM Supply-Chain Attestation/.llmsa/tamper/`

with:

1. `results.csv`
2. `results.json`
3. `results.md`

## Required manifest fields

`run-manifest.json` must include at least:

- generated timestamp (UTC),
- git SHA,
- working tree status,
- Go version,
- platform and architecture,
- Python version.

## 5) Statistical Reporting Rules

For each scenario/workload:

1. Report sample count.
2. Report p50 and p95.
3. Do not publish single-number claims without sample context.
4. Treat trend over repeated runs as more meaningful than one run.

When reporting security yield:

1. Include confusion-style summary (`expected`, `observed`, `mismatch count`).
2. Include at least one failure example and one pass example.
3. Include a limitations section in every report.

## 6) Reproducibility Commands

```bash
# Benchmark and reproducibility outputs
./scripts/benchmark.sh

# Seeded tamper-evidence suite
./scripts/tamper-tests.sh
```

Optional snapshot for public-footprint tracking:

```bash
./scripts/public-footprint-snapshot.sh
```

## 7) Claims Discipline

Allowed claim style:

- "On this workload and environment, we observed X."
- "Seeded tamper suite detected Y/Z cases."

Disallowed claim style:

- "Universal performance guarantee."
- "Complete security guarantee."
- "Regulatory compliance guarantee."

## 8) Threats to Validity

1. Workloads are synthetic and may not represent all production distributions.
2. Local and CI runner environments differ from dedicated production clusters.
3. Seeded attack cases are finite and cannot model every adversarial strategy.
4. Performance can vary by registry latency and signing backend.

## 9) Publication Checklist

Before publishing a benchmark post or report:

1. Include commit SHA and run date.
2. Link raw artifacts.
3. Link scripts used to generate data.
4. Include known limitations and non-claims.
5. Preserve exact command lines used.
