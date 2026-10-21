# Current Hexta ERD

Code snapshot: `af06cf4` on `codex/order-workflows`, 2026-10-20. Includes all seven Atlas API migrations and Python migration 001. **21 public business tables + 1 ai_runtime table**. Migration metadata is omitted. This is a code/migration diagram, not a new live database introspection.

Solid lines represent migration-declared foreign keys. Dashed lines represent application/domain references without a declared FK; their cardinality describes intended references, not a database guarantee. Mermaid normally uses line style for identifying relationships; this document explicitly repurposes it as the FK/application-reference legend. PK/FK attribute labels refer only to actual constraints. All attributes are shown; type precision is retained in comments. Composite uniqueness is listed separately.

There is no users table. identities.id identifies a login identity, while user_id is the account identifier shared by identity records, sessions, membership, ownership and actor fields. user_id is not unique in identities and is not a referenced account primary key. No FK edges to a fictional users table are drawn.

## Complete relationship overview

```mermaid
erDiagram
    direction TB
    identities {
        bigint id PK
        text user_id
    }
    sessions {
        bigint id PK
        text user_id
    }
    platform_operators {
        varchar user_id PK
    }
    tenants ||--o{ tenant_members : membership
    tenants |o--o{ tenant_registration_requests : approved_tenant
    tenants ||--o{ tenant_invitations : invitations
    tenants ||..o{ customers : tenant_scope
    tenants ||..o{ products : tenant_scope
    tenants ||..o{ orders : tenant_scope
    products ||--o{ product_variants : variants
    customers |o..o{ orders : customer
    orders ||--o{ order_items : lines
    product_variants ||..o{ order_items : product_reference
    product_variants ||..o| inventory_items : tenant_variant
    inventory_items ||..o{ stock_movements : movement_history
    tenants ||..o{ ai_drafts : tenant_scope
    tenants ||..o{ audit_logs : tenant_scope
    tenants ||--o{ workflow_definitions : definitions
    workflow_definitions ||--o{ workflow_versions : immutable_versions
    workflow_versions ||--o{ workflow_steps : ordered_steps
    tenants ||--o{ work_items : tenant_scope
    orders ||--o{ work_items : linked_work
    workflow_versions ||--o{ work_items : pinned_version
    workflow_steps ||--o{ work_items : current_step
    work_items ||--o{ work_events : transition_history
    tenants ||..o{ ai_runtime_runs : context_only
```

## Accounts and onboarding

```mermaid
erDiagram
    direction TB
    identities {
        bigserial id PK "required"
        text user_id "required"
        varchar provider "character varying(50); required"
        varchar identifier "character varying(255); required"
        varchar password "character varying(255); nullable"
        timestamptz created_at "nullable"
        timestamptz updated_at "nullable"
    }
    sessions {
        bigserial id PK "required"
        text user_id "required"
        text token "required"
        text provider "required"
        varchar device_info "character varying(500); nullable"
        varchar ip_address "character varying(45); nullable"
        varchar user_agent "character varying(1000); nullable"
        boolean is_active "required"
        timestamptz created_at "nullable"
        timestamptz expires_at "required"
        timestamptz revoked_at "nullable"
    }
    platform_operators {
        varchar user_id PK "varchar(255); required"
        timestamptz created_at "required"
    }
    tenants {
        varchar id PK "character varying(36); required"
        varchar name "character varying(255); required"
        varchar slug "character varying(255); required"
        varchar plan "character varying(50); required"
        varchar status "character varying(50); required"
        varchar owner_id "character varying(255); required"
        timestamptz created_at "nullable"
        timestamptz updated_at "nullable"
    }
    tenant_members {
        bigserial id PK "required"
        varchar tenant_id FK "character varying(36); required"
        varchar user_id "character varying(255); required"
        varchar role "character varying(50); required"
        timestamptz created_at "nullable"
        timestamptz updated_at "nullable"
    }
    tenant_registration_requests {
        varchar id PK "varchar(36); required"
        varchar applicant_id "varchar(255); required"
        varchar name "varchar(100); required"
        varchar slug "varchar(100); required"
        varchar status "varchar(20); required"
        varchar tenant_id FK "varchar(36); nullable"
        varchar reviewed_by "varchar(255); required"
        text review_reason "required"
        timestamptz reviewed_at "nullable"
        timestamptz created_at "required"
        timestamptz updated_at "required"
    }
    tenant_invitations {
        varchar id PK "varchar(36); required"
        varchar tenant_id FK "varchar(36); required"
        varchar email "varchar(255); required"
        varchar role "varchar(50); required"
        varchar invited_by "varchar(255); required"
        varchar token_hash "varchar(64); required"
        varchar status "varchar(20); required"
        timestamptz expires_at "required"
        varchar accepted_by "varchar(255); required"
        timestamptz created_at "required"
        timestamptz updated_at "required"
    }
    tenants ||--o{ tenant_members : "tenant_id"
    tenants |o--o{ tenant_registration_requests : "tenant_id nullable"
    tenants ||--o{ tenant_invitations : "tenant_id"
```

