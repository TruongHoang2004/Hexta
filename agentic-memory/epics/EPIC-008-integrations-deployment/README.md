---
epic: EPIC-008
title: Direct platform connector and eventual VPS deployment
status: in-progress
completed: false
updated: 2026-11-03
---

# Direct platform connector and eventual VPS deployment

## Scope

Verify live integration capabilities and deploy only after local representative flow and sizing evidence exist.

## Decisions and context

- Owner superseded Pancake on 2026-11-03: direct Facebook through a separate Go I/O service, durable inbox/outbox and BasePlugin/platform plugins. Owner has a Page and Meta app; actual access remains unverified.
- Demo eventually runs on VPS; no host capacity/domain chosen. Measure runnable components before sizing.
- Only verify documented webhook authentication/send/reconciliation capabilities; do not invent signatures or delivery guarantees.

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

- [2026-11-03 connector ingress](sessions/2026-11-03_connector-ingress.md) — local/in-review; signed webhook through isolated Kafka to member inbox verified. No live Meta/send/merge claim.

## Remaining work and next handoff

Finish tenant-authorized provisioning, verify Meta live receipt, add provider-specific fetch only after capability confirmation, and implement API permits/outbound attempts before sending. No live sends or deployment implied.
