# AI protobuf contracts

Use [the existing compatibility rule](../rules/api-service/protobuf.md) and [generation commands](README.md) when changing the root AI contract. Preserve field numbers and reserve removed names/numbers; breaking changes require an explicit version/consumer plan.

Trace Go API and Python AI consumers. Regenerate both checked-in binding sets with documented pins; do not hand-edit them or infer a running integration from generated code. Test serialization/default behavior for older callers, including generation opt-in. Inspect generated diff scope and affected RPC tests.

The identity protobuf under `packages/shared/proto/` has separate ownership and generation provenance; use [shared-package guidance](../packages/shared/AGENTS.md). Cross-owner contract work uses the [cross-service workflow](../.agents/skills/cross-service-change/SKILL.md).