## Catalog, sales and inventory

```mermaid
erDiagram
    direction TB
    tenants {
        varchar id PK "character varying(36); required"
        varchar name "character varying(255); required"
        varchar slug "character varying(255); required"
        varchar plan "character varying(50); required"
        varchar status "character varying(50); required"
        varchar owner_id "character varying(255); required"
        timestamptz created_at "nullable"
        timestamptz updated_at "nullable"
    }
    customers {
        varchar id PK "character varying(36); required"
        varchar tenant_id "character varying(36); required"
        varchar full_name "character varying(255); required"
        varchar phone "character varying(32); nullable"
        varchar email "character varying(255); nullable"
        text address "nullable"
        timestamptz created_at "nullable"
        timestamptz updated_at "nullable"
        timestamptz deleted_at "nullable"
    }
    products {
        varchar id PK "character varying(36); required"
        varchar tenant_id "character varying(36); required"
        varchar name "character varying(255); required"
        text description "nullable"
        varchar category "character varying(100); nullable"
        varchar status "character varying(32); required"
        timestamptz created_at "nullable"
        timestamptz updated_at "nullable"
        timestamptz deleted_at "nullable"
    }
    product_variants {
        varchar id PK "character varying(36); required"
        varchar tenant_id "character varying(36); required"
        varchar product_id FK "character varying(36); required"
        varchar sku "character varying(64); required"
        varchar variant_name "character varying(128); required"
        numeric price "numeric(15,2); required"
        numeric cost_price "numeric(15,2); required"
        varchar barcode "character varying(64); nullable"
        timestamptz created_at "nullable"
        timestamptz updated_at "nullable"
        timestamptz deleted_at "nullable"
        boolean price_provided "required"
    }
    orders {
        varchar id PK "character varying(36); required"
        varchar tenant_id "character varying(36); required"
        varchar order_number "character varying(64); required"
        varchar customer_id "character varying(36); nullable"
        varchar employee_id "character varying(36); nullable"
        varchar status "character varying(32); required"
        varchar payment_status "character varying(32); required"
        numeric subtotal "numeric(15,2); required"
        numeric discount "numeric(15,2); required"
        numeric total_amount "numeric(15,2); required"
        varchar source "character varying(32); required"
        varchar draft_id "character varying(36); nullable"
        text notes "nullable"
        timestamptz created_at "nullable"
        timestamptz updated_at "nullable"
        timestamptz deleted_at "nullable"
        bigint version "required"
        boolean requires_delivery "required"
        text delivery_address "required"
    }
    order_items {
        varchar id PK "character varying(36); required"
        varchar tenant_id "character varying(36); required"
        varchar order_id FK "character varying(36); required"
        varchar variant_id "character varying(36); required"
        bigint quantity "required"
        numeric price "numeric(15,2); required"
        numeric line_total "numeric(15,2); required"
        timestamptz created_at "nullable"
        timestamptz updated_at "nullable"
        boolean price_provided "required"
    }
    inventory_items {
        varchar id PK "character varying(36); required"
        varchar tenant_id "character varying(36); required"
        varchar variant_id "character varying(36); required"
        bigint on_hand_qty "required"
        bigint reserved_qty "required"
        bigint available_qty "required"
        bigint safety_threshold "required"
        timestamptz created_at "nullable"
        timestamptz updated_at "nullable"
    }
    stock_movements {
        varchar id PK "character varying(36); required"
        varchar tenant_id "character varying(36); required"
        varchar inventory_id "character varying(36); required"
        varchar movement_type "character varying(32); required"
        bigint quantity "required"
        bigint balance_after "required"
        varchar reference_type "character varying(64); nullable"
        varchar reference_id "character varying(36); nullable"
        text notes "nullable"
        timestamptz created_at "nullable"
    }
    ai_drafts {
        varchar id PK "character varying(36); required"
        varchar tenant_id "character varying(36); required"
        varchar user_id "character varying(255); required"
        varchar intent "character varying(64); required"
        text prompt "required"
        jsonb draft_payload "required"
        varchar status "character varying(32); required"
        timestamptz expires_at "required"
        timestamptz created_at "nullable"
        timestamptz updated_at "nullable"
    }
    audit_logs {
        varchar id PK "character varying(36); required"
        varchar tenant_id "character varying(36); required"
        varchar user_id "character varying(255); required"
        varchar action "character varying(64); required"
        varchar entity_type "character varying(64); required"
        varchar entity_id "character varying(36); required"
        varchar source "character varying(32); required"
        varchar draft_id "character varying(36); nullable"
        jsonb before_state "nullable"
        jsonb after_state "required"
        varchar ip_address "character varying(45); nullable"
        text user_agent "nullable"
        timestamptz created_at "nullable"
    }
    products ||--o{ product_variants : "product_id"
    orders ||--o{ order_items : "order_id"
    tenants ||..o{ customers : "tenant_id"
    tenants ||..o{ products : "tenant_id"
    tenants ||..o{ orders : "tenant_id"
    customers |o..o{ orders : "customer_id nullable"
    product_variants ||..o{ order_items : "variant_id"
    product_variants ||..o| inventory_items : "tenant_id + variant_id"
    inventory_items ||..o{ stock_movements : "inventory_id"
    tenants ||..o{ ai_drafts : "tenant_id"
    tenants ||..o{ audit_logs : "tenant_id"
```

