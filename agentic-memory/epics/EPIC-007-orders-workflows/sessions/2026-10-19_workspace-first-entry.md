---
epic: EPIC-007
session: 2026-10-19_workspace-first-entry
status: in-progress
completed: false
date: 2026-10-19
---

# Workspace-first application entry

## Scope

Owner requested explicit workspace selection before order management. Preserve the existing uncommitted manual-sales implementation.

## Outcome and decisions

Added /workspaces as a dedicated chooser without the business sidebar. Login, registration, OAuth completion, authenticated auth-page redirects and workspace entry links now land there. Workspace cards explicitly select a business and open /orders. Creation can create/open a workspace. Removed automatic first-workspace fallback and old localStorage restoration. A deliberate selection is retained only in sessionStorage across reloads, cleared on login/logout; membership list still validates the stored selection. Direct management navigation without a selected workspace redirects to the chooser without mounting data screens. In-workspace navigation shows the active business and a Change workspace link. The chooser exposes the current account user ID for joining an existing workspace.

## Verification and evidence

Web lint and TypeScript checks passed. No backend/schema/API changes. Local Docker rebuild was attempted twice, but automatic approval review failed because its selected model was at capacity; the action was not executed and this was not an unsafe-action determination. No alternative privileged route was used. The running container still serves the preceding UI, so new-flow browser verification remains pending.

## Remaining work and handoff

Run scripts/local.sh up --build -d --wait web after approval review is available (or have the owner run it), then verify sign-in -> chooser -> selected workspace orders, direct URL without selection, reload and Change workspace. Review/merge remain pending. No completed state is claimed.
