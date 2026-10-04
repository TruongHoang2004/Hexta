# Skill Execution & Documentation Protocol

This protocol defines the standardized lifecycle, input prerequisites, document traversal, and artifact persistence required when activating any skill across the Hexta monorepo.

Whenever an AI agent or developer invokes a skill (such as `/github-task-runner`, `/issue-creator`, `/dev-plan`, etc.), the agent **MUST NOT** perform blind execution. Instead, the agent must verify that all required inputs, documents, and conditions exist before proceeding.

---

## 1. Core 4-Phase Skill Lifecycle

Every skill execution must follow this 4-phase sequence:

```
┌────────────────────────┐
│ 1. Pre-Flight Check    │ ── Check required inputs, tickets, labels, or documentation
└───────────┬────────────┘
            │ (Pass)
            ▼
┌────────────────────────┐
│ 2. Context Loading     │ ── Read GEMINI.md, agentic-memory/rules/, relevant code & designs
└───────────┬────────────┘
            │
            ▼
┌────────────────────────┐
│ 3. Core Execution      │ ── Execute skill actions (code, CLI script, testing, git commands)
└───────────┬────────────┘
            │
            ▼
┌────────────────────────┐
│ 4. Artifact & State    │ ── Persist memory (agentic-memory/), transition issue/PR states
└────────────────────────┘
```

---

## 2. Skill Prerequisites & Artifact Matrix

This matrix specifies the **exact required inputs** needed before starting a skill, the **documents to consult**, and the **resulting artifacts** that must be generated:

| Skill | Trigger / Command | Required Prerequisites (Inputs) | Documents to Read First | Generated Artifacts / State Transition |
|---|---|---|---|---|
| **`issue-creator`** | `/issue-creator`<br>or user requests ticket | Problem statement, user requirements, or bug description; scope boundaries. | 1. `GEMINI.md`<br>2. `agentic-memory/rules/development-rules.md` | - GitHub Issue created via `create_issue.sh`<br>- Labels: `ready`, `priority:*`, type label. |
| **`github-task-runner`** | `/github-task-runner` | At least **one** GitHub issue with status label `ready` and structured acceptance criteria (`- [ ]`). | 1. Selected Issue description<br>2. `GEMINI.md`<br>3. Existing designs in `agentic-memory/designs/` | - Issue transitioned `ready` ➔ `in-progress`<br>- Plan in `agentic-memory/plans/`<br>- Review in `agentic-memory/reviews/`<br>- Changelog in `agentic-memory/changelogs/`<br>- Feature branch & PR opened (`Closes #<id>`)<br>- Issue transitioned `in-progress` ➔ `in-review`. |
| **`dev-plan`** | `/dev-plan` | An assigned Issue, user prompt detailing feature goal, or technical objective. | 1. `agentic-memory/rules/agentic-memory.md`<br>2. `GEMINI.md`<br>3. Existing code in impacted services | - `agentic-memory/plans/YYYY-MM-DD_<feature>_plan.md` |
| **`dev-design`** | `/dev-design` | Architectural change, DB schema evolution, API contract definition, or cross-service feature. | 1. `agentic-memory/plans/...`<br>2. Schema definitions (`migrations/`, `proto/`)<br>3. `services/api/internal/...` | - `agentic-memory/designs/YYYY-MM-DD_<feature>_design.md` (Mermaid diagrams, DTO specs, DB schema) |
| **`basic-code-review`**<br>& **`dev-review`** | `/dev-review` | Code changes staged or committed in git diff. | 1. Active git diff (`git diff HEAD~1` or working tree)<br>2. Associated plan (`agentic-memory/plans/`)<br>3. `GEMINI.md` | - `agentic-memory/reviews/YYYY-MM-DD_<feature>_review.md` |
| **`dev-changelog`** | `/dev-changelog` | Completed code modifications and test verification results. | 1. Staged / committed git diff<br>2. Test execution outputs (`go test`, `npm run lint`) | - `agentic-memory/changelogs/YYYY-MM-DD_<feature>_changelog.md` |
| **`commit-message-generator`** | `/commit-message-generator` | Staged changes in git (`git diff --cached`). | 1. Staged git diff<br>2. Issue number reference | - Conventional Commit message formatted as `type(scope): summary (refs #<num>)`. |
| **`issue-lifecycle-manager`** | `/issue-lifecycle-manager` | Existing GitHub issues or Pull Requests needing status sync, triage, or stale lock recovery. | 1. `agentic-memory/rules/github-workflow.md` | - Issue labels updated (`ready`, `in-review`, closed cleanup)<br>- Stale locks unlocked. |
| **`pr-review-agent`** | `/pr-review-agent` | An open Pull Request targeting `main`. | 1. PR diff & commit history<br>2. Memory artifacts linked in PR body | - PR review comment / approval<br>- Auto squash-merge or tag `needs-human-review`. |

---

## 3. Pre-Flight Failure Handling

If a skill is invoked and its **Required Prerequisites** are missing:

1. **DO NOT GUESS OR BYPASS**: Do not create arbitrary code or empty PRs.
2. **HALT & NOTIFY**: Clearly inform the user what is missing.
3. **RECOMMEND THE UPSTREAM SKILL**:
   - If `/github-task-runner` is called but no issue has label `ready`:
     > *"No task is currently marked `ready`. Would you like to use `/issue-creator` to formulate an issue first, or run `manage_issue_lifecycle.py --triage` to promote an existing backlog item?"*
   - If `/dev-review` or `/dev-changelog` is called without any code diff:
     > *"No file changes detected in working tree or git history to review or document. Please complete code changes first."*
   - If `/dev-design` is requested with ambiguous requirements:
     > *"Requirements for this architectural change are ambiguous. Please clarify [specific questions] or formulate an issue via `/issue-creator` first."*

---

## 4. Documentation Storage Convention

All persistent reasoning must be recorded in `agentic-memory/` according to standard categories:
- Plans: `agentic-memory/plans/YYYY-MM-DD_<feature-name>_plan.md`
- Designs: `agentic-memory/designs/YYYY-MM-DD_<feature-name>_design.md`
- Reviews: `agentic-memory/reviews/YYYY-MM-DD_<feature-name>_review.md`
- Changelogs: `agentic-memory/changelogs/YYYY-MM-DD_<feature-name>_changelog.md`
- Rules: `agentic-memory/rules/<rule-name>.md` and `.agents/rules/<rule-name>.md`

All files, headings, code symbols, and comments in memory documents **MUST be in English**.
