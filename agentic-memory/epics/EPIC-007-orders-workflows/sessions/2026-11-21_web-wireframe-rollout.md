---
epic: EPIC-007
session: 2026-11-21_web-wireframe-rollout
status: in-progress
completed: false
date: 2026-11-21
---

# Web wireframe rollout (MVP 1–4) — phase 1: i18n, error text, primitives, shell

## Scope

Bring the PRD 2.1 wireframes (`docs/specs/prd/`, low-fidelity canvas kept outside the repo) into `apps/web` for the Built slices MVP 1–4, in phases: (1) foundations, (2) small backward-compatible API additions, (3–6) screens per MVP. MVP 5–6 stay TARGET. Desk work (phase 6) also affects [EPIC-004](../../EPIC-004-messaging/README.md). This session covers phase 1 only.

## Starting evidence

Branch `feat/web-shell-i18n` from `origin/main` `0f0c192`. PRD split committed separately on `docs/prd-mvp-split` (`fa70c41`). Before this change the web UI was English-only with hardcoded copy, a flat 11-link sidebar shown to every role (including the static inbox preview), no role in the workspace context and no support code in error text.

## Outcome and decisions

- DECIDED (owner, 2026-11-21): bilingual UI vi/en, Vietnamese default; scope MVP 1–4 plus small API additions.
- IMPLEMENTED: dependency-free typed dictionaries in `apps/web/lib/i18n/` (`vi.ts` is the key source; `en.ts` must match the `Messages` shape at compile time), `I18nProvider`/`useT`/`useSetLocale`, locale cookie `hexta.locale` read by the root layout (`<html lang>`). All routes now render dynamically because the root layout reads a cookie.
- IMPLEMENTED: `lib/errors.ts` (`describeError`, `useErrorText`, `isConflict`) maps 401/403/404/409 to localized text and appends the API `trace_id` as a support code; 4xx business details are shown as returned. Existing screens still use `errorText` from `lib/services/sales.ts` until migrated.
- IMPLEMENTED: app-local primitives in `components/sales/primitives.tsx` (`StatusChip`, `EmptyState`, `ConflictBanner`, `SegmentedFilter`, `Tabs`, `ConfirmDialog` on native `<dialog>`, `ChecklistItem`); no new dependency.
- IMPLEMENTED: workspace shell with grouped sidebar (main + settings), role label, `role`/`isOwner` in `useWorkspace()` (derived from `owner_id` until the API exposes the caller role), settings tab layout, VI/EN switch in `AuthNav`, Vietnamese font subset. The inbox preview link was removed from navigation; the route remains until phase 6.

## Verification and evidence

- `pnpm --filter web lint`: pass.
- `pnpm --filter web build`: pass (after `pnpm install --frozen-lockfile`; the workspace had missing package dependencies and no SDK `dist`).
- `next start` smoke check: `/login` renders `<html lang="vi">` by default and `lang="en"` with cookie `hexta.locale=en`.
- NOT_RUN: signed-in browser acceptance (sidebar, role label, settings tabs, keyboard, dark mode, 390 px width).

## Remaining work and handoff

Phase 1 screens are only partially translated (shell, nav, auth controls). Next per plan: phase 6 (Desk), phase 4 (order checklist + 409), API additions 2b/2c, then phase 5 and the rest. Migrate each screen from `errorText` to `useErrorText` as it is reworked. Not pushed; no PR yet.

## Follow-up — phase 6 (Desk) and API addition 2c

Branch `feat/web-desk` (stacked on `feat/web-shell-i18n`).

- IMPLEMENTED (API, backward compatible): `GET /tenants/:id/conversations?filter=bot` returns open turns whose projection is `BOT_ACTIVE` (`control = system`, unassigned, not `needs_human`). DTO `oneof`, service validation, repository query, controller Swagger text and regenerated `services/api/docs` updated. "Waiting for staff" keeps using the existing `needs_human` filter.
- IMPLEMENTED (web): `components/messaging/inbox.tsx` rebuilt to the wireframe layout — conversation list with Bot / Waiting / Mine / All filters and state chips, thread with header actions (claim, hand over via dialog with the five reason codes, return to bot), localized delivery states including a distinct "delivery unknown" chip (no automatic resend), read-only reason above the composer, customer side panel with draft-order/consultant links and owner assignment. Send idempotency (`request_key`, `client_message_id`, uncertain retry) and `expected_version` handling are unchanged. Expired-connection banner localized.
- Removed the static design mock: route `tenants/[tenantId]/inbox/preview` and `components/inbox/*` (sample data only, no other consumers).
- OPEN: system lines for claim/handoff events need an events feed the API does not expose; the draft-order button opens Orders without linking the conversation (no backend link); assignee names still show IDs until addition 2a.

Verification: `sh scripts/check-gofmt.sh`, `go vet ./services/api/...`, `go test -race ./services/api/...` pass; DB scenarios (including the extended `TestMessagingPaginationAndFilters` bot assertions) SKIPPED/NOT_RUN without `API_TEST_DATABASE_URL`. `make swagger` regenerated (description-only diff). `pnpm --filter web lint` and `build` pass (after clearing stale `.next/dev/types`). Browser acceptance NOT_RUN.

## Follow-up — phase 4 (orders) and API addition 2e

Branch `feat/web-orders` (stacked on `feat/web-desk`).

