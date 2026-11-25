---
epic: EPIC-004
session: 2026-11-26_needs-human-turn-expiry
status: in-progress
completed: false
date: 2026-11-26
---

# Needs-human flag, 24-hour turn expiry and waiting notifications

## Scope

PRD MVP 5 FR-5.6 / FR-5.7 (business rule 3, "Turn expiry" acceptance), MVP 4 FR-4.12 (notifications for waiting conversations) and the S-07 Desk additions for the needs-human badge/panel, closed-turn divider and waiting count. Bot proposal cards are out of scope (separate flow). Product rules: `designs/system/01-product-decisions.md` "Turn expiry" and "Chatbot and Desk — owner-approved flow 2026-11-09".

## Starting evidence

Branch from `main` at `b2d886b`. `needs_human` already existed on turns/views but was cleared by claim, had no reason, and nothing closed turns. The conversation view only held open turns and ordered snapshots by turn version only.

## Outcome and decisions

IMPLEMENTED (unmerged):

- `messaging_turns.needs_human_reason` / `closed_at` and view `needs_human_reason` (migration `20261126300000_messaging_turn_expiry`). Handoff records the reason; claim keeps the flag; `POST /turns/{id}/resolve` is allowed only to the assigned sale, clears the flag, restarts the inactivity window and keeps staff control (no control-generation bump). `HandlingState` now treats sales + assignee as `STAFF_ACTIVE` regardless of the flag.
- Turn expiry job (`TurnExpiryService`, `bootstrap.BuildTurnExpiry`) in `cmd/messaging-command`, every minute, batch 100: thread `FOR UPDATE SKIP LOCKED`, pending intents locked, reread, conditional close on the exact observed version/activity. Closing bumps version and control generation, cancels pending intents older than the connector TTL + 1 min (their sending messages become `unknown`), audits `messaging.turn_expired` as `system:turn-expiry` and keeps history. A recent in-flight send defers closing.
- Activity: customer messages (source time, never backwards), simulated replies, and confirmed channel sends at the provider outcome time (`DeliveryService` → `RecordSendActivity`, no version bump). Pending/failed/unknown/duplicate results do not count; channel send creation no longer touches activity.
- Ingress placement (`types.InboundPlacement`): a message not newer than the closed turn's activity, or older than a full window, is kept on the closed turn as history and opens nothing; otherwise a new unassigned turn opens.
- Read model keeps each thread's latest turn (closed included); an older turn's snapshot never overwrites a newer turn. Filters other than `all` list open turns; new `waiting` filter; `GET /conversations/summary` returns `waiting_staff_count`.
- Desk: header button polling the summary (30 s, visible tab) that opens the waiting filter, count in the filter chip, needs-human badge and right panel with reason/resolve, turn dividers, closed-turn read-only reason, vi/en copy.

Decisions for the owner (not previously recorded):

- Explicit return-to-bot still clears needs-human (owner or assignee), as before; resolve is not required first.
- New turns created without a bot still start `WAITING_STAFF` with `needs_human=true` (no reason), so a claimed conversation stays open until the sale resolves it. Changing new-turn defaults belongs to the reply-mode work.
- Owner unassignment still sets needs-human so queued work never expires unattended.

## Verification and evidence

Disposable `postgres:17.7-alpine3.23` container via `API_TEST_DATABASE_URL`; no application database touched.

- `sh scripts/check-gofmt.sh` PASS; `go vet` (API/shared) PASS; `go test -race ./services/api/... ./packages/shared/...` PASS including new DB tests in `turn_expiry_service_test.go`.
- `make migrate-hash svc=api`, `make migrate-validate svc=api` PASS; `atlas migrate diff --env api` reports no drift.
- `make swagger` regenerated.
- `pnpm --filter web lint`, `tsc --noEmit`, `pnpm --filter web build` PASS; SDK lint/test/build PASS (unchanged).
- Browser acceptance of the Desk changes: NOT_RUN.

## Remaining work and handoff

Merge into `main` (atlas.sum re-hash expected at integration), browser acceptance on S-07, and owner confirmation of the three decisions above.
