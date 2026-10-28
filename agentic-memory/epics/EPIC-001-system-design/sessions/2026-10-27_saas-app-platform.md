---
epic: EPIC-001
session: 2026-10-27_saas-app-platform
status: in-progress
completed: false
date: 2026-10-27
---

# SaaS app ecosystem discovery

## Thesis-first roadmap follow-up

Owner cannot build all three products simultaneously and prioritized Hexta for a 1Dec2026 thesis milestone, with features in Notion. Applied dev-plan and notion-spec-to-implementation; read HEX-1/4/6, queried sprint milestones and outstanding tasks. [Roadmap](../../../plans/2026-10-27_thesis-first-ecosystem_plan.md) preserves all 13 FR/5 NFR groups, proposes a five-week thesis-first allocation with capacity/feature-freeze checkpoints, then supervised Hexta use and one companion app at a time. PostPilot-first is a tentative recommendation, not an owner decision. Notion marks 1Dec as an internal full-draft/MVP handoff, not confirmed defense logistics. No Notion writes, task creation, runtime changes, live PR/CI refresh or new completion claims.

## Scope

Owner described Hexta as the SaaS center for businesses, with ViralForge and PostPilot as connected apps, and requested a proposed architecture drawing. Primary epic: EPIC-001. This is a discussion artifact, not an implementation commitment.

## Starting evidence

Current branch codex/demo-work-units. Existing untracked EPIC-007 evidence preserved. Read all three projects' README/architecture material and selected current routes, domain/pipeline source and recent memory. Live PR/CI and provider state were not refreshed and are not asserted.

## Outcome and decisions

Owner DECIDED platform direction; [design proposal](../../../designs/2026-10-27_saas-app-platform_design.md) recommends shared tenant identity/app access, separate app data ownership, Hexta sales/operations as a module, and explicitly approved ViralForge-to-PostPilot artifact handoff. SSO, entitlements, billing, events and tenant hardening are TARGET/OPEN. Existing runtime architectures remain unchanged.

## Verification and evidence

Documentation-only local link and whitespace checks; no application tests or runtime changes. Cross-project review is bounded source inspection, not a security or production acceptance audit.

## Remaining work and handoff

Discuss packaging, tenant experience and app boundaries with owner before implementation planning. Review/merge remains unverified; parent epic remains in-progress/completed false. No modifications to PostPilot or ViralForge.

## Owner correction in follow-up

Owner found the initial layering confusing: Hexta emphasizes orders/business operations, PostPilot fanpage management, ViralForge content creation. Updated the proposal and canonical architecture with the correction; retain earlier sections as historical proposals. Do not treat a separate Hexta Platform/Sales product split as approved. Shared account architecture and detailed fanpage/conversation scope remain OPEN. Documentation only; git whitespace check passed, no runtime tests.
