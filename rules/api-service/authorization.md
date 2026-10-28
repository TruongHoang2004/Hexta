# Tenant and business authority

Scope: tenant/resource permission, platform operators and AI candidate authority. Use the root [rule router](../README.md) for selection and provenance/status conventions; do not load sibling modules merely because they are linked.

| Rule ID | Origin | Requirement | Verification |
| --- | --- | --- | --- |
| SEC-005 | E | Authenticated identity MUST be derived server-side. Tenant membership, actor action and resource tenant MUST be authorized for every business read/command. Platform operator authority MUST remain separate from tenant ownership; no automatic first-user privilege escalation. | T wrong actor/tenant/operator, revoked state; R endpoints |
| SEC-006 | E | Model output/internal RPC credentials MUST NOT grant end-user business authority. Go MUST apply current-state business gates; provider/channel secrets MUST remain outside DTOs/AI snapshots/logs. | R data flow; T untrusted candidate |

Database current state is authoritative for business decisions. Load cache rules only when cached data/path or freshness can affect that decision; a valid service token or generated candidate is not end-user permission.
