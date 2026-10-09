# Hexta — Data Models & State Machines Architecture

- **Document Version**: 1.0.0
- **Status**: Approved Baseline
- **Author**: Solution Architecture Team
- **Related Documents**: [`docs/vision.md`](file:///Users/truonghoang/Documents/dev/personal/Hexta/docs/vision.md), [`docs/architecture/system-architecture.md`](file:///Users/truonghoang/Documents/dev/personal/Hexta/docs/architecture/system-architecture.md), [`docs/architecture/multi-tenancy-and-security.md`](file:///Users/truonghoang/Documents/dev/personal/Hexta/docs/architecture/multi-tenancy-and-security.md)

---

## 1. Domain Entity Relationship Diagram (ERD)

The relational schema in PostgreSQL 17 is organized around tenant-scoped bounded contexts: Identity & Tenant, Product & Catalog, Inventory, Order Management, Customer, HR, and Audit/AI.

```mermaid
erDiagram
    TENANT ||--o{ USER : "employs"
    TENANT ||--o{ CUSTOMER : "serves"
    TENANT ||--o{ PRODUCT : "catalogs"
    TENANT ||--o{ INVENTORY_ITEM : "stores"
    TENANT ||--o{ ORDER : "processes"
    TENANT ||--o{ EMPLOYEE : "manages"
    TENANT ||--o{ AUDIT_LOG : "records"
    TENANT ||--o{ AI_DRAFT : "generates"

    USER ||--o| EMPLOYEE : "mapped_to"

    PRODUCT ||--|{ PRODUCT_VARIANT : "has"
    PRODUCT_VARIANT ||--o| INVENTORY_ITEM : "tracked_as"
    
    INVENTORY_ITEM ||--o{ STOCK_MOVEMENT : "logs"
    INVENTORY_ITEM ||--o{ STOCK_RESERVATION : "holds"

    CUSTOMER ||--o{ ORDER : "places"
    ORDER ||--|{ ORDER_ITEM : "contains"
    ORDER ||--o{ STOCK_RESERVATION : "creates"
    ORDER_ITEM }o--|| PRODUCT_VARIANT : "references"
    EMPLOYEE ||--o{ ORDER : "assigned_to"

    TENANT {
        uuid id PK
        string name
        string slug UK
        string status
        jsonb settings
        timestamptz created_at
    }

    USER {
        uuid id PK
        uuid tenant_id FK
        string email
        string password_hash
        string role
        string status
        timestamptz created_at
    }

    CUSTOMER {
        uuid id PK
        uuid tenant_id FK
        string full_name
        string phone UK
        string email
        string address
        timestamptz created_at
    }

    PRODUCT {
        uuid id PK
        uuid tenant_id FK
        string name
        string description
        string category
        string status
        timestamptz created_at
    }

    PRODUCT_VARIANT {
        uuid id PK
        uuid tenant_id FK
        uuid product_id FK
        string sku UK
        string variant_name
        numeric price
        numeric cost_price
        string barcode
    }

    INVENTORY_ITEM {
        uuid id PK
        uuid tenant_id FK
        uuid variant_id FK
        integer on_hand_qty
        integer reserved_qty
        integer available_qty
        integer safety_threshold
    }

    STOCK_MOVEMENT {
        uuid id PK
        uuid tenant_id FK
        uuid inventory_id FK
        string movement_type
        integer quantity
        integer balance_after
        string reference_type
        uuid reference_id
        timestamptz created_at
    }

    ORDER {
        uuid id PK
        uuid tenant_id FK
        string order_number UK
        uuid customer_id FK
        uuid employee_id FK
        string status
        string payment_status
        numeric subtotal
        numeric discount
        numeric total_amount
        string source
        uuid draft_id
        timestamptz created_at
    }

    ORDER_ITEM {
        uuid id PK
        uuid tenant_id FK
        uuid order_id FK
        uuid variant_id FK
        integer quantity
        numeric unit_price
        numeric line_total
    }

    EMPLOYEE {
        uuid id PK
        uuid tenant_id FK
        uuid user_id FK
        string employee_code UK
        string full_name
        string position
        string department
        string status
        timestamptz created_at
    }

    AI_DRAFT {
        uuid id PK
        uuid tenant_id FK
        uuid user_id FK
        string intent
        text prompt
        jsonb draft_payload
        string status
        timestamptz expires_at
        timestamptz created_at
    }

    AUDIT_LOG {
        uuid id PK
        uuid tenant_id FK
        uuid user_id FK
        string action
        string entity_type
        uuid entity_id
        string source
        uuid draft_id
        jsonb before_state
        jsonb after_state
        timestamptz created_at
    }
```

---

## 2. Relational Schema Specifications (PostgreSQL / GORM)

### 2.1. Product & Catalog Models
```go
type Product struct {
    BaseTenantModel
    Name        string           `gorm:"type:varchar(255);not null;index"`
    Description string           `gorm:"type:text"`
    Category    string           `gorm:"type:varchar(100);index"`
    Status      string           `gorm:"type:varchar(32);default:'active';index"` // active, archived
    Variants    []ProductVariant `gorm:"foreignKey:ProductID"`
}

type ProductVariant struct {
    BaseTenantModel
    ProductID   uuid.UUID       `gorm:"type:uuid;not null;index"`
    SKU         string          `gorm:"type:varchar(64);not null;index"`
    VariantName string          `gorm:"type:varchar(128);not null"` // e.g. "Size L / White"
    Price       decimal.Decimal `gorm:"type:numeric(15,2);not null"`
    CostPrice   decimal.Decimal `gorm:"type:numeric(15,2);default:0"`
    Barcode     string          `gorm:"type:varchar(64);index"`
}
```

### 2.2. Inventory Models
```go
type InventoryItem struct {
    BaseTenantModel
    VariantID       uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_tenant_variant"`
    OnHandQty       int       `gorm:"not null;default:0"` // Physical quantity in warehouse
    ReservedQty     int       `gorm:"not null;default:0"` // Quantity held for pending orders
    AvailableQty    int       `gorm:"not null;default:0"` // OnHandQty - ReservedQty
    SafetyThreshold int       `gorm:"not null;default:5"` // Threshold for low-stock alert
}

