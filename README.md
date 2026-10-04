# Hexta Monorepo

[![Go Version](https://img.shields.io/badge/Go-1.24%2B-00ADD8?style=flat&logo=go)](https://go.dev/)
[![Node Version](https://img.shields.io/badge/Node.js-20%2B_LTS-339933?style=flat&logo=node.js)](https://nodejs.org/)
[![pnpm Version](https://img.shields.io/badge/pnpm-10%2B-F69220?style=flat&logo=pnpm)](https://pnpm.io/)
[![Next.js](https://img.shields.io/badge/Next.js-16-black?style=flat&logo=next.js)](https://nextjs.org/)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-17-4169E1?style=flat&logo=postgresql)](https://www.postgresql.org/)
[![License](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](LICENSE)

An enterprise-grade, multi-tenant e-commerce platform built as a high-performance polyglot monorepo with Go microservices, Next.js frontend applications, and containerized distributed infrastructure.

---

## Table of Contents
1. [Overview & Architecture](#overview--architecture)
2. [Prerequisites & Tooling](#prerequisites--tooling)
3. [Service & Port Reference Matrix](#service--port-reference-matrix)
4. [How to Run System in Local Machine](#how-to-run-system-in-local-machine)
   - [Mode A: Full Docker Quickstart](#mode-a-full-docker-quickstart)
   - [Mode B: Hybrid Local Development (Recommended)](#mode-b-hybrid-local-development-recommended)
5. [Environment Variables & Configuration](#environment-variables--configuration)
6. [Database Migrations with Atlas](#database-migrations-with-atlas)
7. [Verification, Health Checks & Documentation](#verification-health-checks--documentation)
8. [Testing & Quality Checks](#testing--quality-checks)
9. [Troubleshooting & FAQs](#troubleshooting--faqs)

---

## Overview & Architecture

Hexta is architected with a decoupled frontend, high-throughput Go backend services, and scalable event-driven infrastructure.

```
┌────────────────────────────────────────────────────────────────────────┐
│                        Next.js Frontend Apps                           │
│  Storefront (:3000) │ Admin (:3001) │ Docs (:3002) │ Landing (:3003)   │
└───────────────────────────────────┬────────────────────────────────────┘
                                    │ (HTTP REST / Axios SDK)
                                    ▼
┌────────────────────────────────────────────────────────────────────────┐
│                     Hexta API Gateway (Go / Gin)                       │
│                           Port: 8080                                   │
│  - Presentation (HTTP Controller, DTOs, Swagger Docs)                  │
│  - Core Services (Auth, Google OAuth, Tenant, AI Orchestration)        │
│  - Repository Layer (GORM) & Dependency Injection (Uber Fx)            │
└───────┬──────────────┬──────────────┬──────────────┬───────────┬───────┘
        │              │              │              │           │
        ▼              ▼              ▼              ▼           ▼
┌──────────────┐ ┌───────────┐ ┌────────────┐ ┌────────────┐ ┌──────────┐
│  PostgreSQL  │ │   Redis   │ │Elasticsearch││  MinIO S3  │ │  Kafka   │
│  Port: 5433  │ │ Port: 6379│ │ Port: 9200 │ │  Port: 9000│ │Port: 9092│
└──────────────┘ └───────────┘ └────────────┘ └────────────┘ └──────────┘
```

### Monorepo Structure
- **`apps/`**: Next.js 16 applications
  - `apps/web`: Primary customer e-commerce storefront (Port `3000`)
  - `apps/admin`: Operational management dashboard (Port `3001`)
  - `apps/docs`: Developer & API documentation portal (Port `3002`)
  - `apps/landing`: Product showcase and marketing site (Port `3003`)
- **`services/`**: Go microservices
  - `services/api`: Core API Gateway, authentication, tenant management, and AI services (Port `8080`)
- **`packages/`**: Shared libraries
  - `packages/sdk`: `@ubi/sdk` TypeScript client SDK with silent token refresh interceptors
  - `packages/ui`: `@hexta/ui` shared React/Tailwind component library
  - `packages/shared`: Go shared packages (`gitlab.com/ecommercehub1/shared`) for telemetry and errors
- **`infrastructure/`**: Docker Compose configuration for PostgreSQL, Redis, MinIO, Elasticsearch, Kafka, and Qdrant
- **`migrations/`**: Declarative schema migrations managed via Atlas CLI

---

## Prerequisites & Tooling

Ensure the following runtimes and command-line tools are installed on your workstation (macOS, Linux, or Windows WSL2):

| Tool | Minimum Version | Installation / Verification |
|---|---|---|
| **Docker & Docker Compose** | Docker v24+, Compose v2+ | `docker compose version` |
| **Go** | v1.24+ (or v1.22+) | `go version` |
| **Node.js** | v20+ LTS | `node -v` |
| **pnpm** | v10+ | `pnpm -v` (install via `corepack enable` or `npm i -g pnpm`) |
| **Atlas CLI** | Latest | `atlas version` ([Installation Guide](https://atlasgo.io/getting-started)) |
| **Make** | GNU Make 3.81+ | `make --version` |

---

## Service & Port Reference Matrix

All external ports exposed during local execution:

| Service | Host Port | Protocol | Default Credentials / URL | Purpose |
|---|---|---|---|---|
| **PostgreSQL 17** | `5433` | TCP | `postgres / postgres`<br/>DBs: `api`, `user`, `catalog`, `dev`, `order` | Relational persistence |
| **Redis** | `6379` | TCP | No password default | Cache and session storage |
| **Elasticsearch 8** | `9200` | HTTP | `http://localhost:9200` (xpack disabled) | Full-text search engine |
| **MinIO API** | `9000` | HTTP | User: `minioadmin`<br/>Pass: `minioadmin` | S3-compatible asset storage |
| **MinIO Console** | `9001` | HTTP | `http://localhost:9001` | Web storage dashboard |
| **Apache Kafka** | `9092` | TCP | PLAINTEXT | Event streaming broker |
| **Qdrant Vector DB** | `6333` | HTTP | `http://localhost:6333/dashboard` | AI vector embeddings & search |
| **RedisInsight** | `5540` | HTTP | `http://localhost:5540` | Redis visual inspector |
| **API Service (Go)** | `8080` | HTTP | `http://localhost:8080/api/v1` | Core REST API Gateway |
| **Web Storefront** | `3000` | HTTP | `http://localhost:3000` | Next.js Storefront app |
| **Admin Portal** | `3001` | HTTP | `http://localhost:3001` | Next.js Admin app |
| **Docs Portal** | `3002` | HTTP | `http://localhost:3002` | Next.js Docs app |
| **Landing Site** | `3003` | HTTP | `http://localhost:3003` | Next.js Landing app |

> [!NOTE]
> PostgreSQL is intentionally mapped to host port **`5433`** (`5433:5432`) to prevent conflicts with any local PostgreSQL instance you may already have running on the standard `5432` port.

---

## How to Run System in Local Machine

You can choose between two development workflows depending on your requirements:

### Mode A: Full Docker Quickstart
Spin up the entire stack (Infrastructure + Backend API + RedisInsight) inside Docker with a single command:

```bash
# 1. Start all containers in the background
make local-up

# 2. View streaming logs
docker compose -f infrastructure/local-all.yml logs -f

# 3. Teardown all containers when finished
make local-down
```

---

### Mode B: Hybrid Local Development (Recommended)
This mode runs the data tier in Docker while running Go backend and Next.js frontend applications natively on your host machine. This enables hot-reload, instant rebuilds, and interactive debugging via Delve.

#### Step 1: Start Infrastructure Services
Start PostgreSQL, Redis, MinIO, Elasticsearch, Kafka, and Qdrant in detached mode:
```bash
make infra-up
```
Verify containers are healthy:
```bash
docker compose -f infrastructure/docker-compose.yml ps
```

#### Step 2: Apply Database Migrations
Initialize database schemas using Atlas:
```bash
# Apply migrations via the Atlas Docker container (Recommended)
make migrate-apply svc=api

# Or apply migrations directly using your local Atlas CLI:
make migrate-apply-local svc=api
```

#### Step 3: Run Backend Service (`services/api`)
1. Create your local environment file:
   ```bash
   cp services/api/.env.example services/api/.env
   ```
2. Run the Go API server:
   ```bash
   cd services/api
   go run cmd/main.go
   ```
   *The API will start and listen on `http://localhost:8080`.*

3. *(Optional)* Run backend in debug mode with [Delve](https://github.com/go-delve/delve):
   ```bash
   make debug
   ```

#### Step 4: Install Dependencies & Run Frontend Applications
1. Install workspace dependencies from the monorepo root:
   ```bash
   pnpm install
   ```
2. Create web application environment file:
   ```bash
   cp apps/web/.env.example apps/web/.env
   ```
3. Start the desired application:
   ```bash
   # Launch Web Storefront (Port 3000)
   make ts-dev filter=web

   # Or launch any other app
   pnpm --filter admin run dev -- -p 3001
   pnpm --filter docs run dev -- -p 3002
   pnpm --filter landing run dev -- -p 3003
   ```

---

## Environment Variables & Configuration

Configuration files and environment templates are located across each subsystem:

### 1. Infrastructure (`infrastructure/.env`)
Copy the template to customize Docker container parameters:
```bash
cp infrastructure/.env.example infrastructure/.env
```
Default configuration values:
```dotenv
POSTGRES_USER=postgres
POSTGRES_PASSWORD=postgres
POSTGRES_MULTIPLE_DATABASES=user,api,catalog,dev,order
MINIO_ROOT_USER=minioadmin
MINIO_ROOT_PASSWORD=minioadmin
```

### 2. Backend Service (`services/api/.env`)
Copy the template to provide local secrets:
```bash
cp services/api/.env.example services/api/.env
```
Key configuration fields:
```dotenv
# JWT Secrets
JWT_ACCESS_TOKEN_SECRET=dev-jwt-access-secret-minimum-32-characters-long!
JWT_ACCESS_TOKEN_EXPIRE=3600
JWT_REFRESH_TOKEN_SECRET=dev-jwt-refresh-secret-minimum-32-characters-long!
JWT_REFRESH_TOKEN_EXPIRE=604800

# Google OAuth2 Credentials
GOOGLE_CLIENT_ID=your-google-client-id.apps.googleusercontent.com
GOOGLE_CLIENT_SECRET=your-google-client-secret
GOOGLE_REDIRECT_URL=http://localhost:8080/api/v1/auth/google/callback

# AI Provider Configuration (openai | gemini | anthropic)
AI_PROVIDER=openai
AI_API_KEY=your-ai-api-key
AI_MODEL=gpt-4o-mini
AI_BASE_URL=https://api.openai.com/v1
AI_TIMEOUT=30
AI_MAX_TOKENS=2048
AI_TEMPERATURE=0.7
```

### 3. Frontend Web App (`apps/web/.env`)
```bash
cp apps/web/.env.example apps/web/.env
```
Content:
```dotenv
NEXT_PUBLIC_API_URL=http://localhost:8080
```

---

## Database Migrations with Atlas

We use **Atlas** with GORM to handle declarative migrations.

| Command | Description |
|---|---|
| `make migrate-apply svc=api` | Apply all pending migrations via the migrator container |
| `make migrate-apply-local svc=api` | Apply pending migrations using host `atlas` binary |
| `make migrate-diff svc=api name=<name>` | Compare GORM models against dev database and generate diff |
| `make migrate-status svc=api` | Inspect current migration status and applied revisions |
| `make migrate-hash svc=api` | Recompute the `atlas.sum` integrity hash file |

---

## Verification, Health Checks & Documentation

### 1. API Health Check
Test that the API Gateway is running and able to communicate with PostgreSQL and Redis:
```bash
curl -s http://localhost:8080/api/v1/health | jq
```
Expected response:
```json
{
  "code": "SUCCESS",
  "message": "Success",
  "data": {
    "status": "up",
    "details": {
      "database": "ok",
      "redis": "ok"
    }
  }
}
```

### 2. Swagger API Documentation
Access the interactive OpenAPI / Swagger UI:
- **URL**: [http://localhost:8080/api/v1/swagger/index.html](http://localhost:8080/api/v1/swagger/index.html)

### 3. Ping Test
```bash
curl -s http://localhost:8080/api/v1/ping
# {"message":"pong"}
```

---

## Testing & Quality Checks

Run monorepo test suites and compilation checks:

```bash
# 1. Run all Go tests
go test ./services/api/...

# 2. Build all TypeScript workspaces (apps & packages)
make ts-build
# or: pnpm run --recursive build

# 3. Lint frontend applications
pnpm run --recursive lint
```

---

## Troubleshooting & FAQs

### 1. Port 5432 / 5433 Connection Refused
- **Issue**: Attempting to connect to PostgreSQL on default `5432` fails.
- **Cause**: The container exposes port `5433` on the host to prevent conflicts with native PostgreSQL installations.
- **Fix**: Connect using `postgres://postgres:postgres@localhost:5433/api?sslmode=disable`.

### 2. Atlas Checksum Mismatch (`atlas.sum` error)
- **Issue**: `checksum mismatch for migration file ...`
- **Cause**: Migration files were edited manually after hash calculation.
- **Fix**: Run `make migrate-hash svc=api` to update the checksums.

### 3. Resetting All Local Data & Volumes
- **Issue**: Corrupt database state or stale Redis keys.
- **Fix**: Run `make clean` to stop containers and wipe volume storage:
  ```bash
  make clean
  make infra-up
  make migrate-apply svc=api
  ```

### 4. Google OAuth Redirect Mismatch
- **Issue**: Google login returns `redirect_uri_mismatch`.
- **Cause**: Google Cloud Console authorized redirect URIs do not match the local environment.
- **Fix**: Add `http://localhost:8080/api/v1/auth/google/callback` to the **Authorized redirect URIs** in your Google Cloud Console OAuth 2.0 Client settings.
