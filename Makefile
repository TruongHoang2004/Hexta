# Main Makefile for CommerceHub Backend

.PHONY: debug infra-up infra-down local-up local-down clean

# Run all services (Backend + Infra + Logging) in Docker
local-up:
	@cd infrastructure && docker compose -f local-all.yml up -d

# Stop all services in Docker
local-down:
	@cd infrastructure && docker compose -f local-all.yml down

# Start all services in debug mode locally using Delve
debug:
	@./scripts/debug-all.sh

# Start only infrastructure (Postgres, Redis, Elasticsearch, Minio)
infra-up:
	@cd infrastructure && docker compose up -d

# Stop infrastructure
infra-down:
	@cd infrastructure && docker compose down

# Stop everything and clean up volumes
clean:
	@./scripts/debug-all.sh stop 2>/dev/null || true
	@cd infrastructure && docker compose down -v
