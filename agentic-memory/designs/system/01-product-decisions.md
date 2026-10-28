# Product decisions and domain vocabulary

Status: DECIDED unless marked TARGET/OPEN. These requirements are durable session context, not evidence that the application already implements them.

## Purpose and scope

The immediate goal is a demo/MVP for experimentation. The demo will eventually run on a VPS; VPS capacity and domain have not been selected. Build and verify locally first. The owner supplies API keys later. No real provider key, live Pancake entitlement or customer sending is required to exercise the local fake flow.

Hexta must fit businesses with different operating steps. Packing/shipping are optional workflow capabilities, not universal requirements. Keep messaging independent of orders; closing a consultation never automatically creates, confirms, cancels or fulfills an order. An eventual Pancake replacement is a later direction, not the current MVP scope.

## Terminology

| Term | Meaning |
| --- | --- |
| Tenant / business | Isolated business workspace, identified server-side through authorized membership/connector mapping |
| Owner | Business owner; only actor allowed to grant/revoke employee permissions and change reply settings |
| Sale | Employee handling customers; coarse existing role names do not yet implement all sales permissions |
| Connector | Tenant/page binding to Pancake or a clearly separated simulator |
| Thread | Stable provider conversation identity |
| Consultation turn | One Hexta handling period inside a thread; a later message after closure begins a new turn |
| Control | System or sale owns the right to drive replies; distinct from assignment |
| Needs human | Escalation flag that blocks automation and ordinary expiry |
| Script | Managed response template with immutable versions and one eligible active version |
| Proposal | Generated/edited reply candidate; not a sent message and not a legacy order draft |
| AI run | Durable execution record in Python; successful execution is not authorization to send |
| Outbound intent / attempt | Authorized sending request versus one external transport attempt |
| Work item | A business process instance that may reference an order, with optional configurable steps |

## Conversation policy

- Every sale in a tenant can see that tenant's conversations. Only the assigned sale can manually send; self-claim an unassigned open turn atomically.
- The owner can reassign a turn to a current tenant member. Assignment/control changes invalidate obsolete permissions and pending automatic work.
- When a sale takes a turn, the system must stop automatic sending. This includes queued AI results, not just future incoming messages. Already-dispatched external requests may be impossible to retract; reconcile their outcomes.
- Returning to system handling is explicit. Do not silently resume automation on resolution, timeout, reconnect or generation completion. TARGET default: retain the assignee on explicit return so the sale can still review proposals.
- A business owner controls the proposal setting. Proposal enabled => generate for sale review; proposal disabled => eligible automatic reply. Name the internal modes `suggest` and `auto_send`; explain the toggle in UI to avoid reversal.
- Not every message should invoke a generative model. Route among approved script, AI generation, clarification and human handoff. The owner asked for a model-based evaluation before deciding whether a script suffices; the current exact matcher is only a deterministic fake.
- Human handoff must stop automation. Unassigned items remain in the assignment queue; do not randomly choose a sale. Customer-facing fallback responses follow business configuration and channel authorization.

## Turn expiry

Use a single 24-hour inactivity period for MVP; do not require the business owner to close turns manually. Customer messages and confirmed successful sale/system messages reset activity from the relevant source event time. Drafts, failed sends, duplicates and unknown outcomes do not.

`needs_human=true` prevents automatic expiry. The assigned sale marks the issue resolved; resolution restarts the timer but does not return control to the system. At eligible expiry close the turn, cancel unstarted sends and retain history. A later genuinely new customer message opens a new unassigned turn using current business reply configuration. Delayed/replayed historical events must not manufacture a new turn or extend inactivity to processing time.

## Scripts and delegated permissions

Only the business owner grants/revokes permissions. Four independent grants:

| Permission | Enables | Does not imply |
| --- | --- | --- |
| `script.create` | Create an inactive script/draft | Edit other scripts, activate, archive, grant rights |
| `script.edit` | Create/edit draft content; active edits produce a new version | Activate or archive |
| `script.activate` | Activate a selected version / deactivate current version | Edit, archive or grant rights |
| `script.archive` | Archive script while retaining history | Physical deletion, edit or grant rights |

New scripts start inactive. Existing active content remains immutable; publishing switches an active-version pointer. Historical proposals retain their source versions. Coarse owner/admin/member roles in the current code do not prove these grants exist.

## Orders and flexible workflows

Orders have a separate UI/module. A draft may be incomplete. Explicit confirmation requires customer name and contact, at least one product, positive quantities, an explicitly supplied nonnegative unit price, assigned sale, and an address when delivery is selected. Do not treat missing price as free. These are approved requirements, not a description of today's CreateOrder DTO.

The assigned sale confirms an order. Assigned operational handlers update cancellation/return cases; the owner finalizes them. Preserve original order and case history. No payment collection or refund execution was approved. Existing payment flags, inventory reservation and `auto_confirm` code are legacy capabilities requiring scope audit before reuse.

Workflow definitions should be versioned; optional steps such as packing/delivery are chosen by each business. Instances retain the selected version. Returning unassigned work sends it to the assignment queue. Fine-grained exception transitions and templates beyond these decisions remain TARGET/OPEN.

## AI and infrastructure decisions

- Separate Python service for LangChain/LangGraph; simple graph for MVP, no multi-agent swarm or mandatory vector store.
- Go <-> Python internal gRPC selected; browser uses Go HTTP/SSE.
- Support OpenAI, Anthropic and Gemini via server-side API keys; one selected provider/model per run, no automatic cross-provider failover. The owner will supply keys. No per-tenant key-entry UI is required for MVP.
- Self-host flow observability; Langfuse is the proposed implementation. Self-hosting the application/traces does not mean running local model inference or needing a GPU.
- Kafka requested for high-volume events; detailed topics/reliability are TARGET in the Kafka module. Business states and drafts still belong in PostgreSQL.

## Still open

Provider/model choices, cost/concurrency budgets, deployment secret provisioning, real Pancake webhook/send capabilities, media processing/storage details, VPS/domain sizing, retention/backup policy, production broker topology and measurable SLOs. Ten-tree dataset task HEX-42 was still missing in the earlier planning context; verify Notion status before scheduling around it.

## Tenant commercial boundary and onboarding — owner correction 2026-10-19

DECIDED: tenant is the paying business customer and top-level customer-product boundary; orders, chat, products, customers, members and settings belong within it. Customers apply to register a tenant; Hexta reviews and approves before owner access is granted. Tenant owners invite employees by email; recipients sign in and accept before active membership. Account registration alone grants no tenant access. The initial owner's membership is provisioned by approval as the onboarding bootstrap interpretation. No unrestricted create-and-open workflow is approved by this correction.

TARGET route/data/authority changes and remaining OPEN delivery/billing details are recorded in [the tenant-first framework plan](../../ARCHIVE_INDEX.md#pruned-2026-10-19-tenant-first-framework-plan). Payment integration, pricing, email delivery/verification, operator bootstrap and suspension policy are not implemented or silently decided by this clarification.

IMPLEMENTED locally/unmerged: approval-first onboarding, route-qualified tenant resources, operator CLI authority and email-addressed manual invitations with explicit acceptance. Seven-day expiry is an implementation default. Email provider/mailbox verification, billing automation and broader suspension policy remain OPEN. Existing tenants are preserved. See [implementation design](../2026-10-19_tenant-onboarding_design.md).
