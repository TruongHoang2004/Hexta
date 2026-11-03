---
epic: EPIC-005
title: Kafka event backbone
status: in-progress
completed: false
updated: 2026-10-19
---

# Kafka event backbone

## Scope

Implement high-volume durable event fan-out with inbox/outbox, typed contracts and safe retries/replay.

## Decisions and context

- Owner explicitly selected Kafka. Detailed topic/partition/sizing choices are target designs, not implemented facts.
- PostgreSQL remains business truth and draft authority. Kafka retains immutable events for configured windows; no cross-system exactly-once promise.
- Outbox publication may duplicate; consumer inbox commits with DB mutation/job. Replay cannot silently send messages or call paid models.

## Acceptance checklist

- [x] Use-case/topic/key/retry/DLQ/replay/retention/operations draft written for PR #52.
- [ ] Local broker profile with correct internal/host listeners and producer/consumer smoke.
- [x] Local ingress v1 envelope, transactional connector outbox relay and API consumer inbox (live/merge acceptance open).
- [ ] Ingress/conversation and AI request/result vertical slice with durable Go jobs.
- [ ] Outage, duplicate/rebalance/order-gap, hot-tenant and no-side-effect replay tests.

## Design and implementation references

- [06-kafka-events.md](../../designs/system/06-kafka-events.md)
- [2026-10-18_kafka-event-backbone_design.md](../../ARCHIVE_INDEX.md#pruned-2026-10-18-kafka-event-backbone-design)

## Evidence and verification boundary

A design session and the linked local ingress implementation now exist. Draft retention/partition/production size are not owner-approved numeric budgets.

Status records the last verified local/PR snapshot, not live GitHub state. Recheck merge/CI before promoting to done.

## Sessions

- [2026-10-18_event-backbone-draft](sessions/2026-10-18_event-backbone-draft.md) — in-review; completed: false.

## Remaining work and next handoff

Review target contracts, select a bounded ingress slice with messaging, then implement broker/profile and transactional delivery. Do not enable old compose listener config unchanged.

Connector integration: [2026-11-03 primary session](../EPIC-008-integrations-deployment/sessions/2026-11-03_connector-ingress.md). Local text ingress/Kafka/API consumer implemented and tested on isolated infrastructure; AI/outbound and merge remain open.
