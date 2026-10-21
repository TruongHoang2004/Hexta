# Next work backlog from code and Notion

Date: 2026-10-19. Status: revised for the owner's code-first, fastest-demo direction; no implementation or external tracker changes.

The active delivery sequence below supersedes the original T1–T15 priority assignments. The original breakdown remains scope/reference evidence, not the current execution queue.

## Overview and goal

Deliver a usable local conversation simulator vertical slice, then real AI and the remaining approved business workflows. Go owns authorization and sending; Python returns candidates. Existing auth, tenant, legacy orders and durable fake AI are a foundation, not the completed product.

## Evidence and current state

- Inspected branch `codex/system-design-epic-memory`, HEAD `413ee35`, initially clean. PR [52](https://github.com/TruongHoang2004/Hexta/pull/52) merged 2026-10-19T02:09:59Z; backend/frontend/container checks reported SUCCESS. PR [53](https://github.com/TruongHoang2004/Hexta/pull/53) merged 2026-10-19T02:13:18Z; reported lifecycle checks SUCCESS. No new runtime suite was run for this plan.
- `services/api/internal/present/http/router/router.go` exposes auth/tenant/order/legacy AI routes, without conversation/script/workflow/dashboard routes. `apps/web/app/` has auth and tenant pages without a conversation workspace.
- `services/ai/hexta_ai/graph.py` is explicitly fake exact-match script/human routing; durable runs do not implement real model routing. `services/api/cmd/ai-smoke/` is a smoke client, not the business dispatcher.
- `services/api/internal/present/http/dto/order_dto.go` requires draft items and exposes auto_confirm; `ai_agent_service.go` still uses a legacy Go Gemini client. Audit these before reuse against the approved incomplete-draft/explicit-confirmation rules.
- Legacy compose contains a broker definition; the verified local demo profile does not provide the required integrated Kafka event pipeline.
- [HEX-4](https://app.notion.com/p/558e491fd9044a6baf30b4e61d7b92a8) and [HEX-6](https://app.notion.com/p/2e3c653546b342b3a90a994a760f4340) are completed specification work, not delivered features. Newer decisions supersede historical draft-only scope: both suggest and auto_send are approved, with sale takeover stopping automation.
- [HEX-12](https://app.notion.com/p/9a4bf713faec48be89c603e3dd3f54ee), [HEX-13](https://app.notion.com/p/1d3f7629f7214b46843316300f896b83), [HEX-14](https://app.notion.com/p/5d28adcb431b4980904bd354f8a59205), [HEX-18](https://app.notion.com/p/1040593a63e348f79389550bf72214cb) and [HEX-25](https://app.notion.com/p/4d044585efc74c519cee7c2df14c0a83) are not started and have broad criteria with blank bodies. Refine existing work rather than creating duplicate umbrella tickets.
- [HEX-42](https://app.notion.com/p/3fab0ea49f09813bac37d1128b26312e) is not started, due 2026-10-18, awaiting the owner's ten-plant list. [HEX-41](https://app.notion.com/p/3fab0ea49f09813ca29cfec484c2a3ef) completed documentation research; real entitlement/token/slot and runtime verification remain missing.
- A bounded Notion backlog search returned 41 entries; selected requirements and implementation/evaluation tickets were fetched. GitHub connector open search returned no entries; this is not proof of an exhaustive absence of work.

## Task breakdown

Identifiers below are local proposal identifiers, not new GitHub/Notion IDs. P0 means next delivery batch; P1 follows that slice; P2 is later integration/operations. Every task needs review, appropriate checks and merge into main before done.

| ID / priority | Work and affected paths | Dependencies / existing task | Acceptance |
| --- | --- | --- | --- |
| T1 / P0 | Atlas messaging schema: connectors, threads, turns, messages, revisions and audits; `migrations/api/`, API models/repositories | HEX-12; EPIC-004 | Fresh and existing DB migration; tenant-qualified references, duplicate event identity and one-open-turn race tests |
| T2 / P0 | Claim/reassign/control/escalation APIs; API service/controller/DTO/router | T1; HEX-13 | One winner for simultaneous claim; current-member/tenant checks; takeover fences queued work; resolve never resumes automation |
| T3 / P0 | Local Kafka profile, typed envelope, transactional outbox relay and consumer inbox; `infrastructure/compose.demo.yml`, API infrastructure/repositories | T1; EPIC-005 | Host/internal listeners smoke; duplicate publication and broker outage recovery; replay cannot send or call paid models |
| T4 / P0 | Script versions, four independent grants and owner reply settings; API migrations/repository/service/HTTP | T1; HEX-12/13 | Create inactive; active edit creates draft; archive preserves history; owner-only grants/settings; independent permission denial tests |
| T5 / P0 | Durable Go dispatcher to existing Python RPC and result acceptance; API bootstrap/services and RPC snapshot if needed | T2/T3/T4; EPIC-002/004 | Stable run ID, restart recovery; obsolete context/control/config/script results rejected; human edits preserved |
| T6 / P0 | Proposal editing/review, simulated outbound intents/attempts, final gate and 24h expiry | T5 | Both suggest/auto_send work in simulation; stale draft needs review; timeout stays unknown; no blind resend; injected-clock expiry and needs-human exemption |
| T7 / P0 | Conversation workspace and script/settings UI; `apps/web/app/`, API clients/SDK | T2/T4; complete with T6; HEX-14 | User can receive simulated input, claim, inspect/edit/review/send; clear simulation/status/error/stale indicators; two-sale takeover scenario passes |
| T8 / P0 parallel | Owner dataset delivery and evaluation specification/fixtures | HEX-42 + HEX-25 | Ten named plants with sources/media/missing fields; separately labeled synthetic cases; baseline quick replies vs AI measures total time, edits, errors and media mismatch; thresholds chosen before measured evaluation |
| T9 / P1 | Product/customer consultation APIs/UI and source/media mapping; API + web | T1 and available dataset; HEX-18 | Read/write tenant boundaries; explicit missing fields; media attached to correct product; only authorized facts included in AI context |
| T10 / P1 | OpenAI/Anthropic/Gemini adapters and model router/generator, checkpoints and self-host traces; `services/ai/`, local profile | T5/T8/T9; EPIC-006 | Contract fakes first; script/AI/clarify/human branches, one clarification then handoff, no invented price/policy; server keys, one provider per run; paid replay/cost controls; redacted trace outage does not authorize/block sends |
| T11 / P1 | Audit and align legacy orders; incomplete draft CRUD, assigned-sale confirmation, cancellation/return cases | T9; EPIC-007 | Distinguish omitted/zero price; positive quantity; name/contact/sale; address only for delivery; handler updates and owner finalizes cases; no financial execution or conversation-driven confirmation |
| T12 / P1 | Sequential versioned workflows, responsibility handoff and UI | T11 or independent work-item scope after detailed contract; EPIC-007 | Add/remove/rename/reorder; optional packing; direct vs nominated acceptance; old owner retains responsibility while waiting; running instance pins version; reassignment audited |
| T13 / P1 | Dashboard projections and metric definitions | T9/T11/T12 | Time/employee filters reconcile to underlying rows; define asked/bought event sources before counting; missing values explicit; order value is not collected revenue |
| T14 / P2, access preparation parallel | Verify authorized Pancake account, implement real connector and reconcile | T6 + actual entitlement; EPIC-008 / follow-up to HEX-41 | Real payload/auth/page mapping and send outcomes verified; unknown remains unknown if provider lookup unavailable; distinguish accepted/delivered/read; no simulator claim of live sending |
| T15 / P2 | VPS/domain, HTTPS, deployment roles, retention, backup restore and measured capacity | Stable demo, load/storage evidence and selected host/domain | Repeatable deploy and restore; restricted services/secrets/traces; measured limits; no GPU assumed |

## Active delivery sequence — owner revision

The owner explicitly deferred key-dependent tickets and mock-data work. Optimize for a browser-operable product demo that accepts user-entered records and persists actual application state. No dataset preparation, seed campaign, webhook simulator, fake channel sender or fake AI integration is a prerequisite. Existing fake AI may remain available independently but is not on the demo critical path. This is delivery sequencing, not cancellation of approved messaging/AI/Kafka scope.

### Demo A: manual sales and responsibility flow

One coherent path: log in -> select tenant -> enter customer/product -> save incomplete order draft -> assign sale -> explicitly confirm -> create linked work item -> hand over to a nominated employee -> accept/update work -> owner sees current step and responsible person. Records are entered through the product forms; no imported/mock dataset or key is needed. Browser refresh must preserve progress. Confirmation and work creation are explicit, separate actions. Workflow configuration is user-entered rather than hardcoded to packing.

| Batch | Work | Existing mapping / scope | Demo acceptance |
| --- | --- | --- | --- |
| D1 | Workspace shell: tenant navigation, member/invite access, shared loading/error/empty states; reuse auth, tenant API, SDK and UI | HEX-14, existing platform | Login and tenant switch work; owner can identify/invite the employees needed for handoff; no fabricated dashboard numbers |
| D2 | Minimum customer/product CRUD and matching forms with user-entered data; expose existing model fields through tenant-authorized services | HEX-12/13/14; narrowed T9 | Create, edit and reload a customer and product/variant; missing consultation data explicit; foreign tenant references rejected |
| D3 | Align existing order service/DTO/repository with incomplete draft editing and explicit assigned-sale confirmation, and deliver order list/detail/forms in the same slice | HEX-12/13/14; narrowed T11 | Incomplete draft persists; confirmation validates name/contact/items/quantity/explicit price/sale/address only if delivery; omitted price differs from zero; no auto_confirm shortcut bypasses these gates |
| D4 | Minimum configurable sequential workflow and linked work item: versioned ordered steps, owner-nominated employees, direct/accept handoff; simple forms and work list/detail | HEX-18; narrowed T12 | Two employees can hand over and accept work; old responsibility remains until acceptance; owner can reassign; history and version survive reload; no mandatory packing or shipping |
| D5 | Owner overview over real saved records: current step/assignee, waiting/active work and order status counts | HEX-18; narrowed T13 | Counts match list filters; time/employee filtering; no revenue/ad-performance claims or asked-product counts without a defined source |

Deliver D1/D2/D3 as the first usable increment and demo it before D4/D5. Build each batch vertically (migration where needed + API + SDK + UI + relevant checks); do not finish all backend tickets before exposing any usable screen. Reuse existing code after checking correctness; avoid generic CRUD frameworks or unrelated refactors. D4 enables the second demo focused on business responsibility.

### Demo B: manual conversation workspace

After Demo A feedback, deliver tenant-isolated threads/turns/messages, list/detail, atomic self-claim, owner reassignment and manually edited reply drafts. An internal record/note is not an externally sent message. With no real connector or simulator in this slice, do not expose a successful customer-send action or manufacture incoming events. If the UI needs conversation creation, label it as a manually recorded conversation. Add script authoring/versioning/independent grants only when it improves the reviewed flow. Keep actual sending, automatic routing and proposal generation gated until their deferred dependencies are selected.

This narrows T1/T2/T4/T7 without pulling in T3/T5/T6. Preserve tenant scope and assignment checks from the first command. Full 24-hour automation, stale AI result acceptance and external send reconciliation follow when the automated messaging slice is scheduled.

### Deferred queue

- HEX-42 ten-plant dataset, HEX-25 research dataset/AI evaluation, mock/seed-data generation, new simulator fixtures as deliverables, webhook/channel simulation tools.
- Provider keys, real OpenAI/Anthropic/Gemini model integration and paid validation; real Pancake entitlement/token/send validation. HEX-41 research remains completed.
- Kafka integration, distributed relay/inbox, node checkpoints and self-host Langfuse are removed from the first-demo critical path. PostgreSQL transactions own local business commands and audit history; do not introduce a temporary Redis queue. Kafka remains the approved future event backbone.
- Full cancellation/return cases, media pipelines, advanced reporting and VPS/domain provisioning follow the first demo and feedback. No payment/refund execution is added.

### Checks and definition of demo readiness

Each batch retains tenant authorization, current-member/assignee checks, server validation and transactional state changes. For D3/D4 test incomplete/invalid confirmation, simultaneous transitions, cross-tenant references and responsibility transfer; run applicable migration/Go/frontend checks and regenerate Swagger when changing endpoints. Verify the delivered browser path and persistence after reload with manually entered records. Feature tests may use minimal ordinary test inputs; a standalone mock-data preparation project is deferred. A list of coded screens without a working path does not pass the demo gate.

Existing HEX umbrella tickets should be split into these vertical batches before implementation. Do not mark them complete from partial demo delivery. No Notion writes occurred in the planning revision. Subsequent implementation is recorded below. No duration estimate is asserted before inspecting each batch in detail.

## Risks and edge cases

Takeover versus dispatch, duplicate/reordered inbound events, script revocation while generation runs, worker crash after external send, retained human edits and late receipts require explicit acceptance tests. Keep DB transactions short and final send authorization in Go. Real connector capabilities, model budgets/thresholds, storage/retention and metric definitions remain open; do not infer them from target designs. No bulk implementation or tracker write is authorized by this proposal.

## Definition of done and verification

For each implementation slice: scoped unit/integration tests and applicable Go race/vet, Atlas integrity/migration checks, Swagger regeneration for endpoint changes, frontend type/lint/build for UI changes, and a reproducible local simulator smoke. Capture browser end-to-end evidence for T7. Update relevant architecture modules and epic sessions, review and verify merge into main. No new runtime correctness is claimed by this planning session.

## D1–D3 local implementation handoff

D1–D3 are now implemented and locally verified on codex/manual-sales-demo; they are not merged/done. See [the EPIC-007 session](../epics/EPIC-007-orders-workflows/sessions/2026-10-19_manual-sales-demo.md) for exact checks and limitations. The demo ends at confirmation; D4/D5 and the deferred queue remain unimplemented. No Notion ticket or issue lifecycle state was changed.

## Live reconciliation follow-up — 2026-10-19

PR #54 is confirmed merged with successful CI, so earlier unmerged D1–D3 labels are historical. The active next sequence remains manual acceptance/bug fixes -> D4 versioned responsibility workflows -> D5 factual overview -> manual conversation after feedback. See [all-ticket assessment](2026-10-19_ticket-reconciliation_plan.md) for current Notion states/acceptance mapping. No Notion writes or parent-epic completion.

## D4 implementation follow-up — 2026-10-20

After bounded manual D1–D3 browser checks, owner requested moving straight to the next work. The D4 workflow/work-item vertical slice is implemented locally on codex/order-workflows; [session](../epics/EPIC-007-orders-workflows/sessions/2026-10-20_order-workflows.md) records automated/database/build/upgrade checks and pending browser/delivery gates. Earlier D4-unimplemented wording is historical. D5 remains the next separate feature after D4 review; remaining earlier manual cases are still open. No parent task completion or external tracker write.
