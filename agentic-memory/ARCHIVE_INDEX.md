# Agentic Memory Archive Index

Historical lookup ledger for pruned completed/superseded documents. Entries identify replacement/recovery; they are not current implementation instructions or proof that every parent epic is complete. Do not load this ledger by default.

| Date | Category | Issue / Ref | Title / Objective | Archived On |
|---|---|---|---|---|
| 2026-09-12 | `plans` | - | Implementation Plan: Standardize Issue & Task Lifecycle Rules | 2026-10-09 |
| 2026-09-12 | `reviews` | - | Code Review: Google OAuth2 Authentication Flow | 2026-10-09 |
| 2026-09-12 | `reviews` | - | Code Review: Standardize Issue & Task Lifecycle Rules | 2026-10-09 |

## 2026-10-28 — Source-backed memory pruning

Owner authorized removal of obsolete/redundant Markdown after root rule routing was introduced. This ledger is historical lookup only, not default task context. Current rules: [root router](../rules/README.md). Current architecture: [system index](designs/2026-10-18_system-design_design.md).

GitHub read-only verification: issues #1, #3–8, #10, #18, #20–22, #26, #28–29, #32, #34–38 and #45 are closed; #56 remains open. PRs #9/#11/#52–55/#57/#58 are merged into main; #59 remains OPEN/draft. Issue closure is not asserted as current code correctness. Active #59 plans/review/unit evidence and unresolved current audit were retained.

Tracked files below can be inspected with `git show d09083c:<original repository-relative path>`. Untracked intermediate drafts were consolidated into retained rules/design/review; they were never committed and are not claimed recoverable through Git. SHA columns are the first 16 hexadecimal characters of the original SHA-256.

Retained substance: original API/automation/branding/thesis delivery is in CHANGELOG and code; original HEX-8/9 and AI/Kafka contracts are in modular system design; permissive/legacy order AI/stock behavior is not newly approved. The 2026-10-28 progression was API rules v0.1/73 IDs → mapper/shared-cache/root rules v0.2/91 IDs → selective routing v0.3/91 IDs. Mapper/cache mechanics remain proposed, root placement and mapper/wrapper direction owner-decided. Current audit retains fresh check results and six OPEN findings. No session/parent epic was promoted to done.

| Original file (under agentic-memory/) | Reason / replacement | Recovery / SHA prefix |
| --- | --- | --- |

| <a id="pruned-2026-10-05-codebase-audit"></a>`audits/2026-10-05_codebase-audit.md` | **Codebase Audit Report: 2026-10-05**. Routine scan superseded by the 2026-10-19 scan. [Current reference](audits/2026-10-19_codebase-audit.md) | Git; `b73a041131569a17` |

| <a id="pruned-2026-10-09-codebase-audit"></a>`audits/2026-10-09_codebase-audit.md` | **Codebase Audit Report: 2026-10-09**. Routine scan superseded by the 2026-10-19 scan. [Current reference](audits/2026-10-19_codebase-audit.md) | Git; `cb38609ae0ded5be` |

| <a id="pruned-2026-10-10-codebase-audit"></a>`audits/2026-10-10_codebase-audit.md` | **Codebase Audit Report: 2026-10-10**. Routine scan superseded by the 2026-10-19 scan. [Current reference](audits/2026-10-19_codebase-audit.md) | Git; `6faccdc5848f0b17` |

| <a id="pruned-2026-10-12-codebase-audit"></a>`audits/2026-10-12_codebase-audit.md` | **Codebase Audit Report: 2026-10-12**. Routine scan superseded by the 2026-10-19 scan. [Current reference](audits/2026-10-19_codebase-audit.md) | Git; `9919fe26ffbd6013` |

| <a id="pruned-2026-10-09-issue-34-changelog"></a>`changelogs/2026-10-09_issue-34_changelog.md` | **Changelog: Issue #34 — Core Business Schema Migrations & GORM Models**. Per-task changelog already synthesized in CHANGELOG.md. [Current reference](CHANGELOG.md) | Git; `70e4f5502a8c4dc2` |

| <a id="pruned-2026-10-09-issue-35-changelog"></a>`changelogs/2026-10-09_issue-35_changelog.md` | **Changelog: Inventory Repository, Stock Reservation, and Movement Tracking**. Per-task changelog already synthesized in CHANGELOG.md. [Current reference](CHANGELOG.md) | Git; `fccb5537ebbb1e6c` |

