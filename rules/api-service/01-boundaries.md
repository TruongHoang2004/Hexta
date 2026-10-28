# Layer and dependency boundaries

Scope: production Go source in `services/api`; tests/tools follow the explicit qualifications below. Baseline evidence: [inventory](00-inventory.md), `internal/bootstrap`, `core/service`, `present/http`, `repository`.

| Rule ID | Origin | Requirement | Verification |
| --- | --- | --- | --- |
| ARC-001 | E | Entry points MUST load/validate configuration and start/stop the application or execute their named command. They MUST NOT implement HTTP use cases, SQL queries or business transitions. | S imports/calls; R command scope |
| ARC-002 | E | HTTP presentation MUST call services for business work. Controllers/router/middleware MUST NOT import repository, GORM, database/sql, Redis clients or provider adapters, and MUST NOT issue persistence queries. | S import/AST; R indirect bypass |
| ARC-003 | E | Business services MUST NOT contain SQL/query builders, GORM operations, Gin/HTTP response handling or infrastructure client imports. They MUST express business work through repository/service/external capability interfaces. | S import/AST; R semantics |
| ARC-004 | E | Repository MUST contain persistence concerns only: queries/writes, ORM mapping, transactions, locks, persistence guards, persistence errors and audit persistence. Structural conversion MUST be delegated to aggregate mapper modules; new cache wrappers are forbidden under the owner follow-up. It MUST NOT orchestrate business use cases, decide actor permissions, render responses or call business services/controllers. | S imports; R responsibility |
| ARC-005 | E | Infrastructure MUST own low-level connection/client/protocol mechanics. It MUST NOT decide business permission, workflow progression, order confirmation or external sending eligibility. | S imports/calls; R semantics |
| ARC-006 | P | Core types MUST be free of framework/client/ORM imports and persistence tags. ORM types MUST live under repository; core/use-case types and transport DTOs MUST be separate when their contracts differ. Exact migration layout requires the review decision in 09-adoption-review.md. | S tags/imports; R mappings |
| ARC-007 | O | Bootstrap MUST own Fx graph assembly, adapter/interface binding and process resource lifecycle. Business packages MUST NOT import bootstrap or create an Fx application. | S imports/calls |
| ARC-008 | E | Shared packages/common/utils MUST NOT become a back door around boundaries. Pure helpers MAY be shared; helpers that query, authorize or call providers MUST reside in their owning layer. | S dependency graph; R helper purpose |

## Proposed production import matrix

“Allowed” lists architectural dependencies, not all standard/shared library imports. A dependency must still be needed for that package's responsibility.

| Source | Allowed application dependencies | Forbidden dependencies |
| --- | --- | --- |
| `cmd` | bootstrap, config, shared startup/logging; named CLI use-case entry | Direct business SQL/transition implementation |
| `bootstrap` | constructors/ports in every layer | Business policy implementation |
| HTTP router/middleware/controller | services, core values, DTO/response/validator, aggregate HTTP mapper, shared auth-context/errors/log | repository, infrastructure clients, ORM/DB/cache drivers |
| HTTP DTO/response/validator | transport-neutral core values, scalar/time/decimal types | persistence models, repository/infrastructure clients |
| core/service | core model/types, repository interfaces under current layout, read/cache capability ports, shared errors/log | presentation, bootstrap, raw drivers, infrastructure clients |
| core model/types | pure value/type dependencies such as time/decimal | presentation, repository, infrastructure, config globals, ORM |
| repository | core data, persistence models, aggregate persistence mapper, DB/ORM | core/service, presentation, bootstrap |
| infrastructure | config, protocol/value types; narrow implemented capability ports | presentation/use-case orchestration or concrete business services |

Repository interfaces may remain in `repository` under this compatibility draft. Relocating them to a core ports package is a separate proposal, not a prerequisite to documenting the owner's existing architecture. Infrastructure must not import the service package merely to reference a port when doing so forms a service ↔ adapter cycle; bootstrap can bind structurally satisfied interfaces.

Aggregate HTTP/persistence/cache mappers follow [MAP rules](10-aggregate-mappers.md); do not import the full mapper module into core. Selected read adapters may compose repository loaders with the shared cache manager through ports as specified in [cache rules](11-cache-management.md). This is an explicit supporting adapter, not permission for SQL in core or a cache-enabled full repository wrapper. Existing wrappers are migration debt, not the target import matrix.

Tests MAY import multiple layers for integration/Fx wiring. `internal/testutil` MAY provision isolated databases through Atlas. Generated protocol code is excluded from handwritten style/import-placement rules but not secret/schema-ownership rules. Explicit migration tooling MAY execute approved Atlas migrations; runtime startup MAY NOT. Any production exception needs named scope, rationale, rule ID and owner review; being “legacy” does not make it compliant.
