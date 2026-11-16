---
name: code-review
description: "Review a specified Hexta diff, revision or bounded source area for actionable correctness, architecture, security and test gaps; does not merge or publish reviews."
---

# Code Review

Evidence-based review of correctness, security, reliability and behavioral compatibility. Prioritize actionable defects over style.

## 1. Establish Review Scope

Required context: review target and baseline (diff, revision, or explicit source scope).

- Determine the intended comparison from the user's request. Never assume `HEAD~1` is the baseline.
- A clean working tree does not prevent source review.
- If the baseline is ambiguous, inspect context and ask when necessary.
- Identify changed behavior, affected components, and actual consumers.

## 2. Load Applicable Rules

Read the [Rule Router](../../../rules/README.md) before reviewing implementation details.

- Follow its references to load only frontend, backend, AI, security, testing, or architecture rules applicable to the affected components and execution paths.
- Do not duplicate domain rules here, invent missing rules, or assume unavailable modules exist.
- Keep proposed policy separate from approved requirements.
- If the router is unavailable, report it and continue from verified requirements and observable behavior.

## 3. Analyze Risky Behavior

Trace relevant execution paths, including:

- Authentication, authorization, and tenant isolation.
- Input validation and trust boundaries.
- Error handling, sensitive data, and secret exposure.
- Transactions, concurrency, idempotency, and event replay.
- Cancellation, retries, recovery, and resource cleanup.
- API contracts, backward compatibility, and downstream consumers.

Apply checks by actual risk, not mechanically per category. Inspect implementations, callers, dependencies and neighboring tests.

## 4. Verify Findings

Before reporting a defect:

1. Identify the affected execution path and exact source location.
2. Establish concrete triggering conditions and observable impact.
3. Check existing guards, upstream validation, and downstream handling.
4. Inspect relevant tests and available evidence.
5. Determine whether the issue is introduced, worsened, or pre-existing.
6. Distinguish confirmed defects from hypotheses requiring verification.

Avoid speculative findings without a plausible execution path.

## 5. Validate Changes

- Run scoped tests, static analysis or reproductions when justified and authorized; prefer targeted over repository-wide checks.
- Never claim tests passed unless they were executed successfully.
- Document missing prerequisites, failed checks, and unexamined scope.
- For instruction changes, also inspect loading, routing and adapter behavior.

## 6. Report Findings

Report actionable findings first, ordered by severity:

- **P0 — Critical:** Immediate security, data integrity, or availability risk.
- **P1 — High:** Significant functional, security, or reliability defect.
- **P2 — Medium:** Concrete defect with limited or conditional impact.
- **P3 — Low:** Minor but actionable defect with demonstrable impact.

Each finding gives severity, description, exact location, trigger with expected vs actual behavior, impact and evidence, and a fix direction when useful. Separate confirmed findings from open questions.

Then summarize scope and baseline, applied rules with sources, validation and evidence, and residual risks, limitations and unexamined areas.

If none are found, say so without implying the code is defect-free. Do not invent findings or praise to fill a template.

## 7. Operational Boundaries

Code review is read-only by default.

- Do not modify source code or tests.
- Do not create remote comments, approvals, or review submissions.
- Do not rebase, merge, push, or alter repository history.
- Save a review document only when requested or clearly useful and authorized.
- Operational PR actions require a separate explicitly authorized workflow.
