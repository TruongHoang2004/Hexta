# Protobuf generation and compatibility

Scope: root proto sources and generated API RPC bindings. Use the root [rule router](../README.md) for selection and provenance/status conventions; do not load sibling modules merely because they are linked.

| Rule ID | Origin | Requirement | Verification |
| --- | --- | --- | --- |
| INF-004 | E | Checked-in generated protobuf code MUST be regenerated from root proto sources using documented tooling, never hand-edited. Contract changes MUST preserve compatible field numbers or explicitly version a breaking change. Generated bindings do not establish a running business integration. | S generator/source diff; T compatibility/smoke |

Generation commands/pins live in the repository-root proto/README.md. Load it when regenerating; do not load it for ordinary Go builds.
