<!-- BEGIN:nextjs-agent-rules -->

# This is NOT the Next.js you know

This version has breaking changes — APIs, conventions, and file structure may all differ from your training data. Read the relevant guide in `node_modules/next/dist/docs/` (resolved from this file's directory; in monorepos the `next` package may not be visible from the repo root) before writing any code. Heed deprecation notices.

This block is written and re-added by `next dev` — verify at `node_modules/next/dist/server/lib/generate-agent-files.js`. Removing it from a diff only re-creates the uncommitted change; committing it with your work keeps the tree clean.

<!-- END:nextjs-agent-rules -->

# Web application

Read [shared frontend instructions](../AGENTS.md). Routes and layouts live in `app/`.

## Domain modules (required)

Owner decision 2026-11-21: business code is grouped into domain modules matching the sidebar. A module owns every page of one domain.

| Module | Owns |
| --- | --- |
| `modules/conversations` | inbox (desk), consultant, storefront/checkout simulation |
| `modules/sales` | orders, customers, products/catalog |
| `modules/operations` | work items, workflows |
| `modules/settings` | members, channels, service plan |

The workspace shell is layout, not a domain: `components/layout/` owns the dashboard shell, sidebar and its groups (`nav-config.ts`), workspace switcher and `useWorkspace` context.

Layout per module: `components/` (UI, one component per file), `hooks/` (data and state), `services/` (API calls through [the configured SDK](lib/sdk.ts)), `types.ts`, optional `utils.ts` (pure helpers, no React), and `index.ts` exporting only what routes or other modules use.

- `app/**/page.tsx` stays thin: render a screen imported from a module's `index.ts`.
- Modules import each other only through `index.ts`. Code used by several modules lives in `components/` (shared web UI such as auth nav, brand, page header) or `lib/` (SDK, i18n, errors).
- A new domain gets a new module and a sidebar group in the same change; do not park it in an unrelated module.
- Do not add domain code to `components/<domain>` or `lib/services/<domain>.ts`. `lib/services/onboarding.ts` serves account-level pages outside a workspace (registration, invitations, platform review).

## Other web rules

- Do not hardcode backend hosts. SDK changes also require [SDK guidance](../../packages/sdk/AGENTS.md).
- Preserve Zustand auth-state ownership in `store/useAuthStore.ts` and existing local React state/effects for page interactions. Do not introduce a server-state library implicitly.
- Preserve tenant route/resource scope, version/conflict and idempotency behavior. Browser identity/state is not business authorization. Read API authentication rules through the router when session/cookie/logout behavior changes; refresh credentials remain API-owned HttpOnly cookies and logout must succeed server-side before success is reported.
- `app/globals.css` currently owns theme variables and Tailwind mapping, including shared UI source scanning. Inspect shared primitives and app-local `components/primitives.tsx` before adding styling. Do not claim a separate token package exists.
- Preserve applicable Suspense boundaries for `useSearchParams`. Verify affected Next.js APIs against installed documentation.
- Validate with applicable `pnpm --filter web lint` and `pnpm --filter web build`; SDK changes use its checks too. Record browser acceptance separately from lint/build.
