---
epic: EPIC-003
title: Complete current local Docker runtime
status: in-review
completed: false
updated: 2026-10-28
---

# Complete current local Docker runtime

## Delivery correction — verified 2026-10-28

PR #52 and #57/#58 are now verified merged into main. Pending-merge wording below describes historical session evidence, not current remote state. No fresh runtime/CI result or parent completion follows from this correction.

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
- [x] PR #52 merge into main verified on 2026-10-28; no fresh CI/parent completion inferred.

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

Delivery follow-up: [PR #57](https://github.com/TruongHoang2004/Hexta/pull/57) is open against main; initial CI started. Atlas slice is in-review, completed false; merge not verified. D4 browser acceptance is tracked separately and remains NOT RUN.

## Merge verification follow-up — 2026-10-20

PR #57 merged as 2abd305 at 08:54:27Z with all returned checks successful, superseding the publication snapshot above. Atlas code is delivered; parent acceptance is unchanged. [Cross-epic audit session](../EPIC-001-system-design/sessions/2026-10-20_code-rules-audit.md) verifies syntax/history integrity and records existing rule gaps; D4 manual acceptance remains separate.

## Test/CI compliance follow-up — 2026-10-20, local/unmerged

The shared legacy repository test helper now delegates to Atlas-created disposable databases, so all five old repository tests run with API_TEST_DATABASE_URL. Backend CI checks formatting; API config fails on invalid signing keys before startup. Full Go race suite passed against isolated PostgreSQL with no skips; demo volume unchanged. [Shared session](../EPIC-001-system-design/sessions/2026-10-20_rule-compliance-fixes.md).
