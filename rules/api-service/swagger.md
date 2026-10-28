# Swagger generation and route contracts

Scope: endpoint documentation, registered paths and generated Swagger. Use the root [rule router](../README.md) for selection and provenance/status conventions; do not load sibling modules merely because they are linked.

| Rule ID | Origin | Requirement | Verification |
| --- | --- | --- | --- |
| HTTP-007 | E | Every new/changed HTTP endpoint MUST have accurate Swagger method, full registered path, parameters, security and success/failure shapes. Run repository-root `make swagger` and include regenerated artifacts; generated Swagger MUST NOT be edited manually. | S generated diff/route contract; T docs contract |

Run `make swagger` from repository root when endpoints/contracts change. Source annotations must accurately name the registered method/path, parameters, applicable BearerAuth and success/failure response types. Generated files are not handwritten.