## Versioned workflows and work

```mermaid
erDiagram
    direction TB
    tenants {
        varchar id PK "character varying(36); required"
        varchar name "character varying(255); required"
        varchar slug "character varying(255); required"
        varchar plan "character varying(50); required"
        varchar status "character varying(50); required"
        varchar owner_id "character varying(255); required"
        timestamptz created_at "nullable"
        timestamptz updated_at "nullable"
    }
    orders {
        varchar id PK "character varying(36); required"
        varchar tenant_id "character varying(36); required"
        varchar order_number "character varying(64); required"
        varchar customer_id "character varying(36); nullable"
        varchar employee_id "character varying(36); nullable"
        varchar status "character varying(32); required"
        varchar payment_status "character varying(32); required"
        numeric subtotal "numeric(15,2); required"
        numeric discount "numeric(15,2); required"
        numeric total_amount "numeric(15,2); required"
        varchar source "character varying(32); required"
        varchar draft_id "character varying(36); nullable"
        text notes "nullable"
        timestamptz created_at "nullable"
        timestamptz updated_at "nullable"
        timestamptz deleted_at "nullable"
        bigint version "required"
        boolean requires_delivery "required"
        text delivery_address "required"
    }
    workflow_definitions {
        varchar id PK "character varying(36); required"
        varchar tenant_id FK "character varying(36); required"
        varchar name "character varying(120); required"
        bigint version "required"
        varchar create_hash "character varying(64); required"
        timestamptz created_at "required"
        timestamptz updated_at "required"
    }
    workflow_versions {
        varchar id PK "character varying(36); required"
        varchar tenant_id FK "character varying(36); required"
        varchar definition_id FK "character varying(36); required"
        bigint number "required"
        varchar name "character varying(120); required"
        timestamptz created_at "required"
    }
    workflow_steps {
        varchar id PK "character varying(36); required"
        varchar tenant_id FK "character varying(36); required"
        varchar workflow_version_id FK "character varying(36); required"
        integer ordinal "required"
        varchar name "character varying(120); required"
        varchar handoff_mode "character varying(16); required"
    }
    work_items {
        varchar id PK "character varying(36); required"
        varchar tenant_id FK "character varying(36); required"
        varchar order_id FK "character varying(36); required"
        varchar workflow_version_id FK "character varying(36); required"
        integer current_step FK "required"
        varchar responsible_id "character varying(255); nullable"
        varchar recipient_id "character varying(255); nullable"
        varchar state "character varying(16); required"
        bigint version "required"
        varchar created_by "character varying(255); required"
        varchar request_key "character varying(36); required"
        varchar create_hash "character varying(64); required"
        timestamptz created_at "required"
        timestamptz updated_at "required"
    }
    work_events {
        varchar id PK "character varying(36); required"
        varchar tenant_id FK "character varying(36); required"
        varchar work_item_id FK "character varying(36); required"
        bigint version "required"
        varchar actor_id "character varying(255); required"
        varchar action "character varying(32); required"
        varchar from_responsible_id "character varying(255); nullable"
        varchar responsible_id "character varying(255); nullable"
        varchar recipient_id "character varying(255); nullable"
        integer from_step "required"
        integer current_step "required"
        varchar state "character varying(16); required"
        timestamptz created_at "required"
    }
    tenants ||--o{ workflow_definitions : "tenant_id"
    workflow_definitions ||--o{ workflow_versions : "tenant_id + definition_id"
    workflow_versions ||--o{ workflow_steps : "tenant_id + workflow_version_id"
    tenants ||--o{ work_items : "tenant_id"
    orders ||--o{ work_items : "tenant_id + order_id"
    workflow_versions ||--o{ work_items : "tenant_id + workflow_version_id"
    workflow_steps ||--o{ work_items : "tenant_id + workflow_version_id + current_step"
    work_items ||--o{ work_events : "tenant_id + work_item_id"
    tenants ||..o{ orders : "tenant_id"
```