| <a id="pruned-2026-10-09-issue-36-changelog"></a>`changelogs/2026-10-09_issue-36_changelog.md` | **Changelog: Order State Machine, Pricing, and REST Endpoints**. Per-task changelog already synthesized in CHANGELOG.md. [Current reference](CHANGELOG.md) | Git; `4383d3ad9e8e5fb8` |

| <a id="pruned-2026-10-09-issue-37-changelog"></a>`changelogs/2026-10-09_issue-37_changelog.md` | **Changelog: AI Agent Engine with Tool Calling Registry and Draft Proposal Flow**. Per-task changelog already synthesized in CHANGELOG.md. [Current reference](CHANGELOG.md) | Git; `03872fae87ef39d2` |

| <a id="pruned-2026-10-10-issue-29-changelog"></a>`changelogs/2026-10-10_issue-29_changelog.md` | **Changelog: Unit Test Suites for Critical Core Services and SDK**. Per-task changelog already synthesized in CHANGELOG.md. [Current reference](CHANGELOG.md) | Git; `afec3dbe5b22fcdb` |

| <a id="pruned-2026-09-12-issue-18-local-dev-guide-design"></a>`designs/2026-09-12_issue-18-local-dev-guide_design.md` | **Technical Design: System Architecture & Local Development Orchestration**. Closed GitHub issue. [Current reference](CHANGELOG.md) | Git; `2ca5b77ddf46a664` |

| <a id="pruned-2026-09-12-issue-20-migrate-go-module-paths-design"></a>`designs/2026-09-12_issue-20-migrate-go-module-paths_design.md` | **Technical Design: Go Module Path & Package Namespace Migration**. Closed GitHub issue. [Current reference](CHANGELOG.md) | Git; `9d94bfb97f06f78f` |

| <a id="pruned-2026-09-12-issue-22-issue-lifecycle-sync-design"></a>`designs/2026-09-12_issue-22-issue-lifecycle-sync_design.md` | **Technical Design: Issue Lifecycle Manager & Parallel State Synchronization Workflow**. Closed GitHub issue. [Current reference](CHANGELOG.md) | Git; `aae9f1a4d74962c5` |

| <a id="pruned-2026-09-13-issue-21-clean-legacy-gitlab-references-design"></a>`designs/2026-09-13_issue-21-clean-legacy-gitlab-references_design.md` | **Technical Design: Documentation Standardization & Legacy Reference Elimination**. Closed GitHub issue. [Current reference](CHANGELOG.md) | Git; `5f0646f337a54b15` |

| <a id="pruned-2026-09-13-issue-26-pr-review-agent-design"></a>`designs/2026-09-13_issue-26-pr-review-agent_design.md` | **Technical Design: Autonomous PR Review Agent & Conflict Escalation Engine**. Closed GitHub issue. [Current reference](CHANGELOG.md) | Git; `f51005d8d58eaabf` |

| <a id="pruned-2026-10-15-order-cancellation-returns-design"></a>`designs/2026-10-15_order-cancellation-returns_design.md` | **Order Cancellation and Returns — Proposed Design**. Superseded by modular current architecture or root rule router. [Current reference](designs/system/07-workflows.md) | Git; `964e71b84e37d3fe` |

| <a id="pruned-2026-10-18-durable-ai-runs-design"></a>`designs/2026-10-18_durable-ai-runs_design.md` | **Durable AI runs and local application stack**. Superseded by modular current architecture or root rule router. [Current reference](designs/system/05-grpc-ai.md) | Git; `609b19fb7b8e1a12` |

| <a id="pruned-2026-10-18-hex-8-core-architecture-design"></a>`designs/2026-10-18_hex-8-core-architecture_design.md` | **HEX-8: Core MVP architecture, logical schema and API contracts**. Superseded by modular current architecture or root rule router. [Current reference](designs/2026-10-18_system-design_design.md) | Git; `e64827f79bf65e4e` |

| <a id="pruned-2026-10-18-hex-9-technology-selection-design"></a>`designs/2026-10-18_hex-9-technology-selection_design.md` | **HEX-9: MVP technology selection proposal**. Superseded by modular current architecture or root rule router. [Current reference](designs/system/02-architecture.md) | Git; `501ba99d1530c472` |

