# What We Do Not Claim

This file defines explicit non-claims to keep public messaging technically honest.

## Security Non-Claims

1. `llmsa` does not prevent runtime prompt injection or jailbreak attacks by itself.
2. `llmsa` does not guarantee model quality; it guarantees traceability and policy-enforced evidence checks.
3. `llmsa` does not replace threat modeling, secure coding, or independent security reviews.
4. `llmsa` does not provide revocation guarantees for all historical signatures in its current default flow.

## Operations Non-Claims

1. `llmsa` does not guarantee zero-latency overhead.
2. `llmsa` benchmark results are not universal; they are workload and environment dependent.
3. Admission enforcement is deployment-time control, not full runtime integrity monitoring.

## Compliance Non-Claims

1. Using `llmsa` does not automatically satisfy regulatory certification requirements.
2. Policy packs are reference implementations, not legal or compliance advice.

## External Messaging Rule

Any public post, release note, or case study should include:

1. evidence links,
2. scope boundaries,
3. this non-claims context.
