---
name: dev-plan
description: Use this skill when the user asks to plan a feature, break down a task, analyze requirements, or structure an implementation roadmap. It guides the agent through planning and saves the resulting document to agentic-memory/plans/.
---

# Development Planning Skill (`dev-plan`)

Use this skill whenever you need to formulate an implementation plan, break down engineering tasks, or define roadmaps.

## Workflow

### 1. Research & Understand Requirements
- Read relevant files, configurations, and documentation.
- Check existing rules in `agentic-memory/rules/` and `GEMINI.md`.
- Identify key constraints, dependencies, and stakeholders.

### 2. Formulate the Plan
Structure the plan into the following sections:
1. **Overview & Goal**: Summary of the feature/fix and business outcome.
2. **Current State Analysis**: Existing architecture, code paths, and bottlenecks.
3. **Task Breakdown (Step-by-Step)**:
   - Chronological, concrete development tasks.
   - Files to create, modify, or delete.
4. **Risk Assessment & Edge Cases**:
   - Potential failure modes, concurrency issues, edge cases.
   - Mitigation strategies.
5. **Definition of Done (DoD)**:
   - Automated tests passing.
   - Manual verification criteria.

### 3. Persist into `agentic-memory/plans/`
- Determine `<feature-name>` (e.g., `google-oauth-flow`, `user-cart-service`).
- Get today's date in `YYYY-MM-DD` format.
- Write the plan to:
  `agentic-memory/plans/YYYY-MM-DD_<feature-name>_plan.md`
- In your conversation response, provide a brief summary and a clickable link to the saved plan file.
