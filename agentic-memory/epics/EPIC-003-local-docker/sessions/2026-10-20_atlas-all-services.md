---
epic: EPIC-003
session: 2026-10-20_atlas-all-services
status: in-review
completed: false
date: 2026-10-20
---

# Atlas-only migrations for API and AI

## Scope and starting evidence

Owner explicitly requested no automatic migration and Atlas setup for all services. Started from af06cf4; created codex/atlas-all-services while preserving unrelated ERD/reconciliation/sales-evidence work. Inspected current models, seven API migrations, Python startup/store/DDL, all six Go test AutoMigrate fixtures, Docker/Make/scripts/CI and relevant epic/design context. Database owners are API/public and AI/ai_runtime; old user/catalog directories are absent. [Design](../../../designs/2026-10-20_atlas-all-services_design.md). Cross-epic runtime impact: EPIC-002.

## Outcome and decisions

Root atlas.hcl defines api/ai, explicit SQL desired schemas in migrations/schema, independent versioned histories/revision schemas, and isolated dev databases. Existing seven API migration files/checksums unchanged. New AI initial migration adopts the exact pre-Atlas runs table with catalog column/default/PK/check/index validation, retains runs and fails incompatible adoption. Only AI guarded adoption uses allow-dirty; API never does. Removed Python RunStore.migrate and startup DDL; startup checks table readiness before its worker. Removed the old Python SQL file and obsolete API-local Atlas config.

Replaced every Go AutoMigrate fixture with disposable PostgreSQL databases using actual Atlas API migrations. API_TEST_DATABASE_URL is required; absent URL explicitly skips database tests, configured Atlas/database failure fails. SALES_TEST_DATABASE_URL remains a compatibility input; the workflow actual-schema check now uses the same helper. Fixed a legacy test query to use a case-stable product substring, preserving its tenant-isolation purpose without changing application search behavior. Added assertions for seed writes. Removed existing API startup logging of the raw connection URL.

Make/script support api/ai/all, validate targets and all required URLs before migration execution, provide diff/status/validate/hash and explicit apply. Docker image contains Atlas/config/histories; local startup runs separate migrate/migrate-ai jobs before app services. db-init provisions current AI schemas/metadata and dedicated dev databases. Test profile migrates before Python tests and mounts Atlas sources read-only for adoption tests. CI provisions PostgreSQL/Atlas before Go tests, forbids automatic migration calls, validates both histories, checks desired-schema drift and runs Docker AI tests. AGENTS/runbooks/README now specify Atlas-only patterns.

## Verification and evidence

- API/shared Go race suite passed using API_TEST_DATABASE_URL with separate Atlas-migrated databases. Go vet passed. The actual-schema workflow constraint test additionally passed through the unified helper; concurrency tests now run with API_TEST_DATABASE_URL.
- Host Python: 27 passed. Final Docker Python suite: 27 passed, including migration-compatible legacy retention, incompatible legacy rejection, independent API/AI histories in a shared DB, readiness-only startup and missing-schema failure, plus existing durable-run tests.
- Fresh API history (seven migrations) and new AI history applied successfully in isolated PostgreSQL. Reapplying both histories is a no-op; metadata schemas remain independent. API checksums unchanged.
- make migrate-validate passed. make migrate svc=all and make migrate-status svc=all succeeded; status reports API 20261020075908 (7 applied/0 pending), AI 20261020100000 (1 applied/0 pending).
- make migrate-diff svc=api name=verification and svc=ai both report desired schema synced/no changes, using dedicated dev databases. No extra verification migration generated. Invalid user target and missing all-target URLs reject before migration work; invalid Make target rejects before Docker.
- Rebuilt local API/web/AI/migration images and started required services successfully; five long-running services healthy and db-init/migrate/migrate-ai exit 0. Existing demo aggregate counts stayed tenants 11, orders 8, customers 6, AI runs 1 across upgrade. Legacy adoption regression additionally verifies retained request/digest/result bytes.
- Shell syntax, Compose/CI YAML parse and git diff --check passed. No automatic migration calls remain in Go/Python source. No external provider actions or remote tracker writes.

## Repairs during verification

Initial restricted-role CREATE SCHEMA IF NOT EXISTS failed PostgreSQL database CREATE checks despite existing schema. Migration now checks pg_namespace before conditional schema creation; no database-wide grant added. Fixed Compose one-shot health-wait handling by running initialization as an explicit one-shot after PostgreSQL health. Removed premature AI dev schema creation so realm-mode Atlas dev replay starts clean. Docker adoption tests use explicit sslmode=disable and configured project root; Atlas empty-string revision errors are treated as successful when appropriate. All affected checks rerun and passed; failures were not baselined or bypassed.

## Remaining work and handoff

Local/uncommitted/unmerged. Review and deliver separately; parent epics not complete. Run make migrate svc=all and make migrate-status svc=all to reproduce explicit apply/status. Edit the service desired SQL, generate/review versioned migration and validate before delivery. Deployment must supply private migration credentials; no production role/host provisioning claimed. Separate earlier ERD/manual-acceptance work is preserved. Disposable test resources are cleaned after verification; demo services/data remain running.

## Final cleanup

Stopped the task-owned hexta-atlas-tests container (auto-removed) and Compose test-postgres. Application services/volume and dedicated future Atlas dev databases retained. Rechecked both Atlas histories, all changed Compose/CI YAML and whitespace. An initial system-Python YAML check lacked PyYAML; the existing AI virtualenv parsed all four files successfully. Discarded unrelated generated go.mod direct/indirect annotation churn; dependency versions unchanged.

## PR preparation follow-up

Owner explicitly requested a PR. Prepare a scoped Atlas commit and publish the combined D4 + Atlas branch against current main. The earlier thesis commit is already squash-merged; rebase unpublished feature commits onto current main to exclude duplicate thesis changes. Preserve unrelated ERD/reconciliation/manual evidence outside commits. Existing tests remain the evidence; browser D4 acceptance remains NOT RUN. Final PR URL/state is recorded after publication; completed stays false until merge.

## PR publication

Owner-authorized [PR #57](https://github.com/TruongHoang2004/Hexta/pull/57) is OPEN against main, containing D4 and Atlas-only migrations. Unpublished commits rebased onto main 4a32f5a: D4 is now 3bfab63 and Atlas 74dd65a; application contents unchanged by rebase, which adds only main's existing audit record. Duplicate already-merged thesis changes excluded. All 23 unrelated local files restored byte-for-byte after temporary stashing. Initial CI jobs started on 74dd65a; no success or merge claimed. Publication metadata is committed as a follow-up. Status in-review, completed false; D4 manual browser acceptance remains NOT RUN.
