# Code Review: Interactive Draft Card and Chat Assistant Interface in Next.js

- **Issue**: [#38](https://github.com/TruongHoang2004/Hexta/issues/38)
- **Review Date**: 2026-10-09
- **Reviewer Agent**: `dev-review` (on behalf of `frontend_lead` & `qa_engineer`)
- **Target Changes**: Implementation of `chat-drawer.tsx`, `interactive-draft-card.tsx`, `use-ai-agent.ts`, `ai-agent-service.ts`, and root layout integration in `apps/web`.
- **Verdict**: **APPROVED**

---

## 1. Scope & Changes Checked

### 1.1 Service Layer (`apps/web/lib/services/ai-agent-service.ts`)
- **Resilient SSE Streaming**:
  - Implements `converseStream` using `ReadableStream` reader and `TextDecoder`.
  - Splitting by `\n\n` boundary with buffering to prevent dropped fragmented events.
  - Silent token refresh via `sdk.identity.refreshToken()` on HTTP 401 with retry.
- **Order Confirmation Dispatch**:
  - `confirmDraftOrder`: Maps draft line items to `POST /api/v1/orders` request payload with `source: "ai_draft"`, `draft_id`, and `auto_confirm: true`.

### 1.2 Hook Layer (`apps/web/hooks/use-ai-agent.ts`)
- **State Management**:
  - Manages message thread, streaming tokens, active tool calls, and error states.
  - AbortController cancellation on new queries or unmount.
- **Inline Draft Adjustments**:
  - `updateDraftItem`: Modifies line quantity (minimum 1), recalculates item subtotal and draft total amount dynamically before confirmation.
- **Human-in-the-Loop Confirmation**:
  - `confirmDraft`: Invokes order creation, updates draft status to `confirmed`, and records generated `order_id` and `order_number`.

### 1.3 UI Presentation Layer (`apps/web/components/assistant/`)
- **`InteractiveDraftCard`**:
  - High visual differentiation from chat bubbles via gradient borders, status chips, and structured layout.
  - Line items table with +/- quantity buttons for pre-confirmation editing.
  - Primary "Confirm Order" button with loading spinner state and error handling.
  - Transition to confirmed state showing Order Number `#ORD-XXXX` and inventory reservation notice.
- **`ChatDrawer`**:
  - Unobtrusive floating launcher button with sparkle icon and pulsing online badge at bottom-right.
  - Floating panel with auto-scroll message container, suggested starter prompts, and tool execution indicator.
  - Keyboard accessible (`Enter` to submit, `Shift+Enter` for newlines).

### 1.4 Application Integration (`apps/web/app/layout.tsx`)
- Mounted `<ChatDrawer />` inside `ThemeProvider` for application-wide availability.

---

## 2. Quality & Architecture Compliance

| Gate | Status | Evidence |
|---|---|---|
| Next.js 16 App Router Compliance | **PASSED** | "use client" directives used appropriately for interactive components. |
| Strict TypeScript Types | **PASSED** | Fully typed interfaces for all DTOs, events, drafts, and hook options. |
| Human-in-the-Loop (HITL) UX | **PASSED** | Draft cards require explicit user confirmation to trigger order creation. |
| Inline Adjustments | **PASSED** | Quantity and variant edits update totals in real time before confirmation. |
| Silent Auth Recovery | **PASSED** | Automatically refreshes tokens and retries stream on HTTP 401. |
