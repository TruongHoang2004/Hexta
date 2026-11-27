package messaging

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/TruongHoang2004/Hexta/packages/shared/events"
	"github.com/google/uuid"
)

func message() MessageReceived {
	now := time.Now().UTC().Truncate(time.Second)
	return MessageReceived{Meta: events.Meta{EventID: uuid.NewString(), CorrelationID: uuid.NewString(), TenantID: "tenant", OccurredAt: now}, Platform: events.PlatformFacebook, ConnectionID: uuid.NewString(), ConnectionVersion: 1, AccountID: "page", ConversationKey: "psid", MessageID: "m1", SourceIdentity: "message:m1", Sender: "psid", Recipient: "page", Direction: DirectionInbound, Text: "hello", SourceAt: now, Acquisition: "webhook", Purpose: "live"}
}

func TestMessageReceivedRoundTripsFlatJSON(t *testing.T) {
	e := message()
	b, err := json.Marshal(e)
	if err != nil {
		t.Fatal(err)
	}
	var raw map[string]any
	_ = json.Unmarshal(b, &raw)
	for _, field := range []string{"event_id", "tenant_id", "occurred_at", "connection_id", "external_conversation_key", "text"} {
		if _, ok := raw[field]; !ok {
			t.Fatal("missing flat field", field)
		}
	}
	for _, field := range []string{"event_type", "schema_version", "attachments", "data"} {
		if _, ok := raw[field]; ok {
			t.Fatal("unexpected field", field)
		}
	}
	var got MessageReceived
	if err = json.Unmarshal(b, &got); err != nil || got.Validate() != nil || got.Identity() != e.Identity() || got.Key() != e.Key() {
		t.Fatal("round trip", err)
	}
}

func TestMessageReceivedKeysPerConnectionConversation(t *testing.T) {
	a, b := message(), message()
	b.ConnectionID = a.ConnectionID
	if a.Key() != b.Key() {
		t.Fatal("same conversation must share a partition key")
	}
	b.ConnectionID = uuid.NewString()
	if a.Key() == b.Key() || a.Identity() == b.Identity() {
		t.Fatal("two connections of one tenant must stay independent")
	}
}

func TestMessageReceivedValidate(t *testing.T) {
	media := message()
	media.Text, media.UnsupportedMedia = "", []string{"image", "sticker"}
	echo := message()
	echo.Direction = DirectionOutbound
	for name, e := range map[string]MessageReceived{"text": message(), "media": media, "echo": echo} {
		if err := e.Validate(); err != nil {
			t.Error(name, err)
		}
	}
	cases := map[string]func(*MessageReceived){
		"meta":             func(e *MessageReceived) { e.EventID = "" },
		"platform":         func(e *MessageReceived) { e.Platform = "fixture" },
		"connection":       func(e *MessageReceived) { e.ConnectionID = "x" },
		"version":          func(e *MessageReceived) { e.ConnectionVersion = 0 },
		"source":           func(e *MessageReceived) { e.SourceIdentity = "" },
		"direction":        func(e *MessageReceived) { e.Direction = "sideways" },
		"acquisition":      func(e *MessageReceived) { e.Acquisition = "push" },
		"purpose":          func(e *MessageReceived) { e.Purpose = "test" },
		"empty":            func(e *MessageReceived) { e.Text = "" },
		"too long":         func(e *MessageReceived) { e.Text = strings.Repeat("a", MaxTextRunes+1) },
		"text with media":  func(e *MessageReceived) { e.UnsupportedMedia = []string{"image"} },
		"unknown media":    func(e *MessageReceived) { e.Text, e.UnsupportedMedia = "", []string{"template"} },
		"too much media":   func(e *MessageReceived) { e.Text, e.UnsupportedMedia = "", make([]string, MaxUnsupportedMedia+1) },
		"invalid encoding": func(e *MessageReceived) { e.Text = "\xff" },
	}
	for name, mutate := range cases {
		e := message()
		mutate(&e)
		if e.Validate() == nil {
			t.Error(name)
		}
	}
}
