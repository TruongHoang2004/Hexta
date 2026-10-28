# Configuration and secrets

Scope: config loader, environment handling and startup validation. Use the root [rule router](../README.md) for selection and provenance/status conventions; do not load sibling modules merely because they are linked.

| Rule ID | Origin | Requirement | Verification |
| --- | --- | --- | --- |
| INF-001 | E | Config MUST be loaded/validated before use. Secret values MUST come from explicit configuration and MUST NOT have hardcoded signing-key fallbacks or be printed in errors/docs/test output. Existing global AppConfig is observed; new global/service-locator dependence is not required by this draft. | S config/constants; T missing/invalid settings |

JWT-specific configuration changes additionally require authentication rules. A non-auth configuration change does not require loading OAuth/cookie rules.
