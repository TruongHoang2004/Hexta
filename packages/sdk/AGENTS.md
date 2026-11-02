# TypeScript SDK

`src/index.ts` is a handwritten Axios client, not generated OpenAPI output. `dist/` is tsup output; edit source and use the package build rather than editing bundles. Inspect `src/index.test.ts` for existing Vitest/axios-mock-adapter regressions.

- Preserve exported API/type compatibility, response-envelope handling and configured API URLs. Trace consumers in `apps/web/lib/sdk.ts` and `lib/services/` before changing contracts; some app-specific endpoint wrappers live there.
- Use strict TypeScript without explicit `any`; `pnpm --filter @hexta/sdk lint` invokes the custom type checker.
- Refresh credentials stay in API-owned HttpOnly cookies, never JSON or JavaScript storage. Storage adapters currently hold access tokens only. Preserve server-confirmed logout, refresh coordination and existing failure/retry semantics. Select authentication rules when touching these behaviors.
- Align DTO constraints, endpoint paths and errors with the Go API. Read the cross-service workflow for changes spanning owners; do not invent an SDK generator.
- Run relevant `pnpm --filter @hexta/sdk lint`, `pnpm --filter @hexta/sdk test` and `pnpm --filter @hexta/sdk build`; validate affected app consumers. Contract tests must cover behavior, not merely duplicate implementation.