| <a id="pruned-2026-10-18-kafka-event-backbone-design"></a>`designs/2026-10-18_kafka-event-backbone_design.md` | **Draft: Kafka event backbone for Hexta**. Superseded by modular current architecture or root rule router. [Current reference](designs/system/06-kafka-events.md) | Git; `60e6c8b5c258d734` |

| <a id="pruned-2026-10-28-api-rule-catalog-design"></a>`designs/2026-10-28_api-rule-catalog_design.md` | **API engineering rule catalog design**. Superseded by modular current architecture or root rule router. [Current reference](../rules/README.md) | Consolidated draft; `0d226b421ecce108` |

| <a id="pruned-2026-10-28-selective-rule-routing-design"></a>`designs/2026-10-28_selective-rule-routing_design.md` | **Selective rule routing**. Superseded by modular current architecture or root rule router. [Current reference](../rules/README.md) | Consolidated draft; `3224b827f3330ece` |

| <a id="pruned-2026-10-18-architecture-consolidation"></a>`epics/EPIC-001-system-design/sessions/2026-10-18_architecture-consolidation.md` | **Architecture Consolidation**. Administrative/intermediate delivery record superseded by verified merged PRs #52-55 and retained changelog/current architecture. [Current reference](CHANGELOG.md) | Git; `f30bb9316da1aa6f` |

| <a id="pruned-2026-10-19-brand-favicon"></a>`epics/EPIC-001-system-design/sessions/2026-10-19_brand-favicon.md` | **Icon-only brand and browser favicon**. Administrative/intermediate delivery record superseded by verified merged PRs #52-55 and retained changelog/current architecture. [Current reference](CHANGELOG.md) | Git; `55fa2d015a7d9eb4` |

| <a id="pruned-2026-10-19-demo-pr-handoff"></a>`epics/EPIC-001-system-design/sessions/2026-10-19_demo-pr-handoff.md` | **Demo pull request and manual test handoff**. Administrative/intermediate delivery record superseded by verified merged PRs #52-55 and retained changelog/current architecture. [Current reference](CHANGELOG.md) | Git; `719bac83d9c055b1` |

| <a id="pruned-2026-10-19-documentation-pr"></a>`epics/EPIC-001-system-design/sessions/2026-10-19_documentation-pr.md` | **Documentation delivery PR**. Administrative/intermediate delivery record superseded by verified merged PRs #52-55 and retained changelog/current architecture. [Current reference](CHANGELOG.md) | Git; `d66b4fa605184e53` |

| <a id="pruned-2026-10-19-epic-session-organization"></a>`epics/EPIC-001-system-design/sessions/2026-10-19_epic-session-organization.md` | **Epic Session Organization**. Administrative/intermediate delivery record superseded by verified merged PRs #52-55 and retained changelog/current architecture. [Current reference](CHANGELOG.md) | Git; `f6930959fcb75691` |

| <a id="pruned-2026-10-19-pr-53-conflicts"></a>`epics/EPIC-001-system-design/sessions/2026-10-19_pr-53-conflicts.md` | **Resolve PR 53 conflicts with main**. Administrative/intermediate delivery record superseded by verified merged PRs #52-55 and retained changelog/current architecture. [Current reference](CHANGELOG.md) | Git; `0cf753c5b570fa7e` |

| <a id="pruned-2026-10-19-theme-aware-branding"></a>`epics/EPIC-001-system-design/sessions/2026-10-19_theme-aware-branding.md` | **Theme-aware Hexta branding**. Administrative/intermediate delivery record superseded by verified merged PRs #52-55 and retained changelog/current architecture. [Current reference](CHANGELOG.md) | Git; `5bdbcdf0754ae89e` |

| <a id="pruned-2026-10-19-thesis-continuation"></a>`epics/EPIC-001-system-design/sessions/2026-10-19_thesis-continuation.md` | **Thesis continuation from Notion**. Administrative/intermediate delivery record superseded by verified merged PRs #52-55 and retained changelog/current architecture. [Current reference](CHANGELOG.md) | Git; `77b5c861c53be7d1` |

| <a id="pruned-2026-10-19-thesis-delivery"></a>`epics/EPIC-001-system-design/sessions/2026-10-19_thesis-delivery.md` | **Thesis delivery**. Administrative/intermediate delivery record superseded by verified merged PRs #52-55 and retained changelog/current architecture. [Current reference](CHANGELOG.md) | Git; `4bc451eb4bbf00c3` |

