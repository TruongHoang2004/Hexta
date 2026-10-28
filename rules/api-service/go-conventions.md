# Go workspace and source conventions

Scope: Go module/workspace resolution, formatting and naming. Use the root [rule router](../README.md) for selection and provenance/status conventions; do not load sibling modules merely because they are linked.

| Rule ID | Origin | Requirement | Verification |
| --- | --- | --- | --- |
| INF-005 | E | Go workspace modules MUST use root go.work for local resolution. New go.mod replace directives require documented necessity. The current API shared-module replace is an existing exception/debt to review, not permission to add more. | S module diff; R rationale |
| INF-006 | E | Handwritten Go source MUST follow gofmt, descriptive Go naming and snake_case file naming. Repository artifacts/comments MUST be English. Do not rename exported Go symbols to lower camelCase contrary to Go visibility semantics. | S format/naming; R language/clarity |
