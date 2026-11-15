# Technical Design: Kết nối nền tảng Facebook (MVP v1.0)

- **Status:** IMPLEMENTED (2026-11-15, [PR #69](https://github.com/TruongHoang2004/Hexta/pull/69) in review); live Meta acceptance NOT_RUN. Implementation amendments are folded into this text (log in §7) and were accepted by the owner on 2026-11-15. Evidence: [implementation session](../epics/EPIC-008-integrations-deployment/sessions/2026-11-15_facebook-platform-connection.md).
- **Date:** 2026-11-14
- **PRD:** [PRD — Kết nối nền tảng Facebook (Notion, Đã chốt v1.0)](https://app.notion.com/p/PRD-K-t-n-i-n-n-t-ng-Facebook-418b0ea49f09816f8f8be1015298dd7d)
- **Baseline:** Extends [2026-11-03_connector-service_design.md](2026-11-03_connector-service_design.md) (`IMPLEMENTED` local/PR #66: Facebook webhook text/echo ingress -> connector outbox -> Kafka `hexta.ingress.received.v1` -> `services/api/cmd/ingress` -> Inbox).
- **Boundary Note:** Covers **only** Facebook platform connection (OAuth/Page management, inbound reception, outbound transport). Conversation routing and reply orchestration (Chatbot, AI, Desk, Sale workflow) are **out of scope**.

---

## 1. Architecture & Ownership Boundaries

| Component | Owns | Must NOT Do |
| --- | --- | --- |
| **`services/api` (Go)** | Owner-only channel authorization, `messaging_connections` mirror (1 Page = 1 Tenant), single-use OAuth CSRF state (`channel_oauth_states`), outbound intent outbox & single-attempt dispatch, Inbox/Settings HTTP APIs. | Call Meta Graph API directly, store raw Page Access Tokens in API DB, or embed Chatbot/Desk routing logic. |
| **`services/connector` (Go)** | Facebook plugin (`internal/plugin/facebook`): Meta OAuth token exchange, encrypted Page token storage, `PUT /internal/connections` control plane, webhook HMAC + receipt/outbox, Meta Send API execution & result outbox. | Access API PostgreSQL schema, make tenant/business authorization decisions, or auto-retry failed/unknown sends. |
| **`packages/shared`** | Additive JSON event contracts (`ingress`, `outbound`) and Kafka transport helpers. | Hold business logic. |
| **`apps/web` (Next.js)** | Owner Channel Settings UI (OAuth flow, Page picker, status badges, pause/disconnect, reconnect), expired-connection banner, media placeholder UI, outbound send status. | Hold Meta App Secret, Page Access Tokens, or connector management tokens. |

---

## 2. Inter-Service Communication & API Contracts

### 2.1 HTTP Endpoints — `services/api` (Public / Tenant-Scoped)

| Action | Method & Path | AuthZ | Request DTO | Response DTO |
| --- | --- | --- | --- | --- |
| **Add** | `POST /api/v1/tenants/{tenantId}/channels/facebook/oauth/start` | Tenant `Owner` | `StartFacebookOAuthRequest { RedirectURI string }` | `StartFacebookOAuthResponse { AuthURL string, ExpiresAt time.Time }` |
| **Add** | `GET /api/v1/channels/facebook/oauth/callback` | Authenticated `Owner` that started the flow (single-use `state`, consumed before the code exchange) | Query: `code string, state string, error? string` | `FacebookOAuthPagesResponse { TenantID string, OAuthSessionID string, ExpiresAt time.Time, Pages []SelectableFacebookPageDTO }` (tenant comes from the state, since Meta needs one fixed redirect URI `/channels/facebook/callback`) |
| **Add** | `POST /api/v1/tenants/{tenantId}/channels/facebook/connections` | Tenant `Owner` | `ConnectFacebookPagesRequest { OAuthSessionID string, PageIDs []string }` | `ConnectFacebookPagesResponse { Connections []FacebookConnectionDTO }` |
| **Add** | `GET /api/v1/tenants/{tenantId}/channels/facebook/connections` | Tenant `Member` (Owner sees management controls; Sale sees status/banner) | — | `ListFacebookConnectionsResponse { Connections []FacebookConnectionDTO }` |
| **Add** | `PATCH /api/v1/tenants/{tenantId}/channels/facebook/connections/{connectionId}` | Tenant `Owner` | `UpdateFacebookConnectionRequest { Action string /* "pause" \| "resume" \| "disconnect" */ }` | `FacebookConnectionDTO` |
| **Modify** | Inbox message reads (`GET /api/v1/tenants/{tenantId}/conversations/{id}/messages`) and new real send `POST /api/v1/tenants/{tenantId}/turns/{id}/messages` | Read: tenant `Member`. Send: the assigned sale under sales control, connection `active` | Send: `SendMessageRequest { ClientMessageID string /* UUID */, Text string /* 1–2000 */, ExpectedVersion int64, ContextRevision int64 }` (version/context keep the stale-reply guard) | Message DTO adds `ContentType string` (`"text" \| "unsupported_media"`), `PlaceholderMeta *PlaceholderMetaDTO`, `DeliveryStatus string` (`"received" \| "sending" \| "sent" \| "failed" \| "unknown"`), `DeliveryErrorReason string` |

#### Key HTTP DTO Definitions (`services/api/internal/present/http/dto`)
```go
type SelectableFacebookPageDTO struct {
    PageID       string `json:"page_id"`
    PageName     string `json:"page_name"`
    Availability string `json:"availability"` // "available" | "already_connected_this_tenant" | "owned_by_other_tenant"
}

type FacebookConnectionDTO struct {
    ID               string    `json:"id"`
    TenantID         string    `json:"tenant_id"`
    Platform         string    `json:"platform"` // "facebook"
    AccountID        string    `json:"account_id"` // page_id
    AccountName      string    `json:"account_name"`
    Version          int64     `json:"version"`
    Status           string    `json:"status"` // "active" | "paused" | "disconnected" | "auth_expired"
    LastErrorCode    string    `json:"last_error_code,omitempty"`
    LastErrorMessage string    `json:"last_error_message,omitempty"`
    UpdatedAt        time.Time `json:"updated_at"`
}

type PlaceholderMetaDTO struct {
    MediaKinds []string `json:"media_kinds"` // e.g. ["image"], ["sticker"], ["audio"]
}
```

### 2.2 HTTP Endpoints — `services/connector` (Internal Machine-Only)

| Action | Method & Path | Auth | Request DTO | Response DTO |
| --- | --- | --- | --- | --- |
| **Add** | `POST /internal/facebook/oauth/exchange` | Service Token | `ExchangeOAuthCodeRequest { Code string, RedirectURI string }` | `ExchangeOAuthCodeResponse { OAuthSessionID string, ExpiresAt time.Time, Pages []ExchangedPageDTO { PageID, PageName } }` *(raw Page tokens stay in Connector ephemeral encrypted store)* |
| **Modify** | `PUT /internal/connections` (versions strictly increasing, gaps allowed; replaying the current version is a no-op) | Service Token | `UpsertConnectionRequest { ID, TenantID, Platform, AccountID, AppID, Version, Enabled, OAuthSessionID? string, CredentialRef? string, UnsubscribeOnDisable? bool }` | `UpsertConnectionResponse { ID, Version, Enabled, Subscribed bool }` |

### 2.3 gRPC (`proto/` & `services/ai`)
- **No changes.** Python AI runtime is not involved in Facebook OAuth, ingress reception, or outbound transport.

---

## 3. Kafka Topics & Event Contracts (`packages/shared`)

| Action | Topic | Producer → Consumer | Purpose |
| --- | --- | --- | --- |
| **Modify (Additive)** | `hexta.ingress.received.v1` | `services/connector` → `services/api/cmd/ingress` | Add `event_type = "unsupported_media"` with `attachments[].type` metadata (`image`/`video`/`audio`/`file`/`sticker`) and empty `text` (no binary media stored). |
| **Add** | `hexta.outbound.intent.v1` | `services/api` (outbox relay) → `services/connector` | Authorized single-attempt outbound text delivery command. Written as an `outbound.intent` row in the existing `domain_events` outbox and published raw (no envelope), keyed `connection_id:recipient_id` for per-conversation ordering. |
| **Add** | `hexta.outbound.result.v1` | `services/connector` → `services/api` (result consumer) | Delivery outcome (`sent` / `failed` / `unknown`) + error classification (`window_24h` / `auth_expired` / `rate_limit` / `invalid_recipient` / `transport` / `expired` / `connection_inactive`). The last two are rejected before any provider call: intent TTL (10 min) passed, or connection version/enabled no longer matches. |
| **Add** | `hexta.connector.health.v1` | `services/connector` → `services/api` (result consumer) | Asynchronous Page credential revocation/expiry notification (`auth_expired`) detected outside an outbound send (e.g., webhook/subscription check). |

### Kafka Event DTOs (`packages/shared/outbound/event.go`)
```go
const (
    IntentTopic = "hexta.outbound.intent.v1"
    ResultTopic = "hexta.outbound.result.v1"
    HealthTopic = "hexta.connector.health.v1"
)

type IntentEvent struct {
    SchemaVersion     int       `json:"schema_version"` // 1
    IntentID          string    `json:"intent_id"`
    AttemptID         string    `json:"attempt_id"`
    CorrelationID     string    `json:"correlation_id"`
    TenantID          string    `json:"tenant_id"`
    ConnectionID      string    `json:"connection_id"`
    ConnectionVersion int64     `json:"connection_version"`
    Platform          string    `json:"platform"` // "facebook"
    AccountID         string    `json:"account_id"` // page_id
    RecipientID       string    `json:"recipient_id"` // PSID
    Text              string    `json:"text"`
    IssuedAt          time.Time `json:"issued_at"`
    ExpiresAt         time.Time `json:"expires_at"`
}

type ResultEvent struct {
    SchemaVersion     int       `json:"schema_version"` // 1
    IntentID          string    `json:"intent_id"`
    AttemptID         string    `json:"attempt_id"`
    TenantID          string    `json:"tenant_id"`
    ConnectionID      string    `json:"connection_id"`
    Status            string    `json:"status"` // "sent" | "failed" | "unknown"
    ErrorCategory     string    `json:"error_category,omitempty"` // "window_24h" | "auth_expired" | "rate_limit" | "invalid_recipient" | "transport" | "expired" | "connection_inactive"
    ErrorCode         string    `json:"error_code,omitempty"`
    ProviderMessageID string    `json:"provider_message_id,omitempty"`
    OccurredAt        time.Time `json:"occurred_at"`
}

type ConnectionHealthEvent struct {
    SchemaVersion     int       `json:"schema_version"` // 1
    EventID           string    `json:"event_id"`
    TenantID          string    `json:"tenant_id"`
    ConnectionID      string    `json:"connection_id"`
    ConnectionVersion int64     `json:"connection_version"`
    Status            string    `json:"status"` // "auth_expired"
    ErrorCode         string    `json:"error_code"`
    OccurredAt        time.Time `json:"occurred_at"`
}
```

---

## 4. Database Schema & Atlas Migrations

All schema changes are authored in `migrations/schema/{api,connector}.sql` and generated via `make migrate-diff svc=<api|connector> name=<name>` (connector: host `atlas migrate diff --env connector` until the migrator container gets `CONNECTOR_DEV_URL`). A data migration backfills existing disabled connections to `paused`.

### 4.1 API Schema (`migrations/schema/api.sql`)
1. **Modify `messaging_connections`:**
   - Add columns: `account_name text NOT NULL DEFAULT ''`, `status text NOT NULL DEFAULT 'active'` (`active | paused | disconnected | auth_expired`), `last_error_code text`, `last_error_message text`, `updated_at timestamptz NOT NULL DEFAULT now()`.
   - Add partial unique index: `CREATE UNIQUE INDEX uq_messaging_connections_active_page ON messaging_connections (platform, account_id) WHERE status <> 'disconnected';` (enforces **1 Page = 1 Tenant**).
2. **Modify `messaging_messages`:**
   - Add `content_type text NOT NULL DEFAULT 'text'` (`text | unsupported_media`), `placeholder_meta jsonb`, `delivery_status text NOT NULL DEFAULT 'received'` (`received | sending | sent | failed | unknown`), `delivery_error_reason text`.
   - `status` gains `channel` (a real provider reply, tracked by `delivery_status`); body may be empty only for `unsupported_media`.
   - Unique `(tenant_id, id)` for the intent foreign key.
3. **Modify `messaging_threads`:** add `connection_id uuid` + `recipient_id text` (PSID), both NULL for simulator threads, composite FK to `messaging_connections`. Set on inbound; pre-existing Facebook threads are bound on their next inbound message. Needed because the thread key is a hash, so the outbound recipient is otherwise unrecoverable.
4. **Add `channel_oauth_states`:** `(id, tenant_id, actor_id, platform, state_hash UNIQUE, redirect_uri, expires_at, consumed_at, oauth_session_id, oauth_session_expires_at, pages jsonb, created_at)`. Stores only a SHA-256 hash of a random 256-bit state; consumed once (approved rule SEC-004). The row binds the connector session and granted Pages to the owner who completed OAuth.
5. **Add `outbound_intents` & `outbound_attempts`:**
   - `outbound_intents`: `(id uuid PK, tenant_id uuid NOT NULL, connection_id uuid NOT NULL, message_id uuid NOT NULL, client_message_id text NOT NULL, request_hash text NOT NULL, recipient_id text NOT NULL, status text NOT NULL, error_category text, error_code text, created_at timestamptz, updated_at timestamptz)` with tenant composite FK and unique `(tenant_id, client_message_id)`.
   - `outbound_attempts`: `(id uuid PK, tenant_id uuid NOT NULL, intent_id uuid NOT NULL, connection_version bigint NOT NULL, status text NOT NULL, provider_message_id text, created_at timestamptz, completed_at timestamptz)`.

### 4.2 Connector Schema (`migrations/schema/connector.sql`)
1. **Add `oauth_sessions` (ephemeral encrypted token staging):**
   - `(id uuid PK, app_id text NOT NULL, payload_ciphertext bytea NOT NULL, key_version int NOT NULL, expires_at timestamptz NOT NULL, created_at timestamptz NOT NULL)`.
2. **Modify `connections`:**
   - Add `credential_ciphertext bytea`, `key_version int`, `token_updated_at timestamptz` (AES-256-GCM encrypted Page Access Token; wiped on disconnect).
3. **Add `outbound_attempts` (connector-side idempotency):**
   - `(attempt_id uuid PK, intent_id uuid NOT NULL, connection_id uuid NOT NULL, status text NOT NULL, provider_message_id text, error_category text, error_code text, created_at timestamptz NOT NULL)`. A redelivered attempt still `dispatching` becomes `unknown` and is never resent.
4. **Add `outbound_events`:** result/health outbox (`topic`, `event_key`, `payload`, lease `state`/`fence`/`attempts`/`available_at`), committed with the attempt outcome.

---

## 5. Core Flows & Failure Semantics

1. **OAuth & Connect (Slice A):**
   - `oauth/start` (stores hashed single-use state) -> Meta OAuth -> `oauth/callback` (consumes the state, rechecks owner) -> Connector exchanges `code`, caches encrypted Page tokens in `oauth_sessions` (short TTL, e.g., 10 min) -> API flags `available` / `already_connected_this_tenant` / `owned_by_other_tenant` -> Owner submits `POST /connections` -> API DB transaction enforces partial unique index -> Connector `PUT /internal/connections` decrypts from `oauth_sessions`, persists in `connections`, and calls `/{page_id}/subscribed_apps`.
2. **Pause / Disconnect (Slice A):**
   - API transaction sets `status = 'paused' | 'disconnected'`, `enabled = false`, increments `version`, and cancels all `pending` `outbound_intents` for that `connection_id`. Syncs `PUT /internal/connections` to Connector; on `disconnect`, Connector wipes `credential_ciphertext` and unsubscribes webhook.
3. **Inbound Placeholder (Slice B):**
   - Connector emits `unsupported_media` with attachment kinds -> API ingress creates message with `content_type = 'unsupported_media'` -> Web renders placeholder.
4. **Outbound Send (Slice C):**
   - `POST /turns/{id}/messages`: API verifies the assigned sale, turn version/context, connection `active` + `client_message_id` idempotency -> writes the message (`status=channel`, `delivery_status=sending`) + `outbound_intents` + `outbound_attempts` -> outbox relays `hexta.outbound.intent.v1` -> Connector deduplicates `attempt_id`, calls Meta `POST /{version}/{page_id}/messages` -> emits `hexta.outbound.result.v1`:
     - HTTP 200 -> `sent`.
     - Meta `10` / `2022` -> `failed` (`window_24h`), no auto-retry.
     - Meta `190` / `200–299` -> `failed` (`auth_expired`), plus a `hexta.connector.health.v1` event, API sets `messaging_connections.status = 'auth_expired'` -> Web shows Inbox & Settings banner.
     - Network timeout / 5xx after dispatch -> `unknown`, no silent resend.

---

## 6. Verification Strategy

- **Unit & Contract Tests:** single-use OAuth state (owner binding, replay, expiry, denial), Owner-only RBAC, partial unique index (1-Page-1-Tenant), `pending` intent cancellation on pause/disconnect, `unsupported_media` mapping, outbound error classification (`window_24h`, `auth_expired`, `unknown`).
- **Isolated Integration (`compose.connector-test.yml`):** Mock Meta OAuth & Graph API HTTP server -> Connector -> Kafka -> API -> Inbox status transitions.
- **Live Meta Acceptance (`NOT_RUN` until App Review):** Real Meta App OAuth, Page subscription, live text & media webhook, live 24h send & error verification.

---

## 7. Amendment log (accepted by owner, 2026-11-15)

Implementation-time changes to the original target text, now folded into §1–§6 above. Kept as history. Rationale and evidence: [implementation session](../epics/EPIC-008-integrations-deployment/sessions/2026-11-15_facebook-platform-connection.md).

| Area | Target text | Implemented | Reason |
| --- | --- | --- | --- |
| OAuth state (§1, §2.1, §6) | Signed state | Random 256-bit single-use state; SHA-256 hash stored in `channel_oauth_states` with tenant, owner, redirect URI, 10-min expiry. The row also binds the connector `oauth_session_id` and granted Pages to that owner. | Approved rule SEC-004 requires state to be consumed at callback; a signature alone cannot be single-use. Binding the session stops a leaked session ID being used by another tenant. |
| Callback response (§2.1) | `{OAuthSessionID, Pages}` | Adds `tenant_id`, `expires_at` | Meta needs one fixed redirect URI (`/channels/facebook/callback`), so the tenant comes from the state. |
| Thread binding (§4.1) | Not specified | `messaging_threads.connection_id` + `recipient_id` (PSID), bound on inbound | The existing thread key is a hash; the outbound recipient was otherwise unrecoverable. |
| Send endpoint (§2.1) | "existing `/inbox/...`" | `POST /api/v1/tenants/{tenantID}/turns/{id}/messages` with `client_message_id`, `text`, `expected_version`, `context_revision` | Existing inbox routes are `/conversations` and `/turns/...`. Version/context keep the existing stale-reply guard. Text limit 2,000 (Messenger). |
| Message status (§4.1) | `delivery_status` only | Also `status='channel'` for real replies | Separates real provider replies from `simulated`/`received` rows. |
| Error categories (§3) | 5 categories | Adds `expired` (intent TTL 10 min) and `connection_inactive` (version/enabled mismatch at dispatch) | Additive; both are rejected before any provider call. |
| Connector versions (§2.2) | Version field | Strictly increasing, gaps allowed; replaying the current version is a no-op | An API retry after a failed connector sync must converge. |
| Outbound intent relay (§3) | Outbox relay | Existing `domain_events` outbox; `outbound.intent` rows are published raw to `hexta.outbound.intent.v1`, keyed `connection_id:recipient_id` | Reuses the approved outbox and keeps per-conversation ordering. |
| Connector layout | — | Facebook adapter in `services/connector/internal/plugin/facebook` | Owner request. |
