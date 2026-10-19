---
epic: EPIC-007
title: Approved order scope and flexible business workflows
status: not-started
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

- [ ] Compatibility/scope audit and migration plan for legacy order services/DTOs.
- [ ] Incomplete draft update and explicit guarded confirmation.
- [ ] Versioned optional workflows, handoff/accept/assignment and audit history.
- [ ] Cancellation/return quantity/ownership/finalization cases and tests.

## Design and implementation references

- [03-database.md](../../designs/system/03-database.md)
- [04-http-api.md](../../designs/system/04-http-api.md)
- [07-workflows.md](../../designs/system/07-workflows.md)

## Evidence and verification boundary

Existing merged order foundations are not completion of this approved-scope adaptation. No new implementation session yet.

Status records the last verified local/PR snapshot, not live GitHub state. Recheck merge/CI before promoting to done.

## Sessions

No work session recorded yet. Implementation has not started; do not invent historical progress.

## Remaining work and next handoff

Audit existing order code before reuse; agree detailed case transitions/stock implications without silently importing payment or auto-confirm scope.
