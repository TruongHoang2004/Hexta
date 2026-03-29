package errors

import (
	"context"
	"net/http"
	"testing"
)

func TestInit(t *testing.T) {
	Init("test-service", "trace-id-key")
	if serviceName != "test-service" {
		t.Errorf("expected serviceName to be 'test-service', got '%s'", serviceName)
	}
	if traceIdKey != "trace-id-key" {
		t.Errorf("expected traceIdKey to be 'trace-id-key', got '%v'", traceIdKey)
	}
}

func TestGetTraceId(t *testing.T) {
	key := "trace-id"
	Init("test-service", key)

	ctx := context.WithValue(context.Background(), key, "12345")
	tid := GetTraceId(ctx)
	if tid != "12345" {
		t.Errorf("expected trace id '12345', got '%s'", tid)
	}

	ctxNoTrace := context.Background()
	tidEmpty := GetTraceId(ctxNoTrace)
	if tidEmpty != "" {
		t.Errorf("expected empty trace id, got '%s'", tidEmpty)
	}
}

func TestErrors(t *testing.T) {
	key := "trace-id"
	Init("my-service", key)
	ctx := context.WithValue(context.Background(), key, "trace-123")

	t.Run("BadRequest", func(t *testing.T) {
		err := ErrBadRequest(ctx)
		if err.Code != ErrorCodeBadRequest {
			t.Errorf("expected code %s, got %s", ErrorCodeBadRequest, err.Code)
		}
		if err.HTTPStatus != http.StatusBadRequest {
			t.Errorf("expected status %d, got %d", http.StatusBadRequest, err.HTTPStatus)
		}
		if err.TraceID != "trace-123" {
			t.Errorf("expected traceID 'trace-123', got '%s'", err.TraceID)
		}
		if string(err.Source) != "my-service" {
			t.Errorf("expected source 'my-service', got '%s'", err.Source)
		}
	})

	t.Run("SystemError", func(t *testing.T) {
		detail := "db connection failed"
		err := ErrSystemError(ctx, detail)
		if err.Code != ErrorCodeSystemError {
			t.Errorf("expected code %s, got %s", ErrorCodeSystemError, err.Code)
		}
		if err.Detail != detail {
			t.Errorf("expected detail '%s', got '%s'", detail, err.Detail)
		}
	})
}
