---
epic: EPIC-001
title: System design and repository context
status: in-review
completed: false
updated: 2026-10-19
---

# System design and repository context

## Scope

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
