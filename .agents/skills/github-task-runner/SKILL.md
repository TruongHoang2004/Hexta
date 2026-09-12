---
name: github-task-runner
description: Use this skill to automatically fetch, prioritize, and process tasks from GitHub issues. It checks out feature branches, drives the implementation workflow with agentic-memory, runs verification, and opens Pull Requests.
---

# GitHub Task Runner Skill (`github-task-runner`)

This skill automates end-to-end task execution from GitHub Issues top-down based on priority (`priority:high` > `priority:medium` > oldest open issue).

---

## Workflow Steps

### Step 1: Fetch and Identify Next Issue
Execute the issue selector script:
```bash
./.agents/scripts/get_next_issue.sh
```

- If `"status": "no_tasks"`, report that no pending tasks are ready and stop.
- If `"status": "ready"`, extract the issue object:
  - `number`: Issue ID
  - `title`: Issue Title
  - `body`: Requirements & acceptance criteria
  - `branch_name`: Feature branch name (`task/issue-<number>-<slug>`)

---

### Step 2: Transition Issue to In-Progress & Checkout Branch
1. Transition the issue label from `ready` to `in-progress` to indicate active development:
   ```bash
   gh issue edit <number> --remove-label "ready" --add-label "in-progress"
   ```
2. Checkout the feature branch:
   ```bash
   git checkout -b <branch_name>
   ```

---

### Step 3: Plan the Implementation (`dev-plan`)
1. Analyze the issue requirements, impacted code, and acceptance criteria.
2. Formulate the implementation plan and save it to:
   `agentic-memory/plans/YYYY-MM-DD_issue-<number>_plan.md`
3. Check `agentic-memory/rules/development-rules.md` and `GEMINI.md` to ensure architecture compliance (5-layer Go architecture, Gin, Uber Fx, Next.js Tailwind, etc.).

---

### Step 4: Technical Design (`dev-design`)
*(Mandatory for tasks introducing new database models, API contracts, architectural abstractions, or significant refactoring)*
1. Structure the technical design specification with system architecture, data models (GORM/Atlas), API contracts, security, and performance considerations.
2. Formulate and save the design document to:
   `agentic-memory/designs/YYYY-MM-DD_issue-<number>_design.md`
3. Include the design link in the Pull Request description under Memory Artifacts.

---

### Step 5: Implement Code Changes
1. Make code modifications adhering strictly to project standards and the approved plan/design.
2. All code comments, documentation, and identifiers must be written in **English**.

---

### Step 6: Verify Changes & Code Review (`dev-review`)
1. Run automated build and test commands:
   - For backend: `go test ./...` or `go build ./...`
   - For frontend: `pnpm test` / `pnpm run --recursive build` if applicable.
2. Perform a comprehensive code audit and save the review report to:
   `agentic-memory/reviews/YYYY-MM-DD_issue-<number>_review.md`

---

### Step 7: Generate Changelog & Verification Steps (`dev-changelog`)
Document all modified files, technical rationale, and verification steps in:
`agentic-memory/changelogs/YYYY-MM-DD_issue-<number>_changelog.md`

---

### Step 8: Commit, Push, and Create Pull Request
1. Stage all changes including memory records:
   ```bash
   git add .
   ```
2. Commit with Conventional Commits using the `commit-message-generator` skill:
   ```bash
   git commit -m "<type>(<scope>): <summary> (refs #<number>)"
   ```
3. Push feature branch to origin:
   ```bash
   git push -u origin <branch_name>
   ```
4. Open a Pull Request referencing the issue:
   ```bash
   gh pr create --title "<type>(<scope>): <summary>" --body "Closes #<number>

   ## Summary
   Brief summary of changes.

   ## Memory Artifacts
   - Plan: [agentic-memory/plans/YYYY-MM-DD_issue-<number>_plan.md]
   - Design: [agentic-memory/designs/YYYY-MM-DD_issue-<number>_design.md] (if applicable)
   - Review: [agentic-memory/reviews/YYYY-MM-DD_issue-<number>_review.md]
   - Changelog: [agentic-memory/changelogs/YYYY-MM-DD_issue-<number>_changelog.md]"
   ```
5. Transition issue label from `in-progress` to `in-review`:
   ```bash
   gh issue edit <number> --remove-label "in-progress" --add-label "in-review"
   ```
6. Post a completion comment on the GitHub issue:
   ```bash
   gh issue comment <number> --body "Automated work complete. Pull Request opened for review. Relevant artifacts recorded in \`agentic-memory/\`."
   ```
7. **Definition of Done**: The ticket is only considered **Done** once the Pull Request has been reviewed and merged into `main`. The `Closes #<number>` directive in the PR description will trigger GitHub to close the issue automatically upon merge.
