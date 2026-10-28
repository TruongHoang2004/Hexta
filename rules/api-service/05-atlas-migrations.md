# Atlas-only database lifecycle

Authority: root `AGENTS.md` database rules, root `atlas.hcl`, root Makefile, `scripts/migrate.sh`, `migrations/{schema,api,ai}`, `internal/testutil/database.go`. These rules apply to API runtime/tests and explicitly referenced shared migration tooling. They do not authorize modifications to an application database during a rule check.

| Rule ID | Origin | Requirement | Verification |
| --- | --- | --- | --- |
| MIG-001 | E | Every physical API schema change MUST be represented in `migrations/schema/api.sql` and versioned under `migrations/api` through Atlas tooling. Root `atlas.hcl` is canonical; obsolete service-local `--env gorm` commands MUST NOT be used. | S changed files/tool command; R SQL review |
| MIG-002 | E | Applications and tests MUST NEVER run GORM AutoMigrate, GORM Migrator schema operations, schema-creating model hooks, startup DDL or an alternative automatic migration path. Restarting a service MUST NOT create/alter/drop business schema. | S semantic call/DDL scan; T startup on existing schema |
| MIG-003 | E | Schema migrations MUST be generated using root `make migrate-diff svc=api name=<change>`, reviewed, and validated using `make migrate-validate`. For data-only/custom SQL, create the versioned file through Atlas (`make migrate-init svc=api name=<change>`), author/review SQL and refresh/validate the Atlas checksum. File naming or checksum presence alone does not prove Atlas provenance. | S history/diff; R generation evidence/SQL |
| MIG-004 | E | Applying migrations MUST be an explicit migration job/command (`make migrate svc=api` or `make migrate-apply-local svc=api` with private configuration). Readiness/connection initialization MUST NOT apply DDL. A check-only hook MUST NOT apply to the application's database. | S startup/tool paths; T jobs/startup |
| MIG-005 | E | API/public and AI/ai_runtime MUST retain separate histories, ownership and revision schemas (`atlas_schema_revisions`, `atlas_ai_revisions`). A service MUST NOT mutate another service's tables/schema as a migration shortcut. | S Atlas config/history; T fresh/drift checks |
| MIG-006 | P | Published/applied migration files MUST remain immutable. Corrections MUST use a new migration; changing atlas.sum MUST NOT conceal edits to applied history. Unpublished local generation corrections require explicit review and evidence that no target has applied that version. | S base/history diff; R deployment evidence |
| MIG-007 | E | DB integration tests MUST use an explicitly configured test server and disposable databases migrated with Atlas. They MUST NOT AutoMigrate or migrate/reset the normal application DB. Missing DB prerequisites MUST be reported as NOT_RUN/skipped, not DB verification PASS. | T fixture/source/output |
| MIG-008 | O | Schema-changing delivery MUST verify history integrity, desired-schema/history agreement, fresh application and supported existing-data upgrade as applicable. SQL review MUST cover constraints, tenant relationships, nullability, indexes, defaults, backfill and destructive/locking effects. | T explicit disposable jobs; R upgrade/SQL |

## Required migration sequence

From repository root, with a descriptive snake_case name:

```sh
make migrate-diff svc=api name=add_example_constraint
make migrate-validate
```

Before generation, edit the desired SQL; after generation, inspect the SQL. These commands may start the configured local migration dependencies; they are implementation steps, not part of this documentation deliverable.

For custom/data migrations, `make migrate-init` is the explicit supported exception to schema-diff generation. It is still Atlas-managed versioned migration, not arbitrary runtime DDL. GORM tags do not replace desired SQL. Test fixture CREATE/DROP DATABASE for isolated provisioning and explicit infrastructure role/schema initialization are tooling-only exceptions; neither allows application table DDL outside Atlas histories.
