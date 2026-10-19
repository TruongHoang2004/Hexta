# HTTP API design

## Implemented contract inventory

Base is `/api/v1`. Actual routing: [router.go](../../../services/api/internal/present/http/router/router.go); DTOs: [dto directory](../../../services/api/internal/present/http/dto). Current Swagger is generated under `services/api/docs`; verify annotations against controllers before relying on it. Swagger is registered at `/api/v1/swagger/index.html`, not at host-root `/swagger/index.html`.

| Method/path after base | Input | Status / scope |
| --- | --- | --- |
| GET /ping | none | IMPLEMENTED public liveness response |
| GET /health | none | IMPLEMENTED DB/Redis health; inspect returned data.status, not only HTTP connectivity |
| POST /auth/register | email, password | IMPLEMENTED account/session tokens; password min 6 in current DTO |
| POST /auth/login | email, password | IMPLEMENTED account/session tokens |
| POST /auth/refresh | refresh_token | IMPLEMENTED token refresh |
| POST /auth/logout | optional session_id | IMPLEMENTED authenticated session logout; ownership checks apply |
| GET /auth/google/login | none | IMPLEMENTED OAuth redirect/state; local profile has no OAuth credentials |
| GET /auth/google/callback | provider code/state | IMPLEMENTED account linking; not configured in local smoke |
| POST /tenants | name, slug, plan | IMPLEMENTED authenticated tenant create |
| GET /tenants | none | IMPLEMENTED list caller's tenants |
| GET /tenants/{id} | tenant ID | IMPLEMENTED tenant query |
| PUT /tenants/{id} | optional name, plan, status | IMPLEMENTED update; inspect service role checks |
| GET /tenants/{id}/users | tenant ID | IMPLEMENTED members list |
| POST /tenants/{id}/invites | user_id, role | IMPLEMENTED current membership API; not the target fine-grained script-grant API |
| POST /orders | CreateOrderRequest | IMPLEMENTED legacy create; not the target incomplete-draft contract |
| GET /orders | tenant context/pagination | IMPLEMENTED legacy order list |
| GET /orders/{id} | tenant context + order ID | IMPLEMENTED legacy order query |
| PATCH /orders/{id}/status | status | IMPLEMENTED legacy state transition |
| PATCH /orders/{id}/payment | payment_status | IMPLEMENTED legacy metadata; payment execution is outside approved MVP |
| POST /ai/agent/converse | prompt, conversation_id?, history[] | IMPLEMENTED legacy Gemini order-assistant SSE; not Python messaging gRPC |
| GET /ai/agent/drafts/{id} | tenant context + draft ID | IMPLEMENTED legacy Redis order draft lookup |
| GET /swagger/*any | none | IMPLEMENTED public reference under base |

Current order/legacy AI controllers accept `X-Tenant-ID` or tenant query/body context depending on action; these values are claims to authorize, not proof of membership. Read the exact controller/service before extending it. Target messaging routes use an explicit tenant path with server-side membership/resource checks. No security audit of every existing endpoint is implied by this inventory.

## Existing DTO details and compatibility gaps

- Register/login: email required/valid, password required min 6; response data carries access_token, refresh_token, session_id, user_id (refresh response has tokens). Preserve existing SDK/session behavior; design changes to cookie handling explicitly.
- Tenant create: name/slug required, length 2..100, optional plan. Tenant update: optional name/plan/status. Invite: required user_id, optional role.
- CreateOrderRequest: tenant_id, optional customer_id/employee_id, required nonempty items `{variant_id,quantity>=1,unit_price}`, discount, source, draft_id, notes, auto_confirm. Decimal validation currently differs from the approved explicit-nonnegative-price requirement; do not assume zero/missing semantics are correct without tests.
- Order response: identity/tenant/number, customer/employee, status/payment_status, subtotal/discount/total_amount, source/draft/notes/items and timestamps. List wraps items,total,page,size,total_pages.
- Legacy AI request history entries are role/content. SSE events are token, draft_proposed, tool_call, error, done. Draft includes customer/contact/address/items/price/payment_method/expiry. This is not the future reply-proposal schema.
- Standard Go success wrapper is `response.Response[T]` with `data` and `meta`. Pagination DTOs differ by existing endpoint; keep compatibility until a migration plan standardizes them. Use typed shared errors from `packages/shared/pkg/errors`; do not invent one universal legacy error JSON shape.

## Target HTTP contract conventions

All routes below are TARGET, absent from the baseline router unless separately listed as implemented. Go owns authentication, tenant authorization and validation. Adopt transport DTOs and Swagger per AGENTS.md. Request IDs/correlation IDs are observability identifiers, not authorization.

Mutations that create work accept `Idempotency-Key`; persist `(tenant,actor/operation,key,request_hash,result_ref)` transactionally. Exact repeat returns the original semantic result; same key/different payload => 409. Mutable resources accept `expected_version`; stale state => 409 with current authorized metadata. Identity comes from auth, never request actor IDs. A resource belonging to another tenant is not disclosed.

Error meanings: malformed input 400, unauthenticated 401, authenticated but disallowed action 403, absent/inaccessible resource 404, version/idempotency/assignment conflict 409, typed field validation 422 where the new contract adopts it, quota 429, accepted async work 202, temporary dependency outage 503. Existing legacy handlers may differ; do not silently rewrite them. No new handler should leak raw SQL/provider/credential details.

Cursor pagination for conversations/messages/scripts must have a stable order and maximum page size. SSE is tenant-authorized, sends entity IDs/revisions for refetch, supports reconnect or snapshot fallback and must not expose connector secrets/raw model traces. New browser notification SSE is distinct from legacy token SSE.

## Target route catalog

| Method and route | Request | Result / guard |
|---|---|---|
| POST /integrations/pancake/{connector}/webhooks | provider payload | 200 after durable receipt; tenant resolved from connector |
| POST /simulation/{connector}/events | normalized fixture | development/demo credentials, simulated connector only |
| GET /tenants/{tenant}/conversations | cursor, assignee/unassigned/needs_human filters | paginated thread/current-turn summaries |
| GET /tenants/{tenant}/conversations/{id}/messages | cursor, turn | message and delivery states, no provider secrets |
| POST /tenants/{tenant}/turns/{id}/claim | expected_version | updated turn or 409; atomic self-claim |
| POST /tenants/{tenant}/turns/{id}/assign | assignee_id, expected_version | owner permission; current membership required |
| POST /tenants/{tenant}/turns/{id}/return-to-system | expected_version | assigned sale; does not replay old sends |
| POST /tenants/{tenant}/turns/{id}/resolve | expected_version | assigned sale, clears escalation and resets timer |
| POST /tenants/{tenant}/turns/{id}/proposals | context_revision | 202 routing job; assigned sale for manual regeneration |
| PATCH /tenants/{tenant}/proposals/{id} | body, media_ids, expected_version, reviewed_context_revision | save edits; preserve stale indication until reviewed |
| POST /tenants/{tenant}/turns/{id}/send | proposal_id or manual body/media, expected_version, context_revision | 202 outbound intent; assigned sale and current context |
| GET /tenants/{tenant}/outbound-intents/{id} | none | accepted/failed/unknown; no delivered/read inference |
| POST /tenants/{tenant}/outbound-intents/{id}/reconcile | expected_version | 202 read-only provider check; not resend |
| GET /tenants/{tenant}/events | last event ID | SSE updates or authorized snapshot refetch |
| GET /tenants/{tenant}/scripts | cursor, archive state | active/draft metadata |
| POST /tenants/{tenant}/scripts | name, applicability, template, variables, media | script.create, inactive version |
| POST /tenants/{tenant}/scripts/{id}/versions | content + expected_version | script.edit, immutable draft |
| POST /tenants/{tenant}/scripts/{id}/activate | version_id, expected_version | script.activate, same script+tenant |
| POST /tenants/{tenant}/scripts/{id}/deactivate | expected_version | script.activate; preserve version history |
| POST /tenants/{tenant}/scripts/{id}/archive | expected_version | script.archive; no physical delete |
| PATCH /tenants/{tenant}/reply-settings | mode, expected_version | owner only; invalidate obsolete automatic jobs |
| PUT /tenants/{tenant}/members/{user}/permissions | exact permission set + expected_version | business owner only; target must be a current tenant member; audit grants and revocations |
| POST /tenants/{tenant}/workflow-definitions/{id}/versions | ordered steps | owner; new immutable version for new work |
| POST /tenants/{tenant}/work-items/{id}/handoff | next recipient, expected_version | actor/recipient checks; direct or awaiting acceptance |
| POST /tenants/{tenant}/work-items/{id}/accept | expected_version | nominated recipient only |
| POST /tenants/{tenant}/work-items/{id}/assign | user_id, expected_version | owner; does not change workflow version |
| POST /orders, GET /orders, GET /orders/{id} | draft creation permits partial fields; paginated reads | review compatibility with existing stricter create DTO; never implicitly auto-confirm |
| PATCH /orders/{id} | partial draft fields, expected_version | permit incomplete business information while validating supplied values/types/tenant references |
| POST /orders/{id}/confirm | expected_version | assigned sale only; customer name + contact, at least one product with positive quantity and explicit nonnegative unit price, assigned sale, and address if delivery is required; scope audit stock coupling first |
| POST /orders/{id}/exception-cases | kind, reason, line quantities | operational case, not refund |
| POST /exception-cases/{id}/receive | line quantities/condition/version | employee assigned to the case updates progress; guard duplicate receipt and original line quantity |
| POST /exception-cases/{id}/resolve | result, expected_version | business owner confirms final cancellation/return outcome; progress update permission does not grant finalization |

## Command examples (TARGET)

Claim: `POST /api/v1/tenants/{tenant}/turns/{turn}/claim`, authenticated sale, `Idempotency-Key`, body `{"expected_version":12}`. Only one simultaneous claimant succeeds; return the incremented version, assignee and sales control. No owner manual close endpoint is required by the approved inactivity policy.

Send: `POST /api/v1/tenants/{tenant}/turns/{turn}/send`, body `{"proposal_id":"...","expected_version":13,"context_revision":8}`. Return 202 with intent ID/state. Assigned sale and current reviewed context are required. The dispatcher repeats the gate at initiation. If prior dispatch outcome is unknown, a reconcile action does not enqueue a new send.

Reply settings: `PATCH .../reply-settings`, body `{"mode":"suggest","expected_version":2}`; owner-only. `suggest` means proposals enabled; `auto_send` means proposal review disabled but all automatic-send gates still apply.

Permissions: `PUT .../members/{user}/permissions`, exact set drawn from script.create/edit/activate/archive plus expected_version. Only owner grants/revokes; target user must be a current member. Omitted grants under replace-set semantics are revoked, so UI must present the entire set explicitly.

Proposal edit: `PATCH .../proposals/{id}`, body with body/media_ids, expected_version and reviewed_context_revision. Compare draft_version independently of context revision; retain human edits when new customer input makes the basis stale. Regeneration creates/replaces only unedited eligible candidate state by an explicit policy.

Order PATCH allows partial draft updates; confirm is separate. Confirm must enforce the approved product rules even if the old CreateOrderRequest currently accepts auto_confirm. Exception receive is assigned-handler progress; exception resolve is owner finalization. Neither executes a financial refund.

## Webhook and simulator requirements

Real webhook provider payload/ack/authentication are OPEN pending entitlement and documented Pancake capability checks. Tenant comes from configured connector. Limit payload size, preserve source timestamps, record dedup identity/revision and commit receipt/outbox before acknowledgement. Unsupported/ambiguous payloads are quarantined without fabricating identities.

Simulator endpoints must require development credentials and an explicit simulated connector. Simulator data can exercise the full target event pipeline but cannot reach a real sender. A simulated send result must remain visibly marked, never presented as delivered to a customer.

Customer/product CRUD, metrics/dashboard definitions, media upload and richer workflow endpoints need their own compatible DTO specification. They are not registered public CRUD APIs simply because corresponding DB tables exist.
