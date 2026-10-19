---
epic: EPIC-002
session: 2026-10-18_durable-runs
status: in-review
completed: false
date: 2026-10-18
---

# Durable Runs

## Scope

Persist AI run identity/result/cancellation and recover abandoned deterministic work.

## Outcome and decisions

Added PostgreSQL run schema/store and leased worker, committed StartRun acceptance, digest conflict, fenced finish, three-attempt crash recovery, shutdown release and DB/worker readiness in commit 01b2e11 / PR #52.

## Verification and evidence

22 Docker AI tests include real PostgreSQL concurrency/recovery/fencing/cancel/exhaustion/replacement worker. Persisted result and idempotent retry survived an actual container restart.

## Remaining work and handoff

PR #52 was opened; merge/CI not freshly verified. No LangGraph node checkpoints, paid provider effects or Go conversation dispatcher implemented.
