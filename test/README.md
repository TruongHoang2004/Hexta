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
| **Mock providers** | [`mockplatform/`](mockplatform/), [`cmd/mockplatform/`](cmd/mockplatform/) | stdlib HTTP | Simulated Meta Graph API and Telegram Bot API for local runs and load tests. |
| **Channel load** | [`cmd/loadgen/`](cmd/loadgen/) | stdlib HTTP + pgx | Signed webhook load into the connector with lost/duplicate/latency checks. |
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

## Mock providers and load generator

`mockplatform` stands in for Meta and Telegram so the connector can run without provider accounts. It implements what the connector calls (Graph `dialog/oauth`, `oauth/access_token`, `me/accounts`, `{page}/subscribed_apps`, `{page}/messages`; Bot API `getMe`, `setWebhook`, `deleteWebhook`, `sendMessage`), builds provider-shaped webhooks independently of the connector parser, and signs Facebook webhooks with the shared app secret.

| Endpoint | Purpose |
| --- | --- |
| `GET /` | Customer chat page for manual runs |
| `POST /mock/customers/{facebook,telegram}` | Send a customer message `{account_id, customer, text}` or a photo (`media: "image"`) to the connector |
| `GET /mock/messages?platform&account&customer&after` | Customer messages and sends received from the connector |
| `GET`/`PUT /mock/faults` | Send latency (global or per Page/bot ID), error rate and kind: `server_error`, `rate_limit`, `blocked`, `auth_expired`, `chat_not_found`, `window_24h` |
| `GET /mock/state`, `GET /mock/stats` | Pages, bots and webhook state; counters |
| `GET /mock/telegram/webhook?bot=` | A bot's webhook URL and secret, for `loadgen` only |

Faults apply to sends only. Page tokens are derived from the app secret, so they survive a mock restart; Telegram webhooks do not. Nothing logs tokens, secrets or message text.

`loadgen` posts webhooks at a fixed open-loop rate (at most 512 in flight; the report shows schedule lag if it falls behind). Each text starts with `lg:<run>:<seq>`. With `-db` it polls `messaging_messages` read-only during and after the attack and reports lost and duplicate messages and webhook → visible latency, precise to `-poll` (default 200ms). The API stores the provider timestamp, not its commit time, so latency is measured from the first poll that saw the row. Each poll rereads the run's rows; long T1 runs need an API-side commit timestamp first.

Run both inside the local Compose network; see [the demo runbook](../infrastructure/README.demo.md#channels-mock-providers-and-live-telegram). For T1/T2, raise `-rate`, `-duration` and `-customers`; for T5, stop Kafka (`./scripts/local.sh stop kafka`) during a run and start it again. Outbound load (T3) is not implemented yet.

```sh
go test -race ./mockplatform/...
```

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
All tests must follow the project rules in [`AGENTS.md`](../AGENTS.md) and use English for test cases, logs, and assertions.