- IMPLEMENTED (API, backward compatible): `GET /tenants/:id/orders?status=draft|confirmed&assignee=me|unassigned`. Filters are applied in the tenant-scoped repository query (so totals and pagination stay correct); unknown values return 400; tenant membership is still checked first. `types.OrderListFilter` added; Swagger regenerated. New DB test `TestSalesOrderListFilters` covers own/peer `me`, unassigned, status, invalid values and a foreign tenant (403).
- IMPLEMENTED (web): order list with All / Drafts / Confirmed / Unassigned / Mine filters, status chips, localized empty states and "order value" wording; editor with a live confirmation checklist (saved, assigned sale, customer name, contact, items, prices, delivery address) computed from the current form, Confirm disabled with the reason, a separate "Claim order" for unassigned drafts, and a conflict banner that reloads the latest version on HTTP 409. The server gate remains authoritative. Pager localized.
- OPEN: multi-variant product editing needs an API change (product requests carry a single variant); not done.

Verification: gofmt, `go vet ./services/api/...`, `go test -race` for service/repository/presentation packages pass; DB scenarios SKIPPED/NOT_RUN without `API_TEST_DATABASE_URL`. `pnpm --filter web lint` and `build` pass. Browser acceptance NOT_RUN. Remaining phase-4 screens (customers/products copy, consultant quick replies) not yet reworked.

## Follow-up — phase 4 remainder (catalog, consultant widget)

- Customers/products screen localized with empty states; products show a distinct "price not set" chip versus "0 · free"; the variant name falls back to the localized default when left blank.
- Consultant widget localized and errors use the shared support-code text. Quick replies were already rendered as buttons.
- OPEN: a "dropped from the catalog" notice needs the API to report which AI items were removed (currently removed silently); storefront and checkout pages remain English.

Verification: `pnpm --filter web lint` and `build` pass. Browser acceptance NOT_RUN.

## Follow-up — phase 5 (work and workflows) and API addition 2b

Branch `feat/web-work` (stacked on `feat/web-orders`).

- IMPLEMENTED (API, backward compatible): `GET /tenants/:id/work-items?view=all|mine|awaiting_me`. `mine` = unfinished items the caller is responsible for (including while waiting for a recipient); `awaiting_me` = waiting items nominating the caller. Unknown views return 400. Unlike the plan's draft, `all` is not restricted to owners: members could already list all work and no requirement removes that. `TestWorkflowAcceptanceAndPinnedVersions` extended; Swagger regenerated.
- IMPLEMENTED (web): Work screen with My work / Waiting for me / All tabs (default My work), step progress track, handoff dialog explaining direct versus acceptance, Accept/Complete actions, owner "cancel pending handoff" (assign back to the current responsible, which the API already treats as cancelling the nomination) and reassignment, localized history and 409 reload banner; work creation moved into a collapsible owner panel. Workflows screen: step builder with direct/acceptance toggle, step counter n/30, running-work version note, explanation of handoff types; localized.
- OPEN: workflow version history needs the API to return older versions (definitions currently preload only the current version), so no history list is shown.

Verification: gofmt, `go vet ./services/api/...`, `go test -race` for service/repository/controller packages pass; DB scenarios SKIPPED/NOT_RUN without `API_TEST_DATABASE_URL`. `pnpm --filter web lint` and `build` pass. Browser acceptance NOT_RUN.

## Follow-up — phase 3 (workspace and access) and API addition 2a

Branch `feat/web-workspace` (stacked on `feat/web-work`).

- IMPLEMENTED (API, backward compatible): `GET /tenants/:id/users` adds optional `email` per member, read from the member's local or Google identity (local preferred) and scoped by a join on the tenant's members. Only current members can list (unchanged 403 otherwise). New repository method `TenantMemberEmails` (DB wrapper passthrough), service unit test and repository DB assertion (including an outsider identity that must not appear). Swagger regenerated. SDK `TenantMember.email?` added; SDK lint/test/build pass.
- DECISION TO CONFIRM (owner): every current member can now see teammates' account emails (needed to name assignees in Desk, Orders and Work). Previously only user IDs were visible.
- NOT DONE by design: planned addition 2d (`my_role` on `GET /tenants`) — the picker derives owner/member from `owner_id` and the signed-in user, so no API change was needed.
- IMPLEMENTED (web): shared `useMemberName` (You / email / short ID / former member) used in Desk (list, read-only reason with the handler's name, assignment), Orders and Work; members now load for all roles in Desk. Workspace picker localized with a role chip, accessible status chips and Inbox as the entry route. Invitation page shows distinct wrong-account (403) and no-longer-valid (404/409) states with a switch-account action. Operator review page has Pending / Approved / Rejected tabs with counts and a reject dialog that requires a reason. Members page shows emails, role chips, invitation status chips, a copy button for the one-time link; all localized.
- Remaining English screens: register (S-03), channels (S-18), plan (S-20), storefront/checkout, login/register.

Verification: gofmt, `go vet ./services/api/...`, `go test -race` for service/repository/controller pass (DB scenarios SKIPPED/NOT_RUN without `API_TEST_DATABASE_URL`); `pnpm --filter @hexta/sdk lint|test|build` pass; `pnpm --filter web lint` and `build` pass. Browser acceptance NOT_RUN.
