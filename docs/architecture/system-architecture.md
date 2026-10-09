# Hexta — System Architecture Document (SAD)

- **Document Version**: 1.0.0
- **Status**: Approved Baseline
- **Author**: Solution Architecture Team
- **Target Audience**: Engineering Leads, Backend/Frontend Developers, Thesis Review Committee
- **Related Vision Document**: [`docs/vision.md`](file:///Users/truonghoang/Documents/dev/personal/Hexta/docs/vision.md)

---

## 1. Executive Summary & Architectural Drivers

### 1.1. Context & Business Vision
Hexta is an AI-native, unified business operating platform tailored for Small and Medium Enterprises (SMEs). Traditional enterprise resource planning (ERP) platforms suffer from fragmented modules, steep learning curves, and manual form-heavy operations. Conversely, naive generative AI wrappers pose hallucinations and data integrity hazards.

Hexta bridges this dichotomy through a core architectural tenet:
> **AI acts as the natural intent and interaction layer; the Business Workflow Engine remains the deterministic, immutable source of truth.**

```
┌────────────────────────────────────────────────────────┐
│                   Natural Language UI                  │
└───────────────────────────┬────────────────────────────┘
                            │ Conversational Intent
                            ▼
┌────────────────────────────────────────────────────────┐
│             AI Agent Orchestrator (LLM)                │
│    - Intent Extraction      - Slot Filling             │
│    - Tool Calling           - Draft Proposal           │
└───────────────────────────┬────────────────────────────┘
                            │ Interactive Draft Card
                            ▼
┌────────────────────────────────────────────────────────┐
│         Human-in-the-Loop Confirmation Gate            │
└───────────────────────────┬────────────────────────────┘
                            │ Approved Command
                            ▼
┌────────────────────────────────────────────────────────┐
│          Deterministic Business Workflow               │
│    - Domain State Machine   - Validation Rules         │
│    - Inventory Reservation  - Multi-tenant Isolation   │
│    - Outbox Event Bus       - Audit & Lineage Log      │
└────────────────────────────────────────────────────────┘
```

### 1.2. Primary Architectural Drivers
1. **Multi-Tenancy**: Logical isolation of customer data with zero cross-tenant data leaks.
2. **Determinism & Auditability**: Every state modification must be traceable (who, when, what, via web form or AI draft).
3. **Low-Latency AI Interactions**: Streaming responses (Server-Sent Events) and sub-3-second end-to-end response times for natural language queries.
4. **Modularity & Monorepo Governance**: Decoupled service domains following Clean Architecture and Uber Fx dependency injection.

---

## 2. C4 Model — System Architecture

### 2.1. Level 1: System Context Diagram

```mermaid
C4Context
    title System Context Diagram for Hexta Platform

    Person(owner, "Business Owner / Operator", "Monitors business health, queries insights, inspects anomalies.")
    Person(staff, "Operations Staff", "Processes orders, tracks stock, performs daily operations via natural language and forms.")
    Person(admin, "Platform Admin", "Manages tenant subscriptions, platform telemetry, and system configs.")

    System(hexta, "Hexta Platform", "Unified multi-tenant business operating platform with AI interaction layer.")

    System_Ext(llm, "External LLM Provider", "Google Gemini / OpenAI APIs for NLP, intent classification, and summarization.")
    System_Ext(ecom, "E-commerce Platforms", "Shopee, TikTok Shop, Lazada (Stage 2 connectors).")
    System_Ext(bank, "Banking & Payments", "VietQR, Payment gateways for transaction reconciliation.")

    Rel(owner, hexta, "Queries metrics, views executive dashboards", "HTTPS / WSS")
    Rel(staff, hexta, "Creates order drafts, confirms operations, views stock", "HTTPS / WSS")
    Rel(admin, hexta, "Manages tenants and platform health", "HTTPS")

    Rel(hexta, llm, "Sends prompt context & tool schemas, receives tool calls", "HTTPS / TLS")
    Rel(hexta, ecom, "Syncs orders and inventory webhooks", "HTTPS")
    Rel(hexta, bank, "Reconciles payments", "HTTPS")
```

---

### 2.2. Level 2: Container Diagram

```mermaid
flowchart TD
    subgraph Clients["Client Layer (apps/)"]
        Web["Web Application (:3000)<br/>Next.js 16 + React 19 + Tailwind<br/>(Staff Operations & Executive Cockpit)"]
        Admin["Admin Portal (:3001)<br/>Next.js 16<br/>(Platform & Tenant Governance)"]
    end

    subgraph Gateway["Backend Layer (services/)"]
        API["Hexta Core API (:8080)<br/>Go 1.24 + Gin + Uber Fx<br/>(Auth, Tenant, OMS, IMS, HRM, AI Gateway)"]
    end

    subgraph DataStore["Data & Persistence Layer (infrastructure/)"]
        Postgres[("PostgreSQL 17 (:5433)<br/>OLTP Relational Store<br/>Tenant-scoped schemas & RLS")]
        Redis[("Redis 7 (:6379)<br/>Cache, Session Store,<br/>Distributed Locks, Rate Limiting")]
        Qdrant[("Qdrant Vector DB (:6333)<br/>Semantic Tool Search,<br/>RAG Knowledge Embeddings")]
        Kafka[("Apache Kafka (:9092)<br/>Event Streaming Broker & Domain Events")]
        MinIO[("MinIO S3 (:9000)<br/>Invoices, Receipts, File Attachments")]
    end

    subgraph External["External Integrations"]
        GeminiAPI["Google Gemini API<br/>(Language Model Services)"]
    end

    Web -->|"REST API / SSE (Port 8080)"| API
    Admin -->|"REST API (Port 8080)"| API
    
    API -->|"GORM / pgxpool"| Postgres
    API -->|"go-redis"| Redis
    API -->|"gRPC / HTTP"| Qdrant
    API -->|"Sarama / Kafka Producer"| Kafka
    API -->|"AWS S3 SDK Go"| MinIO
    API -->|"google-genai SDK"| GeminiAPI
```

---

### 2.3. Level 3: Component Diagram (`services/api`)

The Go backend adheres strictly to the 5-layer Clean Architecture outlined in [`GEMINI.md`](file:///Users/truonghoang/Documents/dev/personal/Hexta/GEMINI.md):

```
services/api/
├── cmd/
│   └── main.go                         # Uber Fx runtime assembler & lifecycle hooks
├── config/                             # Viper / Env config loader
└── internal/
    ├── present/http/
    │   ├── controller/                 # HTTP controllers (Gin handlers, validation)
    │   ├── dto/                        # Request / Response DTOs
    │   ├── middleware/                 # AuthN, TenantContext, RateLimit, Telemetry
    │   └── router/                     # Public & Private router registration
    ├── core/
    │   ├── domain/                     # Pure domain entities, values, state machines
    │   ├── service/                    # Business orchestration services
    │   └── port/                       # Inbound & Outbound interfaces
    ├── repository/                     # Database access (GORM, Redis wrappers)
    ├── infrastructure/                 # External clients (DB, Redis, Gemini, Kafka)
    └── bootstrap/                      # Uber Fx module builders (BuildDatabase, BuildService, etc.)
```

```mermaid
graph TD
    subgraph HTTP_Presentation["1. Presentation Layer (HTTP / Gin)"]
        Router["Router Registry"]
        MidTenant["TenantContext Middleware"]
        MidAuth["JWT Auth Middleware"]
        OrderCtrl["OrderController"]
        StockCtrl["InventoryController"]
        HRCtrl["HRController"]
        AICtrl["AIController"]
    end

    subgraph Core_Services["2. Core Domain Services (Uber Fx)"]
        AuthSvc["AuthService"]
        TenantSvc["TenantService"]
        OrderSvc["OrderService<br/>(State Machine & Calculation)"]
        StockSvc["InventoryService<br/>(Reservation & Allocation)"]
        HRSvc["HRService<br/>(Staff & Workload)"]
        AISvc["AIAgentService<br/>(Intent, Tool Calling, Draft Generator)"]
        AuditSvc["AuditService<br/>(Lineage & Action Logger)"]
    end

    subgraph Repositories["3. Repository Layer (GORM)"]
        TenantRepo["TenantRepository"]
        OrderRepo["OrderRepository"]
        StockRepo["InventoryRepository"]
        HRRepo["HRRepository"]
        AuditRepo["AuditRepository"]
    end

    subgraph DB_Infrastructure["4. Infrastructure Layer"]
        GormDB["PostgreSQL Connection Pool"]
        RedisClient["Redis Client"]
        GeminiClient["Gemini LLM Client"]
    end

    Router --> MidAuth --> MidTenant
    MidTenant --> OrderCtrl
    MidTenant --> StockCtrl
    MidTenant --> HRCtrl
    MidTenant --> AICtrl

    OrderCtrl --> OrderSvc
    StockCtrl --> StockSvc
    HRCtrl --> HRSvc
    AICtrl --> AISvc

    AISvc --> OrderSvc
    AISvc --> StockSvc
    AISvc --> GeminiClient

    OrderSvc --> StockSvc
    OrderSvc --> AuditSvc
    OrderSvc --> OrderRepo
    StockSvc --> StockRepo
    HRSvc --> HRRepo
    TenantSvc --> TenantRepo
    AuditSvc --> AuditRepo

    OrderRepo --> GormDB
    StockRepo --> GormDB
    HRRepo --> GormDB
    TenantRepo --> GormDB
    AuditRepo --> GormDB
```

---

## 3. Dynamic Runtime Workflows

### 3.1. Workflow 1: AI-Assisted Operation (Conversational Order Creation)
Demonstrates the safe separation between AI intent translation and deterministic execution.

```mermaid
sequenceDiagram
    autonumber
    actor Staff as Operations Staff
    participant Web as Web Application (Next.js)
    participant AI as AI Controller & Agent Service
    participant LLM as Google Gemini API
    participant OMS as Order Service
    participant IMS as Inventory Service
    participant Audit as Audit & Lineage Service
    participant DB as PostgreSQL 17

    Staff->>Web: Submits: "Customer John (0988776655) buys 2 Blue Shirts size M"
    Web->>AI: POST /api/v1/ai/agent/draft {prompt: "..."} with Tenant JWT
    AI->>AI: Inject Tenant Schema & Tool Definitions
    AI->>LLM: generateContent with Function Calling schemas
    LLM-->>AI: ToolCall: create_order_draft(customer, items, qty)
    AI->>IMS: CheckStockAvailability(tenant_id, sku, qty=2)
    IMS-->>AI: Stock OK (In-stock: 15, Available: 13)
    AI-->>Web: Return 200 OK with Structured Draft JSON (Interactive UI Card)
    
    Web->>Staff: Render Interactive Draft Card (Price, Quantity, Customer Name)
    Staff->>Web: Inspects, modifies notes, clicks "Confirm & Submit"
    
    Web->>OMS: POST /api/v1/orders {draft_payload} (Standard Business API)
    activate OMS
    OMS->>OMS: Validate Business Invariants (Price > 0, Active Tenant)
    OMS->>IMS: ReserveStock(tenant_id, items)
    activate IMS
    IMS->>DB: UPDATE inventory SET reserved_qty = reserved_qty + 2 WHERE sku = ...
    IMS-->>OMS: Stock Reserved Successfully
    deactivate IMS
    OMS->>DB: INSERT INTO orders (status='confirmed', source='ai_draft', ...)
    OMS->>Audit: RecordAuditLog(action='ORDER_CREATED', source='ai_agent', draft_id=...)
    Audit->>DB: INSERT INTO audit_logs (...)
    OMS-->>Web: Order Created #ORD-10492
    deactivate OMS
    Web-->>Staff: Display Success Notification & Live Order Details
```

---

### 3.2. Workflow 2: Natural Language Query (Executive NLQ)
Demonstrates secure, read-only analytics execution respecting Tenant RBAC.

```mermaid
sequenceDiagram
    autonumber
    actor Owner as Business Owner
    participant Web as Web App (Executive Cockpit)
    participant AI as AI Controller
    participant LLM as Google Gemini API
    participant QueryEngine as Analytics / Aggregation Engine
    participant DB as PostgreSQL (Read Replica / Tenant Schema)

    Owner->>Web: Queries: "What is my top selling product this week?"
    Web->>AI: POST /api/v1/ai/query/stream (SSE)
    AI->>AI: Verify RBAC: HasPermission(user, "analytics.read")
    AI->>LLM: Prompt + Metric Catalog Tools (e.g. get_sales_aggregate)
    LLM-->>AI: Call Tool: get_sales_aggregate(period='7d', group_by='product')
    AI->>QueryEngine: ExecuteSalesAggregate(tenant_id, range=7d)
    QueryEngine->>DB: SELECT sku, SUM(quantity), SUM(total_price) ... WHERE tenant_id = ?
    DB-->>QueryEngine: Result Set
    QueryEngine-->>AI: Aggregate Data Payload
    AI->>LLM: Feed Data into LLM for Natural Language Synthesis
    LLM-->>AI: Streamed textual summary + Chart specification
    AI-->>Web: Server-Sent Events (Text stream + JSON Chart config)
    Web-->>Owner: Renders analytical narrative + Bar Chart widget
```

---

## 4. Multi-Tenancy & Data Isolation Model

Hexta implements **Pooled Multi-Tenancy with Row-Level Discrimination and Context Propagation**:

1. **Tenant Identification**: Every authenticated request passes through `middleware.TenantContext()` which extracts the `tenant_id` from the verified JWT claims and sets it into the standard Go `context.Context`.
2. **Repository Boundary**:
   - Every GORM database model embeds a `TenantID uuid.UUID` field with indexing:
     ```go
     type BaseTenantModel struct {
         ID        uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
         TenantID  uuid.UUID      `gorm:"type:uuid;not null;index"`
         CreatedAt time.Time      `gorm:"not null"`
         UpdatedAt time.Time      `gorm:"not null"`
         DeletedAt gorm.DeletedAt `gorm:"index"`
     }
     ```
   - Repository base wrappers inject `WHERE tenant_id = ?` into all `SELECT`, `UPDATE`, and `DELETE` queries.
3. **Database-Level Protection (PostgreSQL RLS)**:
   - For defense-in-depth, PostgreSQL Row-Level Security (RLS) policies are active:
     ```sql
     ALTER TABLE orders ENABLE ROW LEVEL SECURITY;
     CREATE POLICY tenant_isolation_policy ON orders
         FOR ALL
         USING (tenant_id = current_setting('app.current_tenant_id')::uuid);
     ```

---

## 5. Technology Stack Rationale Matrix

| Technology | Layer | Choice Rationale |
| :--- | :--- | :--- |
| **Go 1.24+** | Backend Services | Native concurrency (goroutines), minimal memory footprint, rapid cold-start times, compiled type safety. |
| **Gin Framework** | HTTP Routing | Mature, battle-tested HTTP router with high throughput and low allocs. |
| **Uber Fx** | Dependency Injection | Modular dependency graph resolution; explicit lifecycle management (`OnStart`, `OnStop`) avoiding global state. |
| **Next.js 16 (App Router)** | Client Applications | Server Components for instant data loading, Client Components for dynamic chat interaction, unified TypeScript SDK. |
| **PostgreSQL 17** | Relational Persistence | Robust transactional guarantees (ACID), native JSONB support, Row-Level Security (RLS), performant indexing. |
| **Atlas CLI** | Migration Engine | Declarative schema management; eliminates manual migration drift across development and CI/CD pipelines. |
| **Google Gemini API** | AI / LLM Engine | Native Function Calling support, low latency, Vietnamese comprehension, large token context window. |
| **Qdrant Vector DB** | Semantic Search | Rust-based high-performance vector search engine for RAG knowledge bases and tool indexing. |
| **Apache Kafka** | Event Streaming | Partitioned event broker ensuring reliable decoupling and event replayability across modules. |

---

## 6. Observability, Security & Non-Functional Architecture

### 6.1. Telemetry & Observability
- **Distributed Tracing**: OpenTelemetry instrumentation integrated into Gin HTTP middleware (`otelgin`) and outbound HTTP/DB clients.
- **Metrics**: Prometheus collector exposed at `/metrics` collecting request counts, latencies (p50, p95, p99), DB connection pool stats, and LLM call latencies.
- **Structured Logging**: Uber Zap logger outputting structured JSON logs with correlation IDs (`trace_id`, `tenant_id`, `user_id`).

### 6.2. Security Architecture
- **Stateless Authentication**: Short-lived Access Tokens (15 min) + Refresh Tokens stored in Redis with automatic rotation.
- **Input Sanitization**: Strict DTO validation using Go `validator/v10` prior to passing parameters to business logic.
- **LLM Prompt Injection Defense**: System prompts enforce rigid schema boundaries. User input is treated strictly as conversational data, never as executable instructions.
- **Prompt PII Protection**: Masking of customer PII (e.g. phone numbers, national IDs) prior to transmission to external LLMs where feasible.

---

## 7. Architectural Roadmap & Alignment

```
Current Status:
[x] Auth & Session Infrastructure
[x] Tenant Core Service & Repository
[x] Atlas Migration Foundation

MVP Target (Thesis Delivery):
[ ] Order Domain (Catalog, Pricing, OMS State Machine)
[ ] Inventory Domain (Stock Levels, Stock Reservations)
[ ] Basic HR Domain (Staff Profiles, Workload Mapping)
[ ] AI Agent Engine (Intent Classifier, Tool Registry, Draft Card Generator)
[ ] Executive Dashboard & Chat UI (Next.js apps/web)
[ ] Audit Trail & Data Lineage Engine

Phase 2 (Post-Thesis Commercialization):
[ ] E-commerce Connectors (Shopee, TikTok Shop)
[ ] Payment Webhooks & Auto-Reconciliation
[ ] Multi-warehouse Inventory Routing

Phase 3 (Enterprise Intelligence):
[ ] Proactive Anomaly Detection
[ ] Automated Replenishment Triggers
[ ] Predictive Sales & Cashflow Modeling
```