| <a id="pruned-2026-10-28-api-rule-catalog"></a>`epics/EPIC-001-system-design/sessions/2026-10-28_api-rule-catalog.md` | **Strict API rule catalog draft**. Same-conversation intermediate rule/audit session consolidated into root rule modules, retained mapper/cache design and current review. [Current reference](../rules/README.md) | Consolidated draft; `3de56fc422d155f3` |

| <a id="pruned-2026-10-28-architecture-compliance"></a>`epics/EPIC-001-system-design/sessions/2026-10-28_architecture-compliance.md` | **Architecture and pattern compliance audit**. Same-conversation intermediate rule/audit session consolidated into root rule modules, retained mapper/cache design and current review. [Current reference](../rules/README.md) | Consolidated draft; `d591fef482aea84b` |

| <a id="pruned-2026-10-28-mappers-cache-root-rules"></a>`epics/EPIC-001-system-design/sessions/2026-10-28_mappers-cache-root-rules.md` | **Mapper, cache and rule-location follow-up**. Same-conversation intermediate rule/audit session consolidated into root rule modules, retained mapper/cache design and current review. [Current reference](../rules/README.md) | Consolidated draft; `0e7378eb991e1bb7` |

| <a id="pruned-2026-10-28-selective-rule-routing"></a>`epics/EPIC-001-system-design/sessions/2026-10-28_selective-rule-routing.md` | **Selective rule modules and one router**. Same-conversation intermediate rule/audit session consolidated into root rule modules, retained mapper/cache design and current review. [Current reference](../rules/README.md) | Consolidated draft; `5d5e1b77a905a5e9` |

| <a id="pruned-2026-09-12-issue-1-ai-config-plan"></a>`plans/2026-09-12_issue-1-ai-config_plan.md` | **Implementation Plan: AI Service Provider Configuration Support**. Closed GitHub issue. [Current reference](CHANGELOG.md) | Git; `3d7b9fd5c606198d` |

| <a id="pruned-2026-09-12-issue-10-codebase-audit-plan"></a>`plans/2026-09-12_issue-10-codebase-audit_plan.md` | **Implementation Plan: Scheduled Codebase Audit Workflow & Automated Issue Generation**. Closed GitHub issue. [Current reference](CHANGELOG.md) | Git; `f651a5b6c3128591` |

| <a id="pruned-2026-09-12-issue-18-local-dev-guide-plan"></a>`plans/2026-09-12_issue-18-local-dev-guide_plan.md` | **Implementation Plan: Comprehensive Local Development Guide**. Closed GitHub issue. [Current reference](CHANGELOG.md) | Git; `e9c8e1cb75273f30` |

| <a id="pruned-2026-09-12-issue-20-migrate-go-module-paths-plan"></a>`plans/2026-09-12_issue-20-migrate-go-module-paths_plan.md` | **Implementation Plan: Migrate Go Module Paths and Internal Imports from GitLab to GitHub**. Closed GitHub issue. [Current reference](CHANGELOG.md) | Git; `2450b1a0246ecc26` |

| <a id="pruned-2026-09-12-issue-22-issue-lifecycle-sync-plan"></a>`plans/2026-09-12_issue-22-issue-lifecycle-sync_plan.md` | **Implementation Plan: Issue Lifecycle Manager & Parallel State Synchronization Workflow**. Closed GitHub issue. [Current reference](CHANGELOG.md) | Git; `0fa547ffa320df3b` |

| <a id="pruned-2026-09-12-issue-3-ci-cd-workflows-plan"></a>`plans/2026-09-12_issue-3-ci-cd-workflows_plan.md` | **Implementation Plan: GitHub Actions CI/CD Workflows for Monorepo**. Closed GitHub issue. [Current reference](CHANGELOG.md) | Git; `c435a9af5013a1cc` |

