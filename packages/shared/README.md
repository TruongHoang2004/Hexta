# Hexta Shared Packages (`packages/shared`)

Centralized shared libraries, domain utilities, validation logic, observability tooling, and Protocol Buffers contracts for the [Hexta](https://github.com/TruongHoang2004/Hexta) platform.

---

## 📦 Package Overview

| Package | Path | Description |
| :--- | :--- | :--- |
| **Errors** | [`pkg/errors`](pkg/errors) | Standardized domain error handling, HTTP status code mapping, and trace-correlated errors. |
| **Logger** | [`pkg/logger`](pkg/logger) | Production-ready Uber Zap wrapper with structured JSON output and contextual trace logging. |
| **Telemetry** | [`pkg/telemetry`](pkg/telemetry) | OpenTelemetry (OTel) SDK initialization, OTLP gRPC export, and Gin tracing middleware. |
| **Validator** | [`pkg/validator`](pkg/validator) | Custom validation rules (e.g., decimals, VAT tax patterns) for `go-playground/validator`. |
| **Common Utilities** | [`pkg/common`](pkg/common) | Shared casting and utility helpers. |
| **Protobuf Contracts** | [`proto`](proto) / [`gen/go`](gen/go) | Protocol Buffers definitions and generated Go code (e.g. Identity v1). |

---

## 🚀 Installation & Usage

### 1. Error Handling (`pkg/errors`)
Provides a consistent `*errors.Error` type used across service and repository layers.

```go
import "github.com/TruongHoang2004/Hexta/packages/shared/pkg/errors"

// Initialize package with service name and context trace ID key at startup
errors.Init("api-service", "trace_id")

// Return structured domain errors
func (s *userService) FindByID(ctx context.Context, id string) (*User, error) {
    user, err := s.repo.FindByID(ctx, id)
    if err != nil {
        return nil, errors.ErrNotFound(ctx, "User", "not found")
    }
    return user, nil
}
```

### 2. Structured Logging (`pkg/logger`)
Wraps Uber Zap with support for contextual logging (extracting trace IDs from OpenTelemetry spans):

```go
import "github.com/TruongHoang2004/Hexta/packages/shared/pkg/logger"

// Create a production or development logger
log := logger.NewLogger(logger.LoggerOption{
    IsProd: true, // JSON in prod, console encoder in dev
})

// Standard logging
log.Info("Service started on port %d", 8080)

// Context-aware logging (includes trace_id from OTel span if present)
log.InfoC(ctx, "Processed transaction successfully")
```

### 3. OpenTelemetry Tracing (`pkg/telemetry`)
Initializes the OpenTelemetry TracerProvider and registers trace context propagation:

```go
import (
    "context"
    "github.com/TruongHoang2004/Hexta/packages/shared/pkg/telemetry"
)

func main() {
    ctx := context.Background()
    tp, err := telemetry.InitTracer(ctx, "api-service")
    if err != nil {
        log.Fatalf("failed to initialize tracer: %v", err)
    }
    defer tp.Shutdown(ctx)
}
```

### 4. Custom Request Validation (`pkg/validator`)
Extends `go-playground/validator/v10` with custom types and domain validations:

```go
import (
    "github.com/TruongHoang2004/Hexta/packages/shared/pkg/validator"
    "github.com/gin-gonic/gin/binding"
)

func SetupValidator(log *logger.Logger) {
    v := validator.NewValidator()
    validator.RegisterDecimalTypeFunc(v)
    validator.RegisterValidations(v, log)
}
```

---

## 🧪 Testing

Run unit tests across all shared packages:

```bash
go test ./... -v
```

---

## 🔗 Repository & Contributing

This library is part of the [Hexta monorepo](https://github.com/TruongHoang2004/Hexta).
When introducing modifications or new shared utilities:
1. Ensure all code adheres to English-only identifiers and documentation.
2. Include comprehensive unit tests (`*_test.go`).
3. Follow the 5-layer architecture rules defined in [`GEMINI.md`](../../GEMINI.md).
