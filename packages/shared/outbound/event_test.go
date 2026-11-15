package outbound

import (
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

func validIntent() IntentEvent {
	now := time.Now()
	return IntentEvent{SchemaVersion: 1, IntentID: uuid.NewString(), AttemptID: uuid.NewString(), CorrelationID: uuid.NewString(), TenantID: uuid.NewString(), ConnectionID: uuid.NewString(), ConnectionVersion: 1, Platform: "facebook", AccountID: "page", RecipientID: "psid", Text: "hello", IssuedAt: now, ExpiresAt: now.Add(time.Minute)}
}

func TestIntentValidation(t *testing.T) {
	if err := validIntent().Validate(); err != nil {
		t.Fatal(err)
	}
	for name, mutate := range map[string]func(*IntentEvent){
		"version":   func(e *IntentEvent) { e.SchemaVersion = 2 },
		"attempt":   func(e *IntentEvent) { e.AttemptID = "x" },
		"platform":  func(e *IntentEvent) { e.Platform = "zalo" },
		"empty":     func(e *IntentEvent) { e.Text = "" },
		"long":      func(e *IntentEvent) { e.Text = strings.Repeat("a", MaxTextRunes+1) },
		"lifetime":  func(e *IntentEvent) { e.ExpiresAt = e.IssuedAt },
		"recipient": func(e *IntentEvent) { e.RecipientID = "" },
	} {
		e := validIntent()
		mutate(&e)
		if e.Validate() == nil {
			t.Fatal("accepted invalid intent:", name)
		}
	}
}

func TestResultValidation(t *testing.T) {
	r := ResultEvent{SchemaVersion: 1, IntentID: uuid.NewString(), AttemptID: uuid.NewString(), TenantID: uuid.NewString(), ConnectionID: uuid.NewString(), Status: StatusSent, OccurredAt: time.Now()}
	if r.Validate() != nil {
		t.Fatal("rejected sent result")
	}
	r.Status = StatusFailed
	if r.Validate() == nil {
		t.Fatal("failed result requires a category")
	}
	r.ErrorCategory = ErrorWindow24h
	if r.Validate() != nil {
		t.Fatal("rejected classified failure")
	}
	r.Status = "delivered"
	if r.Validate() == nil {
		t.Fatal("accepted unknown status")
	}
}
