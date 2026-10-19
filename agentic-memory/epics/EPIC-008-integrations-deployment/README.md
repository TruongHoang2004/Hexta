---
epic: EPIC-008
title: Real Pancake connector and eventual VPS deployment
status: not-started
completed: false
updated: 2026-10-19
---

# Real Pancake connector and eventual VPS deployment

## Scope

Verify live integration capabilities and deploy only after local representative flow and sizing evidence exist.

## Decisions and context

- MVP initially uses Pancake, with eventual replacement a later direction. Owner currently lacks access/entitlement to verify the real connector.
- Demo eventually runs on VPS; no host capacity/domain chosen. Measure runnable components before sizing.
- Only verify documented webhook authentication/send/reconciliation capabilities; do not invent signatures or delivery guarantees.

## Acceptance checklist

- [ ] Actual Pancake entitlement, payload/revision/auth/send/lookup behavior verified.
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

No work session recorded yet. Implementation has not started; do not invent historical progress.

## Remaining work and next handoff

Continue simulator-first work while external access and deployment inputs are absent. No live sends, paid calls or host purchases implied.
