# Kubernetes Admission (Planned for v1.0)

MVP does not include the validating webhook. This document reserves contract expectations:
- Verify deployment image digest against required attestation set.
- Enforce policy bundle before admission allow.
- Fail closed on verifier errors.
