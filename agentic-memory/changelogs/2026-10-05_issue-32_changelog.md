# Task Changelog: Default Light Mode with Seamless Dark Mode Toggle Support

- **Issue**: [#32](https://github.com/TruongHoang2004/Hexta/issues/32)
- **Date**: 2026-10-05
- **Branch**: `task/issue-32-feat-ui-set-light-mode-as-default-theme-`
- **Scope**: `packages/ui`, `apps/web`

---

## 1. Summary of Changes
Switched the primary interface theme to Light Mode as the default experience across the web application while fully preserving dark mode functionality with a responsive theme toggle mechanism.

## 2. Key Technical Decisions
1. **Encapsulated Theme Provider & Toggle (`packages/ui`)**:
   - Implemented `ThemeProvider` wrapping `next-themes` provider to standardize theme context across monorepo apps.
   - Implemented `ThemeToggle` component with Sun/Moon transition icons, client-mount hydration guard, and accessible labeling.
   - Exported both components in `packages/ui/src/index.ts`.
2. **Tailwind CSS v4 Dark Variant & Variable Mapping (`apps/web/app/globals.css`)**:
   - Added `@custom-variant dark (&:where(.dark, .dark *));` to ensure Tailwind utilities (`dark:...`) resolve properly with class-based toggling.
   - Established crisp light theme tokens on `:root` as the primary default and dark tokens under `.dark`.
3. **Root Layout Configuration (`apps/web/app/layout.tsx`)**:
   - Configured `ThemeProvider` with `attribute="class"`, `defaultTheme="light"`, and `enableSystem={false}` to guarantee light mode is the initial default.
4. **Accessible Navigation Toggles (`apps/web`)**:
   - Integrated `<ThemeToggle />` into `AuthNav` for landing page visitors and authenticated users.
   - Added `<ThemeToggle />` into the workspace dashboard overview header in `apps/web/app/(dashboard)/tenant/page.tsx`.

## 3. Impacted Files
- `packages/ui/src/components/theme-provider.tsx` (new)
- `packages/ui/src/components/theme-toggle.tsx` (new)
- `packages/ui/src/index.ts`
- `apps/web/app/globals.css`
- `apps/web/app/layout.tsx`
- `apps/web/components/auth-nav.tsx`
- `apps/web/app/(dashboard)/tenant/page.tsx`
- `agentic-memory/plans/2026-10-05_issue-32_plan.md`
- `agentic-memory/reviews/2026-10-05_issue-32_review.md`

## 4. Verification & Testing
- `pnpm --filter "./packages/*" run build`: Success in <1s.
- `pnpm --filter web run lint`: Zero errors or warnings.
- `pnpm --filter web run build`: Successfully generated production Turbopack build with all routes prerendered.
