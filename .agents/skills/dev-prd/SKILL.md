---
name: dev-prd
description: Use this skill when creating or refining a Product Requirements Document (PRD) in Hexta. Guides Notion cross-referencing, 7-step discovery, 13-section What/Why specification, and Notion page persistence.
---

# Product Requirements Document Skill (`dev-prd`)

Use this skill to define **What & Why** for a new feature or product slice before moving to Technical Design (`dev-design`).

## Core Principles

1. **Strict Separation (`What & Why` vs `How`):**
   - Never include technical implementation choices (databases, message brokers, gRPC/HTTP internals, schemas, DTOs, or service names) in a PRD. Those belong in `dev-design`.
   - Workflow stops at **PRD → Technical Design** (`dev-design`); do not generate implementation prompts unless explicitly requested.
2. **Notion + Repo Cross-Reference First:**
   - Before drafting, check Notion (`Hexta — Tài liệu dự án`, `Hexta — Tasks & Backlog`) and `agentic-memory/designs/system/01-product-decisions.md` to verify where the capability sits in Hexta and whether it aligns with current product decisions.
3. **Interactive Discovery Before Finalizing:**
   - Follow the 7-step process: *Problem Discovery → User Research → Goals & Scope → Requirements Definition → User Flow & UX → Acceptance Criteria & Metrics → Review & Sign-off*.
   - Clarify scope (In-scope / Out-of-scope / MVP), permissions, and edge cases with the user before locking the document.

## 13-Section PRD Structure (Scale to Feature Size)

1. **Overview** — High-level summary of the capability.
2. **Problem Statement** — Current pain point and why it matters; explicitly mark unverified assumptions.
3. **Target Users** — Roles and what each role needs to do.
4. **Goals & Success Metrics** — Business goals and measurable MVP success criteria.
5. **Scope** — Explicit `In-Scope (MVP)` and `Out-of-Scope` (prevent scope creep).
6. **User Stories** — Role / Action / Value statements.
7. **Functional Requirements** — Table with `ID`, `Requirement`, and `Priority` (`Must` / `Should` / `Could`).
8. **Non-functional Requirements** — Security, isolation, reliability, recoverability, performance.
9. **User Flow & UX** — Step-by-step flows and required UI states (empty, active, error, read-only).
10. **Business Rules** — Authoritative rules and edge cases (permissions, conflicts, failure states).
11. **Dependencies & Risks** — External dependencies, operational risks, and mitigations.
12. **Acceptance Criteria** — `Given / When / Then` scenarios covering both **happy paths** and **error/edge cases** (missing input, unauthorized, duplicate/conflict, external failure).
13. **Open Questions** — Unresolved product decisions blocking sign-off.

## Persistence & Readiness Check

- Create/update a dedicated page in **Notion** under **`Hexta — Tài liệu dự án`** for each PRD, with explicit status: **Bản nháp / Đang rà soát / Đã chốt**.
- Before transitioning to `dev-design`, verify the 5 readiness checks:
  1. What the system must and must not do is clear.
  2. Who is authorized for each action is defined.
  3. Business rules and edge cases are resolved.
  4. Completion criteria (`Given / When / Then`) are testable.
  5. Open questions are resolved or explicitly deferred.
