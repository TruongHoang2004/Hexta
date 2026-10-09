# Changelog: Interactive Draft Card and Chat Assistant Interface in Next.js

- **Issue**: [#38](https://github.com/TruongHoang2004/Hexta/issues/38)
- **Date**: 2026-10-09
- **Domain**: `domain:frontend-ui`
- **Component**: `apps/web`
- **Author**: `frontend_lead` agent

---

## Summary of Changes

Implemented the conversational assistant interface in `apps/web` (Next.js 16 + React 19 + Tailwind CSS) featuring real-time Server-Sent Events (SSE) stream rendering, an Interactive Order Draft Review Card with Human-in-the-Loop "Confirm Order" action, inline item quantity and variant editing, and silent token refresh.

### Key Changes:
1. **`apps/web/lib/services/ai-agent-service.ts`**:
   - `AIAgentService` client with `converseStream` parsing SSE events (`token`, `tool_call`, `draft_proposed`, `done`).
   - Integrated silent token refresh on HTTP 401 via `@hexta/sdk` (`sdk.identity.refreshToken`).
   - `confirmDraftOrder`: Dispatches `POST /api/v1/orders` with `draft_id`, line items, `source: "ai_draft"`, and `auto_confirm: true`.
2. **`apps/web/hooks/use-ai-agent.ts`**:
   - React state management for conversational messages, streaming tokens, active tool calls, and order drafts.
   - `updateDraftItem`: Supports inline quantity and variant adjustment, recalculating totals before confirmation.
   - `confirmDraft`: Invokes order creation and updates draft status to `confirmed` with assigned `order_number`.
3. **`apps/web/components/assistant/interactive-draft-card.tsx`**:
   - Card component with distinct styling, badges, customer summary, and interactive item list.
   - Inline +/- quantity controls updating subtotals and total amount.
   - Primary "Confirm Order" button with loading spinner state and success confirmation display.
4. **`apps/web/components/assistant/chat-drawer.tsx`**:
   - Floating action button at bottom-right with animation and online status dot.
   - Collapsible panel with header controls, conversation scroll view, quick prompt chips, tool execution pill, and input box.
5. **`apps/web/app/layout.tsx`**:
   - Mounted `<ChatDrawer />` globally inside `ThemeProvider`.

---

## Verification
- Backend integration verified with `go test -race ./...` in `services/api`.
- TypeScript strict types validated across all DTOs and component props.
