---
name: code-review
description: "Review a specified Hexta diff, revision or bounded source area for actionable correctness, architecture, security and test gaps; does not merge or publish reviews."
---

# Code review

Required context: review target and baseline (diff/revision or explicit source scope). A clean working tree does not prevent source review. Establish the comparison from the user's request; do not assume HEAD~1 is the intended baseline.

1. Identify changed behavior and actual consumers. Select applicable domain/rule modules through [the router](../../../rules/README.md); keep proposed policy separate from approved requirements.
2. Trace risky paths: authorization/tenant isolation, validation, errors/secrets, transactions/concurrency/replay, cancellation/recovery and contract compatibility as applicable. Inspect implementation and neighboring tests, not just names or checklist keywords.
3. Run justified scoped checks within authorization; label missing prerequisites and unexamined scope. Instruction reviews inspect loading/adapter behavior as well as text.
4. Report actionable findings with severity, source locations, concrete trigger/impact and evidence. Separate pre-existing issues from introduced/worsened behavior and distinguish confirmed findings from questions. Do not invent praise or findings to fill a template.

Output: findings first, followed by review scope, evidence and residual gaps; state when no actionable findings were found. Save a distinct review document only when requested/useful and writes are allowed. Review does not authorize fixes, remote comments, approvals, rebases or merges; operational PR work uses its separate authorized workflow.
