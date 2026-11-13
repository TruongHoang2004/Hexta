---
epic: EPIC-008
session: 2026-11-03_connector-design
status: in-review
completed: false
date: 2026-11-03
---

# Connector design

## Scope and starting evidence

Design-only request: direct Facebook, separate connector, abstraction before plugins, webhook/poll normalization into Kafka. Main at 2e5f18e; unrelated registration/privacy work preserved. Inspected simulator service/router and workspace: no connector or operational ingestion consumer found.

## Outcome and decisions

[Design](../../../designs/2026-11-03_connector-service_design.md) distinguishes owner decisions from proposed Go interfaces, durable receipts/outbox, polling leases, event identities and private APIs. API retains business authority. Related epics: EPIC-004 and EPIC-005. Canonical system modules mark earlier Pancake targets superseded.

## Verification and evidence

Source and official Kafka/Meta-owned API collection reviewed. Direct Meta docs returned 429; exact webhook/history capabilities remain unverified. Documentation whitespace and local-link checks passed. Application/integration/live tests NOT_RUN (design only); no code/deployment/merge claim.

## Remaining work

Review language/operational proposals; implement contracts and fixture adapter before Facebook webhook. First acceptance: real Page text appears once in inbox. Polling provider support, AI and outbound remain subsequent slices. No parent completion claim.

## Subsequent owner decision

The owner rejected this design. Its proposed implementation details are not approved. See [reference-based follow-up](2026-11-03_connector-omni-review.md); this record retains the original session evidence.
