package messaging

import (
	"encoding/json"
	"testing"
	"time"
)

func TestConversationChanged(t *testing.T) {
	e := ConversationChanged{ThreadID: "thread", TurnID: "turn", TurnVersion: 1, LastActivityAt: time.Now()}
	b, _ := json.Marshal(e)
	var got ConversationChanged
	if json.Unmarshal(b, &got) != nil || got.Validate() != nil || got.Key() != "thread" {
		t.Fatal("round trip")
	}
	for name, mutate := range map[string]func(*ConversationChanged){
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
