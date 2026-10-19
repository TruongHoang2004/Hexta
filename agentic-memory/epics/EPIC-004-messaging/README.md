---
epic: EPIC-004
title: Conversation domain, scripts and reply/send policy
status: not-started
completed: false
updated: 2026-10-19
---

# Conversation domain, scripts and reply/send policy

## Scope

Implement tenant-isolated threads/turns/messages, assignment, script grants, proposals, result acceptance and simulator dispatch.

## Decisions and context

- All tenant sales see conversations; only assigned sale sends. Claim is atomic; owner can reassign. Takeover stops unstarted autosends, return to system is explicit.
- Owner proposal on means suggest/review; off means eligible auto_send. Routing chooses script/AI/clarify/human. Only owner grants four independent script permissions.
- Turn expiry is 24h; successful customer/sale/system activity counts, drafts/failure/unknown/duplicates do not. Needs-human exempts expiry. Resolve resets clock without resuming system control.
- Preserve human draft edits and check current context/control/script eligibility at final dispatch. In-flight remote sends may require reconciliation.

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

No implementation session yet. Requirements anchors are listed in EPIC-001; real Pancake remains separate EPIC-008 dependency.

Status records the last verified local/PR snapshot, not live GitHub state. Recheck merge/CI before promoting to done.

## Sessions

No work session recorded yet. Implementation has not started; do not invent historical progress.

## Remaining work and next handoff

Start a bounded simulator/inbox/outbox vertical slice jointly with Kafka. Existing legacy AI order drafts are not messaging proposals.
