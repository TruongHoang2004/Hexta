---
epic: EPIC-007
title: Approved order scope and flexible business workflows
status: in-progress
completed: false
updated: 2026-10-19
---

# Approved order scope and flexible business workflows

## Scope

Adapt legacy order code to approved incomplete-draft/explicit-confirmation rules and optional workflow steps.

## Decisions and context

- Keep messaging separate from orders; packing/shipping are not mandatory for every business.
- Incomplete order drafts allowed. Assigned sale confirms after named customer/contact, product, positive quantities, explicit nonnegative price and delivery address when applicable.
- Assigned handler progresses cancellation/return; owner finalizes. Preserve original order/line history; no payment/refund execution scope.
- Existing payment flags, inventory reservation and auto_confirm are legacy behavior to audit, not approved new MVP requirements.

## Acceptance checklist

- [x] Local compatibility/scope audit and additive migration for legacy order services/DTOs; unmerged.
- [x] Local incomplete draft update and explicit guarded confirmation, PostgreSQL race checks and browser demo; unmerged.
- [ ] Versioned optional workflows, handoff/accept/assignment and audit history.
- [ ] Cancellation/return quantity/ownership/finalization cases and tests.

## Design and implementation references

- [03-database.md](../../designs/system/03-database.md)
- [04-http-api.md](../../designs/system/04-http-api.md)
- [07-workflows.md](../../designs/system/07-workflows.md)

## Evidence and verification boundary

The manual D1–D3 sales slice is implemented and locally verified on codex/manual-sales-demo; merge/review remain pending. Existing merged foundations do not complete the parent epic.

Status records the last verified local/PR snapshot, not live GitHub state. Recheck merge/CI before promoting to done.

## Sessions

- [2026-10-19_manual-sales-demo](sessions/2026-10-19_manual-sales-demo.md) — in-progress; completed: false.

## Remaining work and next handoff

Review/merge the manual sales slice. Next product slice is configurable responsibility handoff (D4), after feedback on the order demo. Cancellation/return cases and dashboard remain separate; do not inherit legacy stock/payment effects.

## Workspace-first follow-up

[2026-10-19_workspace-first-entry](sessions/2026-10-19_workspace-first-entry.md) — in-progress, completed: false. Explicit chooser and route gating implemented; lint/type checks pass. Runtime rebuild/browser verification pending because automatic approval review failed at model capacity, not because the action was deemed unsafe.

## Tenant-first planning dependency

The owner's corrected onboarding/product hierarchy is recorded in [EPIC-001's framework session](../EPIC-001-system-design/sessions/2026-10-19_tenant-first-framework.md). Preserve the sales implementation, but complete the approved tenant registration/approval/invitation foundation and tenant-scoped navigation before expanding workflows. No duplicate implementation session is recorded here.

## Tenant-first implementation follow-up

[Implementation session](../EPIC-001-system-design/sessions/2026-10-19_tenant-first-implementation.md) — in-progress; completed: false. Approval-first onboarding, manual email-addressed invitation acceptance, independent operator authority and tenant-qualified sales navigation/API are implemented locally. Automated checks pass; live privileged approval verification and delivery review/merge remain open. This supersedes previous pending-build/direct-create entry descriptions without rewriting their sessions.

## Shared branding follow-up

[EPIC-001 branding session](../EPIC-001-system-design/sessions/2026-10-19_theme-aware-branding.md) records the light/dark shared logo applied to the workspace shell. No sales behavior changes; local/unmerged.

Manual acceptance and PR delivery are tracked in the [primary handoff session](../EPIC-001-system-design/sessions/2026-10-19_demo-pr-handoff.md); use the [checklist](../../../infrastructure/TESTING.demo.md) for sales, tenant and invitation checks.

Sales/tenant/branding delivery is in [PR #54](https://github.com/TruongHoang2004/Hexta/pull/54), open with CI pending at publication. Manual acceptance and merge are not complete; broader workflow scope remains in-progress.
