---
epic: EPIC-004
session: 2026-11-03_inbox-components
status: in-review
completed: false
date: 2026-11-03
---

# Tenant inbox design — PR #61

- Delivery: [PR #61](https://github.com/TruongHoang2004/Hexta/pull/61), OPEN/unmerged when checked on 2026-11-03. Owner requested reference-based design, tenant integration and readable component boundaries.
- IMPLEMENTED: `/tenants/[tenantId]/inbox/preview`, **Inbox preview** menu entry and expanded layout. Components/hooks/types/sample data live in `apps/web/components/inbox/`; original [HTML concept](../../../designs/prototypes/2026-11-03_admin-inbox.html) remains available.
- Behavior: queue/channel filters, search, thread selection, per-conversation reply/note drafts, local previews, closed-reply guard and customer context; narrow screens use list/detail navigation. Lifecycle, assignment, control and escalation stay separate. Workspace tenant remount clears preview state.
- Boundary: illustrative data only; no persistence, provider calls or real sending. Rebased from `468995b` onto `2e5f18e`, preserving upstream simulator `/inbox`. Initial backend-absence statements applied only to the older snapshot; the final design route uses `/inbox/preview`.
- Verification: prototype `node --check`, Prettier, web lint, generated Next.js types, TypeScript and build passed; lint/build rerun after rebase with both routes present. Initial missing `LayoutProps` resolved through `next typegen`. Browser/responsive/accessibility acceptance NOT_RUN; CI success and merge not asserted.
- Handoff: review with an authenticated tenant; future data/send integration must preserve backend authorization, versions and idempotency. Parent epic criteria remain open.
- Memory cleanup: owner requested consolidation of this PR's repetitive Markdown. This record combines concept, integration, refactor and publication evidence; original versions recoverable at `d0e7747`. Cleanup checks cover local links, stale references and whitespace; application code unchanged.
