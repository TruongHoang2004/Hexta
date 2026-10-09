package repository

import (
	"testing"
)

func TestSessionCacheWrapper_Initialization(t *testing.T) {
	wrapper := NewSessionCacheWrapper(nil, nil)
	if wrapper == nil {
		t.Fatalf("expected non-nil session cache wrapper instance")
	}
}
