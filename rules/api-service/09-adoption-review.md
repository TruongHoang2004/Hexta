# Adoption decisions and known deviations

This document separates observed source behavior from proposed strict policy. It does not approve exceptions or authorize a refactor. The owner's explicit examples (SQL-free services, controller validation, persistence-only repositories, Atlas-only migrations/no automigration) are preserved throughout the catalog.

## Decisions to review

| Topic | Draft recommendation | Why review is needed |
| --- | --- | --- |
| Architecture identity | Retain layered Go architecture with Fx and repositories; tighten boundaries incrementally | Existing code does not establish full DDD/hexagonal/CQRS. No architecture rewrite requested |
| Repository scope | Permit SQL and transaction/lock/version primitives; delegate conversion to aggregate mappers, replace cache wrappers; forbid business authorization/orchestration | A literal “query strings only” rule would exclude necessary persistence responsibilities already present |
| Interface location | Keep existing repository interfaces allowed; require interface-based service dependencies; consider core/ports relocation separately | Concrete dependencies are widespread; forcing interface relocation is a new architecture decision |
| Model location | Honor existing AGENTS rule that ORM models belong in repository; proposed destination for shared ORM types: `internal/repository/model`, with transport-neutral core types | Current `core/model` violates the declared rule; moving ORM types requires mapping/import migration and must not be done in this documentation task |
| Mapper ownership | Owner requires one mapper module per aggregate; proposed boundary-specific subpackages keep conversions pure and avoid import cycles | Dedicated ownership is decided; precise layout is proposed in MAP rules |
| Root rule location | Owner selected root rules/; catalog moved, links and AGENTS navigation updated | No duplicate policy at the former memory path |
| Public DTOs | Require explicit request/response DTOs and mapping, no persistence-derived response types | Stricter than current aliases; protects API compatibility |
| Response formats | Preserve `Response[T]` for business success and existing shared ErrorResponse for failures; explicitly list native protocol exceptions | Literal “all responses” conflicts with OAuth redirects, metrics, Swagger and the current error converter |
| Legacy violations | Assess the whole API and report all failures; initially block new/worsened violations while recording old debt, if owner approves that rollout | Whole-codebase compliance and the merge gate are different; old debt must remain visible as FAIL |
| Session-cache policy | Propose DB-authoritative security state; use shared manager + policy-owned read adapters for approved display data | Owner rejected continuing the wrapper approach; exact manager/protocol and migration remain for review |
| Migration provenance | Diff-generated schema migrations; Atlas-created custom/data migrations are permitted with review | Some data migrations cannot be inferred from desired-schema diff alone; no runtime DDL exception |

## Concrete current gaps

| Gap/evidence | Affected rule IDs | Status |
| --- | --- | --- |
| `core/model/*.go` contains GORM imports/tags, while workflow ORM models live under repository | ARC-006, REP-002 | Existing divergence; no move performed |
| Auth imports concrete Redis; health service stores raw GORM/cache | ARC-003, SVC-001 | Existing boundary failures |
| Workflow/onboarding/overview constructors accept concrete repositories | SVC-002 | Proposed strictness not met |
| Sales response aliases and workflow response derivations expose persistence shapes | HTTP-004 | Proposed DTO isolation not met |
| Identity/session queries omit WithContext | REP-003 | Existing request-context gap |
| baseRepository exposes raw error details through shared conversion | REP-010, ERR-002 | Existing error-sanitization gap |
| Legacy OrderService reserves/deducts stock separately from final order/status write | SVC-005, REP-006 | Dormant legacy hazard; not the active SalesController order route |
| Auth constructor consumes DB session repository despite registered cache wrapper | CACHE-003 | Replace obsolete wiring through reviewed migration; keep DB-authoritative revocation |
| Tenant cache invalidation uses only updated slug; deletion ignores backend errors and cached tenant status may be stale | CACHE-009–012 | Old alias, authority freshness and fill-race/outage handling required; no performance benchmark claimed |
| API go.mod has a local shared replace; module/workspace Go versions differ | INF-005 | Existing replace exception and toolchain alignment review; not automatically a build failure |
| Service-local Makefile still references obsolete Atlas gorm env and old app names; Swagger title says CommerceHub | MIG-001, HTTP-007 | Tooling/documentation drift; root Atlas/Makefile remain canonical |
| Router uses bare `/ping` JSON and route/header-conflict `gin.H` error | HTTP-005, HTTP-006 | Existing envelope/classification gap |
| Core input types contain transport validation tags; custom validators include legacy non-English comments | HTTP-001, INF-006 | Placement/language cleanup needed |

This list is a source-backed starting inventory, not an exhaustive automated violation count. See the [prior review](../../agentic-memory/reviews/2026-10-28_architecture-compliance_review.md) for the detailed failure paths and bounded test evidence. No historical tests are claimed as fresh verification of this draft.

## Approval and activation sequence

1. Owner reviews the layer responsibilities, proposed rules and choices above.
2. Incorporate corrections and reconcile AGENTS.md/development-rules navigation and duplicated statements. Record version/approval, retaining historical evidence.
3. Implement full-source assessment and a reviewed legacy-debt/gating policy, with rule IDs and test evidence. Do not weaken the contract to fit current code.
4. Implement and validate a hook using the approved catalog; trust/enable through supported product controls. Add CI enforcement independently where appropriate.

Owner follow-up moves the catalog to root rules/, adds per-aggregate mappers and selects replacement of repository wrappers. See [mapper rules](10-aggregate-mappers.md), [cache design/rules](11-cache-management.md) and the follow-up design record. Other review choices remain pending.

Current deliverable stops at the reviewable draft. There are no hook files, runtime code changes, migration applications, commits or merge/completion claims.
