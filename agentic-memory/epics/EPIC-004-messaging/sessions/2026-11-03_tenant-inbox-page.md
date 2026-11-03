---
epic: EPIC-004
session: 2026-11-03_tenant-inbox-page
status: in-review
completed: false
date: 2026-11-03
---

# Tenant workspace inbox page

## Scope and starting evidence

Owner requested integration of the approved inbox concept into tenant management. Branch `TruongHoang2004/ui-ux`, starting commit `468995b`; preserved previous concept and memory edits. Read frontend/verification rules, nested web instructions and installed Next.js client/page guidance after installing locked dependencies. No API/schema/authentication contract changes or remote lifecycle operations.

## Outcome and decisions

- IMPLEMENTED: `/tenants/[tenantId]/inbox`, React inbox component and tenant navigation entry. Existing workspace tenant resolution/access state gates rendering; tenant-keyed workspace remount clears local draft state on tenant changes.
- Queue/channel filtering, search, selection, per-conversation reply/note drafts, local previews, closed-reply guard and customer context. Narrow screens show list/detail navigation and expandable customer context. Workspace expands on the inbox route to accommodate panels.
- Data remains explicitly sample-only. No provider calls, real sends, persistence, assignment commands or AI behavior. Production tenant-isolated data and send permissions remain TARGET.

## Verification

- `pnpm install --frozen-lockfile`: passed; generated changes to tracked dependency launchers restored.
- `pnpm --filter web lint`: passed.
- Initial `pnpm --filter web exec tsc --noEmit` needed generated `LayoutProps` from a fresh install; `pnpm --filter web exec next typegen` then TypeScript passed.
- `pnpm --filter web build`: passed, including dynamic inbox route and TypeScript.
- `git diff --check`: passed.
- Authenticated browser interaction, responsive visual and accessibility acceptance NOT_RUN. No production messaging acceptance or GitHub merge verified.

## Remaining work and handoff

Local code in review, completed false pending delivery/merge. Review page with a real tenant session. Wire ingestion/query/send contracts only when messaging backend is available; all existing epic acceptance criteria remain open.
