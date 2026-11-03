---
epic: EPIC-004
session: 2026-11-03_inbox-components
status: in-review
completed: false
date: 2026-11-03
---

# Inbox component refactor

## Scope and starting evidence

Owner requested readable, maintainable component boundaries after reviewing the tenant inbox implementation. Preserved existing local workspace/page/prototype/memory changes on `TruongHoang2004/ui-ux` (HEAD `468995b`). Previous session: [tenant page](2026-11-03_tenant-inbox-page.md). Frontend and verification rules apply; Go layer boundaries are not applicable to this React-only change. Installed Next.js client guidance already consulted in this session chain.

## Outcome and decisions

- Split the monolithic inbox into header, filters, conversation list, conversation panel, message thread, message composer, customer details and shared avatar components. Each component declares typed props and explicit callbacks.
- `inbox.tsx` composes the layout; `use-inbox.ts` owns selection, filters, drafts and local previews. Shared types, sample data and the pure queue predicate have dedicated modules.
- Formatted all inbox TypeScript/JSX for readable markup. Preserved existing sample-only behavior, responsive classes, tenant mount boundaries and closed-reply guard. No API/schema/auth/provider change.

## Verification

- `pnpm dlx prettier --write 'apps/web/components/inbox/*.{ts,tsx}'`: passed.
- `pnpm --filter web lint`: passed.
- `pnpm --filter web exec tsc --noEmit`: passed.
- `pnpm --filter web build`: passed, including inbox route.
- `git diff --check`: passed.
- Browser/visual interaction acceptance not run. No merge verified; local delivery remains in-review/completed false.

## Remaining work and handoff

Review component split and verify authenticated browser interactions. Real provider/data/send integration and parent messaging acceptance criteria remain open.

## Publication follow-up — reconciliation with current main

- Fetched `origin/main` at `2e5f18e`; it now includes the simulator/manual messaging implementation and its existing `/tenants/[tenantId]/inbox` route. Earlier statements about backend absence describe the original `468995b` workspace snapshot only.
- Rebased this change onto current main, preserving upstream simulator code, route, navigation and memory. Final design route is `/tenants/[tenantId]/inbox/preview`, linked as **Inbox preview**; only that route expands the workspace layout. The original `/inbox` page remains the simulator UI.
- Reran web lint and production build after reconciliation: passed, both inbox routes included. Browser acceptance remains NOT_RUN. No production integration or merge completion asserted.

- Published [PR #61](https://github.com/TruongHoang2004/Hexta/pull/61) against `main`; verified OPEN/non-draft. CI was queued at publication; no CI-success or merge assertion.
