---
epic: EPIC-001
session: 2026-10-19_brand-favicon
status: in-review
completed: false
date: 2026-10-19
---

# Icon-only brand and browser favicon

## Scope and starting evidence

Owner requested an icon-only logo and browser tab icon. Clean starting tree on codex/manual-sales-demo at 37b9934. Previous branding is present in code; no merge claim is made.

## Outcome and decisions

IMPLEMENTED locally: public/brand/hexta-icon.png is a transparent icon-only asset generated using built-in ImageGen from the approved light logo. Prompt: retain only the blue hexagonal H emblem, preserve geometry, remove wordmark, center on a square transparent canvas with small safe margins. Generated master is 1254 x 1254 RGBA. Sharp packages the master into app/icon.png (192 x 192) and replaces app/favicon.ico (16/32/48px PNG frames). Next.js file metadata conventions supply icon links automatically. The same blue icon is used on light/dark browser tabs.

## Verification and evidence

- Web lint and TypeScript noEmit passed.
- Browser inspection on local dev server verified icon links for favicon.ico and icon.png, with versioned URLs; icon.png rendered the Hexta emblem.
- ICO directory validated for three frames; PNG master has alpha. git diff --check passed.

## Remaining work and handoff

Review/merge pending; status in-review and completed false. Parent epic remains in-progress. Rebuild existing Docker deployment to serve updated assets.
