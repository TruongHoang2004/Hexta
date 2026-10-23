---
epic: EPIC-007
session: 2026-10-23_u08-overview-api
status: in-progress
completed: false
date: 2026-10-23
---

# U08 owner overview read API

## Scope

Implement U08 from [the work-unit plan](../../../plans/2026-10-23_demo-work-units_plan.md), following the root-reviewed [U07 owner overview contract](../../../designs/2026-10-23_owner-overview_design.md). Scope is one owner-only read API with full matching counts and a bounded work projection. U01–U06 browser/operator gates remain independent; no inferred acceptance or parent completion.

## Starting evidence

Shared branch `codex/demo-work-units`, HEAD `656a6c4040b31e8d28049338e3e094df5a6db88a` (merged PR #58 baseline supplied by root). Preserved other units' source, session, evidence and shared-document changes. Read canonical AGENTS.md, EPIC-007/index/latest U06 session, system design/API/workflow modules, actual Fx/router/controller/repository/model/schema/test conventions and the released U07 design. No nested service AGENTS.md found. Root reviewed and released U07 before source edits. Root owns consolidated epic/index/system/CHANGELOG updates and final runtime integration; this unit uses the dev-changelog workflow to provide its evidence and handoff without concurrent shared-document writes.

## Outcome and decisions

**IMPLEMENTED locally:** `GET /api/v1/tenants/{tenantID}/overview` in separate core types, service, repository, DTO and controller files, with Fx providers/router wiring and generated Swagger. It returns normalized filters, complete order status counts, work state counts/unassigned subset and deterministic paginated pinned-version work rows. No schema, index, cache, command, browser or runtime change.

- Service gates active tenant/current caller membership/exact tenant owner identity; a membership role label alone cannot grant access. Optional employee membership is checked after caller authorization. Employee IDs remain opaque strings.
- Controller rejects explicit empty/repeated known query keys, malformed timestamps and pagination syntax. Service validates paired half-open bounds, page/size/assignment and contradictory employee/unassigned filters. UTC normalization preserves timestamp fractions.
- Repository starts PostgreSQL `REPEATABLE READ READ ONLY`; plain tenant/member reads, optional employee check, independent order/work aggregates and row query share that snapshot. It deliberately avoids command helpers with `FOR SHARE`.
- Both work queries reuse one predicate builder. Rows join same-tenant physical order, pinned workflow version and current ordinal without event/all-step preload or per-row queries. Work remains present when its source order is old or soft-deleted. Orders exclude soft deletion and cannot be multiplied by multiple work items.
- Repository errors return a generic structured system error without raw SQL/driver details. Empty rows serialize as `[]`; nullable bounds/responsibility/recipient remain explicit nulls.
- Existing tenant route/header middleware also applies to the new route. Focused HTTP harness tests the controller/service rather than the full authentication middleware/header conflict path; root integration retains that distinction.

Changed/new components:

| Component | Files |
| --- | --- |
| Core | `services/api/internal/core/types/overview.go`, `core/service/overview_service.go`, `core/service/overview_service_test.go` |
| Persistence | `services/api/internal/repository/overview_repository.go`, `overview_repository_test.go` |
| HTTP | `services/api/internal/present/http/dto/overview_dto.go`, `controller/overview_controller.go`, `controller/overview_controller_test.go`, `router/router.go` |
| Composition | `services/api/internal/bootstrap/repository.go`, `service.go`, `controller.go`, `overview_test.go` |
| Generated contracts | `services/api/docs/docs.go`, `swagger.json`, `swagger.yaml` |

## Verification and evidence

Production code compiled with `go test ... -run '^$'`. The default host Go cache had a sandbox-denied entry; the unit used `GOCACHE=/private/tmp/hexta-go-cache`. The first sandboxed database run could not connect to the configured local test server. The approved local-network rerun privately sourced/exported `/private/tmp/hexta-work-unit-tests.env`, never printing values, and passed:

```sh
GOCACHE=/private/tmp/hexta-go-cache go test -race -count=1 -v \
  ./services/api/internal/core/service \
  ./services/api/internal/present/http/controller \
  ./services/api/internal/repository \
  ./services/api/internal/bootstrap -run '^TestOverview'
```

**PASS: all six tests, no skips or race report.** Each database test uses canonical testutil to create a random disposable database, apply actual Atlas history and remove only that database afterward. No application data or shared Docker runtime was changed.

| Test | Verified boundary |
| --- | --- |
| `TestOverviewWorkedCohortsAndPagination` | Exact full/date/A/B/unassigned worked counts, assigned cohort, all time, counts beyond page size, second/beyond-last pages, non-null empty array, pinned v1 labels after v2 publication, nominee excluded from B responsibility, several work per order, old/soft-deleted source order retained in W, tied UUID order |
| `TestOverviewAuthorizationAndLegacyEmpty` | Missing actor401, member/nonmember403, missing tenant404, invalid tenant400, owner empty filter, invalid member filters400, inactive403, role-only owner label403, known legacy statuses/other bucket, NULL creation timestamps excluded from bounded O, empty other tenant, removed owner membership403. Foreign/former/missing filter strings all take the same current-membership rejection; this fixture does not provision every historical account state |
| `TestOverviewConsistentReadSnapshot` | Test-only Gorm row callback/channel pauses after work aggregate; another committed update changes state/responsibility. First response keeps old aggregate/row facts, next request sees new committed state. No production test hook |
| `TestOverviewHTTPQueryAndAuthorization` | Empty/repeated six known keys, missing/equal/inverted bounds, date-only/offsetless times, invalid assignment/pagination/member, contradictory filters, structured status responses, opaque current member, UTC fractional echo, maximum valid pagination, ignored unknown key and empty DTO |
| `TestOverviewSnapshotOptionsAndReadOnly` | SQL SHOW confirms repeatable read/read-only, attempted UPDATE is rejected, failed snapshot returns system error |
| `TestOverviewControllerDependenciesAreWired` | Fx validates real repository/service/controller/validator/cache provider graph without opening live clients |

Also **PASS**:

```sh
GOCACHE=/private/tmp/hexta-go-cache go vet \
  ./services/api/internal/core/service ./services/api/internal/repository \
  ./services/api/internal/present/http/controller ./services/api/internal/bootstrap \
  ./services/api/internal/present/http/router
make swagger
```

`gofmt` applied to owned Go source/tests/wiring. A Python assertion over generated swagger.json checked overview route, seven parameters (path plus six query), Bearer security, 200/400/401/403/404/500 responses and nested response/work-row schema properties. PASS. Swagger log: `/private/tmp/hexta-u08-swagger.log` (local ephemeral evidence, not a repository artifact).

## Remaining work and handoff

Root has frozen backend source for full Go integration regression/vet and independent review. U09 was notified of the exact types/JSON and passing focused API checks. Root coordinates the final API/web runtime rebuild and browser verification; none ran in U08. No commit, push, PR, merge or privilege action by this unit. No standalone API PR was opened because root is integrating the user-authorized units on the shared branch. Session remains in-progress/completed false until broader verification/review and confirmed main merge. Root serially updates canonical API/workflow modules, epic/index and consolidated CHANGELOG using this evidence. Browser/UI acceptance and original operator gaps remain open rather than inferred from these tests.

### Integration correction — docs route-registration caller

Root's broader regression found a compile failure outside the initial focused package list: `services/api/docs/contract_test.go` directly invoked `router.RegisterRoutes` with its previous signature. Added the new `OverviewController` argument to that existing test wiring, without changing API behavior or regenerating Swagger. `rg -n 'RegisterRoutes\(' services/api` found only the router definition and this direct call. Focused verification: `GOCACHE=/private/tmp/hexta-go-cache go test -race -count=1 ./services/api/docs` **PASS** (2.207s), and `GOCACHE=/private/tmp/hexta-go-cache go vet ./services/api/docs` **PASS**. Root retains ownership of the ongoing broader regression; its initial compile finding is preserved rather than treating the earlier focused checks as full-suite evidence. Source frozen again after this correction.

## Root integration follow-up — 2026-10-23

Root integrated Go race/vet rerun and final Docker default API/web build/start passed. Bounded positive overview browser results now exist; see primary execution session. Merge remains pending. [Primary execution evidence](2026-10-23_work-unit-execution.md). Completed remains false.
