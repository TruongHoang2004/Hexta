# Infrastructure and Backend Services Management

.PHONY: build-all up down restart logs ps clean

# Build all services (backend and infrastructure)
# Use GITLAB_TOKEN=<your_token> make build-all if needed
build-all:
	docker compose build --parallel --build-arg GITLAB_TOKEN=$(GITLAB_TOKEN)

# Start all services in detached mode
up:
	docker compose up -d

# Stop all services
down:
	docker compose down

# Restart the full stack
restart: down up

# View real-time logs for all services
logs:
	docker compose logs -f

# List all running containers
ps:
	docker compose ps

# Remove all volumes and stopped containers
clean:
	docker compose down -v
