---
epic: EPIC-005
session: 2026-10-18_event-backbone-draft
status: in-review
completed: false
date: 2026-10-18
---

# Event Backbone Draft

## Scope

Draft Kafka use cases for high message/event volumes at owner request.

## Outcome and decisions

Documented topics/keys for ingress, conversation/config, AI request/results, proposals/outbound/receipts, expiry, separate orders and projections. Added outbox/inbox, retry/DLQ/replay, retention, backpressure and phased acceptance plan.

## Verification and evidence

Design cross-checked with official Kafka/Debezium references. No producer/consumer, broker load test or local Kafka profile implemented. Draft included in PR #52.

## Remaining work and handoff

Review detailed target contracts before selecting an implementation slice. A completed design session will not complete the parent Kafka implementation epic.
