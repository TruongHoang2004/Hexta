# Frontend applications

Applies to all apps; app-local instructions add only meaningful differences. Inspect the target manifest and installed Next.js documentation required by its generated block. Current apps use App Router, React, strict TypeScript and Tailwind; do not infer other frontend frameworks or data libraries.

## Discover before building

Before implementing UI, inspect [shared UI exports](../packages/ui/src/index.ts), search target-app components, and read candidate component APIs and existing usage. Prefer compatible existing components over new primitives. Explain an actual API/behavior gap before introducing a shared component; keep business compositions in the owning app. Semantic HTML is allowed and encouraged where appropriate; reuse is not a ban on native elements. Use [component discovery](../.agents/skills/frontend-component-discovery/SKILL.md) for a substantial UI addition, or reuse its already-completed result.

## Components, data and styles

- Keep server/client boundaries intentional; add client behavior only where needed. Inspect neighboring routes, layouts and state ownership before extending them.
- Use strict TypeScript without explicit `any`. Match current app state/data patterns; do not assume all apps have the web app's SDK or Zustand integration.
- Preserve semantic structure, accessible names/labels, keyboard operation, visible focus and status announcements. Check loading, empty, error, disabled and success states, responsive layout and light/dark appearance where supported.
- Match existing tokens, spacing and component variants before adding arbitrary styling. Read [UI guidance](../packages/ui/AGENTS.md) when modifying the shared package, not for every consumer edit.
- Keep async errors visible and prevent stale responses from replacing current state. Preserve cancellation and retry semantics; retries must not duplicate mutations.
- Align input constraints with the actual API contract; frontend validation never grants authorization. Cross-contract changes select affected service/package rules through the root router.

## Validation

Inspect the target package scripts. Run applicable app lint/build and focused checks for changed behavior; exercise keyboard, responsive and error paths when relevant. No app interaction test harness is established by a passing SDK test. Record manual checks and unavailable coverage accurately; do not invent test commands. Preserve Next.js-generated instruction blocks.
