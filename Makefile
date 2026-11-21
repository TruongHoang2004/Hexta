# Main Makefile for Hexta Platform

.PHONY: debug infra-up infra-down local-up local-down clean up down restart ps logs reset

# ==============================
# Full local stack in Docker (scripts/local.sh + infrastructure/compose.demo.yml)
# web, api, ingress, ai, connector, postgres, redis, kafka + init/migration jobs
# ==============================

# Build and start everything, waiting until all services are healthy
up:
	@./scripts/local.sh up --build -d --wait
	@echo "Web: http://localhost:3000  API: http://localhost:8080/api/v1/health  AI: http://localhost:58000/health/ready"

# Stop everything, keeping database and Kafka volumes
down:
	@./scripts/local.sh down

restart: down up

ps:
	@./scripts/local.sh ps -a

# Follow logs (optional: make logs s="api connector")
logs:
	@./scripts/local.sh logs -f --tail 100 $(s)

# DESTRUCTIVE: stop and delete all local data volumes (accounts, tenants, AI runs, Kafka topics)
reset:
	@printf "Delete all local Hexta data volumes? [y/N] " && read ans && [ "$$ans" = y ]
	@./scripts/local.sh --profile test --profile tools down -v

# Run all services (Backend + Infra + Logging) in Docker
local-up:
	@cd infrastructure && docker compose -f local-all.yml up -d

# Stop all services in Docker
local-down:
	@cd infrastructure && docker compose -f local-all.yml down

# Start all services in debug mode locally using Delve
debug:
	@./scripts/debug-all.sh

# Start only infrastructure (Postgres, Redis, Elasticsearch, Kafka, Qdrant)
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

.PHONY: migrate migrate-init migrate-diff migrate-apply migrate-apply-local migrate-status migrate-hash migrate-validate migration-db

# Current database-owning services are api and ai. Use a private migrator URL.
# Docker commands reuse the local stack; host commands use exported URLs.
DOCKER_MIGRATE = ./scripts/local.sh run --build --rm --no-deps --entrypoint atlas migrator

check-svc:
	@case "$(svc)" in api|ai|connector) ;; *) echo "svc must be api, ai or connector" >&2; exit 1;; esac

migrate-init: check-svc
	@test -n "$(name)" || { echo "name is required" >&2; exit 1; }
	$(DOCKER_MIGRATE) migrate new "$(name)" --env "$(svc)"

migrate-diff: check-svc migration-db
	@test -n "$(name)" || { echo "name is required" >&2; exit 1; }
	$(DOCKER_MIGRATE) migrate diff "$(name)" --env "$(svc)"

migration-db:
	@case "$(if $(svc),$(svc),all)" in api|ai|connector|all) ;; *) echo "svc must be api, ai, connector or all" >&2; exit 1;; esac
	./scripts/local.sh up -d --wait postgres
	./scripts/local.sh run --rm --no-deps db-init

migrate migrate-apply: migration-db
	@case "$(if $(svc),$(svc),all)" in api|ai|connector|all) ;; *) echo "svc must be api, ai, connector or all" >&2; exit 1;; esac
	./scripts/local.sh run --build --rm --no-deps -e MIGRATE_TARGET="$(if $(svc),$(svc),all)" migrator

migrate-apply-local:
	MIGRATE_TARGET="$(if $(svc),$(svc),all)" ./scripts/migrate.sh apply

migrate-status: migration-db
	./scripts/local.sh run --build --rm --no-deps -e MIGRATE_TARGET="$(if $(svc),$(svc),all)" migrator status

migrate-hash:
	MIGRATE_TARGET="$(if $(svc),$(svc),all)" ./scripts/migrate.sh hash

migrate-validate:
	MIGRATE_TARGET="$(if $(svc),$(svc),all)" ./scripts/migrate.sh validate

# ==============================
# API Documentation (Swagger)
# ==============================

.PHONY: swagger
swagger:
	@cd services/api && swag init -g cmd/main.go -o docs --parseDependency --parseInternal

# ==============================
# Frontend / TypeScript (pnpm)
# ==============================

.PHONY: ts-install ts-add ts-add-dev ts-add-root ts-build ts-dev

# Install all workspace dependencies
ts-install:
	@pnpm install

# Add a package to a specific workspace (Usage: make ts-add pkg=axios filter=web)
# Note: filter can be the package name (e.g., web, @hexta/sdk) or folder path (e.g., ./apps/web)
ts-add:
	@if [ -z "$(pkg)" ] || [ -z "$(filter)" ]; then echo "Error: pkg and filter are required (e.g. make ts-add pkg=axios filter=web)"; exit 1; fi
	@pnpm add $(pkg) --filter $(filter)

# Add a dev dependency to a specific workspace (Usage: make ts-add-dev pkg=typescript filter=@hexta/sdk)
ts-add-dev:
	@if [ -z "$(pkg)" ] || [ -z "$(filter)" ]; then echo "Error: pkg and filter are required (e.g. make ts-add-dev pkg=typescript filter=@hexta/sdk)"; exit 1; fi
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

# ==============================
# Graduation Thesis (LaTeX -> PDF)
# ==============================

.PHONY: thesis thesis-pdf thesis-watch thesis-clean thesis-distclean

THESIS_DIR := graduation thesis
THESIS_MAIN := main.tex
THESIS_PATH := /Library/TeX/texbin:$(PATH)

# Compile graduation thesis (main.tex -> main.pdf) using XeLaTeX via latexmk
thesis thesis-pdf:
	@PATH="$(THESIS_PATH)" command -v latexmk >/dev/null 2>&1 || { echo "Error: latexmk not found. Install MacTeX/TeX Live or add /Library/TeX/texbin to PATH." >&2; exit 1; }
	@cd "$(THESIS_DIR)" && PATH="$(THESIS_PATH)" latexmk -xelatex -interaction=nonstopmode -halt-on-error "$(THESIS_MAIN)"

# Watch and auto-recompile graduation thesis on file changes
thesis-watch:
	@PATH="$(THESIS_PATH)" command -v latexmk >/dev/null 2>&1 || { echo "Error: latexmk not found. Install MacTeX/TeX Live or add /Library/TeX/texbin to PATH." >&2; exit 1; }
	@cd "$(THESIS_DIR)" && PATH="$(THESIS_PATH)" latexmk -xelatex -pvc -interaction=nonstopmode "$(THESIS_MAIN)"

# Remove intermediate LaTeX build artifacts, keeping main.pdf
thesis-clean:
	@cd "$(THESIS_DIR)" && PATH="$(THESIS_PATH)" latexmk -c "$(THESIS_MAIN)"

# Remove all LaTeX build artifacts including main.pdf
thesis-distclean:
	@cd "$(THESIS_DIR)" && PATH="$(THESIS_PATH)" latexmk -C "$(THESIS_MAIN)"
