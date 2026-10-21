---
epic: EPIC-001
title: System design and repository context
status: in-progress
completed: false
updated: 2026-10-20
---

# System design and repository context

## Scope

Latest discovery: [DDD domain map session](sessions/2026-10-20_ddd-domain-map.md), in-progress/completed false. Nine candidate business contexts and adoption guidance proposed; documentation only, local/unmerged. Parent acceptance and implementation states unchanged.

Consolidate approved product/system decisions, provide a reliable architecture reference, and keep agent context organized by epic/session.

## Decisions and context

- Demo/MVP first; retain implemented versus decided/target distinctions. Go owns business authorization and sending, Python executes AI; messaging and orders remain separate.
- The modular system design is the architecture reference. AGENTS.md owns common instructions; epic directories own progress and handoff. GEMINI.md is an entry redirect to AGENTS.md.
- Requirements/flows/technology documentation was previously approved; that does not mean target features exist in code.

## Acceptance checklist

- [x] Eight system-design modules cover product, architecture, DB, HTTP, gRPC/AI, Kafka, workflows and operations.
- [x] Physical migration catalog, router inventory and Protobuf contract are reconciled to the inspected baseline.
- [x] Common rules moved to AGENTS.md and epic/session memory structure written.
- [ ] Commit/review/merge documentation follow-up; verify its final PR state before marking this epic done.

## Design and implementation references

- [2026-10-18_system-design_design.md](../../designs/2026-10-18_system-design_design.md)
- [01-product-decisions.md](../../designs/system/01-product-decisions.md)
- [02-architecture.md](../../designs/system/02-architecture.md)

## Evidence and verification boundary

HEX-6 2e3c653546b342b3a90a994a760f4340; HEX-7 8d451c4ec8f54ca58cf2d9628a0b72af; HEX-8 cfa16b89d09f4211a7fd9e7dc1b5d8e8; HEX-9 af0f29236ac948a286e2e59b7379a854; PRD 3f6b0ea49f098117aa38d0421ce163b0.

Status records the last verified local/PR snapshot, not live GitHub state. Recheck merge/CI before promoting to done.

## Sessions

- [2026-10-18_architecture-consolidation](sessions/2026-10-18_architecture-consolidation.md) — in-review; completed: false.
- [2026-10-19_epic-session-organization](sessions/2026-10-19_epic-session-organization.md) — in-review; completed: false.

- [2026-10-19_documentation-pr](sessions/2026-10-19_documentation-pr.md) — in-review; completed: false.

- [2026-10-19_pr-53-conflicts](sessions/2026-10-19_pr-53-conflicts.md) — in-review; completed: false.

## Remaining work and next handoff

PR #52 is confirmed merged on 2026-10-19; PR #53 now targets main. Resolve/review its documentation merge with main and confirm #53 merge/CI before completing this epic.

## 2026-10-19 planning follow-up

PR #52 and #53 merge is now verified; their reported checks succeeded. Earlier pending-merge notes above remain historical. New [backlog session](sessions/2026-10-19_next-work-backlog.md) is in-review, completed: false, with a [proposed implementation sequence](../../plans/2026-10-19_next-work-backlog_plan.md). The new planning artifact still requires delivery; implementation epics are not completed by this reconciliation.

## Owner-directed demo-first revision

[Reprioritization session](sessions/2026-10-19_demo-first-reprioritization.md) — in-review; completed: false. The active backlog now prioritizes manual workspace/customer/product/order code, followed by responsibility workflows and a real-data overview. Key/mock-data projects and Kafka/AI infrastructure are deferred from the first-demo critical path. This supersedes the simulator-first recommendation for immediate execution, without removing approved future scope.

## Manual sales implementation follow-up

