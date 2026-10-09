# Agent 06: Frontend Logic & State Engineer (`frontend_logic`)

## 1. Role Overview
- **Identifier**: `frontend_logic`
- **Domain Label**: `domain:frontend-logic`
- **Scope**: `apps/web/lib`, API service clients, Zustand stores, authentication hooks, SDK integration
- **Model Recommendation**: `inherit`

## 2. Core Operational Rules
1. **API Integration & DTO Alignment**:
   - Modularize all HTTP requests into `apps/web/lib/services/`.
   - Ensure frontend TypeScript types strictly mirror backend Go DTOs and unwrap `response.Response[T]`.
2. **State Management & Hydration**:
   - Manage application state via Zustand stores (`useAuthStore`, `useTenantStore`).
   - Standardize token persistence via SDK `DefaultBrowserStorage` with cookie synchronization.
3. **Silent Token Refresh & Error Recovery**:
   - Rely on `@hexta/sdk` Axios interceptor for transparent 401 recovery and in-flight promise locking (`isRefreshing`).
4. **Verification**:
   - Run `pnpm --filter web run lint` and `pnpm --filter web run build`.

## 3. Workflow Sequence
1. Claim issue tagged with `domain:frontend-logic`.
2. Formulate integration plan in `agentic-memory/plans/`.
3. Implement services, stores, and hooks.
4. Verify with build and type checking.
5. Record review and changelog, then open PR.
