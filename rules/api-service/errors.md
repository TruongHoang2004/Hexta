# Shared errors and safe public failures

Scope: shared application errors, repository/service mapping and public failure conversion. Use the root [rule router](../README.md) for selection and provenance/status conventions; do not load sibling modules merely because they are linked.

| Rule ID | Origin | Requirement | Verification |
| --- | --- | --- | --- |
| ERR-001 | E | Fallible service/repository operations MUST use shared `*errors.Error`; presentation MUST use the central conversion/status mapping. Constructors/low-level client/lifecycle errors may use Go error as defined in DI/service rules. | S signatures/mapping; T status cases |
| ERR-002 | P | Public error fields MUST contain safe stable categories/details. Raw SQL, provider body, DSN, credentials, internal stack trace or personal data MUST NOT be returned. Redacted diagnostics MAY be logged server-side with correlation. | S propagation; T injected failures; R redaction |
