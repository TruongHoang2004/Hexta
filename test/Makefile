# ==============================
# Testing operations
# ==============================

.PHONY: test test-v test-cache-clean tidy fmt stress

# Run all e2e tests
test:
	go test ./e2e/...

# Run all e2e tests verbosely
test-v:
	go test ./e2e/... -v

# Clean go test cache
test-cache-clean:
	go clean -testcache

# Tidy Go modules
tidy:
	go mod tidy

# Format code
fmt:
	go fmt ./...

# Run stress test
stress:
	go run ./cmd/main.go