The locally implemented D1–D3 first-demo slice is recorded once in [EPIC-007's primary session](../EPIC-007-orders-workflows/sessions/2026-10-19_manual-sales-demo.md). Shared architecture modules and the active plan now reflect the current sales boundary. This is unmerged delivery and does not complete planning or implementation epics.

## Tenant-first framework correction

[2026-10-19_tenant-first-framework](sessions/2026-10-19_tenant-first-framework.md) — in-progress; completed: false. Owner confirmed approval-first tenant registration and owner-issued email invitations with explicit recipient acceptance. Tenant is the paying business boundary; the proposed nested route/permission framework replaces the previous unrestricted create-and-open onboarding direction. No implementation in this session.

## Tenant-first implementation follow-up

[Implementation session](sessions/2026-10-19_tenant-first-implementation.md) — in-progress; completed: false. Approval-first onboarding, manual email-addressed invitation acceptance, independent operator authority and tenant-qualified sales navigation/API are implemented locally. Automated checks pass; live privileged approval verification and delivery review/merge remain open. This supersedes previous pending-build/direct-create entry descriptions without rewriting their sessions.

## Theme-aware branding

[Branding session](sessions/2026-10-19_theme-aware-branding.md) — in-review; completed: false. Approved light/dark logo PNGs and shared theme-aware web component integrated; lint, TypeScript and landing/login browser checks passed. Local/unmerged.

## Icon-only branding and favicon

[Favicon session](sessions/2026-10-19_brand-favicon.md) — in-review; completed: false. Icon-only transparent master, 192px app icon and multi-size ICO implemented and verified locally.

## PR and evening test handoff

[Delivery session](sessions/2026-10-19_demo-pr-handoff.md) tracks the icon commit, full demo PR and [manual checklist](../../../infrastructure/TESTING.demo.md). Review/manual acceptance/merge remain pending.

[PR #54](https://github.com/TruongHoang2004/Hexta/pull/54) is open/mergeable at publication; checks pending. Handoff session is in-review, completed false; parent epic remains in-progress.

## Thesis continuation

[Thesis session](sessions/2026-10-19_thesis-continuation.md) — in-review; completed: false. Existing Vietnamese LaTeX chapters populated from current Notion decisions and code evidence; structure and formatting configuration preserved. Compilation and heading/citation checks passed; author details, empirical results and final PDF visual review remain.

## Thesis author correction and official emblem

[Evidence/logo session](sessions/2026-10-19_thesis-evidence-logo.md) — in-review; completed: false. Owner clarified Notion is their own work; bibliography/citations removed. Insufficiently supported sections left empty, historical counts and invented methodology removed, and official-site UET emblem added at the existing cover slot. Existing LaTeX project compiles; heading order and layout files preserved. This supersedes the earlier thesis session's populated bibliography/results decisions.

## Thesis installation appendix removed

[Removal session](sessions/2026-10-19_thesis-remove-installation.md) — in-review; completed: false. Owner requested UI-only end-user scope; installation appendix and dangling reference removed. Existing main.tex compiles. No claim of completed cloud deployment.

## Thesis delivery request

[Delivery session](sessions/2026-10-19_thesis-delivery.md) — in-review; completed: false at commit. PR #54 is verified merged; thesis changes are delivered separately from current main. The owner explicitly authorized commit and merge; GitHub PR state records final delivery.

## Live ticket/merge reconciliation — 2026-10-19

[Reconciliation session](sessions/2026-10-19_ticket-reconciliation.md) — in-review, completed false. PR #54 is confirmed merged with successful checks; thesis PR #55 is also merged. Earlier pending-merge notes are historical. [All 42 ticket assessments and next work](../../plans/2026-10-19_ticket-reconciliation_plan.md) preserve partial versus complete scope. No Notion edits; parent epic remains in-progress.

## Sprint 2 repaired on Notion

[Repair session](sessions/2026-10-19_sprint-2-repair.md) records the owner-authorized assignment of HEX-27/23/20/18 to 19–25 October with focused scope and verified reciprocal links. Sprint is in-progress; no feature completion implied.

## Current code ERD — 2026-10-20

[ERD session](sessions/2026-10-20_current-code-erd.md) adds a 22-table code/migration diagram at af06cf4 with all 215 columns, physical/application relationship distinction and detailed domains. No schema changes; local/unmerged documentation, parent status unchanged.

## Post-merge standards audit — 2026-10-20

[Session](sessions/2026-10-20_code-rules-audit.md), [review](../../reviews/2026-10-20_code-rules-audit_review.md): PR #57 merge and green CI verified at main 2abd305. Fresh syntax/type/unit checks pass within documented DB skips; six existing security/contract/architecture/type/formatting gaps remain OPEN. Review artifacts local/unmerged; parent completion unchanged.

## Delegated compliance fixes — 2026-10-20

[Session](sessions/2026-10-20_rule-compliance-fixes.md): four module subagents implemented R1–R6 and supplemental logging/path cleanup; independent review caught and verified fixes for logout integration and Swagger prefixes. Go race suite80 tests/no skips, docs contract race test, vet, SDK35 tests/types/build, web lint/tsc, Atlas/format/shell/YAML checks pass. Local/uncommitted on codex/rules-compliance; browser E2E not rerun; parent completion unchanged.

## PR publication — 2026-10-21

[Publication session](sessions/2026-10-21_compliance-pr.md) prepares the compliance branch and pending documentation for review. Fresh checks pass within documented PostgreSQL/browser limits; parent status/completion unchanged.
