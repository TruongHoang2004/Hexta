---
epic: EPIC-008
title: Direct platform connector and eventual VPS deployment
status: in-progress
completed: false
updated: 2026-11-15
---

# Direct platform connector and eventual VPS deployment

## Scope

Verify live integration capabilities and deploy only after local representative flow and sizing evidence exist.

## Decisions and context

- Owner superseded Pancake on 2026-11-03: direct Facebook through a separate Go I/O service, durable inbox/outbox and BasePlugin/platform plugins. Owner has a Page and Meta app; actual access remains unverified.
- Demo eventually runs on VPS; no host capacity/domain chosen. Measure runnable components before sizing.
- Verify provider webhook authentication/fetch/send capabilities; do not invent delivery guarantees. The first connector design was rejected; the Omni Bridge-based replacement remains a proposal.

## Acceptance checklist

- [ ] Actual Facebook app/Page entitlement, payload/revision/auth/send/lookup behavior verified.
- [ ] Live connector mappings/credentials, quotas and unknown-send reconciliation.
- [ ] Media storage/processing and privacy/retention requirements resolved.
- [ ] VPS/domain/resources, HTTPS/private networks, broker/trace budgets and backup restore tested.

## Design and implementation references

- [02-architecture.md](../../designs/system/02-architecture.md)
- [08-operations-security.md](../../designs/system/08-operations-security.md)

## Evidence and verification boundary

Pancake research 3fab0ea49f09813ca29cfec484c2a3ef. Media/retention/backup destination/SLOs remain open; HEX-42 dataset was missing in earlier context and needs a fresh Notion check.

Status records the last verified local/PR snapshot, not live GitHub state. Recheck merge/CI before promoting to done.

## Sessions

- [2026-11-03 connector ingress](sessions/2026-11-03_connector-ingress.md) — [PR #66](https://github.com/TruongHoang2004/Hexta/pull/66), OPEN/non-draft/in-review; signed webhook through isolated Kafka to member inbox verified. No live Meta/send/merge claim.
- [2026-11-15 Facebook platform connection](sessions/2026-11-15_facebook-platform-connection.md) — PR in review, not merged: owner OAuth + Page picker, pause/resume/disconnect, media placeholders, single-attempt outbound send with result/health events. Isolated mock-Meta E2E verified; live Meta NOT_RUN.
- [2026-11-14 audit hardening](../EPIC-001-system-design/sessions/2026-11-14_audit-hardening.md) — [PR #68](https://github.com/TruongHoang2004/Hexta/pull/68) in-review: connector acceptance no longer uses a global advisory lock; admission limit is now approximate.

## Remaining work and next handoff

Tenant-authorized provisioning and single-attempt outbound are implemented locally (2026-11-15 session); design deviations are folded into the design and await owner confirmation in PR review. Then verify live Meta OAuth/receipt/send after app access, add an outbox-exhaustion sweeper and credential health checks, and add provider-specific fetch only after capability confirmation. No live sends or deployment implied.
