---
epic: EPIC-004
title: Conversation domain, scripts and reply/send policy
status: in-progress
completed: false
updated: 2026-11-03
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

Local simulator inbox implementation started on 2026-11-02; [session](sessions/2026-11-02_messaging-inbox.md) records scope and bounded verification. Direct Facebook transport is tracked in EPIC-008; live acceptance remains separate.

Status records the last verified local/PR snapshot, not live GitHub state. Recheck merge/CI before promoting to done.

## Sessions

- [2026-11-02 simulator inbox](sessions/2026-11-02_messaging-inbox.md) — in-progress/completed false; local code, no merge.

- [2026-11-03 inbox design](sessions/2026-11-03_inbox-components.md) — [PR #61](https://github.com/TruongHoang2004/Hexta/pull/61), in-review/completed false. Sample UI at `/inbox/preview`; simulator `/inbox` preserved. Verification and next steps are in the session.

## Remaining work and next handoff

Review the bounded simulator/manual inbox, run two-profile browser acceptance and explicitly migrate/rebuild the local application. Follow with turn lifecycle, scripts/proposals and durable outbound/Kafka integration. Existing legacy AI order drafts are not messaging proposals.

Connector integration: [2026-11-03 primary session](../EPIC-008-integrations-deployment/sessions/2026-11-03_connector-ingress.md). Local text ingress/Kafka/API consumer implemented and tested on isolated infrastructure; AI/outbound and merge remain open.
