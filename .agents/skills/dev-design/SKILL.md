---
name: dev-design
description: Use this skill when designing technical architecture, database schemas, API contracts, or system workflows before writing code. It guides the agent through design best practices and saves the document to agentic-memory/designs/.
---

# Technical Design Skill (`dev-design`)

Use this skill when architecting new services, database schemas, API contracts, or refactoring existing system modules.

## Workflow

### 1. Context & Architecture Alignment
- Select applicable domain guidance through [the root router](../../../rules/README.md). Derive boundaries and contracts from actual code and approved design; do not apply Go-specific templates to Python or frontend-only work.
- Stay strictly within the approved PRD scope; do not pull in out-of-scope domain orchestration (such as unrelated routing or workflow logic).

### 2. Formulate the Technical Design Document
Include only sections relevant to the requested design:
1. **Design Summary**: Problem statement, architectural vision, and design decisions.
2. **System Architecture & Data Flow**:
   - Mermaid diagram illustrating module interactions.
   - Synchronous (HTTP/gRPC) vs Asynchronous (queues/events) flows.
3. **Data Model & Atlas Migrations (explicitly list all schema changes)**:
   - Every added/modified/removed table, column, data type, primary key, foreign key, and partial/unique index in `migrations/schema/{api,connector,ai}.sql`.
   - Note the owning service and `make migrate-diff` target.
4. **Inter-Service Communication, API Contracts & DTOs (explicitly list every change)**:
   - **HTTP Endpoints:** Every added, modified, or removed route (method, path, authZ), with concrete **Request DTO** and **Response DTO** definitions.
   - **gRPC (`proto/`):** Every added, modified, or removed RPC and message DTO (or explicitly state `No changes`).
   - **Kafka Topics:** Every added or modified topic, producer → consumer pair, and exact **event envelope / payload DTO**.
   - Swagger or protobuf ownership and compatibility where applicable; Go business errors use the existing shared errors package.
5. **Security, Caching & Performance Considerations**:
   - Authentication & Authorization checks.
   - Existing cache behavior and invalidation only when affected; no new caching assumption.
   - Transaction boundaries and DB connection handling.

### 3. Persist when authorized

Explicit read-only tasks return the proposal in conversation. Otherwise follow [project memory](../../../agentic-memory/README.md): create this document only when it adds distinct lasting information, and link it from the primary epic session.

Persist into `agentic-memory/designs/`
- Determine `<feature-name>`.
- Get today's date in `YYYY-MM-DD` format.
- Write the design document to:
  `agentic-memory/designs/YYYY-MM-DD_<feature-name>_design.md`
- In your conversation response, provide a brief summary and a clickable link to the saved design file.
