# Changelog: Issue #34 — Core Business Schema Migrations & GORM Models

- **Issue**: [#34](https://github.com/TruongHoang2004/Hexta/issues/34)
- **Date**: 2026-10-09
- **Branch**: `task/issue-34-feat-db-implement-atlas-schema-migration`

---

## 1. Change Summary
Implemented the database persistence foundations for Hexta's core business domains:
- **Product Catalog Domain**: Added `products` and `product_variants` tables and GORM models.
- **Inventory Management Domain**: Added `inventory_items` and `stock_movements` tables and GORM models with 2-phase reservation support.
- **Order Management Domain**: Added `orders` and `order_items` tables and GORM models with status enums and draft lineage.
- **Customer Domain**: Added `customers` table and GORM model with soft-delete support.
- **Audit & AI Lineage Domain**: Added `audit_logs` (immutable JSONB state logging) and `ai_drafts` (interactive LLM proposals).
- **Atlas Migrations**: Generated declarative SQL migration `20261009180000_create_core_business_schema.sql` and updated `atlas.sum`. Registered all models in `services/api/cmd/tools/main.go`.

---

## 2. Impacted Files & Components

| File Path | Action | Description |
| :--- | :---: | :--- |
| [`services/api/internal/core/model/products.go`](file:///Users/truonghoang/Documents/dev/personal/Hexta/services/api/internal/core/model/products.go) | `[NEW]` | `Product` and `ProductVariant` structs with GORM tags. |
| [`services/api/internal/core/model/inventory.go`](file:///Users/truonghoang/Documents/dev/personal/Hexta/services/api/internal/core/model/inventory.go) | `[NEW]` | `InventoryItem` and `StockMovement` structs. |
| [`services/api/internal/core/model/orders.go`](file:///Users/truonghoang/Documents/dev/personal/Hexta/services/api/internal/core/model/orders.go) | `[NEW]` | `Order` and `OrderItem` structs with state machine enums. |
| [`services/api/internal/core/model/customers.go`](file:///Users/truonghoang/Documents/dev/personal/Hexta/services/api/internal/core/model/customers.go) | `[NEW]` | `Customer` struct with tenant-scoped unique phone index. |
| [`services/api/internal/core/model/audit_logs.go`](file:///Users/truonghoang/Documents/dev/personal/Hexta/services/api/internal/core/model/audit_logs.go) | `[NEW]` | `AuditLog` struct with JSONB state columns. |
| [`services/api/internal/core/model/ai_drafts.go`](file:///Users/truonghoang/Documents/dev/personal/Hexta/services/api/internal/core/model/ai_drafts.go) | `[NEW]` | `AIDraft` struct with status and expiration tracking. |
| [`services/api/cmd/tools/main.go`](file:///Users/truonghoang/Documents/dev/personal/Hexta/services/api/cmd/tools/main.go) | `[MODIFY]` | Added all new domain models into `gormschema.Load(...)`. |
| [`migrations/api/20261009180000_create_core_business_schema.sql`](file:///Users/truonghoang/Documents/dev/personal/Hexta/migrations/api/20261009180000_create_core_business_schema.sql) | `[NEW]` | Declarative PostgreSQL migration script. |
| [`migrations/api/atlas.sum`](file:///Users/truonghoang/Documents/dev/personal/Hexta/migrations/api/atlas.sum) | `[MODIFY]` | Recalculated migration checksum tree. |
| [`agentic-memory/plans/2026-10-09_issue-34_plan.md`](file:///Users/truonghoang/Documents/dev/personal/Hexta/agentic-memory/plans/2026-10-09_issue-34_plan.md) | `[NEW]` | Implementation plan artifact. |
| [`agentic-memory/reviews/2026-10-09_issue-34_review.md`](file:///Users/truonghoang/Documents/dev/personal/Hexta/agentic-memory/reviews/2026-10-09_issue-34_review.md) | `[NEW]` | Comprehensive code review artifact. |

---

## 3. Key Technical Decisions
- **Monetary Precision**: Used `decimal.Decimal` (`numeric(15,2)`) across all pricing and total amount fields to prevent IEEE floating point rounding errors.
- **Tenant Scope Enforcement**: Hardcoded `tenant_id varchar(36)` foreign key index on all tables as part of the defense-in-depth model.
- **Declarative Migration with Atlas**: Generated SQL matching GORM model definitions and ensured checksum verification passes `atlas migrate validate`.

---

## 4. Verification Guide
1. **Model Compilation**:
   ```bash
   cd services/api && go build ./...
   ```
2. **Unit Tests**:
   ```bash
   cd services/api && go test -v ./...
   ```
3. **Atlas Migration Integrity**:
   ```bash
   atlas migrate validate --dir file://migrations/api
   ```