| <a id="pruned-2026-09-12-issue-4-security-auth-plan"></a>`plans/2026-09-12_issue-4-security-auth_plan.md` | **Implementation Plan: Security Hardening for Authentication (Issue #4)**. Closed GitHub issue. [Current reference](CHANGELOG.md) | Git; `973e7675dab8863f` |

| <a id="pruned-2026-09-12-issue-5-plan"></a>`plans/2026-09-12_issue-5_plan.md` | **Implementation Plan: Fix Unique Identifier Collision & Multi-Provider Linking (Issue #5)**. Closed GitHub issue. [Current reference](CHANGELOG.md) | Git; `a3faf1bc062230e7` |

| <a id="pruned-2026-09-12-issue-6-plan"></a>`plans/2026-09-12_issue-6_plan.md` | **Implementation Plan: Reconcile Auth Token Storage, Synchronize JWT Claims & Enforce English Language Rule (Issue #6)**. Closed GitHub issue. [Current reference](CHANGELOG.md) | Git; `4fdb45b234c3d926` |

| <a id="pruned-2026-09-12-issue-7-multi-tenant-plan"></a>`plans/2026-09-12_issue-7-multi-tenant_plan.md` | **Implementation Plan: Multi-Tenant Database Models, Service Domain, and REST API**. Closed GitHub issue. [Current reference](CHANGELOG.md) | Git; `66466a81647b7536` |

| <a id="pruned-2026-09-12-issue-8-sdk-token-refresh-plan"></a>`plans/2026-09-12_issue-8-sdk-token-refresh_plan.md` | **Implementation Plan: Silent Token Refresh Interceptor and SDK Package Rebranding**. Closed GitHub issue. [Current reference](CHANGELOG.md) | Git; `d909ce468bef9c26` |

| <a id="pruned-2026-09-13-issue-21-clean-legacy-gitlab-references-plan"></a>`plans/2026-09-13_issue-21-clean-legacy-gitlab-references_plan.md` | **Implementation Plan: Replace Legacy GitLab References & Template READMEs with Project Standards**. Closed GitHub issue. [Current reference](CHANGELOG.md) | Git; `decf231278e6d683` |

| <a id="pruned-2026-09-13-issue-26-pr-review-agent-plan"></a>`plans/2026-09-13_issue-26-pr-review-agent_plan.md` | **Plan: PR Review Agent with Auto-Merge & Conflict Escalation**. Closed GitHub issue. [Current reference](CHANGELOG.md) | Git; `dd55cc370a8e8290` |

| <a id="pruned-2026-10-05-issue-32-plan"></a>`plans/2026-10-05_issue-32_plan.md` | **Implementation Plan: Default Light Mode with Seamless Dark Mode Toggle Support**. Closed GitHub issue. [Current reference](CHANGELOG.md) | Git; `38b71fda24565dd0` |

| <a id="pruned-2026-10-05-issue-creator-and-skill-execution-protocol-plan"></a>`plans/2026-10-05_issue-creator-and-skill-execution-protocol_plan.md` | **Implementation Plan: Issue Creator Skill & Skill Execution Protocol**. Historical automation/health/test work superseded by current skills, test code, architecture review and rule modules. [Current reference](reviews/2026-10-28_architecture-compliance_review.md) | Git; `bd35a31de8a0fce7` |

| <a id="pruned-2026-10-09-health-controller-layer-fix-plan"></a>`plans/2026-10-09_health-controller-layer-fix_plan.md` | **Plan: Resolve 5-Layer Architecture Boundary Violations in Health Controller**. Historical automation/health/test work superseded by current skills, test code, architecture review and rule modules. [Current reference](reviews/2026-10-28_architecture-compliance_review.md) | Git; `71ef561395d128de` |

| <a id="pruned-2026-10-09-issue-34-plan"></a>`plans/2026-10-09_issue-34_plan.md` | **Implementation Plan: Core Business Schema Migrations & GORM Models**. Closed GitHub issue. [Current reference](CHANGELOG.md) | Git; `cfff6620b23c17fc` |

| <a id="pruned-2026-10-09-issue-35-plan"></a>`plans/2026-10-09_issue-35_plan.md` | **Implementation Plan: Inventory Repository, Stock Reservation, and Movement Tracking**. Closed GitHub issue. [Current reference](CHANGELOG.md) | Git; `a2ff10ed52de4211` |

| <a id="pruned-2026-10-09-issue-36-plan"></a>`plans/2026-10-09_issue-36_plan.md` | **Implementation Plan: Order State Machine, Pricing, and REST Endpoints**. Closed GitHub issue. [Current reference](CHANGELOG.md) | Git; `cf58172962f22b43` |

| <a id="pruned-2026-10-09-issue-37-plan"></a>`plans/2026-10-09_issue-37_plan.md` | **Implementation Plan: AI Agent Engine with Tool Calling Registry and Draft Proposal Flow**. Closed GitHub issue. [Current reference](CHANGELOG.md) | Git; `e32e665b916cb8ee` |

| <a id="pruned-2026-10-09-issue-45-plan"></a>`plans/2026-10-09_issue-45_plan.md` | **Implementation Plan: Remove MinIO Service from Docker Compose Infrastructure**. Closed GitHub issue. [Current reference](CHANGELOG.md) | Git; `2db5eb58e2bf4faf` |

| <a id="pruned-2026-10-09-repository-test-coverage-plan"></a>`plans/2026-10-09_repository-test-coverage_plan.md` | **Plan: Implement Test Coverage for Session and Tenant Repositories and Cache Wrappers**. Historical automation/health/test work superseded by current skills, test code, architecture review and rule modules. [Current reference](reviews/2026-10-28_architecture-compliance_review.md) | Git; `a3e9acde841a51c8` |

| <a id="pruned-2026-10-10-issue-29-plan"></a>`plans/2026-10-10_issue-29_plan.md` | **Implementation Plan: Implement Comprehensive Unit Test Suites for Critical Core Services and SDK**. Closed GitHub issue. [Current reference](CHANGELOG.md) | Git; `11c4b37b388e902e` |

| <a id="pruned-2026-10-15-hex-4-mvp-scope-plan"></a>`plans/2026-10-15_hex-4-mvp-scope_plan.md` | **HEX-4 MVP Scope Planning**. Superseded proposal/implementation plan. [Current reference](plans/2026-10-27_thesis-first-ecosystem_plan.md) | Git; `e3d1dd2d6610a842` |

| <a id="pruned-2026-10-18-durable-ai-local-stack-plan"></a>`plans/2026-10-18_durable-ai-local-stack_plan.md` | **Durable runs and local Docker plan**. Superseded proposal/implementation plan. [Current reference](designs/system/05-grpc-ai.md) | Git; `c1b0b4683015d2eb` |

| <a id="pruned-2026-10-18-hex-10-local-ai-bootstrap-plan"></a>`plans/2026-10-18_hex-10-local-ai-bootstrap_plan.md` | **HEX-10: Local AI service bootstrap**. Superseded proposal/implementation plan. [Current reference](designs/system/05-grpc-ai.md) | Git; `71116f543fea17da` |

| <a id="pruned-2026-10-19-next-work-backlog-plan"></a>`plans/2026-10-19_next-work-backlog_plan.md` | **Next work backlog from code and Notion**. Superseded proposal/implementation plan. [Current reference](plans/2026-10-27_thesis-first-ecosystem_plan.md) | Git; `8f0228684b038656` |

| <a id="pruned-2026-10-19-tenant-first-framework-plan"></a>`plans/2026-10-19_tenant-first-framework_plan.md` | **Tenant-first product framework and delivery plan**. Superseded proposal/implementation plan. [Current reference](designs/2026-10-19_tenant-onboarding_design.md) | Git; `091cda574b2b44a4` |

| <a id="pruned-2026-09-12-issue-1-ai-config-review"></a>`reviews/2026-09-12_issue-1-ai-config_review.md` | **Code Review: AI Service Provider Configuration Support**. Closed GitHub issue. [Current reference](CHANGELOG.md) | Git; `c0d68355cb25421b` |

| <a id="pruned-2026-09-12-issue-10-codebase-audit-review"></a>`reviews/2026-09-12_issue-10-codebase-audit_review.md` | **Code Review: Codebase Audit Workflow & Automated Issue Generation**. Closed GitHub issue. [Current reference](CHANGELOG.md) | Git; `ede379d595dc12cc` |

| <a id="pruned-2026-09-12-issue-18-local-dev-guide-review"></a>`reviews/2026-09-12_issue-18-local-dev-guide_review.md` | **Code Review: Comprehensive Local Development Guide**. Closed GitHub issue. [Current reference](CHANGELOG.md) | Git; `035c615a4a910cfe` |

| <a id="pruned-2026-09-12-issue-20-migrate-go-module-paths-review"></a>`reviews/2026-09-12_issue-20-migrate-go-module-paths_review.md` | **Code Review: Go Module Path & Internal Import Migration**. Closed GitHub issue. [Current reference](CHANGELOG.md) | Git; `0d05b1f6ce5c626d` |

| <a id="pruned-2026-09-12-issue-22-issue-lifecycle-sync-review"></a>`reviews/2026-09-12_issue-22-issue-lifecycle-sync_review.md` | **Code Review: Issue Lifecycle Manager & Parallel State Synchronization Workflow**. Closed GitHub issue. [Current reference](CHANGELOG.md) | Git; `2b2af5d1b4624b4e` |

| <a id="pruned-2026-09-12-issue-3-ci-cd-workflows-review"></a>`reviews/2026-09-12_issue-3-ci-cd-workflows_review.md` | **Code Review: GitHub Actions CI/CD Workflows for Monorepo**. Closed GitHub issue. [Current reference](CHANGELOG.md) | Git; `0ea350ae5bb7b998` |

| <a id="pruned-2026-09-12-issue-4-security-auth-review"></a>`reviews/2026-09-12_issue-4-security-auth_review.md` | **Code Review: Authentication Security Hardening (Issue #4)**. Closed GitHub issue. [Current reference](CHANGELOG.md) | Git; `14749e1a40db7a78` |

| <a id="pruned-2026-09-12-issue-5-review"></a>`reviews/2026-09-12_issue-5_review.md` | **Code Review Report: Resolve Unique Identifier Collision on OAuth Login & Enable Multi-Provider Linking**. Closed GitHub issue. [Current reference](CHANGELOG.md) | Git; `218f3b273d00c820` |

| <a id="pruned-2026-09-12-issue-6-review"></a>`reviews/2026-09-12_issue-6_review.md` | **Code Review Report: Reconcile Auth Token Storage, Synchronize JWT Claims & Enforce English Language Rule**. Closed GitHub issue. [Current reference](CHANGELOG.md) | Git; `66d78ceb192c3849` |

| <a id="pruned-2026-09-12-issue-7-multi-tenant-review"></a>`reviews/2026-09-12_issue-7-multi-tenant_review.md` | **Code Review: Multi-Tenant Database Models, Service Domain, and REST API**. Closed GitHub issue. [Current reference](CHANGELOG.md) | Git; `12e0e4eae6319a05` |

| <a id="pruned-2026-09-12-issue-8-sdk-token-refresh-review"></a>`reviews/2026-09-12_issue-8-sdk-token-refresh_review.md` | **Code Review: Silent Token Refresh Interceptor and SDK Rebranding**. Closed GitHub issue. [Current reference](CHANGELOG.md) | Git; `979676fd7dff1de7` |

| <a id="pruned-2026-09-12-pr-11-db-account-linking-review"></a>`reviews/2026-09-12_pr-11-db-account-linking_review.md` | **Code Review Report: PR #11 (Unique Identifier Collision & Account Linking)**. PR #11 verified merged into main. [Current reference](CHANGELOG.md) | Git; `506d33f619d7e0cc` |

| <a id="pruned-2026-09-12-pr-9-ci-cd-workflows-review"></a>`reviews/2026-09-12_pr-9-ci-cd-workflows_review.md` | **Code Review Report: PR #9 (CI/CD GitHub Actions Workflows)**. PR #9 verified merged into main. [Current reference](CHANGELOG.md) | Git; `d12bd014a40de53b` |

| <a id="pruned-2026-09-13-issue-21-clean-legacy-gitlab-references-review"></a>`reviews/2026-09-13_issue-21-clean-legacy-gitlab-references_review.md` | **Code Review: Documentation Standardization & Legacy GitLab Reference Elimination**. Closed GitHub issue. [Current reference](CHANGELOG.md) | Git; `a86bc7a1515a6d1a` |

| <a id="pruned-2026-09-13-issue-26-pr-review-agent-review"></a>`reviews/2026-09-13_issue-26-pr-review-agent_review.md` | **Code Review: Autonomous PR Review Agent & Conflict Escalation Engine**. Closed GitHub issue. [Current reference](CHANGELOG.md) | Git; `55ce6a2052f36a43` |

| <a id="pruned-2026-10-05-issue-32-review"></a>`reviews/2026-10-05_issue-32_review.md` | **Code Review: Default Light Mode with Dark Mode Toggle Support**. Closed GitHub issue. [Current reference](CHANGELOG.md) | Git; `7a8075cf10d4bb31` |

| <a id="pruned-2026-10-05-issue-creator-and-skill-execution-protocol-review"></a>`reviews/2026-10-05_issue-creator-and-skill-execution-protocol_review.md` | **Code Review: Issue Creator Skill & Skill Execution Protocol**. Historical automation/health/test work superseded by current skills, test code, architecture review and rule modules. [Current reference](reviews/2026-10-28_architecture-compliance_review.md) | Git; `eca876da9114ec21` |

| <a id="pruned-2026-10-09-health-controller-layer-fix-review"></a>`reviews/2026-10-09_health-controller-layer-fix_review.md` | **Code Review: Resolve 5-Layer Architecture Boundary Violations in Health Controller**. Historical automation/health/test work superseded by current skills, test code, architecture review and rule modules. [Current reference](reviews/2026-10-28_architecture-compliance_review.md) | Git; `7b4690e1538abad8` |

| <a id="pruned-2026-10-09-issue-34-review"></a>`reviews/2026-10-09_issue-34_review.md` | **Code Review: Core Business Schema Migrations & GORM Models**. Closed GitHub issue. [Current reference](CHANGELOG.md) | Git; `5580dead47509ec1` |

| <a id="pruned-2026-10-09-issue-35-review"></a>`reviews/2026-10-09_issue-35_review.md` | **Code Review: Inventory Repository, Stock Reservation, and Movement Tracking**. Closed GitHub issue. [Current reference](CHANGELOG.md) | Git; `49b3adb1994e16dd` |

| <a id="pruned-2026-10-09-issue-36-review"></a>`reviews/2026-10-09_issue-36_review.md` | **Code Review: Order State Machine, Pricing, and REST Endpoints**. Closed GitHub issue. [Current reference](CHANGELOG.md) | Git; `146d28858a8b0475` |

| <a id="pruned-2026-10-09-issue-37-review"></a>`reviews/2026-10-09_issue-37_review.md` | **Code Review: AI Agent Engine with Tool Calling Registry and Draft Proposal Flow**. Closed GitHub issue. [Current reference](CHANGELOG.md) | Git; `2f619002f0bfb19e` |

| <a id="pruned-2026-10-09-issue-45-review"></a>`reviews/2026-10-09_issue-45_review.md` | **Code Review: Remove MinIO Service from Docker Compose Infrastructure**. Closed GitHub issue. [Current reference](CHANGELOG.md) | Git; `160a90d597f4dc6b` |

| <a id="pruned-2026-10-09-repository-test-coverage-review"></a>`reviews/2026-10-09_repository-test-coverage_review.md` | **Code Review: Repository Layer Test Coverage Expansion**. Historical automation/health/test work superseded by current skills, test code, architecture review and rule modules. [Current reference](reviews/2026-10-28_architecture-compliance_review.md) | Git; `3028b00173169293` |

| <a id="pruned-2026-10-10-issue-29-review"></a>`reviews/2026-10-10_issue-29_review.md` | **Code Review: Unit Test Suites for Critical Core Services and SDK**. Closed GitHub issue. [Current reference](CHANGELOG.md) | Git; `9046ac21ed6811d5` |

| <a id="pruned-2026-10-18-hex-10-local-ai-bootstrap-review"></a>`reviews/2026-10-18_hex-10-local-ai-bootstrap_review.md` | **Code Review: HEX-10 Local AI Bootstrap, Go/Python gRPC Boundary & Review Fixes**. Older reviewed baseline superseded by merged compliance delivery/current audit. Historical verification is retained in sessions/changelog. [Current reference](designs/system/05-grpc-ai.md) | Git; `06be6aaae814cf5c` |

| <a id="pruned-2026-10-20-code-rules-audit-review"></a>`reviews/2026-10-20_code-rules-audit_review.md` | **Current-code syntax and engineering-rule audit**. Older reviewed baseline superseded by merged compliance delivery/current audit. Historical verification is retained in sessions/changelog. [Current reference](reviews/2026-10-28_architecture-compliance_review.md) | Git; `f00e24808d8081c1` |


### Legacy rule copies compacted rather than deleted

The three files under agentic-memory/rules are now compatibility redirects, preserving skill links without retaining a second policy catalog. Their pre-cleanup bodies remain available at d09083c: development-rules.md is superseded by the focused root rule modules (including retained frontend Suspense/API URL conventions); github-workflow.md by root delivery rules and current runner/reviewer skills/scripts; skill-execution-protocol.md by its maintained .agents counterpart under current AGENTS/selective-loading instructions. No runtime automation schedule or auto-merge approval is created by this cleanup.
