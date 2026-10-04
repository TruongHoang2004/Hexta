# Local Development Guide: How to Run the Hexta System Locally

This document provides detailed setup instructions, execution workflows, and operational references for running the Hexta monorepo on a local workstation.

---

## 1. Quick Reference Matrix

| Service | Port | Protocol | Credentials / Details |
|---|---|---|---|
| **PostgreSQL 17** | `5433` | TCP | `postgres / postgres` (DBs: `api`, `user`, `catalog`, `dev`, `order`) |
| **Redis** | `6379` | TCP | Standard cache (no password) |
| **Elasticsearch 8**| `9200` | HTTP | Search service (`http://localhost:9200`) |
| **MinIO S3** | `9000` / `9001` | HTTP | User: `minioadmin`, Pass: `minioadmin` (Console: `localhost:9001`) |
| **Kafka** | `9092` | TCP | Broker at `localhost:9092` |
| **Qdrant** | `6333` | HTTP | Vector DB dashboard at `http://localhost:6333/dashboard` |
| **RedisInsight** | `5540` | HTTP | GUI at `http://localhost:5540` |
| **API Gateway** | `8080` | HTTP | REST API at `http://localhost:8080/api/v1` |
| **Web Storefront** | `3000` | HTTP | Next.js Storefront at `http://localhost:3000` |
| **Admin Portal** | `3001` | HTTP | Next.js Admin at `http://localhost:3001` |
| **Docs Portal** | `3002` | HTTP | Next.js Documentation at `http://localhost:3002` |
| **Landing Site** | `3003` | HTTP | Next.js Marketing Landing at `http://localhost:3003` |

---

## 2. Tooling Prerequisites

Install the following tools before starting:
- **Docker** (v24+) & **Docker Compose** (v2+)
- **Go** (v1.24+ recommended, or v1.22+)
- **Node.js** (v20+ LTS)
- **pnpm** (v10+): `corepack enable && corepack prepare pnpm@latest --activate`
- **Atlas CLI**: `curl -sSf https://atlasgo.sh | sh`
- **GNU Make**

---

## 3. Execution Workflows

### Mode A: Full Containerized Stack
Best for full-system smoke testing and CI integration.

```bash
# Start all infrastructure and backend services in Docker
make local-up

# Follow backend logs
docker compose -f infrastructure/local-all.yml logs -f

# Teardown stack
make local-down
```

---

### Mode B: Hybrid Local Setup (Recommended)
Best for active feature development with hot reload and fast feedback loops.

#### Step 1: Infrastructure
```bash
make infra-up
```

#### Step 2: Database Migrations
```bash
make migrate-apply svc=api
```

#### Step 3: Backend (`services/api`)
```bash
cp services/api/.env.example services/api/.env
cd services/api && go run cmd/main.go
```
The Go API service starts on `http://localhost:8080`.

#### Step 4: Frontend (`apps/web`)
```bash
pnpm install
cp apps/web/.env.example apps/web/.env
make ts-dev filter=web
```
The Web storefront starts on `http://localhost:3000`.

---

## 4. Health Verification & Testing

```bash
# Health check
curl -s http://localhost:8080/api/v1/health | jq

# Interactive Swagger UI
open http://localhost:8080/api/v1/swagger/index.html

# Run backend unit tests
go test ./services/api/...

# Run frontend builds
pnpm run --recursive build
```

---

## 5. Troubleshooting & FAQs

- **Postgres Port Collision**: Always connect to port `5433` on the host machine.
- **Atlas Migration Hash Mismatch**: Execute `make migrate-hash svc=api` to update `atlas.sum`.
- **Wiping Local Volumes**: Execute `make clean` to stop containers and clear persistent storage.
- **Google OAuth Setup**: Ensure Google Cloud Console contains `http://localhost:8080/api/v1/auth/google/callback` in authorized redirect URIs.
