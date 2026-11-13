---
name: code-review
description: "Review a specified Hexta diff, revision or bounded source area for actionable correctness, architecture, security and test gaps; does not merge or publish reviews."
---

# Code Review

Perform evidence-based code reviews focused on correctness, security, reliability, and behavioral compatibility. Prioritize actionable defects over stylistic preferences.

## 1. Establish Review Scope

Required context: review target and baseline (diff, revision, or explicit source scope).

- Determine the intended comparison from the user's request.
- Never assume `HEAD~1` is the baseline.
- A clean working tree does not prevent source review.
- If the baseline is ambiguous, inspect available context and ask for clarification when necessary.
- Identify changed behavior, affected components, and actual consumers.

## 2. Load Applicable Rules

Read the [Rule Router](../../../rules/README.md) to identify applicable domain-specific rules before reviewing implementation details.

- Follow the router's references to load relevant frontend, backend, AI, security, testing, or architecture rules.
- Load only rules applicable to the affected components and execution paths.
- Do not duplicate domain-specific rules inside this skill.
- Do not invent missing rules or assume unavailable modules exist.
- Keep proposed policy separate from approved requirements.
- If the router is unavailable, report the limitation and continue using verified requirements and observable behavior.

## 3. Analyze Risky Behavior

Trace relevant execution paths, including:

- Authentication, authorization, and tenant isolation.
- Input validation and trust boundaries.
- Error handling, sensitive data, and secret exposure.
- Transactions, concurrency, idempotency, and event replay.
- Cancellation, retries, recovery, and resource cleanup.
- API contracts, backward compatibility, and downstream consumers.

Apply checks based on actual risk rather than mechanically inspecting every category.

Inspect implementations, callers, dependencies, and neighboring tests.

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

- Run relevant scoped tests, static analysis, or reproduction steps when justified and authorized.
- Prefer targeted checks over expensive repository-wide validation.
- Never claim tests passed unless they were executed successfully.
- Document missing prerequisites, failed checks, and unexamined scope.
- For instruction changes, inspect loading, routing, and adapter behavior in addition to instruction text.

## 6. Report Findings

Report actionable findings first, ordered by severity:

- **P0 — Critical:** Immediate security, data integrity, or availability risk.
- **P1 — High:** Significant functional, security, or reliability defect.
- **P2 — Medium:** Concrete defect with limited or conditional impact.
- **P3 — Low:** Minor but actionable defect with demonstrable impact.

Each finding must include:

- Severity and concise description.
- Precise source location.
- Triggering conditions and expected versus actual behavior.
- Concrete impact and supporting evidence.
- Recommended direction for resolution when useful.

Separate confirmed findings from open questions.

After findings, summarize:

- Review scope and baseline.
- Applied rules and their source references.
- Validation performed and evidence collected.
- Residual risks, limitations, and unexamined areas.

If no actionable findings are found, state that explicitly without implying the code is defect-free.

Do not invent findings or praise to fill a template.

## 7. Operational Boundaries

Code review is read-only by default.

- Do not modify source code or tests.
- Do not create remote comments, approvals, or review submissions.
- Do not rebase, merge, push, or alter repository history.
- Save a separate review document only when requested or clearly useful and authorized.
- Operational PR actions require a separate explicitly authorized workflow.

