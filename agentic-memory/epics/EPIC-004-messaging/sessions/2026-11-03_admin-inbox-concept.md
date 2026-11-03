---
epic: EPIC-004
session: 2026-11-03_admin-inbox-concept
status: in-review
completed: false
date: 2026-11-03
---

# Admin inbox concept

## Scope

A reviewable interactive inbox layout based on the owner-supplied reference screenshot. Primary scope is visual design, not provider collection or message dispatch.

## Starting evidence

Branch `TruongHoang2004/ui-ux`, starting commit `468995b`, clean working tree. EPIC-004 had no prior session. Canonical workflows describe messaging as TARGET. Existing web routes contain sales/workflow screens; no inbox route was found. No GitHub/CI/merge verification performed.

## Outcome and decisions

- IMPLEMENTED as a standalone prototype: [inbox concept](../../../designs/prototypes/2026-11-03_admin-inbox.html), also displayed through an Orca inline visual.
- TARGET layout: application rail, inbox/channel navigation, searchable conversation list, message thread/composer and customer context.
- Local interactions: search, channel/queue filtering, conversation selection, per-conversation reply/note drafts and preview entries. Closed conversations disable reply previews but allow notes. No persistence or external requests.
- Separate lifecycle, assignment, control and escalation displays reflect approved workflow dimensions. Viewing/admin status does not establish sending permission; preview entries are not actual sends.
- Channel names and customer data are illustrative, not verified provider capabilities. No backend/API/schema contract changes.

## Verification and evidence

- `node --check /tmp/hexta-inbox-check.js`: JavaScript syntax check (extracted inline script).
- `git diff --check`: whitespace check.
- Static inspection: self-contained HTML/CSS/JS, no remote libraries/network requests, user draft text inserted with textContent, accessible search/draft labels and focus outlines.
- Browser rendering/interaction and application lint/build NOT_RUN; standalone design does not modify the Next.js app. Root Playwright dependency unavailable. Visual acceptance remains for owner review.
- Selected root router and verification rules; frontend conventions consulted for possible integration, but no apps/web or SDK edits made.

## Remaining work and handoff

Review layout and desired channel scope with owner; then integrate into the tenant workspace with actual permission and data contracts. Provider ingestion, real sending, responsive production behavior, accessibility/browser acceptance and all existing epic acceptance criteria remain open. Local/unmerged concept; parent epic is in-progress, not complete.
