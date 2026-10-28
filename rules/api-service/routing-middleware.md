# Routing and middleware

Scope: route exposure, actor context and middleware responsibilities. Use the root [rule router](../README.md) for selection and provenance/status conventions; do not load sibling modules merely because they are linked.

| Rule ID | Origin | Requirement | Verification |
| --- | --- | --- | --- |
| HTTP-008 | O | Middleware MUST perform cross-cutting transport work: credential extraction, validated actor context, logging/tracing/recovery/CORS. Resource permission decisions MUST remain in services. Controller MUST derive actor from trusted request context, never body/query actor IDs. | S source; T forged identity/scope |
| HTTP-009 | P | Route registration MUST expose only intentional controllers and methods. Fx-provided constructor existence MUST NOT be treated as an exposed endpoint. A new route MUST be reviewed against service authorization and documentation before registration. | S route inventory; T unauthorized caller |
