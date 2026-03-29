APP_NAME=merchant
MAIN_FILE=cmd/main.go
APP_FILE=bin/merchant

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
	rm -rf bin/

# ==============================
# Database migrations (Atlas)
# ==============================

.PHONY: migrate migrate-init migrate-diff migrate-apply migrate-status

migrate:
	@echo "Please run: make migrate <init|diff|apply|status>"

migrate-init:
	@read -p "Enter migration name: " name; \
	atlas migrate new $$name --env gorm

migrate-diff:
	@read -p "Enter migration name: " name; \
	atlas migrate diff $$name --env gorm

migrate-apply:
	atlas migrate apply --env gorm

migrate-status:
	atlas migrate status --env gorm

migrate-hash:
	atlas migrate hash --env gorm

# ==============================
# Proto Generation
# ==============================
gen-proto:
	PATH=$$(go env GOPATH)/bin:$$PATH protoc --go_out=. --go_opt=paths=source_relative \
	--go-grpc_out=. --go-grpc_opt=paths=source_relative \
	proto/merchant.proto
