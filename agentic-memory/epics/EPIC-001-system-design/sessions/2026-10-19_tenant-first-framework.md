---
epic: EPIC-001
session: 2026-10-19_tenant-first-framework
status: in-progress
completed: false
date: 2026-10-19
---

# Tenant-first framework planning

## Scope and starting evidence

Owner corrected the product hierarchy: tenant is the paying business customer; all business resources belong within it and access follows invitation/membership. Inspected current tenants model, API router and workspace source on codex/manual-sales-demo, preserving existing uncommitted work.

## Outcome and decisions

Owner confirmed: customers apply for tenant registration and Hexta approves before access; tenant owners invite by email and recipients sign in/accept before activation. Saved [the framework plan](../../../ARCHIVE_INDEX.md#pruned-2026-10-19-tenant-first-framework-plan) with actor boundaries, nested routes, approval/invitation models, reuse of current sales code and five implementation slices. Approval provisions initial owner membership as the bootstrap interpretation; employee onboarding follows invitation acceptance. Email delivery/verification, operator bootstrap and billing policy remain open. No production payment/provider/key work is inferred.

## Verification and evidence

Planning only. Source inspection and local Markdown links/whitespace checked. No application code changed, runtime build/test, remote state refresh or tracker write. Existing tenant_id gates remain useful; current direct tenant creation/direct membership invite and plan/status mutation require replacement/restriction under the new workflow.

## Remaining work and handoff

Review framework with the owner before implementation. Select invitation delivery/verification and operator bootstrap in the detailed design. Start with approval authority and old-endpoint bypass closure, then applicant/tenant shell and invitation acceptance. Existing D4 workflow work follows this foundation. Plan and parent epics are unmerged/incomplete.
