# Hexta Infrastructure & Local Environment

Orchestration, backing services, and observability tooling for local development across the [Hexta](https://github.com/TruongHoang2004/Hexta) platform.

---

## 🏗 Stack Overview

The local infrastructure stack runs via Docker Compose and provides all dependencies required by Hexta backend services.

| Service | Port(s) | Description | Default Credentials |
| :--- | :--- | :--- | :--- |
| **PostgreSQL 17** | `5433:5432` | Relational database (`user`, `api`, `catalog`, `dev`, `order`) | `postgres` / `postgres` |
| **Redis** | `6379:6379` | In-memory session store & cache | No auth (local) |
| **Kafka** | `9092:9092` | Event streaming and asynchronous messaging | `PLAINTEXT://localhost:9092` |
| **Elasticsearch** | `9200`, `9300` | Search indexing and document store | No auth (local dev) |
| **MinIO** | `9000`, `9001` | S3-compatible object storage (API + Console) | `minioadmin` / `minioadmin` |
| **Qdrant** | `6333`, `6334` | Vector database for embeddings and similarity search | HTTP `6333`, gRPC `6334` |
| **Grafana** | `3000` | Observability dashboards (traces, logs, metrics) | `admin` / `admin` |
| **Prometheus** | `9090` | Time-series metrics collection | Web UI `9090` |
| **Loki** | `3100` | Centralized log aggregation | HTTP API |
| **Tempo** | `3200`, `4317` | Distributed tracing backend (OTLP gRPC 4317) | OTLP exporter |

---

## 📁 Compose Configurations

- **`docker-compose.yml`**: Core infrastructure stack (Postgres, Redis, Kafka, Elasticsearch, MinIO, Qdrant).
- **`docker-compose.migrate.yml`**: Atlas migration runner for declarative and version-controlled database schemas.
- **`docker-compose.log.yml`**: Full observability pipeline (Grafana, Loki, Promtail, Prometheus, Tempo, OTel Collector).
- **`docker-compose.ui.yml`**: Local management UIs (Adminer, Redis Commander).
- **`local-all.yml`**: Composite stack running all services and infrastructure concurrently.

---

## 🚀 Quickstart & Operational Commands

Manage the infrastructure using the provided [`Makefile`](Makefile) or from the repository root:

### Starting & Stopping Infrastructure
```bash
# Start all core backing services in background
make up

# Check status of running containers
make ps

# Tail real-time logs
make logs

# Stop all services
make down

# Clean stop (removes containers and mounted volume data)
make clean
```

### Full Platform Startup
From the project root:
```bash
# Start all services, infrastructure, and logging together
make local-up

# Stop all services and infrastructure
make local-down
```

---

## 🗄 Database Initialization & Migrations

### Automatic Multi-Database Setup
PostgreSQL initializes multiple logical databases (`user`, `api`, `catalog`, `dev`, `order`) on first startup using the initialization scripts in [`init-scripts/`](init-scripts/).

### Atlas Migrations
Hexta uses [Atlas](https://atlasgo.io/) with GORM models to manage schema versions:

```bash
# Generate a new migration diff from GORM models
make migrate-diff svc=api name=add_user_table

# Apply migrations across all databases
make migrate-apply svc=all

# Apply migrations for a single service
make migrate-apply svc=api

# Inspect migration status
make migrate-status svc=api
```

---

## 🔗 Repository & Contributing

This infrastructure configuration is maintained in the [Hexta monorepo](https://github.com/TruongHoang2004/Hexta).
For development conventions and guidelines, refer to [`GEMINI.md`](../GEMINI.md).
