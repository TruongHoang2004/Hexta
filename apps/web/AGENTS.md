<!-- BEGIN:nextjs-agent-rules -->

# This is NOT the Next.js you know

This version has breaking changes — APIs, conventions, and file structure may all differ from your training data. Read the relevant guide in `node_modules/next/dist/docs/` (resolved from this file's directory; in monorepos the `next` package may not be visible from the repo root) before writing any code. Heed deprecation notices.

This block is written and re-added by `next dev` — verify at `node_modules/next/dist/server/lib/generate-agent-files.js`. Removing it from a diff only re-creates the uncommitted change; committing it with your work keeps the tree clean.

<!-- END:nextjs-agent-rules -->

# Web application

Read [shared frontend instructions](../AGENTS.md). Web-specific business compositions belong in `components/`; inspect sales and messaging implementations before adding patterns. Routes and layouts live in `app/`.

- Keep API access in `lib/services/` and [the configured SDK](lib/sdk.ts); do not hardcode backend hosts in components. SDK changes also require [SDK guidance](../../packages/sdk/AGENTS.md).
- Preserve Zustand auth-state ownership in `store/useAuthStore.ts` and existing local React state/effects for page interactions. Do not introduce a server-state library implicitly.
- Preserve tenant route/resource scope, version/conflict and idempotency behavior. Browser identity/state is not business authorization. Read API authentication rules through the router when session/cookie/logout behavior changes; refresh credentials remain API-owned HttpOnly cookies and logout must succeed server-side before success is reported.
- `app/globals.css` currently owns theme variables and Tailwind mapping, including shared UI source scanning. Inspect shared primitives and app-local sales primitives before adding styling. Do not claim a separate token package exists.
- Preserve applicable Suspense boundaries for `useSearchParams`. Verify affected Next.js APIs against installed documentation.
- Validate with applicable `pnpm --filter web lint` and `pnpm --filter web build`; SDK changes use its checks too. Record browser acceptance separately from lint/build.
