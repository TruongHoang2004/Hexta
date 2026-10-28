# Rule router

This is the single rule-selection index for humans, agents and the future hook. Read this file first; open only the modules selected below. A link is a lookup option, not an instruction to load its target or all siblings. AGENTS.md remains the canonical agent entry point.

## Selection protocol

1. Identify the task, intended files, affected contracts and actual callers. Select the union of matching path rows and behavior triggers; deduplicate files.
2. Read selected modules before editing. Follow another rule link only when its condition applies; do not recursively load indexes, inventories, designs or historical sessions.
3. Re-evaluate selection when new files, dependencies or behavior enter scope. Include matching test/code-generation paths; tests do not bypass production rules.
4. Before finishing, apply [verification](api-service/08-verification.md) to the selected scope and report relevant checks/unknowns. Do not label unselected modules as checked.
5. For an unknown path, inspect its owner/imports and use the closest responsibility row; explicitly record the selection or unresolved ownership. Do not silently skip or default to loading every rule.

Scope strings below are repository-relative globs and responsibility hints, not an installed executable matcher. Proposed package paths apply when created. Content/behavior triggers take precedence over filename assumptions.

## Path / responsibility → rule module

| Editing scope | Read |
| --- | --- |
| Any Go source/module/workspace | [Go conventions](api-service/go-conventions.md) |
| New package, changed imports/constructor/interface, moved layer responsibility | [Layer boundaries](api-service/01-boundaries.md) |
| `services/api/cmd/**`, `internal/bootstrap/**`, Fx binding or lifecycle | [Fx/lifecycle](api-service/fx-lifecycle.md) + [boundaries](api-service/01-boundaries.md) |
| `services/api/internal/present/http/{controller,validator}/**` | [Controller validation](api-service/controller-validation.md) |
| `services/api/internal/present/http/dto/**` | [DTO contracts](api-service/http-dto.md) |
| `services/api/internal/present/http/response/**` | [Response formats](api-service/http-responses.md) |
| `services/api/internal/present/http/router/**`, `middleware/**` | [Routing/middleware](api-service/routing-middleware.md) |
| `services/api/internal/core/service/**` | [Business services](api-service/03-services.md) |
| `services/api/internal/core/{model,types,ports}/**` | [Boundaries/type ownership](api-service/01-boundaries.md); use behavior triggers for affected contracts |
| `services/api/internal/repository/**`, ORM models or business SQL | [Repository/GORM](api-service/04-repositories-gorm.md) |
| `migrations/**`, `atlas.hcl`, migration scripts/jobs, DB fixture provisioning | [Atlas-only migrations](api-service/05-atlas-migrations.md) |
| `services/api/internal/mapper/**`, aggregate/projection conversion | [Aggregate mappers](api-service/10-aggregate-mappers.md) |
| Cache manager, `internal/cachepolicy/**`, cached `internal/readmodel/**`, cache wrappers/payloads | [Cache management](api-service/11-cache-management.md) |
| `services/api/internal/infrastructure/**` external clients, network timeout/retry | [Client adapters](api-service/infrastructure-clients.md) |
| `services/api/config/**`, environment/startup config | [Configuration](api-service/configuration.md) |
| `proto/**`, generated `internal/infrastructure/ai/rpc/**` | [Protobuf](api-service/protobuf.md) |
| `services/api/docs/**`, controller Swagger annotations | [Swagger](api-service/swagger.md) |
| Shared errors, error conversion/status handling | [Errors](api-service/errors.md) |
| JWT/session/refresh/logout/OAuth/cookie/CORS paths in any package | [Authentication](api-service/authentication.md) |
| Tenant membership, owner/operator grant, actor/resource scope, AI business authority | [Authorization](api-service/authorization.md) |
| `common/log/**`, logging/tracing/metrics/health probes | [Observability](api-service/observability.md) |
| `apps/web/**`, `packages/sdk/**` | [Frontend/SDK](frontend.md); API contract/auth modules only if behavior crosses that boundary |
| Issue/PR lifecycle or CI workflow/pipeline behavior | [Delivery workflow](delivery-workflow.md) + rules for the checks/modules actually changed |

## Behavior triggers (add to path selection)

| Change | Additionally read |
| --- | --- |
| Add/remove/expose an endpoint or alter its request/response contract | DTO + response formats + Swagger + authorization; controller/routing when affected |
| Change structural representation at a layer boundary | Aggregate mappers + DTO or repository or cache module for the representation touched |
| Change a fallible operation's errors/status/public detail | Errors + response formats if HTTP serialization changes |
| Change authentication/session/cookie/origin behavior | Authentication, regardless of filename |
| Change actor/tenant/resource permission or AI acceptance/send gate | Authorization, regardless of filename |
| Change transaction, concurrent state, idempotency or persistence guards | Services + repository; cache only if cached values/invalidation depend on the write |
| Change physical schema or add a DB integration fixture | Atlas; affected repository/model rules |
| Change cache snapshot/key/invalidation/freshness | Cache; mapper for snapshot conversion, authorization if authority/freshness changes |
| Change a client/resource constructor or Fx interface binding | Fx/lifecycle; configuration only if settings change |

The names in behavior rows refer to the linked modules in the path table. For a DTO rename, inspect impacted handlers/mappers/docs/tests; do not assume path matching alone describes the blast radius.

## Context budget and broad audits

Normal tasks use the selected modules only. Full-codebase review remains broad in coverage but runs module by module, retaining findings/evidence and selected rule IDs rather than loading every rule into one context. Semantic unknowns remain NOT_RUN, not PASS. Whole-source static checks can eventually run without injecting every rule body into the conversation. No hook/automatic selector exists yet.

[Inventory](api-service/00-inventory.md), [catalog metadata](api-service/README.md), [adoption decisions](api-service/09-adoption-review.md) and design documents are optional: read them for architecture discovery, policy review or implementation rationale, not every task. No default requirement to load the whole API catalog.

Rule provenance: E = approved existing requirement; D = explicit owner direction; O = observed pattern formalized for review; P = proposed refinement. E/D remain effective; O/P are draft unless separately approved. S/T/R in modules mean static/test/reasoning evidence. Stable IDs are preserved; root placement/routing do not approve the entire draft. Modules state responsibilities, not claims of current runtime compliance.

When adding/moving a rule, update this router's path/behavior selection and links in the same change. Keep one authoritative rule body per ID. Never turn a whole directory into an unconditional dependency to avoid choosing relevant rules. Legacy rule files under agentic-memory/rules are compatibility redirects only; project-memory indexes/history are optional context selected separately from engineering rules.
