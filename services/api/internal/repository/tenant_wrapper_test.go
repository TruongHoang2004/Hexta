package repository

import (
	"testing"
)

func TestTenantCacheWrapper_Initialization(t *testing.T) {
	wrapper := NewTenantCacheWrapper(nil, nil)
	if wrapper == nil {
		t.Fatalf("expected non-nil tenant cache wrapper instance")
	}
}
