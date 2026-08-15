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

# ==============================
# Database migrations (Atlas)
# ==============================

.PHONY: migrate-init migrate-diff migrate-apply migrate-status migrate-hash

# Check if svc is provided
check-svc:
	@if [ -z "$(svc)" ]; then echo "Error: svc is required (e.g. make migrate-diff svc=api name=init_schema)"; exit 1; fi

# Helper to run atlas in docker
# We use the migrator image, but override the entrypoint to run specific atlas commands
# The working directory in container is /workspace (which is the platform folder)
DOCKER_MIGRATE = cd infrastructure && docker compose -f docker-compose.yml -f docker-compose.migrate.yml run --rm --entrypoint atlas migrator

# Create new migration
migrate-init: check-svc
	@if [ -z "$(name)" ]; then read -p "Enter migration name: " name_val; else name_val="$(name)"; fi; \
	$(DOCKER_MIGRATE) migrate new $$name_val --config file://migrations/$(svc)/atlas.hcl --env gorm

# Generate migration diff
migrate-diff: check-svc
	@if [ -z "$(name)" ]; then read -p "Enter migration name: " name_val; else name_val="$(name)"; fi; \
	$(DOCKER_MIGRATE) migrate diff $$name_val --config file://migrations/$(svc)/atlas.hcl --env gorm

# Apply migrations for one service or all supported services through the migration image.
# Usage: make migrate-apply svc=all
#        make migrate-apply svc=user
migrate-apply:
	@cd infrastructure && MIGRATE_TARGET="$(if $(svc),$(svc),all)" docker compose -f docker-compose.yml -f docker-compose.migrate.yml run --rm migrator

# Apply migrations locally without Docker
# Usage: make migrate-apply-local svc=all
#        make migrate-apply-local svc=user
migrate-apply-local:
	@if [ -z "$(svc)" ] || [ "$(svc)" = "all" ]; then \
		echo "Applying migrations for user..."; \
		atlas migrate apply --dir "file://migrations/user" --url "postgres://postgres:postgres@localhost:5433/user?sslmode=disable"; \
		echo "Applying migrations for api..."; \
		atlas migrate apply --dir "file://migrations/api" --url "postgres://postgres:postgres@localhost:5433/api?sslmode=disable"; \
		echo "Applying migrations for catalog..."; \
		atlas migrate apply --dir "file://migrations/catalog" --url "postgres://postgres:postgres@localhost:5433/catalog?sslmode=disable"; \
	else \
		echo "Applying migrations for $(svc)..."; \
		atlas migrate apply --dir "file://migrations/$(svc)" --url "postgres://postgres:postgres@localhost:5433/$(svc)?sslmode=disable"; \
	fi

# Show migration status
migrate-status: check-svc
	$(DOCKER_MIGRATE) migrate status --config file://migrations/$(svc)/atlas.hcl --env gorm

# Hash migrations
migrate-hash: check-svc
	$(DOCKER_MIGRATE) migrate hash --dir file://migrations/$(svc)

# ==============================
# Frontend / TypeScript (pnpm)
# ==============================

.PHONY: ts-install ts-add ts-add-dev ts-add-root ts-build ts-dev

# Install all workspace dependencies
ts-install:
	@pnpm install

# Add a package to a specific workspace (Usage: make ts-add pkg=axios filter=web)
# Note: filter can be the package name (e.g., web, @ubi/sdk) or folder path (e.g., ./apps/web)
ts-add:
	@if [ -z "$(pkg)" ] || [ -z "$(filter)" ]; then echo "Error: pkg and filter are required (e.g. make ts-add pkg=axios filter=web)"; exit 1; fi
	@pnpm add $(pkg) --filter $(filter)

# Add a dev dependency to a specific workspace (Usage: make ts-add-dev pkg=typescript filter=@ubi/sdk)
ts-add-dev:
	@if [ -z "$(pkg)" ] || [ -z "$(filter)" ]; then echo "Error: pkg and filter are required (e.g. make ts-add-dev pkg=typescript filter=@ubi/sdk)"; exit 1; fi
	@pnpm add -D $(pkg) --filter $(filter)

# Add a dev package to the workspace root (Usage: make ts-add-root pkg=prettier)
ts-add-root:
	@if [ -z "$(pkg)" ]; then echo "Error: pkg is required (e.g. make ts-add-root pkg=prettier)"; exit 1; fi
	@pnpm add -w -D $(pkg)

# Build all TypeScript projects
ts-build:
	@pnpm run --recursive build

# Start dev server for a specific workspace (Usage: make ts-dev filter=web)
ts-dev:
	@if [ -z "$(filter)" ]; then echo "Error: filter is required (e.g. make ts-dev filter=web)"; exit 1; fi
	@pnpm --filter $(filter) run dev
