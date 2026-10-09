package service

import (
	"context"
	"testing"
)

func TestHealthService_NilDependencies(t *testing.T) {
	svc := NewHealthService(nil, nil)
	result := svc.CheckHealth(context.Background())

	if result.Status != "down" {
		t.Fatalf("expected status 'down' when dependencies are nil, got '%s'", result.Status)
	}

	if result.Details["database"] != "unconfigured" {
		t.Errorf("expected details[database] == 'unconfigured', got '%s'", result.Details["database"])
	}

	if result.Details["redis"] != "unconfigured" {
		t.Errorf("expected details[redis] == 'unconfigured', got '%s'", result.Details["redis"])
	}
}
