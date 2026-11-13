---
epic: EPIC-008
session: 2026-11-03_connector-omni-review
status: in-review
completed: false
date: 2026-11-03
---

# Omni Bridge reference and connector redesign

## Scope and evidence

Owner rejected the initial design and requested local Omni Bridge as reference. Read `/Users/truonghoang/Documents/dev/piscale/omni-bridge` at 8ec0efd with pre-existing local documentation edits; no reference files changed or credentials read. Inspected base/Facebook plugin, consumer interfaces, DI registries, normalized form, webhook handler, historical sync and Kafka outbound listener.

## Outcome

[New proposal](../../../designs/2026-11-03_connector-omni-bridge_design.md): BasePlugin embedding, consumer-owned interfaces, platform registry and common inbound/sync services. Distinguishes the reference's gRPC inbound/Kafka outbound from Hexta's requested Kafka inbound. Direct Kafka ACK before HTTP success is proposed instead of making the rejected receipt/outbox architecture mandatory. Base architecture and delivery proposal remain unapproved. Marked original design rejected and corrected system indexes/authority notes without deleting history.

## Verification

Static reference/source inspection; documentation whitespace and relative-link checks passed. Runtime/application/live integration tests NOT_RUN, no code implementation or deployment. Existing Hexta frontend edits preserved; no merge verification.

## Handoff

Review the replacement proposal, especially delivery availability tradeoffs, then implement abstract/base contracts before the Facebook plugin. No parent completion claim.
