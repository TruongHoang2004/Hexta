---
name: cross-service-change
description: "Coordinate Hexta HTTP, protobuf, schema or shared-export changes that affect multiple owners or consumers, including generated bindings and rollout compatibility."
---

# Cross-service change

Required context: proposed contract/export change and actual producers/consumers. Read each affected domain's instructions through [the router](../../../rules/README.md), not every domain by default.

1. Build an impact map: source of truth, producer, direct/transitive consumers, generated outputs, tests and rollout order. HTTP uses Go DTO/Swagger plus handwritten SDK/app wrappers; root AI protobuf produces Go/Python bindings; shared identity protobuf is a separate family.
2. Establish additive/default/unknown-value behavior and mixed-version expectations. Identify whether schema/data upgrades, exported props/types or token changes require staged rollout. Do not label a breaking change compatible because compilation passes.
3. Change source contracts first, then regenerate only owned outputs using documented tooling. If generator provenance is missing, stop that dependent regeneration and report it rather than guessing or hand-editing generated files.
4. Update affected consumers and focused compatibility tests. For schema work follow Atlas ownership and explicit isolated validation; for UI exports inspect actual app imports.
5. Verify source/output parity and relevant producer/consumer checks. Document ordering, rollback/recovery limits and any required follow-up when simultaneous deployment cannot be assumed.

Output: compact consumer/contract matrix, implementation and generation evidence, rollout constraints and NOT_RUN gaps. Update the relevant architecture owner module and one primary session when writes are authorized. Do not duplicate the full log across epics.
