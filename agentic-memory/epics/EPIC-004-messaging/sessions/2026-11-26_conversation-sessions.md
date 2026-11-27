---
epic: EPIC-004
session: 2026-11-26_conversation-sessions
status: in-progress
completed: false
date: 2026-11-26
---

# Conversation sessions: session list, manual close, outcomes and linked orders

## Scope

Owner-approved 2026-11-26: a customer can have many chat sessions in one conversation. A session is the existing consultation turn (`messaging_turns`); it closes after the fixed 24-hour inactivity window or manually, and a genuinely new customer message opens a new unassigned turn under the current reply mode. Product rules: `designs/system/01-product-decisions.md` "Conversation sessions — owner 2026-11-26", "Turn expiry", "Chatbot and Desk" and "Reply mode decisions". PRDs: `docs/specs/prd/mvp-4-channels-desk.md`, `mvp-5-bot-replies.md`. Order flow itself is out of scope.

## Starting evidence

Branch `worktree-agent-a38cdcc88bd9506cf`, reset to `integrate/mvp5-6` at `ce9fef2`, later merged with `integrate/mvp5-6` at `2eaa46b` (reply mode `off`, `review_pending`). Orders had no conversation, thread or turn reference (only `customer_id`, `employee_id`, legacy `draft_id`); the Desk "Create draft order" button only linked to the Orders page.

## Outcome and decisions

IMPLEMENTED (unmerged):

- Schema (`20261126900000_conversation_sessions`, data backfill `20261126900100_backfill_conversation_sessions`, `20261126900200_conversation_sessions_close_cause`): `messaging_turns.close_cause` (`expired` | `closed_manually`, required exactly when closed), `closed_by`, `outcome` (CHECK set), `note` (≤ 1000 chars), `last_handler_id`; manual closes require closer and outcome. View gains `closed_at`, `close_cause`, `outcome`. `orders.messaging_turn_id` (nullable, composite FK `(tenant_id, messaging_turn_id)` → turns, partial index). Message index `(tenant_id, turn_id, id)`. Backfill marks existing closed turns `expired`, sets `last_handler_id` from the assignee and copies the session fields into the view.
- One closing write (`MessagingRepository.CloseTurn` with `types.TurnClosure`) for expiry and manual close: version and control generation bump, `needs_human`, `needs_human_reason` and `review_pending` cleared, cause/closer/outcome/note recorded, last handler kept. Expiry keeps its stricter observed-state condition (unflagged, no pending review, unchanged activity).
- `POST /turns/{id}/close` (`MessagingCommandService.CloseSession`): assigned sale under staff control or owner; outcome required, note ≤ 1000; `expected_version`; under the thread lock cancels undispatched sends of every author (`CancelUndispatchedSends`, generalizing `CancelPendingBotSends`); dispatched sends are reconciled later and never extend a closed turn. Audit `messaging.turn_closed_manually` records handling state, assignee, outcome, `has_note`, `closed_while_needs_human`, previous reason and cancelled sends (no note text).
- `POST /turns/{id}/outcome` (`UpdateSessionOutcome`): closed sessions only; owner or `last_handler_id`; version bump; audit `messaging.turn_outcome_updated`.
- `GET /conversations/{id}/sessions` (`MessagingQueryService.Sessions`): newest first, cursor = turn ID, limit 1–100; turn (cause, outcome, note, assignee, last handler), `opened_at` (earliest message), `message_count`, `first_message_id`, linked orders (id, number, status). Thread resolved in the tenant first (another tenant gets 404).
- Order link: `SalesService.SaveDraft` accepts `messaging_turn_id` on creation only; the turn must belong to the tenant and be open (share-locked); updates never change it. Order rules otherwise unchanged.
- `ConversationChanged` gains optional `closed_at`, `close_cause`, `outcome` (additive); projector/rebuild store them. A closed snapshot has a higher version than its open snapshots, so replays never resurrect a closed session.
- Desk: "Phiên" panel in the right column (newest first, current session highlighted, load more, jump to first message with bounded older-page loading), close dialog and outcome dialog (`Picker`, `Textarea`, `ConfirmDialog`), turn dividers labelled with each session's range, cause and outcome; "Create draft order" passes `?turn=` for an open session and Orders links the next new draft (with a "Do not link" option). vi/en copy.

Decisions for the owner:

- Outcome set (proposal): `order_placed`, `consulted`, `price_inquiry`, `complaint`, `no_response`, `other`.
- A draft links only to an open session; linking is fixed at creation.
- "Last handler" is the last staff assignee (kept after return to the bot); it, plus the owner, may edit a closed session's outcome. An owner who closes an unassigned session leaves no handler.
- A delayed customer message whose source time falls after the closed session's last activity (even if before the manual close) opens a new session rather than being folded into the closed one; exact replays are deduplicated by ingress identity.
- Manual close cancels undispatched staff replies as well as bot sends.

## Verification and evidence

Recorded in the final report of this session (all on disposable databases on the throwaway container `postgres://postgres:pg@localhost:55439`; no application database touched). Upgrade path checked on a disposable database: an existing closed turn and its view row were backfilled to `expired` with the assignee as last handler.

Browser acceptance of the Desk panel: NOT_RUN.

## Remaining work and handoff

Owner confirmation of the outcome set and the decisions above; browser acceptance of the panel, jump and dialogs; merge into `main` (re-hash `atlas.sum` if other migrations land first; these migrations must sort last).
