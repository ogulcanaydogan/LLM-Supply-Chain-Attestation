# Benchmark Methodology

## Required outputs
- Raw JSON/CSV artifacts.
- Reproducibility manifest with commit and environment metadata.
- Markdown summary with limitation notes.

## MVP benchmark checks
- Determinism: repeated attestation generation hashes match.
- Tamper detection: modified subject must fail verification (`exit 12`).
- Policy enforcement: missing required attestation must fail gate (`exit 13`).
