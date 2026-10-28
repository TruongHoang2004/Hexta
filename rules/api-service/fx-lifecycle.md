# Fx composition and resource lifecycle

Scope: internal/bootstrap, constructor bindings and resource ownership. Use the root [rule router](../README.md) for selection and provenance/status conventions; do not load sibling modules merely because they are linked.

| Rule ID | Origin | Requirement | Verification |
| --- | --- | --- | --- |
| DI-001 | E | Production constructors and bindings MUST be assembled with `fx.Provide`/`fx.Invoke` in bootstrap. Business/controller/repository code MUST NOT construct global DB/cache/provider clients or use a service locator. Pure value/helper constructors are allowed. | S allocation/import scan; R lifetime |
| DI-002 | O | Infrastructure MUST implement low-level capabilities; bootstrap MUST bind them to the interfaces consumers request, including aggregate read ports and the shared cache manager; no full per-repository cache-wrapper bindings in new code. Constructor errors MUST fail startup, and the DI graph MUST be tested when bindings change. | T Fx graph; S bindings |
| DI-003 | O | Long-lived servers, DB/client connections, background tasks and telemetry MUST have explicit owners and Fx lifecycle start/stop cleanup. Shutdown MUST use a bounded context; background work MUST honor cancellation. | S lifecycle; T startup failure/shutdown |

No new factory framework, global service locator or mandatory empty baseService embedding is required. A health probe is an infrastructure capability; placing raw GORM/cache clients in core is a boundary deviation.
