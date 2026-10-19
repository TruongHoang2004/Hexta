---
epic: EPIC-003
session: 2026-10-18_local-stack
status: in-review
completed: false
date: 2026-10-18
---

# Local Stack

## Scope

Provide and verify Docker-only startup of currently runnable application components.

## Outcome and decisions

Added scripts/local.sh, web/API/AI/DB/Redis compose, AI schema role initialization, Atlas public migration and Docker test profile. Fixed existing Fx IDraftCache provider. Wrote full runbook and smoke account/tenant checks.

## Verification and evidence

Five services healthy, init/Atlas exit 0, web build/TypeScript and Go checks passed; account/login/tenant smoke passed. Temporary test DB stopped; application services were left running. PR #52 opened.

## Remaining work and handoff

Confirm merge/CI and inspect actual Docker state. This is not hot reload or the whole future messaging/Kafka/provider feature set; environment secrets stay ignored.
