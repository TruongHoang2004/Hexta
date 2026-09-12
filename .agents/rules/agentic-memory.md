# Agentic Memory Protocol Rule

This rule instructs all AI assistants and subagents working in this codebase on how to manage and persist development reasoning and lifecycle artifacts into `agentic-memory/`.

---

## 1. Core Mandate
Whenever working on non-trivial development tasks (planning, technical design, code review, or refactoring/implementing features):
- **ALWAYS** check existing rules in `agentic-memory/rules/` and `GEMINI.md`.
- **ALWAYS** persist your analysis, plans, designs, reviews, and changelogs into the corresponding folder inside `agentic-memory/`.

---

## 2. Target Locations & Naming Conventions

All documents created by the agent must follow the naming pattern:
`agentic-memory/<category>/YYYY-MM-DD_<feature-name>_<category_singular>.md`

Where `<category>` is one of:
1. **`plans/`**:
   - Filename: `agentic-memory/plans/YYYY-MM-DD_<feature-name>_plan.md`
   - Trigger: Before starting non-trivial tasks, when formulating execution steps or decomposing tasks.
2. **`designs/`**:
   - Filename: `agentic-memory/designs/YYYY-MM-DD_<feature-name>_design.md`
   - Trigger: When creating architecture specifications, database schemas, API contracts, or system diagrams.
3. **`reviews/`**:
   - Filename: `agentic-memory/reviews/YYYY-MM-DD_<feature-name>_review.md`
   - Trigger: When running code reviews, audits, PR checks, or analyzing diffs.
4. **`changelogs/`**:
   - Filename: `agentic-memory/changelogs/YYYY-MM-DD_<feature-name>_changelog.md`
   - Trigger: After completing significant changes, explaining code modifications, or providing verification steps.
5. **`rules/`**:
   - Filename: `agentic-memory/rules/<topic-rules>.md`
   - Trigger: When establishing new project-wide conventions or architectural decisions that must persist across future agent sessions.

---

## 3. Development Skills Integration
Use the specialized skills in `.agents/skills/` when executing corresponding tasks:
- `/dev-plan`
- `/dev-design`
- `/dev-review`
- `/dev-changelog`

---

## 4. Language Standard
- **Mandatory English**: All code, comments, documentation, commit messages, API specifications, and memory files (`plans/`, `designs/`, `reviews/`, `changelogs/`, `rules/`) **MUST be written in English**.

