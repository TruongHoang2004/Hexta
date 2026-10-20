---
epic: EPIC-003
title: Complete current local Docker runtime
status: in-progress
completed: false
updated: 2026-10-20
---

# Complete current local Docker runtime

## Scope

Run the current web/API/AI/PostgreSQL/Redis components, migrations and smoke checks using Docker without host language runtimes.

## Decisions and context

- VPS/domain are not needed for local work. The current profile is a built application, not hot reload or a public VPS deployment.
- scripts/local.sh generates ignored local service/JWT secrets; local PostgreSQL retains state and AI has separate schema ownership.
- Google OAuth and live model/provider connections are not configured. Legacy Go Gemini URL is disabled locally.

## Acceptance checklist

- [x] Five long-running services healthy; initialization and Atlas jobs exit 0.
- [x] Web build/TypeScript, account/login/tenant smoke and AI restart checks passed.
- [x] Docker-only launcher, runbook, temporary test DB and safe shutdown/reset instructions written.
- [x] Existing Fx IDraftCache binding startup failure fixed and regression-tested.
- [ ] Confirm PR #52 merge and CI before marking delivery done.

## Design and implementation references

- [README.demo.md](../../../infrastructure/README.demo.md)
- [local.sh](../../../scripts/local.sh)
- [compose.demo.yml](../../../infrastructure/compose.demo.yml)
- [08-operations-security.md](../../designs/system/08-operations-security.md)

## Evidence and verification boundary

HEX-10 89298bfe7b02407f8eaab9453df66229 was in progress. PR #52 contains this slice. Secrets remain ignored; never print/commit them.

Status records the last verified local/PR snapshot, not live GitHub state. Recheck merge/CI before promoting to done.

## Sessions

- [2026-10-18_local-stack](sessions/2026-10-18_local-stack.md) — in-review; completed: false.

## Remaining work and next handoff

Verify merge/CI; rerun documented startup if runtime/config changes. Inspect actual Docker state instead of assuming the containers remain running.

## Atlas-only service migrations — 2026-10-20

[Session](sessions/2026-10-20_atlas-all-services.md): owner requested Atlas for every service and no automatic schema updates. Root config/desired SQL/API+AI histories, explicit Docker migration jobs, migration-backed PostgreSQL tests and CI are implemented locally on codex/atlas-all-services. Go race/vet, 27 host/Docker AI tests, fresh/adoption/no-op/invalid-state cases, Make apply/status/zero-drift and local preserved-count upgrade passed. No runtime startup DDL or test AutoMigrate remains. Local/unmerged; parent in-progress, completed false. [Runbook](../../../infrastructure/README.demo.md#atlas-for-all-current-services).
