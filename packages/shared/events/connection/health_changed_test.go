package connection

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/TruongHoang2004/Hexta/packages/shared/events"
	"github.com/google/uuid"
)

func TestHealthChanged(t *testing.T) {
	e := HealthChanged{Meta: events.NewMeta(uuid.NewString(), "", time.Now()), ConnectionID: uuid.NewString(), ConnectionVersion: 2, Status: StatusAuthExpired, ErrorCode: "190"}
	b, _ := json.Marshal(e)
	var got HealthChanged
	if json.Unmarshal(b, &got) != nil || got.Validate() != nil || got.Key() != e.ConnectionID {
		t.Fatal("round trip")
	}
	for name, mutate := range map[string]func(*HealthChanged){
		"status":  func(e *HealthChanged) { e.Status = "ok" },
		"version": func(e *HealthChanged) { e.ConnectionVersion = 0 },
		"tenant":  func(e *HealthChanged) { e.TenantID = "tenant" },
	} {
		c := e
		mutate(&c)
		if c.Validate() == nil {
			t.Error(name)
		}
	}
}
