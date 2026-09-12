---
name: dev-design
description: Use this skill when designing technical architecture, database schemas, API contracts, or system workflows before writing code. It guides the agent through design best practices and saves the document to agentic-memory/designs/.
---

# Technical Design Skill (`dev-design`)

Use this skill when architecting new services, database schemas, API contracts, or refactoring existing system modules.

## Workflow

### 1. Context & Architecture Alignment
- Consult `agentic-memory/rules/` and `GEMINI.md`:
  - 5-layer Go architecture (`cmd/` -> `controller/` -> `service/` -> `repository/` -> `infrastructure/`).
  - GORM models with Atlas migrations.
  - Gin DTOs and `response.Response[T]` wrappers.
  - Next.js App Router with TypeScript strict mode.

### 2. Formulate the Technical Design Document
Structure the design document with:
1. **Design Summary**: Problem statement, architectural vision, and design decisions.
2. **System Architecture & Data Flow**:
   - Mermaid diagram illustrating module interactions.
   - Synchronous (HTTP/gRPC) vs Asynchronous (queues/events) flows.
3. **Data Model & Schema (PostgreSQL / GORM / Atlas)**:
   - Entity tables, data types, primary keys, foreign keys, and indexes.
   - Migration script outline.
4. **API Contracts**:
   - HTTP method, route, request DTO, response DTO.
   - Swagger documentation format.
   - Error code mappings using `gitlab.com/ecommercehub1/shared/pkg/errors`.
5. **Security, Caching & Performance Considerations**:
   - Authentication & Authorization checks.
   - Redis caching strategy and cache invalidation.
   - Transaction boundaries and DB connection handling.

### 3. Persist into `agentic-memory/designs/`
- Determine `<feature-name>`.
- Get today's date in `YYYY-MM-DD` format.
- Write the design document to:
  `agentic-memory/designs/YYYY-MM-DD_<feature-name>_design.md`
- In your conversation response, provide a brief summary and a clickable link to the saved design file.
