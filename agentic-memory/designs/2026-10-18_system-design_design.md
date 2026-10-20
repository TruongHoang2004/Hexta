# Hexta system design: canonical entry point

Last reconciled: 2026-10-19. Code baseline: `01b2e11`, merged to main as `9784695` through [PR #52](https://github.com/TruongHoang2004/Hexta/pull/52) on 2026-10-19. Documentation [PR #53](https://github.com/TruongHoang2004/Hexta/pull/53) merged on 2026-10-19. The manual sales demo on codex/manual-sales-demo is locally implemented and verified, but unmerged; current sales additions are recorded in the architecture/DB/API/workflow modules. This specification consolidates owner decisions and the current implementation. It replaces earlier dated designs as the navigation/interpretation baseline, without deleting their history.

## Read order

1. [Epic index](../epics/README.md): select the relevant epic and read its own context/session records.
2. [Product decisions and terminology](system/01-product-decisions.md).
3. [Architecture and ownership](system/02-architecture.md).
4. [Database: existing physical schema and target logical model](system/03-database.md).
5. [HTTP APIs: implemented inventory and target contracts](system/04-http-api.md).
6. [gRPC, AI graph and durable execution](system/05-grpc-ai.md).
7. [Kafka events and delivery design](system/06-kafka-events.md).
8. [Business workflows and concurrency](system/07-workflows.md).
9. [Operations, security, tests and implementation roadmap](system/08-operations-security.md).

## Evidence labels

| Label | Meaning |
| --- | --- |
| IMPLEMENTED | Present in the inspected baseline. This does not mean every business scenario has been runtime-tested. |
| VERIFIED | A named runtime/test check actually passed; see operations for the exact verification boundary. |
| DECIDED | The owner has chosen the requirement/direction; code may still be missing. |
| TARGET | Technical contract/schema proposed to realize decided behavior; refine compatibly during implementation. |
| OPEN | Requires evidence, measurement or an explicit owner choice. Never silently treat as implemented/approved. |

Product decisions come from the owner's conversation and approved Notion requirements. Code/migrations/Protobuf determine implemented behavior. When these differ, retain both as an explicit gap; do not let a legacy endpoint override the approved scope. New explicit owner instructions supersede this snapshot; update this pack when they change a decision.

## System at a glance

Hexta is a tenant-isolated sales/operations application. Messaging assistance is initially integrated through Pancake; order/workflow operations are separate. Go owns business authorization and sending. Python owns AI execution behind gRPC. Kafka is the owner-selected future event backbone. PostgreSQL owns business truth and AI run state. Redis is a cache, not the durable event/job authority. AI observability will be self-hosted.

The current local stack runs Next.js, Go API, Python fake AI, PostgreSQL, Redis, role/schema initialization and Atlas migrations. Authentication/tenant pages and legacy order/AI-draft foundations exist. A local unmerged follow-up adds customer/product/order pages and a guarded incomplete-draft/explicit-confirmation HTTP boundary. The new conversation model, Go conversation worker, real provider adapters, Kafka producers/consumers, Pancake connector and Langfuse are not implemented in this slice.

## Maintenance contract

When changing a schema, endpoint, RPC, event or business rule, update its corresponding module and the affected epic/session records in the same change. Keep target and implemented sections separate. Do not add a new dated design that silently contradicts this pack; link the decision and amend the relevant section. Keep secret values out of all documents. Record validation in CHANGELOG.md with the actual command/result, not an assumed test outcome.

The original consolidation was documentation only. Local unmerged sales/tenant-first follow-ups now add schema, API and UI as explicitly recorded in the modules; future messaging/sending capabilities remain targets.

## D4 local implementation — 2026-10-20

Configurable sequential workflow versions, explicit order-linked work, direct/accepted handoff, owner reassignment/unassigned queue and work history are now implemented locally on codex/order-workflows. Database/API/workflow/operations modules contain the current boundary. [Design](2026-10-20_order-workflows_design.md), [session](../epics/EPIC-007-orders-workflows/sessions/2026-10-20_order-workflows.md). Automated checks/build/upgrade passed; browser verification and review/merge remain. D5 overview and exceptions/messaging remain separate.
