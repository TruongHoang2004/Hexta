---
name: pr-review-agent
description: Autonomous PR review, conflict resolution, and auto-merge agent. Evaluates open pull requests against risk matrices, squash merges low-risk changes, and escalates complex conflicts and high-risk PRs to maintainers.
---

# PR Review & Auto-Merge Agent Skill (`pr-review-agent`)

This skill provides autonomous Pull Request evaluation, risk classification, automated rebase / conflict resolution, squash merging, and maintainer escalation for the Hexta monorepo.

---

## 1. Overview & Operational Principles

The PR Review Agent bridges the gap between completed tasks (`github-task-runner` opening PRs in `in-review`) and integration into `main`. The agent enforces defensive merging:
- **Low Risk / Simple PRs**: Automatically squash-merged if clean and passing checks.
- **Trivial Lockfile Conflicts**: Automatically rebased and resolved in an isolated temporary worktree.
- **High Risk / Complex Conflicts**: Tagged with `needs-human-review`, commented with risk rationale and conflicting file breakdown, and escalated to maintainers.

---

## 2. CLI Usage & Commands

The core script is located at `.agents/scripts/pr_review_agent.py` with an executable wrapper `.agents/scripts/pr_review_agent.sh`.

### Operational Modes

1. **Review and Merge (Primary Mode)**:
   Scans open pull requests, evaluates risk, auto-merges low-risk PRs, and escalates high-risk PRs:
   ```bash
   ./.agents/scripts/pr_review_agent.sh --review-and-merge
   ```

2. **Conflict Resolution Only**:
   Targets only PRs in `CONFLICTING` or `DIRTY` states, attempting automated lockfile and swagger regeneration:
   ```bash
   ./.agents/scripts/pr_review_agent.sh --resolve-conflicts
   ```

3. **Dry-Run Preview**:
   Evaluates all open PRs without modifying GitHub labels, branches, or PR states:
   ```bash
   ./.agents/scripts/pr_review_agent.sh --dry-run
   ```

4. **Targeted Single PR Inspection**:
   Inspect a specific pull request by number:
   ```bash
   ./.agents/scripts/pr_review_agent.sh --pr <number> --dry-run
   ```

5. **JSON Output for CI/CD Pipelines**:
   Output machine-readable evaluation results:
   ```bash
   ./.agents/scripts/pr_review_agent.sh --dry-run --json
   ```

---

## 3. Risk Classification Matrix

| Risk Level | Heuristics & File Paths | Automated Action |
|---|---|---|
| **Low (🟢)** | Documentation (`.md`, `docs/`), README files, `agentic-memory/` artifacts, workflow configurations, simple refactors with changed files $\le 10$ and **no** critical path files. | **Auto-Merge (Squash)** if state is `CLEAN` / `MERGEABLE`. |
| **Medium (🟡)** | Feature code changes with unit tests (`*_test.go`, `*.test.ts`, `test_*.py`), changed files $\le 30$, and **no** critical path files. | **Auto-Merge (Squash)** if clean or rebase succeeds. |
| **High (🔴)** | Changes touching security/auth, database migrations, module definitions (`go.mod`), changed files $> 50$, or core services (`internal/core/service/`) without tests. | **Escalate**: Add `needs-human-review` label, post detailed markdown review comment. **Never auto-merged**. |

---

## 4. Critical Path Files (Always Escalate)

Any PR touching one or more of the following patterns is strictly classified as **High Risk**:
- **Authentication & Security**:
  - `services/api/internal/core/service/auth_service.go`
  - Any file in `**/middleware/auth*`
  - Any file in `**/security*` or `**/oauth*`
  - Tokens and JWT handlers (`auth_token.ts`, `jwt*.go`)
- **Database Migrations & Schemas**:
  - `migrations/api/*.sql` or any `migrations/**`
  - `**/atlas.sum`
- **Monorepo / Module Roots**:
  - `services/api/go.mod`
  - `packages/shared/go.mod`
  - `go.work`
- **Blast Radius**:
  - Total changed files $> 50$

---

## 5. Conflict Resolution Engine

When a PR encounters merge conflicts (`DIRTY` / `CONFLICTING`):
1. The agent spawns an ephemeral isolated git worktree (`.worktrees_tmp_rebase_<pr>`).
2. Runs `git rebase origin/main`.
3. Inspects unmerged files via `git diff --name-only --diff-filter=U`:
   - **Trivial Files** (`go.sum`, `go.work.sum`, `package-lock.json`, `pnpm-lock.yaml`, `docs/docs.go`, `docs/swagger.*`):
     Auto-resolves by regenerating or accepting baseline, continues rebase, and force-pushes with lease (`--force-with-lease`).
   - **Source Code Conflicts**:
     Immediately aborts rebase (`git rebase --abort`), cleans up the worktree, and marks the PR for escalation.

---

## 6. Escalation Protocol

When escalating a PR:
1. Adds the `needs-human-review` label (`#d93f0b`) to the PR.
2. Posts an idempotent comment containing:
   - Risk level and evaluation decision.
   - List of identified risk factors.
   - List of conflicting or critical files.
3. Embedded marker `<!-- pr-review-agent-escalation -->` ensures the agent never spams consecutive comments on the same PR.
