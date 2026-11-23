package delivery

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/TruongHoang2004/Hexta/packages/shared/events"
	"github.com/google/uuid"
)

func request() SendRequested {
	now := time.Now().UTC()
	return SendRequested{Meta: events.Meta{EventID: uuid.NewString(), CorrelationID: uuid.NewString(), TenantID: uuid.NewString(), OccurredAt: now}, IntentID: uuid.NewString(), AttemptID: uuid.NewString(), ConnectionID: uuid.NewString(), ConnectionVersion: 1, Platform: events.PlatformTelegram, AccountID: "bot", RecipientID: "chat", Text: "hi", ExpiresAt: now.Add(time.Minute)}
}

func TestSendRequested(t *testing.T) {
	e := request()
	b, _ := json.Marshal(e)
	var got SendRequested
	if json.Unmarshal(b, &got) != nil || got.Validate() != nil || got.Key() != e.ConnectionID+":chat" {
		t.Fatal("round trip")
	}
	for name, mutate := range map[string]func(*SendRequested){
		"no correlation": func(e *SendRequested) { e.CorrelationID = "" },
		"tenant uuid":    func(e *SendRequested) { e.TenantID = "tenant" },
		"intent":         func(e *SendRequested) { e.IntentID = "x" },
		"platform":       func(e *SendRequested) { e.Platform = "sms" },
		"recipient":      func(e *SendRequested) { e.RecipientID = strings.Repeat("r", 256) },
		"empty text":     func(e *SendRequested) { e.Text = "" },
		"long text":      func(e *SendRequested) { e.Text = strings.Repeat("a", MaxTextRunes+1) },
		"expired":        func(e *SendRequested) { e.ExpiresAt = e.OccurredAt },
	} {
		c := request()
		mutate(&c)
		if c.Validate() == nil {
			t.Error(name)
		}
	}
}
