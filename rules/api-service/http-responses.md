# HTTP response formats

Scope: success/error envelopes and native-protocol exceptions. Use the root [rule router](../README.md) for selection and provenance/status conventions; do not load sibling modules merely because they are linked.

| Rule ID | Origin | Requirement | Verification |
| --- | --- | --- | --- |
| HTTP-005 | E | Business JSON success responses MUST use `response.Response[T]`. Business errors MUST use shared application-error conversion and HTTP status mapping. No ad hoc `gin.H` error shapes may bypass this contract. Existing error envelope is separate from success envelope; do not wrap errors in a success response. | S response paths; T status/body |
| HTTP-006 | P | Non-business transport endpoints MAY use native formats only in an explicit inventory: Swagger UI, Prometheus `/metrics`, OAuth redirects/cookie transitions and SSE when actually routed. `/ping` must be classified/documented before being exempted. Health JSON follows the documented DTO/envelope. | R route inventory; T format |
