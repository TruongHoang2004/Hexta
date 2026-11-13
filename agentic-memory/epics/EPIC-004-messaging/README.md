---
epic: EPIC-004
title: Conversation domain, scripts and reply/send policy
status: in-progress
completed: false
updated: 2026-11-02
---

# Conversation domain, scripts and reply/send policy

## Scope

Implement tenant-isolated threads/turns/messages, assignment, script grants, proposals, result acceptance and simulator dispatch.

## Decisions and context

- All tenant sales see conversations; only assigned sale sends. Claim is atomic; owner can reassign. Takeover stops unstarted autosends, return to system is explicit.
- Owner proposal on means suggest/review; off means eligible auto_send. Routing chooses script/AI/clarify/human. Only owner grants four independent script permissions.
- Turn expiry is 24h; successful customer/sale/system activity counts, drafts/failure/unknown/duplicates do not. Needs-human exempts expiry. Resolve resets clock without resuming system control.
- Preserve human draft edits and check current context/control/script eligibility at final dispatch. In-flight remote sends may require reconciliation.

## First slice implemented locally

- Inbox list/history, simulator customer input, atomic claim, owner assignment and assigned-sale manual simulated reply.
- Tenant/current-membership checks, stale version/context rejection, durable retry keys and transactional audit; separate core/ORM/HTTP representations and aggregate mappers.
- API migration generated and checked against disposable PostgreSQL; source remains local/unmerged. No live provider, AI or parent completion claim.

## Acceptance checklist

- [ ] Atlas domain schema with tenant-aware FKs/uniqueness and one-open-turn constraint.
- [ ] Owner-only independent script grants, settings and immutable active script versions.
- [ ] Thread/turn command/query API and UI, claims/reassignment, escalation/resolve/expiry.
- [ ] Go AI dispatcher plus stale-output acceptance and versioned editable reply proposals.
- [ ] Simulated outbound intents/attempts and final send gate; concurrency/unknown-result tests.

## Design and implementation references

- [03-database.md](../../designs/system/03-database.md)
- [04-http-api.md](../../designs/system/04-http-api.md)
- [07-workflows.md](../../designs/system/07-workflows.md)

## Evidence and verification boundary

Local simulator inbox implementation started on 2026-11-02; [session](sessions/2026-11-02_messaging-inbox.md) records scope and bounded verification. Real Pancake remains separate EPIC-008 dependency.

Status records the last verified local/PR snapshot, not live GitHub state. Recheck merge/CI before promoting to done.

## Sessions

- [2026-11-02 simulator inbox](sessions/2026-11-02_messaging-inbox.md) — in-progress/completed false; local code, no merge.

## Remaining work and next handoff

Review the bounded simulator/manual inbox, run two-profile browser acceptance and explicitly migrate/rebuild the local application. Follow with turn lifecycle, scripts/proposals and durable outbound/Kafka integration. Existing legacy AI order drafts are not messaging proposals.

## Connector design follow-up — 2026-11-03

Direct Facebook now targets a separate connector service; see [primary session](../EPIC-008-integrations-deployment/sessions/2026-11-03_connector-design.md). Normalized Kafka events feed the API consumer; business authority stays in API. Design only, no new runtime verification.

Follow-up: owner rejected the initial connector proposal. [Omni Bridge-based redesign](../EPIC-008-integrations-deployment/sessions/2026-11-03_connector-omni-review.md) remains under review; no delivery mechanism is approved by that draft.

## Active Desk control slice — 2026-11-10

- IMPLEMENTED locally/unmerged: projected BOT_ACTIVE / WAITING_STAFF / STAFF_ACTIVE, durable staff queue, reasoned handoff and explicit bot return, generation/version fencing and transactional audit. Reuses existing schema and current active-member/assignee/owner policy.
- Desk UI exposes waiting work, claim, handoff and return. The simulator starts in staff queue; legacy unassigned staff work remains visible. No online/offline action changes control.
- [Session](sessions/2026-11-10_chatbot-desk-control.md) records isolated PostgreSQL/race, HTTP and frontend evidence. Granular Desk/Page entitlements, bot settings, AI candidate/send gates, cancellation, live connector, notifications, handoff-history UI and browser acceptance remain open. This slice does not complete EPIC-004.
