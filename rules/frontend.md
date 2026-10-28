# Frontend and SDK conventions

Existing approved rules extracted from AGENTS.md for conditional loading. Scope: apps/web and packages/sdk. This file does not add frontend architecture beyond the established instructions.

- Use strict TypeScript; avoid explicit any. Run `pnpm --filter @hexta/sdk lint` for SDK type enforcement.
- Use Tailwind CSS and Next.js App Router under apps/web/app; apply nested AGENTS instructions, including installed Next.js documentation before writing affected code.
- Keep API access in apps/web/lib/services and the shared SDK; preserve backend DTO validation requirements in UI inputs.
- Browser refresh credentials remain API-owned HttpOnly cookies, never browser JSON/JavaScript storage. Logout must succeed server-side before UI reports completion. Read API authentication rules only when touching that behavior.
- Match existing style, English repository artifacts/comments and documented API contracts. API contract changes additionally select DTO/Swagger/response rules through the root router.
- Run relevant SDK tests/types/build, web lint/type/build or focused behavioral tests according to the actual change. Record prerequisites/skips; no automatic claim of full browser acceptance.

- Preserve applicable Suspense boundaries around useSearchParams and centralize API URLs in configured services/SDK; do not hardcode backend hosts in components. Read installed-version Next.js guidance for actual API/convention changes.
