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
