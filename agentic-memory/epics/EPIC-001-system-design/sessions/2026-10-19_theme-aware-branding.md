---
epic: EPIC-001
session: 2026-10-19_theme-aware-branding
status: in-review
completed: false
date: 2026-10-19
---

# Theme-aware Hexta branding

## Scope and starting evidence

Owner approved the generated hexagonal emblem/lowercase wordmark light and dark variants and requested integration. Branch: codex/manual-sales-demo; existing uncommitted sales/onboarding work preserved. No PR or merge claimed. Cross-epic impact: EPIC-007 workspace shell.

## Outcome and decisions

IMPLEMENTED: approved PNG files copied unchanged into apps/web/public/brand/hexta-{light,dark}.png. Shared components/brand-logo.tsx switches assets using the existing class-based dark variant, without client theme hooks. A fixed aspect frame hides source transparent margins without stretching artwork; the accessible label is Hexta. Integrated landing navigation/footer, login/register, workspace/account shell and invitation entry. No API/DB/business-rule changes.

Generation provenance: built-in ImageGen. Light prompt: preserve blue hexagonal H emblem and lowercase hexta, royal blue #2563EB/navy #0F172A, transparent background. Dark edit prompt: preserve geometry/layout; change emblem to #3B82F6 and wordmark to #F8FAFC, transparent background. Original generated PNGs retained; no regeneration during integration.

## Verification and evidence

- pnpm --dir apps/web lint: passed.
- pnpm --dir apps/web exec tsc --noEmit: passed.
- git diff --check: passed.
- Browser verified local dev server at 127.0.0.1:3100: landing logo visible in light and dark after theme toggle; dark login logo visible with accessible Hexta label. No authenticated workspace runtime check performed.
- Source files both 1983 x 793. Generated imagery is raster, not SVG.

## Remaining work and handoff

Local delivery ready for review; merge into main not verified. Parent epic remains in-progress. Existing Docker demo requires rebuilding to serve updated assets.
