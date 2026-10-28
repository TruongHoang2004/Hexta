# Repository, GORM and consistency rules

Evidence: `internal/repository/{sales_repository,workflow_repository,overview_repository,inventory_repository,session_wrapper,tenant_wrapper}.go`. Repository guards protect persistence invariants; they do not replace service authorization.

| Rule ID | Origin | Requirement | Verification |
| --- | --- | --- | --- |
| REP-001 | E | All business SQL/GORM reads and writes MUST reside in repository. Repository MAY contain joins/preloads, filters, pagination, upserts, calls to aggregate persistence mappers, constraint/error interpretation, locks and transaction callbacks. It MUST NOT choose actor permissions, workflow transitions, confirmation/send eligibility or call providers. | S imports/query locations; R policy ownership |
| REP-002 | E | ORM models/tags/hooks MUST be defined in the repository layer. Hooks MUST NOT implement business orchestration, provider calls or DDL. Atlas schema/history, not model auto-sync, defines the physical schema. Existing `core/model` ORM structs are deviations, not a permitted second model location. | S imports/tags/hooks; R mapping |
| REP-003 | O | Every request-derived DB operation MUST attach caller context using `WithContext(ctx)` or a context-bearing transaction. Repository MUST NOT replace it with Background/TODO. Nested queries MUST use the transaction handle, not the root DB. | S data flow; T cancellation/rollback |
| REP-004 | E | Queries of tenant-owned rows MUST include trusted tenant scope, including nested preloads/joins and mutation predicates. IDs alone MUST NOT authorize lookup/update. Identity/session/platform records need their own explicit global/principal scope; they MUST NOT be mechanically forced into tenant filtering. | S query review; T cross-tenant IDs/preloads |
| REP-005 | E | Values MUST be parameterized. Dynamic SQL identifiers/order clauses MUST be selected from code-owned allowlists; user strings MUST NOT become table/column/order fragments. | S SQL construction; T malicious inputs; R allowlists |
| REP-006 | O | Multi-row changes required by one use case MUST commit or roll back together. A conditional mutation MUST check RowsAffected and distinguish not-found/conflict consistently. Stock counters and state/version predicates MUST be enforced atomically, not read-modify-write without a lock/predicate. | T partial failure/concurrency; R transaction scope |
| REP-007 | O | Repository transaction APIs MUST hide raw GORM transactions from business services, propagate callback failures and handle commit failures. Transaction-scoped repository instances MUST NOT escape callbacks or be reused after completion. | S signatures; T callback/commit failure; R lifetime |
| REP-008 | E | Queries MUST avoid per-row N+1 reads; use bounded preloads, joins or batch lookup. Collection queries MUST have explicit pagination or an approved bounded internal use. Any unbounded audit/event hydration MUST be reviewed for payload growth. | S loops/query calls; T query/payload count; R bounds |
| REP-009 | O | Multi-query reports requiring a coherent snapshot MUST use an explicit suitable isolation scope, such as the existing overview read-only repeatable-read transaction. Read endpoints MUST NOT acquire command locks without a justified consistency requirement. | R isolation; T concurrent writer snapshot |
| REP-010 | E | Repository failures exposed upward MUST use shared application errors with safe public details. Raw SQL/driver/connection errors MUST NOT be copied into public error fields. Not-found/conflict mappings MUST be intentional. | S mapping; T DB constraint/driver fault response |


Cache rules CACHE-001–014 now live in [shared cache management](11-cache-management.md). Current wrappers are historical implementation, not the new standard. Repository stays DB-only; pure conversions live in its aggregate mapper.

## Boundary examples

Allowed in repository: `WHERE tenant_id = ? AND version = ?`, `FOR UPDATE`, batch product lookup, FK/unique-conflict mapping, atomic stock predicates and transactional audit insertion requested by the service.

Forbidden in repository: “only the owner may publish,” “this order is ready to confirm,” “advance to the next work step,” or “send this AI reply.” Those are service decisions. Checking input needed to safely execute a persistence primitive does not move the whole business use case into repository.

Preferred write shape: controller validates transport → service authorizes and computes → repository applies scoped/conditional mutations in one transaction → service result → controller DTO mapping. No HTTP response code is selected in a repository, although shared application errors currently carry their defined HTTP mapping.
