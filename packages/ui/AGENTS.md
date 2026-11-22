# Shared UI components

This package exports React source from `src/index.ts`. Inspect exports, existing usage and the implementation before adding a primitive or changing its API. Shared UI owns reusable presentation; tenant/business workflows and service calls belong in applications.

- Preserve compatible exports, props, ref behavior and existing composition, including Button `asChild` via Radix Slot. Prefer explicit variants using existing CVA and `cn` helpers over duplicate components or class-string branches.
- Use existing semantic token classes. Web currently supplies theme variables/mapping in `apps/web/app/globals.css`; there is no separate shared token package. Check actual consuming apps before changing token expectations or Tailwind scanning.
- Keep native element semantics, accessible names, focus, keyboard interaction and disabled states. Composition must not silently remove button/link behavior or create nested interactive controls.
- Components follow the shadcn/ui pattern on Radix primitives (owner decision 2026-11-22); Radix and `cmdk` dependencies live in this package, not in apps.
- Add shared primitives only for a justified reusable responsibility; do not pull app-specific forms, permissions or fetched business state into the package.
- Search imports from `@hexta/ui` and direct source imports before changing exports/variants. Preserve supported React peers and avoid app-only dependencies.
- Inspect available lint/type tooling; this package has no established interaction suite. Validate affected consumer builds and relevant manual accessibility/theme/responsive behavior. Do not claim the package lint script works without checking its tool/config prerequisites. State coverage gaps rather than adding a framework implicitly.
