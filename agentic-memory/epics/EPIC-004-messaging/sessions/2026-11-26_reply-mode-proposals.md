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

## Verification and evidence

Disposable `postgres:17.7-alpine3.23` container via `API_TEST_DATABASE_URL`; no application database touched. gofmt, `go vet`, `go test -race ./services/api/... ./packages/shared/...` PASS; Atlas diff/validate PASS; `pnpm --filter web lint`/`build` PASS. Browser acceptance and live Go↔Python routing NOT_RUN.

## Remaining

Compose `AI_RUN_ADDR` wiring, conversation-list "proposal waiting" badge, merge into `main`.
