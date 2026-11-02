# Go API service

Use [the root router](../../rules/README.md) to select detailed API modules; do not preload the catalog. Begin with the neighboring route/controller, service, repository interface and bootstrap wiring for the changed use case.

- Preserve presentation → service → repository responsibilities and Fx composition in bootstrap. Current repository-owned interfaces are allowed; moving them into core ports is a separate design decision.
- Select Go conventions for workspace/build changes; the existing shared-module `replace` is documented debt, not permission to add another. Use the root Go workspace.
- Select authorization, errors, context/transactions, mapper and cache rules when the affected behavior requires them. Every aggregate owns a dedicated pure mapper module; new per-repository cache wrappers are forbidden. Do not promote proposed mapper/cache mechanisms to approved requirements.
- Existing core ORM types and legacy cache wrappers are deviations to track, not templates to expand. Do not refactor unrelated application code just because instructions describe the target.
- Endpoint changes select DTO/response/Swagger and actual consumers. AI RPC changes additionally select [contracts](../../proto/AGENTS.md); schema changes select [migrations](../../migrations/AGENTS.md).
- Follow [API verification](../../rules/api-service/08-verification.md) for applicable formatting, vet, race tests and contract checks. DB tests require explicitly isolated Atlas-managed databases. Missing prerequisites are NOT_RUN; do not migrate normal application databases during checks.
