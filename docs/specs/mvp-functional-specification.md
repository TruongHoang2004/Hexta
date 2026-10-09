# Hexta — MVP Functional Specification & Requirements Document

- **Document Version**: 1.0.0
- **Status**: Approved Baseline
- **Author**: Solution Architecture Team
- **Related Documents**: [`docs/vision.md`](file:///Users/truonghoang/Documents/dev/personal/Hexta/docs/vision.md), [`docs/architecture/system-architecture.md`](file:///Users/truonghoang/Documents/dev/personal/Hexta/docs/architecture/system-architecture.md), [`docs/architecture/ai-agent-engine-design.md`](file:///Users/truonghoang/Documents/dev/personal/Hexta/docs/architecture/ai-agent-engine-design.md)

---

## 1. Scope & Objective

This document defines the functional requirements and user story specifications for the **Hexta MVP (Thesis Scope)**. The MVP validates the foundational hypothesis:
> **An SME can operate its core business (Orders, Inventory, HR) through a unified data platform with an AI-native natural language interface, without risking business logic integrity.**

---

## 2. User Personas

| Persona | Role | Primary Goal | Pain Point in Traditional Software |
| :--- | :--- | :--- | :--- |
| **Mai (Owner / Operator)** | Business Owner | Needs instant answers on daily revenue, low stock, and pending orders. | Must export multiple Excel spreadsheets or navigate 10 different ERP screens to get an overview. |
| **Nam (Store Clerk)** | Operations Staff | Fast order entry while talking to customers in-store or over chat. | Tedious data entry into complex forms; high risk of misclicking SKUs or mistyping prices. |
| **Linh (Tenant Admin)** | System Admin | Invites staff, assigns roles, and reviews audit logs. | Lack of clear audit trail when staff make mistakes or blame system errors. |

---

## 3. Epic Breakdown & User Stories

### EPIC-1: Tenant Onboarding & Role-Based Access Control

#### US-1.1: Tenant Registration & Setup
- **As a** Business Owner (Mai),
- **I want to** register my organization with a business name and slug,
- **So that** I receive an isolated workspace for my company.
- **Acceptance Criteria (Gherkin)**:
  ```gherkin
  Scenario: Successful Tenant Registration
    Given a new user visits "/register"
    When they submit a valid business name "Tiệm Bánh Mì Xanh", email, and password
    Then the system creates a new Tenant record with status "active"
    And creates an Owner user account associated with that Tenant
    And issues a JWT with claims {tenant_id, role: "TenantOwner"}
  ```

#### US-1.2: Staff Provisioning & Role Assignment
- **As an** Owner,
- **I want to** invite a staff member (Nam) and assign the `OperationsStaff` role,
- **So that** he can perform order drafting but cannot view business revenue.
- **Acceptance Criteria**:
  - Staff user receives an invitation or direct account credentials.
  - Staff user cannot access `/api/v1/analytics/*` or see revenue widgets on the dashboard.

---

### EPIC-2: Product Catalog & Inventory Governance

#### US-2.1: Product & Variant Catalog Management
- **As a** Staff member or Manager,
- **I want to** view, create, and update product items with SKU, barcode, unit price, and cost price,
- **So that** the business maintains a single source of truth for products.
- **Acceptance Criteria**:
  - SKU must be unique within the tenant.
  - Deleting a product performs a soft delete (`deleted_at`), retaining historical order consistency.

#### US-2.2: Two-Phase Inventory Tracking & Stock Reservation
- **As an** Operations Staff member,
- **I want** the system to track `OnHandQty`, `ReservedQty`, and `AvailableQty`,
- **So that** we prevent overselling items that are currently in draft or pending dispatch.
- **Acceptance Criteria**:
  ```gherkin
  Scenario: Stock Reservation on Order Confirmation
    Given Product SKU "BM-01" has OnHandQty = 10 and ReservedQty = 0 (Available = 10)
    When an order with 3 units of "BM-01" is confirmed
    Then ReservedQty becomes 3 and AvailableQty becomes 7
    And OnHandQty remains 10 until shipment fulfillment
  ```

---

### EPIC-3: Order Processing & State Lifecycle

#### US-3.1: Standard Order Creation via Web Form
- **As a** Staff member,
- **I want to** create an order using standard web inputs (customer name, phone, item selector),
- **So that** I have a conventional interface if I prefer manual entry.
- **Acceptance Criteria**:
  - Validates phone format and positive item quantity.
  - Automatically calculates line item totals, subtotal, and total amount.

#### US-3.2: Order Fulfillment & Cancellation Transitions
- **As a** Store Manager,
- **I want to** advance an order from `Confirmed` to `Fulfilled` or `Cancelled`,
- **So that** stock levels and financial statuses accurately reflect reality.
- **Acceptance Criteria**:
  - On `Fulfilled`: `ReservedQty` decreases by item qty, `OnHandQty` decreases by item qty. A `StockMovement` (OUTBOUND) record is created.
  - On `Cancelled`: `ReservedQty` decreases by item qty, restoring `AvailableQty`.

---

### EPIC-4: AI Conversational Operations (Agent-Assisted Drafting)

#### US-4.1: Natural Language Order Drafting
- **As a** Staff member (Nam),
- **I want to** type or dictate an order (e.g. *"Khách anh Hoàng 0909112233 đặt 2 bánh mì đặc biệt và 1 sữa tươi"*),
- **So that** the AI Agent converts it into a structured order draft without me filling out a form.
- **Acceptance Criteria**:
  ```gherkin
  Scenario: Conversational Order Drafting
    Given user inputs prompt "Khách anh Hoàng 0909112233 đặt 2 bánh mì đặc biệt"
    When the AI agent processes the message
    Then it extracts customer name "anh Hoàng", phone "0909112233", SKU for "bánh mì đặc biệt", and quantity 2
    And checks that stock is available for the SKU
    And renders an Interactive Draft Card containing price, total, and "Confirm" button
    And NO permanent order is saved in the database yet
  ```

#### US-4.2: Human-in-the-Loop (HITL) Draft Confirmation
- **As a** Staff member,
- **I want to** review the AI-generated draft card, edit any quantity or address if needed, and click "Confirm",
- **So that** the order is finalized through the business workflow with complete accuracy.
- **Acceptance Criteria**:
  - Clicking "Confirm" calls `POST /api/v1/orders` passing the draft ID.
  - System logs an audit record with `source = 'ai_agent_draft'`.

---

### EPIC-5: Executive NLQ & Real-time Cockpit

#### US-5.1: Natural Language Metrics Query
- **As a** Business Owner (Mai),
- **I want to** ask in chat *"Hôm nay quán bán được bao nhiêu đơn, doanh thu thế nào?"*,
- **So that** I get an immediate summary without downloading spreadsheets.
- **Acceptance Criteria**:
  - System extracts timeframe ("today") and queries aggregated metrics.
  - Responds via SSE with total revenue, order count, and a summary breakdown within 2.5 seconds.
  - Staff without `analytics:revenue` permissions receive a polite unauthorized message.

#### US-5.2: Executive Cockpit Overview
- **As a** Business Owner,
- **I want to** see an executive summary card displaying Today's Revenue, Pending Orders, Low Stock Alerts, and Active Staff,
- **So that** I have instant operational visibility upon opening Hexta.

---

### EPIC-6: Audit Trail & Data Lineage

#### US-6.1: Immutable Operation Logging
- **As an** Auditor or Owner,
- **I want to** inspect the audit log for any order or inventory modification,
- **So that** I can see who initiated the change, what fields were updated, and whether it originated from a web form or an AI prompt.
- **Acceptance Criteria**:
  - Every order creation logs user ID, timestamp, before/after state JSON, and source (`web_form` vs `ai_agent_draft`).
  - Audit records cannot be updated or deleted by any user role (Append-only).

---

## 4. Acceptance Verification Matrix for Thesis Defense

| Test Case ID | Test Description | Success Benchmark |
| :--- | :--- | :--- |
| **TC-SEC-01** | Cross-tenant data isolation test | Request from Tenant A with Tenant B's order ID returns `404 Not Found`. |
| **TC-AI-01** | End-to-end AI Order Creation | Natural language prompt generates correct Draft Card; confirming creates valid Order and reserves stock. |
| **TC-AI-02** | AI Hallucination Guardrail | AI cannot execute order creation without explicit user confirmation payload. |
| **TC-NLQ-01** | Executive Revenue Query | Prompt "Doanh thu hôm nay" returns accurate sum matching database aggregate. |
| **TC-INV-01** | Concurrency Stock Reservation | Two simultaneous orders for the last remaining item result in 1 confirmed and 1 out-of-stock rejection. |
| **TC-AUD-01** | Lineage Traceability | An order created via AI displays linked `draft_id` and original user prompt in the audit inspector. |
