APP_NAME=catalog
MAIN_FILE=cmd/main.go
APP_FILE=bin/catalog

# ==============================
# Development / Build / Run
# ==============================
dev: 
	go run -mod=mod -tags=dev $(MAIN_FILE)

build: 
	go build -o bin/$(APP_NAME) $(MAIN_FILE)

run: 
	./$(APP_FILE)

clean:
	rm -rf bin/ docs/docs.go docs/swagger.json docs/swagger.yaml

# ==============================
# Database migrations (Atlas)
# ==============================

.PHONY: migrate migrate-init migrate-diff migrate-apply migrate-status

# "make migrate <target>"
migrate:
	@echo "Please run: make migrate <init|diff|apply|status>"

# Create new migration
migrate-init:
	@read -p "Enter migration name: " name; \
	atlas migrate new $$name --env gorm

# Generate migration diff
migrate-diff:
	@read -p "Enter migration name: " name; \
	atlas migrate diff $$name --env gorm

# Apply migrations
migrate-apply:
	atlas migrate apply --env gorm

# Show migration status
migrate-status:
	atlas migrate status --env gorm


# ==============================
# Proto Generation
# ==============================
gen-proto:
	PATH=$$(go env GOPATH)/bin:$$PATH protoc --go_out=. --go_opt=paths=source_relative \
	--go-grpc_out=. --go-grpc_opt=paths=source_relative \
	proto/catalog.proto
