---
epic: EPIC-001
session: 2026-10-20_ddd-domain-map
status: in-progress
completed: false
date: 2026-10-20
---

# Initial DDD domain discovery

## Scope

Identify business subdomains and candidate bounded contexts from current code and approved requirements. Documentation only; no refactoring, issue creation, commit or publication requested.

## Starting evidence

Branch codex/rules-compliance, HEAD 2abd305 with extensive existing compliance/documentation work preserved. Loaded EPIC-001/latest compliance session, canonical product/architecture/workflow/ERD, engineering rules and dev-design/skill protocol. No live PR/CI reconciliation; delivery states unchanged.

## Outcome and decisions

Added [domain map](../../../designs/2026-10-20_ddd-domain-map_design.md): nine proposed business contexts, candidate aggregates, core/supporting classification, contracts, Go/Python/provider boundaries and adoption sequence. Six have active foundations, Inventory is legacy, Conversation Handling and Response Scripts are targets. All proposed divisions remain TARGET recommendations, not owner-approved implementation decisions. No new schema/API/behavior.

## Verification and evidence

Compared active routes, service methods, sales confirmation, work creation, workflow models, desired-state schema and fake graph to canonical decisions. Markdown links/fences and whitespace checked. No runtime tests; application unchanged. The missing historical rules/agentic-memory.md reference is superseded by canonical AGENTS.md and the available epic/session/development protocols.

## Remaining work and handoff

Validate boundaries/core priorities with owner, then separately scope refactoring if requested. No automatic code reorganization or migration. Local/unmerged documentation; session and parent epic remain in-progress/completed false.