## Python AI runtime

```mermaid
erDiagram
    direction TB
    ai_runtime_runs {
        text tenant_id PK "required"
        text run_id PK "required"
        bytea request "required"
        bytea digest "required"
        bytea result "required"
        smallint status "required"
        integer attempts "required"
        uuid lease_token "nullable"
        timestamptz lease_until "nullable"
        text error_code "required"
        timestamptz created_at "required"
        timestamptz updated_at "required"
    }
```

## Constraint and interpretation notes

- identities: unique (provider, identifier); password nullable after migration 20260912083721.
- tenant_members: unique (tenant_id, user_id). Registration requests: unique slug where status is not rejected. Pending invitations: unique (tenant_id, email) where status is pending; token_hash unique.
- product_variants: unique (tenant_id, sku). inventory_items: unique (tenant_id, variant_id), so a variant has at most one inventory row within a tenant, but no FK enforces its existence.
- customers: unique (tenant_id, phone) only for nonempty phone and nondeleted rows. orders: unique (tenant_id, order_number) and (tenant_id, id).
- order_items.price is the SQL column exposed as unit_price by the model/JSON API. price_provided distinguishes missing from explicit zero. Customer and employee references may be absent on drafts.
- workflow_definitions and workflow_versions have unique (tenant_id, id). Versions: unique (tenant_id, definition_id, number). Steps: unique (workflow_version_id, ordinal) and (tenant_id, workflow_version_id, ordinal). Ordinals are zero-based, bounded 0–29.
- work_items: unique (tenant_id, id) and (tenant_id, created_by, request_key). Pinned workflow and current step use composite tenant FKs. State is active/waiting/completed; waiting retains a responsible employee and distinct nominated recipient.
- work_events: unique (work_item_id, version); transition/history references use a composite tenant FK. Actor/responsible/recipient fields contain account user IDs but have no membership FK; service authorization checks current membership.
- ai_runtime.runs: composite PK (tenant_id, run_id), no public-table FK. Request/digest/result are bytea, not jsonb. Lease fields and status checks support durable execution; lease_token is not an entity FK.
- audit_logs.entity_type/entity_id and stock_movements.reference_type/reference_id are polymorphic application references; they are not blanket FKs to orders. ai_drafts and audit_logs have draft references at selected callers, not newly declared constraints.
- Inventory/stock/legacy AI draft tables exist physically; this does not imply current manual-sales routes execute stock, payment or AI-draft workflows.
- Selected GORM associations describe simpler single-ID links than the composite workflow FKs installed by Atlas. SQL migrations define the physical constraints shown here. No future conversation, Kafka inbox/outbox, exception or checkpoint tables are included.

## Sources

- [migrations/api/20260220100303_init.sql](../../../migrations/api/20260220100303_init.sql)
- [migrations/api/20260912083721_fix_identities_provider_unique_index.sql](../../../migrations/api/20260912083721_fix_identities_provider_unique_index.sql)
- [migrations/api/20260912160000_create_tenants.sql](../../../migrations/api/20260912160000_create_tenants.sql)
- [migrations/api/20261009180000_create_core_business_schema.sql](../../../migrations/api/20261009180000_create_core_business_schema.sql)
- [migrations/api/20261019030000_sales_demo.sql](../../../migrations/api/20261019030000_sales_demo.sql)
- [migrations/api/20261019040000_tenant_onboarding.sql](../../../migrations/api/20261019040000_tenant_onboarding.sql)
- [migrations/api/20261020075908_order_workflows.sql](../../../migrations/api/20261020075908_order_workflows.sql)
- [migrations/ai/20261020100000_durable_runs.sql](../../../migrations/ai/20261020100000_durable_runs.sql)
- GORM schema inventory: [cmd/tools/main.go](../../../services/api/cmd/tools/main.go).
- Workflow models: [workflowmodel/models.go](../../../services/api/internal/repository/workflowmodel/models.go).

Atlas ownership follow-up (2026-10-20): AI table shape is unchanged, but its SQL is now in the Atlas AI history; runtime startup no longer migrates. Independent Atlas metadata schemas remain excluded from this business ERD.
