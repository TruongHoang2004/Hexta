package utils

import (
	"strings"
	"testing"
)

func TestGenerateOrderID(t *testing.T) {
	id := GenerateOrderID(8)
	if !strings.HasPrefix(id, "NL") {
		t.Errorf("GenerateOrderID(8) = %s; want prefix NL", id)
	}
	if len(id) != 10 { // NL + 8 chars
		t.Errorf("GenerateOrderID(8) length = %d; want 10", len(id))
	}
}

func TestGenerateUniqueOrderID(t *testing.T) {
	id := GenerateUniqueOrderID()
	if !strings.HasPrefix(id, "NL") {
		t.Errorf("GenerateUniqueOrderID() = %s; want prefix NL", id)
	}
	// Length check: NL (2) + Timestamp (14) + Random (8) = 24
	if len(id) != 24 {
		t.Errorf("GenerateUniqueOrderID() length = %d; want 24", len(id))
	}
}

func TestGenerateUniqueOrderID_Uniqueness(t *testing.T) {
	count := 999999
	ids := make(map[string]bool)
	for i := 0; i < count; i++ {
		id := GenerateUniqueOrderID()
		if ids[id] {
			t.Fatalf("Duplicate ID generated: %s", id)
		}
		ids[id] = true
	}
}
