package messaging

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/TruongHoang2004/Hexta/packages/shared/events"
)

func TestConversationChanged(t *testing.T) {
	e := ConversationChanged{Meta: events.NewMeta("tenant", "", time.Now()), ThreadID: "thread", TurnID: "turn", TurnVersion: 1}
	b, _ := json.Marshal(e)
	var got ConversationChanged
	if json.Unmarshal(b, &got) != nil || got.Validate() != nil || got.Key() != "thread" || got.EventID != e.EventID {
		t.Fatal("round trip")
	}
	for name, mutate := range map[string]func(*ConversationChanged){
		"meta":   func(e *ConversationChanged) { e.EventID = "x" },
		"thread": func(e *ConversationChanged) { e.ThreadID = "" },
		"turn":   func(e *ConversationChanged) { e.TurnID = "" },
	} {
		c := e
		mutate(&c)
		if c.Validate() == nil {
			t.Error(name)
		}
	}
}
