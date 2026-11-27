# Shared Go package

This is a Go module containing shared errors, validators, logging/telemetry and identity protobuf contracts. It is not a frontend utility package. Inspect actual imports before changing an exported helper.

- Keep reusable helpers within their responsibility; do not create a shared back door for business queries, provider calls or authorization. Do not import service internals.
- Use [Go conventions](../../rules/api-service/go-conventions.md) and the root workspace. Select error/observability rules only for affected behavior; avoid loading the complete API catalog.
- Preserve exported signatures and error/status behavior for all actual consumers, including API and test modules. Keep tenant/business policy in its owning service.
- `proto/identity.proto` and `gen/go/identity/v1/` are a separate existing contract family from root `proto/hexta/ai/v1/`. Never hand-edit generated bindings. Verify this family's generation command/tool pins before regeneration; the root AI recipe is not an identity generator. If tooling provenance is missing, report that gap before changing generated outputs.
- Run affected shared/API/test-module checks based on imports and behavior, with isolated prerequisites for integration work. Generated code is exempt from handwritten style rules, not compatibility or secret-handling requirements.
- Asynchronous contracts live in `events/`: one file (plus its `_test.go`) per event in its domain package, exactly one Kafka topic per event (no topic carries two event types), and one line in `events/catalog`. Keep existing topic names and wire formats unless the owner approves a change; provision new topics in the Compose files (the catalog test enforces it). `kafkatransport` stays transport-only.
