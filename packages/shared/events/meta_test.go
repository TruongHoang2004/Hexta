package events

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestMetaValidate(t *testing.T) {
	valid := NewMeta("tenant", uuid.NewString(), time.Now())
	if err := valid.Validate(); err != nil {
		t.Fatal(err)
	}
	for name, m := range map[string]Meta{
		"no event id":     {TenantID: "t", OccurredAt: time.Now()},
		"bad event id":    {EventID: "x", TenantID: "t", OccurredAt: time.Now()},
		"no tenant":       {EventID: uuid.NewString(), OccurredAt: time.Now()},
		"long tenant":     {EventID: uuid.NewString(), TenantID: string(make([]byte, 37)), OccurredAt: time.Now()},
		"no time":         {EventID: uuid.NewString(), TenantID: "t"},
		"bad correlation": {EventID: uuid.NewString(), CorrelationID: "x", TenantID: "t", OccurredAt: time.Now()},
	} {
		if m.Validate() == nil {
			t.Error(name)
		}
	}
}

func TestHashIsStableAndOrderSensitive(t *testing.T) {
	if Hash("a", "b") != Hash("a", "b") || Hash("a", "b") == Hash("b", "a") || Hash("ab") == Hash("a", "b") {
		t.Fatal("hash")
	}
}