type StockMovement struct {
    BaseTenantModel
    InventoryID   uuid.UUID `gorm:"type:uuid;not null;index"`
    MovementType  string    `gorm:"type:varchar(32);not null"` // 'INBOUND', 'OUTBOUND', 'ADJUSTMENT'
    Quantity      int       `gorm:"not null"`
    BalanceAfter  int       `gorm:"not null"`
    ReferenceType string    `gorm:"type:varchar(64)"`          // 'ORDER', 'INVENTORY_AUDIT'
    ReferenceID   uuid.UUID `gorm:"type:uuid;index"`
    Notes         string    `gorm:"type:text"`
}
```

### 2.3. Order Models
```go
type Order struct {
    BaseTenantModel
    OrderNumber   string          `gorm:"type:varchar(64);not null;uniqueIndex:idx_tenant_order_num"`
    CustomerID    *uuid.UUID      `gorm:"type:uuid;index"`
    EmployeeID    *uuid.UUID      `gorm:"type:uuid;index"`
    Status        string          `gorm:"type:varchar(32);not null;default:'draft';index"`
    PaymentStatus string          `gorm:"type:varchar(32);not null;default:'unpaid';index"`
    Subtotal      decimal.Decimal `gorm:"type:numeric(15,2);not null;default:0"`
    Discount      decimal.Decimal `gorm:"type:numeric(15,2);not null;default:0"`
    TotalAmount   decimal.Decimal `gorm:"type:numeric(15,2);not null;default:0"`
    Source        string          `gorm:"type:varchar(32);not null;default:'web_form'"` // 'web_form', 'ai_draft'
    DraftID       *uuid.UUID      `gorm:"type:uuid;index"`
    Items         []OrderItem     `gorm:"foreignKey:OrderID"`
}

type OrderItem struct {
    BaseTenantModel
    OrderID   uuid.UUID       `gorm:"type:uuid;not null;index"`
    VariantID uuid.UUID       `gorm:"type:uuid;not null;index"`
    Quantity  int             `gorm:"not null"`
    UnitPrice decimal.Decimal `gorm:"type:numeric(15,2);not null"`
    LineTotal decimal.Decimal `gorm:"type:numeric(15,2);not null"`
}
```

---

## 3. Order Lifecycle State Machine

Orders progress through a strict finite state machine (FSM) ensuring consistent business rules:

```mermaid
stateDiagram-v2
    [*] --> Draft : Created via Form or AI Agent
    
    Draft --> Confirmed : User confirms draft & inventory reserved
    Draft --> Cancelled : Discarded by user
    
    Confirmed --> Processing : Staff starts packaging
    Confirmed --> Cancelled : Customer cancels (Releases Reserved Stock)
    
    Processing --> Fulfilled : Delivered & Paid (Stock Deducted)
    Processing --> Cancelled : Packing failed / Cancelled (Releases Reserved Stock)
    
    Fulfilled --> Returned : Customer returns item (Stock Returned to Inventory)
    Fulfilled --> [*]
    Cancelled --> [*]
    Returned --> [*]
```

### Transition Guard Rules
1. **`Draft -> Confirmed`**:
   - Verification: Every item in the order must have `AvailableQty >= Item.Quantity`.
   - Action: Atomically increment `ReservedQty` by item quantity in a single DB transaction.
2. **`Confirmed -> Cancelled`**:
   - Action: Decrement `ReservedQty` by item quantity.
3. **`Processing -> Fulfilled`**:
   - Action: Atomically decrement `OnHandQty` and decrement `ReservedQty`. Record `StockMovement` (OUTBOUND).
4. **`Fulfilled -> Returned`**:
   - Action: Increment `OnHandQty`. Record `StockMovement` (INBOUND - Return).

---

## 4. Inventory Reservation & Stock Mutation Lifecycle

To prevent overselling under concurrent requests, stock is managed through a **two-phase allocation model**:

```mermaid
stateDiagram-v2
    [*] --> Available : Inbound Stock Received
    
    Available --> Reserved : Order Confirmed (Hold Stock)
    Reserved --> Available : Order Cancelled (Release Hold)
    
    Reserved --> Deducted : Order Fulfilled (Permanent Deduct)
    
    Available --> Adjusted : Manual Physical Audit Correction
    Deducted --> [*]
    Adjusted --> Available
```

### Concurrency Control Pattern (Optimistic vs. Row Lock)
When reserving stock, the repository performs atomic conditional updates:

```sql
UPDATE inventory_items
SET 
    reserved_qty = reserved_qty + :quantity,
    available_qty = available_qty - :quantity,
    updated_at = NOW()
WHERE 
    tenant_id = :tenant_id 
    AND variant_id = :variant_id
    AND available_qty >= :quantity;
```
If rows affected is `0`, the system aborts the transaction with `ErrInsufficientStock`, preventing race conditions without heavy table locking.
