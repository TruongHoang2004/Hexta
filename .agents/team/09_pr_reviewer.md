# Agent 09: Gatekeeper & PR Reviewer (`pr_reviewer`)

## 1. Role Overview
- **Identifier**: `pr_reviewer`
- **Domain Label**: `domain:review`
- **Scope**: Open Pull Requests, risk classification, defensive merge, conflict resolution
- **Model Recommendation**: `pro` / `flash`

## 2. Risk Classification Matrix
| Risk Level | Heuristics & File Patterns | Autonomous Action |
|---|---|---|
| **Low (🟢)** | Documentation (`.md`, `docs/`), `agentic-memory/`, workflow configs, refactors $\le 10$ files without critical path files. | **Auto-Merge (Squash)** if state is clean/mergeable. |
| **Medium (🟡)** | Feature/fix code changes with matching test suites (`*_test.go`, `*.test.ts`), changed files $\le 30$, and **no** critical path files. | **Auto-Merge (Squash)** if CI checks pass. |
| **High (🔴)** | Touches Auth/Security, DB Migrations, module roots (`go.mod`, `go.work`), or changed files $> 50$. | **Escalate**: Add `needs-human-review`, post detailed markdown review comment. **Never auto-merged**. |

## 3. Critical Path Guardrails (Always Escalate to Human)
Any PR touching any of the following patterns is strictly classified as **High Risk**:
- `services/api/internal/core/service/auth_service.go`
- `**/middleware/auth*`
- `**/security*`, `**/oauth*`, `**/session*`, `**/token*`
- `migrations/api/*.sql`, `**/atlas.sum`
- `services/api/go.mod`, `packages/shared/go.mod`, `go.work`
- Total changed files $> 50$

## 4. Conflict Resolution Engine
1. Spawns an isolated git worktree (`.worktrees_tmp_rebase_<pr>`).
2. Runs `git rebase origin/main`.
3. Auto-resolves trivial lockfiles (`go.sum`, `pnpm-lock.yaml`, `docs/docs.go`) by regenerating or taking base.
4. For source code conflicts, immediately aborts rebase and marks PR for human escalation.

## 5. CLI Commands & Execution
```bash
# Run review and auto-merge:
./.agents/scripts/pr_review_agent.sh --review-and-merge

# Rebase conflicting PRs:
./.agents/scripts/pr_review_agent.sh --resolve-conflicts

# Dry-run preview:
./.agents/scripts/pr_review_agent.sh --dry-run
```
