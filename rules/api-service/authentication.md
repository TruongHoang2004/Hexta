# Authentication and browser credential security

Scope: JWT/session validity, refresh/logout, Google OAuth and credentialed CORS. Use the root [rule router](../README.md) for selection and provenance/status conventions; do not load sibling modules merely because they are linked.

| Rule ID | Origin | Requirement | Verification |
| --- | --- | --- | --- |
| SEC-001 | E | Access/refresh signing keys MUST be explicitly configured, distinct and at least 32 bytes; missing/invalid configuration MUST fail startup. Token validation MUST enforce signature, token purpose, expiry and current session state. | T config, cross-purpose/expired/revoked tokens |
| SEC-002 | E | Browser refresh credentials MUST be API-owned HttpOnly, SameSite=Lax cookies, Secure in production. They MUST NOT be returned in browser JSON, URL parameters or JavaScript storage/read paths. Refresh/logout cookie mutations MUST validate trusted browser origin. | T HTTP headers/body/origin; R browser contract |
| SEC-003 | E | Credentialed CORS MUST allow explicit configured frontend origins; wildcard credentialed origins are forbidden. Logout MUST succeed server-side before the UI reports success; API failures MUST remain failures. | T CORS/origin/logout response |
| SEC-004 | E | OAuth state MUST be cryptographically random, bounded in lifetime, validated/consumed at callback and bound to the initiating browser flow. Provider identities MUST follow the defined account-linking policy; bearer/refresh credentials MUST NOT travel in callback query parameters. | T missing/reused/wrong state and account linking |

Passwords use bcrypt in the current service. Test fixtures may use clearly labeled synthetic credentials; never copy deployment secrets. Security changes require behavioral tests, not import checks alone. No production-readiness certification is implied.
