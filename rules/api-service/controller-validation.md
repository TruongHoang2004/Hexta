# Controller binding and validation

Scope: HTTP controllers and presentation validators. Use the root [rule router](../README.md) for selection and provenance/status conventions; do not load sibling modules merely because they are linked.

| Rule ID | Origin | Requirement | Verification |
| --- | --- | --- | --- |
| HTTP-001 | E | Every endpoint MUST bind an explicit request DTO, validate path/query/body inputs at the presentation boundary, then invoke a service. Body-less endpoints need only applicable input validation. DTO validation tags and reusable validators MUST be defined under presentation. | S DTO/call inspection; T invalid inputs; R coverage |
| HTTP-002 | P | Controllers MUST distinguish malformed transport input from valid input rejected by business rules. Services MUST recheck business invariants for non-HTTP callers; controller validation MUST NOT be the sole authorization/business safeguard. | T direct-service and HTTP cases; R |
| HTTP-003 | O | Handlers accepting bodies/lists MUST impose documented size/count/pagination bounds before expensive work. Zero/missing/invalid values MUST follow a defined endpoint policy; response pagination MUST match the executed query. | T boundary values; R contracts |

Syntax/required/enum/body-bound checks belong here. Current membership/permissions/transitions belong in service; uniqueness/FKs/conditional writes belong in repository/schema. Controllers must not query to validate ownership. Load DTO/mapper rules only when conversions/contracts are affected.
