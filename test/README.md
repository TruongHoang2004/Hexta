# Hexta Testing Suite (`test/`)

Integration and performance testing harnesses for the [Hexta](https://github.com/TruongHoang2004/Hexta) platform.

---

## 🧪 Overview

This directory provides automated End-to-End (E2E) integration test suites and high-throughput stress/load testing tools to validate service reliability and performance under load.

| Suite | Path | Framework / Tool | Purpose |
| :--- | :--- | :--- | :--- |
| **End-to-End (E2E)** | [`e2e/`](e2e/) | [`httpexpect/v2`](https://github.com/gavv/httpexpect) | Full HTTP API integration flows against live running services. |
| **Stress Testing** | [`stress/`](stress/) / [`cmd/`](cmd/) | [`vegeta/v12`](https://github.com/tsenart/vegeta) | High-concurrency load and latency testing. |
| **Mock Fixtures** | [`mock/`](mock/) | Go Mock Generators | Reusable test data generators for users and requests. |
| **Test Config** | [`config/`](config/) | Go Config | Target base URL and environment bindings. |

---

## 🚀 Running Tests

### Prerequisites
Ensure the backing infrastructure and target API service are running before executing test suites:
```bash
# 1. Start backing services (PostgreSQL, Redis, etc.)
cd infrastructure && make up

# 2. Run database migrations
make migrate-apply svc=all

# 3. Start the API service (e.g. on port 9090 or configured BaseURL)
cd services/api && make run
```

### End-to-End (E2E) Integration Tests
The E2E suite validates the complete authentication lifecycle:
1. User registration (`POST /api/v1/auth/register`)
2. User login & token generation (`POST /api/v1/auth/login`)
3. Access token validation (`GET /api/v1/auth/validate-token`)
4. Unauthorized rejection without Bearer header
5. Refresh token rotation (`POST /api/v1/auth/refresh`)
6. Authenticated user profile retrieval (`GET /api/v1/users/me`)
7. Paginated user listing (`GET /api/v1/users/`)

Run tests using the test [`Makefile`](Makefile):
```bash
# Run all E2E tests
make test

# Run E2E tests with verbose logging
make test-v

# Invalidate Go test cache before running
make test-cache-clean test
```

### Stress / Load Testing
The stress test engine uses Vegeta to simulate concurrent user registration and API load:
```bash
# Run the stress test suite
make stress
```

Default stress parameters attack the target endpoint at 50 requests/second for 300 seconds and report percentile latencies, success rates, and status code distributions.

---

## 🛠 Useful Commands

| Command | Action |
| :--- | :--- |
| `make test` | Run all E2E integration tests |
| `make test-v` | Run tests with verbose output |
| `make stress` | Execute Vegeta load test |
| `make tidy` | Tidy Go module dependencies |
| `make fmt` | Format test code |
| `make test-cache-clean` | Clear Go test cache |

---

## 🔗 Repository & Contributing

Maintained within the [Hexta monorepo](https://github.com/TruongHoang2004/Hexta).
All tests must follow the project rules in [`GEMINI.md`](../GEMINI.md) and use English for test cases, logs, and assertions.
