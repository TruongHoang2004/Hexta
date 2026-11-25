---
epic: EPIC-004
session: 2026-11-26_reply-mode-proposals
status: in-progress
completed: false
date: 2026-11-26
---

# Reply mode, per-message bot router and Desk proposals

## Scope

PRD MVP 5 FR-5.1, FR-5.2, FR-5.3 (bot handling flow, "Suggest mode" acceptance, business rules 1 and 4), screens S-19 and the S-07 proposal card with its stale state. Product rules: `designs/system/01-product-decisions.md` "Conversation policy" and "Chatbot and Desk — owner-approved flow 2026-11-09"; AI contract `designs/system/05-grpc-ai.md`. Needs-human panel, closed-turn divider and waiting count belong to the needs-human/turn-expiry session.

## Starting evidence

Branch from `main` at `b2d886b`, with the scripts session (`ScriptService.ActiveScriptVersions`) and the needs-human/turn-expiry session merged in locally. No Go production AI run client existed (only `cmd/ai-smoke`); takeover did not cancel pending sends; new turns always started `WAITING_STAFF`.

## Outcome and decisions

IMPLEMENTED (unmerged):

- `tenant_reply_settings` and `messaging_reply_proposals` (migration `20261126200000_reply_mode_and_proposals`). `GET/PUT /tenants/{tenantID}/settings/reply-mode`: members read, owner changes with `expected_version`; audited.
- New-turn default: once the owner has saved a reply mode, new turns start `BOT_ACTIVE`; before that the pre-bot `WAITING_STAFF` default stays.
- Customer messages on `BOT_ACTIVE` turns (ingress and simulator) enqueue a durable routing job in the same transaction; a newer message supersedes the turn's live proposal. `ReplyRouterService` (`bootstrap.BuildReplyRouter`, messaging-command, 1 s idle poll) decides after commit: deterministic `MatchScript` → script answer; otherwise Python `AIService` over gRPC (`infrastructure/ai.RunClient`, `ai.run_addr`), non-blocking StartRun then GetRun polling (≈80 s bound, CancelRun on timeout); AI route `script` must name an active version and uses Go's content; `human`, failure or unconfigured AI hand off through the shared `applyHandoff` transition with a reason. The decision is rechecked under the thread lock (status, control generation, context revision).
- `auto_send` sends through the shared outbound pipeline (`enqueueChannelReply`; simulated message for simulator threads) only while the turn is `BOT_ACTIVE` at the job's generation and revision; `autoSendEligible` is the Q5.2 seam.
- Proposal commands: `POST /turns/{id}/proposals/{proposalID}/send` (as-is or edited, idempotent `client_message_id`), `/discard`, `POST /turns/{id}/regenerate-proposal`. Stale = generation mismatch or closed turn; never sendable. Detail response carries `proposals` with `stale`.
- Control-changing commands and router handoffs cancel the turn's undispatched bot sends (intent pending and outbox event still `pending` → event `failed`, intent/attempt `cancelled`, message `failed/cancelled`); leased/published sends stay pending for reconciliation.
- Web: Settings › Reply mode page and the Desk proposal card (routing, pending, stale/regenerate), bot sender label, vi/en.

Decisions for the owner:

- Sending a proposal from bot handling self-claims the turn for the sender (the only sender stays the assigned sale); the bot then stops routing that turn until it is returned.
- "Bot configured" = a saved reply-mode row; there is no separate on/off switch.
- Regenerate exists (wireframe "Tạo lại"); a person-requested proposal is never auto-sent and never hands off.
- AI generation is opted in but no business facts are supplied, so the AI path currently yields clarification or handoff (OPEN: facts source, model-based script evaluation).

## Owner decisions 2026-11-26 (D1–D3)

The owner replaced the interim "saved row = bot on" rule and the all-candidates auto_send gate; see `designs/system/01-product-decisions.md` "Reply mode decisions — owner 2026-11-26". IMPLEMENTED (unmerged):

- D1: mode `off` (default; missing row reads `off`, existing rows keep their mode). `newTurnHandling` starts BOT_ACTIVE only for suggest/auto_send. `botIntake` (simulator and channel ingress) hands a BOT_ACTIVE turn to staff with reason `bot_off` when a customer writes while the bot is off, so no routing job is created; the router skips script/AI work while off and `apply` rechecks the mode under the thread lock, dropping a queued job and handing its turn off (`botHandoff`, audited `messaging.handoff`). Existing turns are not moved by the switch itself; their pending proposals stay sendable/discardable. Return-to-bot and regenerate are refused with 409 while off; the Desk hides both actions.
- D2: `autoSendEligible` = script route or clarify; AI `generate` answers always stay pending for review.
- D3: `messaging_turns.review_pending` is set when a proposal is left pending on a BOT_ACTIVE turn and cleared on send, discard, supersede (new customer message, regenerate) and every control change (DB CHECK: only on open BOT_ACTIVE turns). It blocks expiry (`ExpiryDue`, candidate scan, conditional close) and is projected through `messaging.conversation.changed` (`review_pending`, additive/omitempty) to the view; the waiting count/filter is `assignee_id IS NULL AND (control = 'sales' OR review_pending)` with a matching partial index. Turn version bumps with every flag change, so replays keep ordering.
- Web S-19: three options (Tắt / Gợi ý / Tự động gửi) on `@hexta/ui` ToggleGroup, auto_send copy says AI-written answers always wait for staff review, confirmations kept; Desk list shows a "proposal awaiting review" chip and the `bot_off` reason label.

Superseded decision above: "Bot configured = a saved reply-mode row" no longer holds; `configured` in the API now only says whether a row exists.

## Verification and evidence

Disposable `postgres:17.7-alpine3.23` container via `API_TEST_DATABASE_URL`; no application database touched. gofmt, `go vet`, `go test -race ./services/api/... ./packages/shared/...` PASS; Atlas diff/validate PASS; `pnpm --filter web lint`/`build` PASS. Browser acceptance and live Go↔Python routing NOT_RUN.

D1–D3 (2026-11-26): disposable DB on container `hexta-integ-pg` only. gofmt, `go build`, `go vet`, `go test -race -count=1 ./services/api/... ./packages/shared/...` (incl. new `reply_mode_decisions_test.go` and docs contract) PASS; `atlas migrate diff`/`make migrate-hash`/`make migrate-validate` and `atlas schema diff` (synced) PASS; `make swagger` regenerated; `pnpm --filter web lint`/`build`, `tsc --noEmit` PASS. Browser acceptance of S-19/Desk NOT_RUN.

## Remaining

Compose `AI_RUN_ADDR` wiring, conversation-list "proposal waiting" badge, merge into `main`.
