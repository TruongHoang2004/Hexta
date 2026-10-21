# Notion ticket reconciliation and next demo plan

Date: 2026-10-19. Scope: all 42 active rows returned by the Hexta task data source, with has_more=false. This is a read-only Notion reconciliation; no status/property/task writes were performed. Repository report delivery is local/unmerged. No application code changed.

## Goal and evidence

Keep the owner's fastest usable demo priority; defer keys, mock-data projects, paid/provider validation and real connectors. Distinguish ticket status, actual implemented scope, historical verification and main merge. Avoid a product-completion percentage from umbrella tickets.

- Notion source: [Hexta Tasks & Backlog](https://app.notion.com/p/e965ef5844a04d0a8d06a174c3bda8b3), collection db7a0686-538c-4cbc-92a7-feaa8364fc65. Snapshot: 10 completed, 1 in progress, 31 not started. Query included IDs/titles/statuses/acceptance/notes; no pagination remained. HEX-6 requirements were fetched to compare functional scope, and source schema was fetched before querying.
- [PR #54](https://github.com/TruongHoang2004/Hexta/pull/54) merged 2026-10-19T07:31:26Z as 5a1e7196f259b3ff07fd4ade75a01054623b79aa; all returned backend/frontend/container/lifecycle checks succeeded. Tenant onboarding, manual sales, branding and manual checklist are merged delivery, superseding earlier pending-merge memory notes.
- [PR #52](https://github.com/TruongHoang2004/Hexta/pull/52) and [#53](https://github.com/TruongHoang2004/Hexta/pull/53) merges reverified. [PR #55](https://github.com/TruongHoang2004/Hexta/pull/55) thesis slice merged 2026-10-19T08:19:33Z. Thesis merge does not fill intentionally blank abstract, related work, empirical results or conclusions.
- Inspected current router, web routes, onboarding/sales service tests and fake Python graph on clean codex/thesis-content at 3b9d182. Reused recorded implementation test/browser evidence; no new runtime/application tests in this review. Existing plan/session boundaries remain authoritative for what was actually run.

## Actual delivered product slice

Approved tenant chooser and nested resources; customer application -> Hexta approval -> initial owner; owner email-addressed private manual links -> explicit recipient acceptance; customer/product create/edit/archive; incomplete drafts and versioned edit; assigned-sale explicit confirmation; tenant/member/owner/assignment/reference validation and audit; SDK/Swagger/additive migrations; reproducible local Docker and Hexta logo/favicon.

Live operator approval/recipient acceptance is covered in isolated HTTP/DB fixtures; end-to-end verification in the running demo remains pending authorized operator setup. No email provider is connected. Sale confirmation has no stock/payment/customer-send side effect. FR-10 is only partially covered: richer customer needs/priority, consultation/media and progress are not all delivered. Tenant-first commercial boundary is a newer owner decision that should be reflected in the existing Notion specification.

No conversation workspace/schema/control lifecycle, production script permissions/versioning, Go business AI dispatcher, actual model routing, Kafka integration, configurable work handoff, cancellation/returns, operations dashboard, real Pancake or remote deployment is delivered. Durable fake Python runs are a technical foundation, not full AI consultation.

## Ticket-by-ticket reconciliation

Proposed states below are review recommendations, not Notion writes. Ready-to-close initial slices do not complete umbrella MVP tickets. Keep historical research/specification completion separate from implementation.

| Task | Notion now | Assessment | Evidence / remaining boundary |
| --- | --- | --- | --- |
| [HEX-1](https://app.notion.com/44f73fb5f2f04f419ca30d4a98ad3044) Problem and target users | Completed | Keep completed: specification/research | Notion completion concerns specification/research, not implementation of the full product. Owner tenant-first correction should be reflected in existing specifications. |
| [HEX-2](https://app.notion.com/a441eeef3e094aa8bf2cdbc1463c2ccc) Week-one deliverable review | Completed | Keep completed: specification/research | Notion completion concerns specification/research, not implementation of the full product. Owner tenant-first correction should be reflected in existing specifications. |
| [HEX-3](https://app.notion.com/41e0543d339b4dbba0aaaaa5a43fc5e4) Thesis objectives/questions | Completed | Keep completed: specification/research | Notion completion concerns specification/research, not implementation of the full product. Owner tenant-first correction should be reflected in existing specifications. |
| [HEX-4](https://app.notion.com/558e491fd9044a6baf30b4e61d7b92a8) MVP scope | Completed | Keep completed: specification/research | Notion completion concerns specification/research, not implementation of the full product. Owner tenant-first correction should be reflected in existing specifications. |
| [HEX-5](https://app.notion.com/3149706d4b654162b0c3c8f8acf1474a) Comparable solution survey | Completed | Keep completed: specification/research | Notion completion concerns specification/research, not implementation of the full product. Owner tenant-first correction should be reflected in existing specifications. |
| [HEX-6](https://app.notion.com/2e3c653546b342b3a90a994a760f4340) Functional/nonfunctional requirements | Completed | Keep completed: specification/research | Notion completion concerns specification/research, not implementation of the full product. Owner tenant-first correction should be reflected in existing specifications. |
| [HEX-7](https://app.notion.com/8d451c4ec8f54ca58cf2d9628a0b72af) User flows and wireframes | Completed | Keep completed: specification/research | Notion completion concerns specification/research, not implementation of the full product. Owner tenant-first correction should be reflected in existing specifications. |
| [HEX-8](https://app.notion.com/cfa16b89d09f4211a7fd9e7dc1b5d8e8) Architecture/schema/API design | Completed | Keep completed: specification/research | Notion completion concerns specification/research, not implementation of the full product. Owner tenant-first correction should be reflected in existing specifications. |
| [HEX-9](https://app.notion.com/af0f29236ac948a286e2e59b7379a854) Technology and method | Completed | Keep completed: specification/research | Notion completion concerns specification/research, not implementation of the full product. Owner tenant-first correction should be reflected in existing specifications. |
| [HEX-10](https://app.notion.com/89298bfe7b02407f8eaab9453df66229) Repository/development setup | In progress | Close candidate: local environment slice | Merged Docker stack, runbook and healthy rebuild evidence; confirm clean-machine setup criterion before closure. Future Kafka/providers are separate. |
| [HEX-11](https://app.notion.com/0e08159ea1f04d17b71287cd451cdea1) Thesis preliminary sections | Not started | In progress | Front matter exists, but abstract and author-owned sections intentionally remain empty. |
| [HEX-12](https://app.notion.com/9a4bf713faec48be89c603e3dd3f54ee) Models and migrations | Not started | In progress | Tenant/application/invitation and sales migrations merged and tested; conversation/script/workflow schema absent. |
| [HEX-13](https://app.notion.com/1d3f7629f7214b46843316300f896b83) Core backend APIs | Not started | In progress | Tenant onboarding and sales APIs merged; conversation/control/script/workflow/send APIs absent. |
| [HEX-14](https://app.notion.com/5d28adcb431b4980904bd354f8a59205) Initial main-flow UI | Not started | In progress | Tenant chooser, applications/review, invitations, customers/products/orders and branding merged; core messaging/workflow UI absent. |
| [HEX-15](https://app.notion.com/1cb726a231744b8ba4d325cb36f4179b) Frontend/backend integration | Not started | Ready to close against stated criterion | At least one approved UI/API flow is merged and browser-verified: customer/product -> draft -> assigned-sale confirmation. Full messaging integration is not claimed. |
| [HEX-16](https://app.notion.com/7f38fd6a83b241828b011a5a341a9583) Initial feature tests | Not started | Ready to close against initial-test scope | Merged initial service/HTTP/SDK tests, PostgreSQL concurrency cases and passing PR CI; live operator approval and usability acceptance remain documented gaps. |
| [HEX-17](https://app.notion.com/ab14cd2ebc7541fcbda54ab069d3df69) Research/design thesis draft | Not started | In progress | Design draft exists; related-work evidence and bibliography intentionally empty. Notion thinking is not literature. |
| [HEX-18](https://app.notion.com/1040593a63e348f79389550bf72214cb) Remaining MVP functionality | Not started | In progress | Sales foundation delivered; configurable workflow, cases, overview and core messaging still missing. Deferred scope remains explicitly tracked. |
| [HEX-19](https://app.notion.com/d4d16b596f824b22897e06af4f589ad4) Access/input/error controls | Not started | In progress | Tenant/current-member/owner/assignment/validation/version guards implemented for delivered flows; future script/send/workflow rules absent. |
| [HEX-20](https://app.notion.com/5d756fb3751b4c3e9b99e7a5f011ad9b) Important unit/integration tests | Not started | In progress | Substantial unit/integration coverage for delivered slices; no tests for future message/workflow paths. |
| [HEX-21](https://app.notion.com/f7ba81b17c1c4bd39c9b25c61fc9f85c) Requirements/design thesis update | Not started | In progress | Requirements/architecture chapters drafted and merged; final consistency/author review remains. |
| [HEX-22](https://app.notion.com/4991b03f6aec4157b4c8de6a0052054f) Core value and essential integrations | Not started | Not complete; later integration slice | No user-facing consultation/AI/channel send end-to-end; fake runtime is a technical foundation. |
| [HEX-23](https://app.notion.com/f1bd17cab1af4a3f8c9074b4290711a6) Edge cases and consistency | Not started | In progress | Draft/version/approval/accept races and wrong-tenant guards covered; messaging/send/workflow races remain. |
| [HEX-24](https://app.notion.com/9386115a38da425db2ef000ad5fabdc0) Staging/demo deployment | Not started | Partial: local demo only | Local Docker is repeatable and accessible on loopback; no remote/staging deployment or full MVP acceptance. |
| [HEX-25](https://app.notion.com/4d044585efc74c519cee7c2df14c0a83) Evaluation data and criteria | Not started | Deferred by owner | Research evaluation dataset/criteria not prepared; routine feature test fixtures do not satisfy this ticket. |
| [HEX-26](https://app.notion.com/aabd17f4c9b74d4a85995f6e563a0d7d) Implementation thesis chapter | Not started | In progress | Code-supported implementation chapter drafted/merged; full reviewed thesis criterion not complete. |
| [HEX-27](https://app.notion.com/bfd36947136444e19da9584b420183fd) End-to-end tests and critical fixes | Not started | In progress | Sales browser smoke performed; full MVP E2E and owner manual acceptance/bug triage pending. |
| [HEX-28](https://app.notion.com/1344dcc67b6744baac708892e9834e0b) Evaluation evidence/screenshots/metrics | Not started | Partial technical evidence only | Some UI screenshots and technical checks exist; no formal evaluation dataset/results/metrics. |
| [HEX-29](https://app.notion.com/80d09300a4e440cbaf802e8d9ca79626) MVP evaluation | Not started | Not started; deferred evaluation | No completed study against predefined research criteria. |
| [HEX-30](https://app.notion.com/9f7cdbfa6a2f4c1baee47232790dc302) 80-90% thesis draft | Not started | Not complete | Experiment/results/abstract/conclusion/references intentionally blank; no 80-90% claim. |
| [HEX-31](https://app.notion.com/e2a07320684b415ba0bbc79805aa3d55) Goal/design/result consistency | Not started | Not started as final review | Architecture/code gaps are documented, but final goal-design-result reconciliation is incomplete. |
| [HEX-32](https://app.notion.com/0d065034e3674a7295375299e4237511) MVP feature freeze | Not started | Not started | No product feature freeze agreed. |
| [HEX-33](https://app.notion.com/25048dc39d5b41edb096335082f86cf3) Final release/deployment | Not started | Not started | No final release candidate or chosen target deployment. |
| [HEX-34](https://app.notion.com/526fa739ca194b84a4ad6c8f597223f2) Installation/usage documentation | Not started | In progress | Local runbook/manual checklist exist; not the final complete MVP usage/install guide. |
| [HEX-35](https://app.notion.com/06958b818feb44b393523082a0cc7843) Complete thesis content/format | Not started | Not complete | Thesis scaffold/design/implementation content exists; final literature/results/author material and formatting review remain. |
| [HEX-36](https://app.notion.com/405069278ca549639f543f5f33d372f6) Demo script/presentation | Not started | In progress | Manual testing checklist and a short sales demo path exist; final demo script/presentation not delivered. |
| [HEX-37](https://app.notion.com/627a969eb5c34349b25704418d705802) Final MVP end-to-end acceptance | Not started | Not started | No final full-MVP release candidate acceptance. |
| [HEX-38](https://app.notion.com/3cb128fc38d649faa45abcb03da18f08) Final deployment/data/docs verification | Not started | Partial local preparation | Local deployment/data preservation/docs verified; final deployment/demo dataset/doc reconciliation pending. |
| [HEX-39](https://app.notion.com/b8cd9f508c7a48b69500f1f4903a3309) Final thesis review | Not started | Not complete | Compilation recorded; full final references/content/format QA not satisfied. |
| [HEX-40](https://app.notion.com/875a21831e6a4d1a9803c282ebc8f87a) Package final deliverables | Not started | Not started | Final consolidated product/source/docs/thesis delivery not complete. |
| [HEX-41](https://app.notion.com/3fab0ea49f09813ca29cfec484c2a3ef) Pancake feasibility survey | Completed | Keep completed: research only | Documentation survey done; actual Pancake entitlement/token/webhook/send integration not verified. |
| [HEX-42](https://app.notion.com/3fab0ea49f09813bac37d1128b26312e) Ten-plant AI dataset | Not started | Deferred by owner | Ten-plant dataset not supplied/prepared; no invented business facts. |

## Recommended immediate tracker reconciliation

1. HEX-15 and HEX-16 meet their stated initial criteria based on merged code, recorded UI/API evidence and passing tests; propose completion with PR/test links and explicit remaining boundaries.
2. HEX-10 is a close candidate for the local environment scope; verify the documented clean-environment install criterion before closing. Kafka/provider/Pancake work must not silently expand this environment ticket.
3. HEX-12/13/14/18/19/20/23 are active partial implementation, not unstarted and not fully complete. Attach the merged tenant/sales scope and list missing messaging/workflow pieces. HEX-24/27/34/36/38 have local/demo preparation, not final full-MVP acceptance.
4. HEX-11/17/21/26 have thesis work, but preserve empty evidence-dependent sections and the author's review boundary. Do not close HEX-28/29/30/35/39 from compilation or a UI screenshot.
5. Existing HEX-1..9 and HEX-41 stay complete for their approved research/design scope. Refresh current tenant/route/invitation decisions in existing specs and clarify HEX-41's older simulator/draft-only note using newer HEX-6/owner direction, rather than reopening research or claiming live integration.
6. Notion lacks a review status; use in-progress plus merge/test notes where not complete. Do not fabricate new ticket IDs or create duplicate umbrella tasks. Proposed D4/D5 slices should be linked to HEX-18/12/13/14/19/20 when future task creation is requested.

## Active next work (supersedes historical implementation statuses in the backlog plan)

| Order | Deliverable / paths | Dependencies and acceptance |
| --- | --- | --- |
| 1 | Validate and stabilize current tenant/sales demo; existing services/web and infrastructure/TESTING.demo.md | Owner manual tests and authorized operator setup for live approval; test matching/wrong email, revoke/reissue, tenant isolation, missing vs zero price, stale edits and confirmation. Triage concrete bugs; preserve existing data. Record PASS/FAIL/NOT RUN; do not infer user acceptance. Maps HEX-27/19/20/23. |
| 2 | D4a: versioned sequential workflow definitions; Atlas migration, Go model/repository/service/DTO/controller, SDK and tenant settings UI | Owner can add/remove/rename/reorder steps, choose recipient and direct/accept policy; packing/shipping optional. New work pins immutable version; editing definitions never changes existing work. Maps HEX-18 and 12/13/14. |
| 3 | D4b: linked work item and responsibility handoff; work service/API, tenant work list/detail, tests/audit | Explicitly create work linked to an order; confirmation alone does not imply a workflow side effect. Owner nominates recipient; direct transfer changes responsibility immediately, accept-required retains old actor until recipient accepts. Owner reassignment audited; current member/tenant/version gates, concurrent accept/reassign and reload tests. Scope sources are HEX-6 FR-11 and existing D4 plan. |
| 4 | D5: factual owner overview; tenant-authorized aggregation API and dashboard | Show order statuses, current work step/assignee, awaiting acceptance/active work, time/employee filters. Counts reconcile with lists. No collected-revenue or asked-product claim without data semantics/source. Maps HEX-18 and FR-13; follows D4 for work metrics. |
| 5 | Manual conversation slice after Demo A feedback | Separate tenant inbox/turns/notes, self-claim and owner assignment, manual drafts; no external-send success or fabricated inbound events. Detailed schema/actor/concurrency contract before code. Full AI/script/send/expiry integration follows separately selected dependencies. |

Each implementation batch is vertical: narrow migration + Go API + client/UI + meaningful authorization/concurrency tests + browser/reload evidence. Regenerate Swagger and validate Atlas when relevant. Done means scoped acceptance and main merge; partial slices never complete the whole MVP.

## Deferred and risks

HEX-25/42 data/evaluation projects; provider keys/adapters and paid model verification; live Pancake entitlement/send validation; Kafka/checkpoints/Langfuse; full returns/cancellations, media pipelines and VPS/domain follow product feedback. These remain approved future scope rather than canceled requirements. Ordinary automated-test fixtures are still required for code correctness.

Concurrency and changing authority matter most for D4: simultaneous handoff/accept/reassign, removed recipients, tenant access changes and old workflow versions. Define version checks and lock ordering before implementation. Never infer role authority from an invitation token or AI output. Metrics and research results require actual observations, not a successful build.

## Verification and handoff

This review fetched live Notion rows and remote merge/check state, inspected code/memory, and checked report links/whitespace. No Notion edits, account grants, new feature implementation or application test reruns. Parent epics remain incomplete. Next executable step is existing manual acceptance/bug triage, then a detailed D4 schema/API/workflow design based on the approved responsibility rules.
