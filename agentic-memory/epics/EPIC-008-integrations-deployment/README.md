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

- 2026-11-03 owner decision supersedes Pancake-first: direct Facebook integration through a separate machine-facing connector service, abstraction before plugins, webhook/poll acquisition and normalized Kafka events. Owner reports Page administration and an existing Meta app; actual permissions remain unverified.
- Demo eventually runs on VPS; no host capacity/domain chosen. Measure runnable components before sizing.
- Verify provider webhook authentication/fetch/send capabilities; do not invent delivery guarantees. The first connector design was rejected; the Omni Bridge-based replacement remains a proposal.

## Acceptance checklist

- [ ] Actual Facebook app permissions, subscriptions, webhook authentication and fetch/send capabilities verified.
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

- [2026-11-03 connector design](sessions/2026-11-03_connector-design.md) — in-review; design only, implementation not started.

## Remaining work and next handoff

Implement abstractions and fixture contracts before the Facebook webhook plugin, then durable Kafka ingestion into the existing inbox. See [connector design](../../designs/2026-11-03_connector-omni-bridge_design.md). No live sends, paid calls or host purchases implied.

## Owner rejection and reference-based redesign

The owner rejected the initial draft. [Omni Bridge-based proposal](../../designs/2026-11-03_connector-omni-bridge_design.md) uses base embedding, consumer interfaces, a registry and shared inbound/sync services. Direct Kafka ACK is a proposal, not an approved durability change. [Follow-up session](sessions/2026-11-03_connector-omni-review.md).
