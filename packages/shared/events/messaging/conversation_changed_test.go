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

// review_pending is additive: snapshots written before it existed decode as false, and a set flag round-trips.
func TestConversationChangedReviewPendingIsBackwardCompatible(t *testing.T) {
	var old ConversationChanged
	if err := json.Unmarshal([]byte(`{"thread_id":"thread","turn_id":"turn","turn_version":1}`), &old); err != nil || old.ReviewPending {
		t.Fatal("a snapshot without review_pending must decode as false", err)
	}
	b, _ := json.Marshal(ConversationChanged{ThreadID: "thread", TurnID: "turn", ReviewPending: true})
	var got ConversationChanged
	if json.Unmarshal(b, &got) != nil || !got.ReviewPending {
		t.Fatal("review_pending must round-trip")
	}
}
