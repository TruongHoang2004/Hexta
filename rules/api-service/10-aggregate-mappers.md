# Aggregate-owned mapper modules

Owner decision: every aggregate must own a separate mapper module so transformation code does not expand controllers/services/repositories. Current code has logical aggregate candidates, not an already enforced DDD module layout. This document specifies the target boundary; it does not claim the target packages exist.

## Rules

| Rule ID | Origin | Requirement | Verification |
| --- | --- | --- | --- |
| MAP-001 | D | Every implemented aggregate with boundary representations MUST own a dedicated mapper module. Mappers MUST NOT accumulate in one global mapper file or be copied into controllers/services/repositories. An aggregate is a consistency boundary, not automatically one database table. | S module inventory; R aggregate ownership |
| MAP-002 | P | The aggregate mapper module MUST separate HTTP, persistence and cache conversions into boundary-specific Go packages. Callers MUST import only the needed package; core MUST NOT import mapper packages that depend on DTO/ORM/cache representations. | S dependency/cycle graph |
| MAP-003 | P | Mapper functions MUST be deterministic, side-effect-free structural transformations. They MUST NOT query, use Redis/providers, authorize, apply business defaults/transitions, generate IDs/timestamps or calculate business prices. They MUST accept all required data as inputs. | S imports/calls; T conversion cases; R semantics |
| MAP-004 | P | Mappers MUST explicitly select fields and preserve zero/null/missing, decimal precision, enum values and intended timestamps. Secret/internal fields MUST NOT leak into HTTP/cache shapes; unknown stored representations MUST fail safely rather than invent a business state. | T round-trip/serialization/unknown values |
| MAP-005 | P | Services MUST accept/return transport-neutral inputs/results. Controllers MUST use the aggregate HTTP mapper; repositories MUST use its persistence mapper; read adapters MUST use its cache mapper. A layer MUST NOT construct another layer's model simply to invoke a mapper. | S types/call sites; R boundary flow |
| MAP-006 | P | Cross-aggregate read projections MUST have a named projection-owned mapper and explicit source inputs. One aggregate mapper MUST NOT load or mutate another aggregate. Lists MAY use small batch mapping helpers in the owning module. | S imports/calls; R projection ownership |
| MAP-007 | P | Mapper tests MUST cover meaningful conversion risks: omitted/zero values, decimal/time precision, unsupported persisted enums and sensitive field exclusion. Pure scalar copies need no mirror-only test per field. | T focused unit tests; R relevance |

**D** denotes an explicit owner decision in this revision. Detailed mechanisms marked P remain proposed for review.

## Proposed incremental package layout

Retain the existing layers while making ownership explicit. Example for the logical order aggregate:

```text
services/api/internal/
  core/model/order/                 # pure aggregate values
  core/types/order/                 # use-case inputs/results
  repository/model/order/           # GORM rows
  present/http/dto/order/           # HTTP contracts
  readmodel/order/cachevalue/       # cache snapshot contract, if enabled
  mapper/order/                    # one aggregate-owned mapper module
    httpmap/mapper.go              # DTO <-> use-case values
    persistencemap/mapper.go       # GORM rows <-> core values
    cachemap/mapper.go             # cache snapshots <-> read values
```

This is logical ownership under `mapper/order`, not a requirement to move all API packages into vertical modules now. If vertical aggregate modules are adopted later, the same mapper subpackages can move with their aggregate; that broader migration is separate. An aggregate without cache needs no empty cachemap package. No public facade may import all three packages, since it would pull transport/persistence dependencies into unrelated callers.

| Package | May depend on | Must not depend on |
| --- | --- | --- |
| `httpmap` | owning DTO, core inputs/results, pure values | GORM rows, repository clients, Redis, service implementations |
| `persistencemap` | owning persistence models and core values | DTO/HTTP, Redis, repository/service implementations |
| `cachemap` | owning cache snapshot schema and core read values | Redis driver, DTO/HTTP, ORM, service implementations |
| core values | pure value packages | any mapper/DTO/persistence/cache package |

Persistence model packages must not import their mapper; repositories depend on both. DTO/cache schema packages must likewise not import their mapper. This avoids conversion-driven cycles. Invalid conversion may return a typed/plain conversion error; its owning layer maps it to the shared safe application error. HTTP validators still validate input syntax and services still enforce business invariants; a mapper replaces neither.

Order/line items are an illustrative ownership unit. Workflow definition/version, work item, tenant, customer and product boundaries must be named in the migration plan from their actual invariants; a file named `sales` is not proof of a single Sales aggregate.
