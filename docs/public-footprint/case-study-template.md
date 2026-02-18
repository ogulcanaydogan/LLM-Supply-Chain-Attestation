# Anonymous Pilot Case Study Template

Use this template to publish one metrics-backed adoption story without disclosing company identity.

## 1. Context

- Industry profile (high level):
- Team size:
- Environment (GitHub Actions, Kubernetes, registry stack):
- Compliance or audit pressure (if shareable):

## 2. Problem Before `llmsa`

- What was not provable before:
- Incident or near-miss pattern:
- Why existing controls were insufficient:

## 3. Integration Steps

1. `llmsa` scope introduced (which attestation types were enabled).
2. CI wiring (`attest`, `sign`, `verify`, `gate`).
3. Registry publish/pull strategy.
4. Admission enforcement rollout mode and timeline.

## 4. Measurement Method

- Date window (UTC):
- Baseline method:
- Tooling and commands used:
- Data collection sources (workflow URLs, run artifacts, reports):

## 5. Metrics (Required)

| Metric | Before | After | Collection Method |
|---|---:|---:|---|
| Tamper detection success rate | | | |
| Verify p95 latency (N statements) | | | |
| Gate violation detection rate | | | |
| Release assurance evidence completeness | | | |

## 6. Reproducibility Commands

```bash
# include exact commands and refs used for the reported metrics
```

## 7. Failure Modes Observed

- What failed during rollout:
- What was tuned or changed:
- Residual risk that remains:

## 8. What This Case Study Does Not Prove

- It does not prove universal performance across all workloads.
- It does not prove protection against runtime compromise after admission.
- It does not replace independent security review.

## 9. Public Evidence Links

- Release:
- Workflow run(s):
- Benchmark artifact(s):
- Policy/report artifact(s):
