# HTTP DTO contracts

Scope: request/response shapes and DTO-to-core conversion. Use the root [rule router](../README.md) for selection and provenance/status conventions; do not load sibling modules merely because they are linked.

| Rule ID | Origin | Requirement | Verification |
| --- | --- | --- | --- |
| HTTP-004 | P | Request/response DTOs MUST be explicit public contracts in `present/http/dto`. They MUST NOT alias/derive/embed ORM entities or use persistence models as nested fields. Controllers MUST delegate conversion of use-case results to the aggregate HTTP mapper; no growing inline mapping blocks. Core inputs MAY be constructed from validated DTO values. | S AST/imports; T serialization |

Delegate structural conversion to the aggregate HTTP mapper. Service input/output types remain transport-neutral; ORM aliases in current response DTOs are migration debt. Validation remains at the controller boundary; mappers do not authorize or apply business defaults.
