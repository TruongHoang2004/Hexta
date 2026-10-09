# Hexta Autonomous Agent Squad (10-Agent Swarm)

This directory documents the 10-Agent Autonomous Engineering Team designed to build, maintain, and evolve the Hexta monorepo 24/7.

---

## 1. Squad Architecture & Domain Matrix

To eliminate race conditions, merge churn, and context drift, the team is divided into distinct specializations with domain-scoped issue queues:

| # | Agent Name | Domain Tag | Scope & File Boundaries | Recommended Model | Primary Skills |
|---|---|---|---|---|---|
| **01** | `architect_lead` | `domain:spec` | Backlog triage, architecture designs, atomic issue decomposition | `pro` | [`issue-creator`](file:///Users/truonghoang/Documents/dev/personal/Hexta/.agents/skills/issue-creator/SKILL.md), [`dev-design`](file:///Users/truonghoang/Documents/dev/personal/Hexta/.agents/skills/dev-design/SKILL.md) |
| **02** | `backend_core` | `domain:backend-core` | `services/api/internal/core/service`, domain logic, business rules | `pro` | [`github-task-runner`](file:///Users/truonghoang/Documents/dev/personal/Hexta/.agents/skills/github-task-runner/SKILL.md), [`dev-plan`](file:///Users/truonghoang/Documents/dev/personal/Hexta/.agents/skills/dev-plan/SKILL.md) |
| **03** | `backend_api` | `domain:backend-api` | `services/api/internal/present/http`, DTOs, Gorm repos, Atlas migrations | `inherit` | [`github-task-runner`](file:///Users/truonghoang/Documents/dev/personal/Hexta/.agents/skills/github-task-runner/SKILL.md), [`dev-review`](file:///Users/truonghoang/Documents/dev/personal/Hexta/.agents/skills/dev-review/SKILL.md) |
| **04** | `backend_infra` | `domain:backend-infra` | `packages/shared`, Redis cache, external clients, Uber Fx DI | `inherit` | [`github-task-runner`](file:///Users/truonghoang/Documents/dev/personal/Hexta/.agents/skills/github-task-runner/SKILL.md), [`dev-plan`](file:///Users/truonghoang/Documents/dev/personal/Hexta/.agents/skills/dev-plan/SKILL.md) |
| **05** | `frontend_ui` | `domain:frontend-ui` | `apps/web/app/components`, Tailwind CSS, UI layouts, responsive | `inherit` | [`github-task-runner`](file:///Users/truonghoang/Documents/dev/personal/Hexta/.agents/skills/github-task-runner/SKILL.md) |
| **06** | `frontend_logic`| `domain:frontend-logic`| `apps/web/lib`, API service clients, auth state, form validation | `inherit` | [`github-task-runner`](file:///Users/truonghoang/Documents/dev/personal/Hexta/.agents/skills/github-task-runner/SKILL.md) |
| **07** | `qa_engineer` | `domain:qa` | `*_test.go`, frontend tests, mock repos, regression test suites | `inherit` | [`basic-code-review`](file:///Users/truonghoang/Documents/dev/personal/Hexta/.agents/skills/basic-code-review/SKILL.md), [`dev-review`](file:///Users/truonghoang/Documents/dev/personal/Hexta/.agents/skills/dev-review/SKILL.md) |
| **08** | `code_scout` | `domain:audit` | Automated codebase audits, security leaks, 5-layer adherence, TODO tracking | `flash` | [`.agents/scripts/audit_codebase.py`](file:///Users/truonghoang/Documents/dev/personal/Hexta/.agents/scripts/audit_codebase.py) |
| **09** | `pr_reviewer` | `domain:review` | PR risk analysis, lockfile rebase, squash auto-merge, maintainer escalation | `pro` / `flash` | [`pr-review-agent`](file:///Users/truonghoang/Documents/dev/personal/Hexta/.agents/skills/pr-review-agent/SKILL.md) |
| **10** | `janitor_memory`| `domain:ops` | Stale lock recovery (>2h), memory consolidation, CHANGELOG updates | `flash_lite` | [`memory-consolidator`](file:///Users/truonghoang/Documents/dev/personal/Hexta/.agents/skills/memory-consolidator/SKILL.md), [`issue-lifecycle-manager`](file:///Users/truonghoang/Documents/dev/personal/Hexta/.agents/skills/issue-lifecycle-manager/SKILL.md) |

---

## 2. Interaction & Lifecycle Workflow

```
   ┌────────────────────────────────────────────────────────┐
   │ 01. Architect Lead: Triage Backlog & Tag Domain Labels  │
   └───────────┬────────────────────────────────┬───────────┘
               │                                │
               ▼ (Ready Issues)                 ▼ (Architecture Designs)
   ┌───────────────────────┐        ┌───────────────────────┐
   │ 02..06 Feature Devs   │        │ 08. Code Scout        │
   │ (Backend & Frontend)  │        │ (Audits & Security)   │
   └───────────┬───────────┘        └───────────┬───────────┘
               │                                │
               ▼ (Open PRs: in-review)          ▼ (Discovered Debt Issues)
   ┌───────────────────────┐        ┌───────────────────────┐
   │ 07. QA Engineer       │        │ 10. Janitor Memory    │
   │ (Adds/Validates Tests)│        │ (Cleans stale locks,  │
   └───────────┬───────────┘        │  compacts memory)     │
               │                    └───────────────────────┘
               ▼
   ┌────────────────────────────────────────────────────────┐
   │ 09. PR Reviewer & Gatekeeper: Evaluate Risk & Merge    │
   └────────────────────────────────────────────────────────┘
```

---

## 3. How to Run the Agents

### Method 1: Subagents Inside Antigravity (Interactive / Pair Programming)
All 10 agents are registered as first-class subagents in Antigravity. You can prompt the parent agent to invoke them individually or concurrently:

```text
"Hãy gọi subagent backend_core để triển khai issue #45 trên workspace isolated branch"
"Hãy gọi subagent code_scout quét toàn bộ codebase và báo cáo phát hiện"
```

The parent agent invokes them via `invoke_subagent` using isolated `Workspace: branch` mode so their git working copies never interfere.

### Method 2: Autonomous CLI Squad Runner (Local / Server Daemon)
Use the automated dispatcher script:

```bash
# Check status and assigned domains of all 10 agents:
./.agents/scripts/run_squad.sh --status

# Run the 3 monitoring agents (Scout, PR Reviewer, Janitor):
./.agents/scripts/run_squad.sh --monitors

# Run a specific agent cycle:
./.agents/scripts/run_squad.sh --agent backend_core

# Dry-run preview of pending tasks across all domains:
./.agents/scripts/run_squad.sh --dry-run
```

### Method 3: Continuous 24/7 Deployment (GitHub Actions / Crontab)
The repository includes automated GitHub Action workflows in `.github/workflows/`:
- `issue-lifecycle-sync.yml`: Runs every 30 mins to recover stale tasks.
- `pr-review-agent.yml`: Runs every 30 mins to review and auto-merge safe PRs.
- `codebase-scout.yml`: Runs weekly or on-demand to create debt issues.
