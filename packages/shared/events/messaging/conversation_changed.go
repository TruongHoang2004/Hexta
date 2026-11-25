package messaging

import (
	"errors"
	"time"
)

// ConversationChangedTopic carries ConversationChanged inside events.Envelope. Producer: API messaging command side
// (outbox). Consumer: API conversation read model.
const ConversationChangedTopic = "hexta.messaging.conversations.v1"

// ConversationChangedEventType is the envelope event_type and the API outbox row type.
const ConversationChangedEventType = "messaging.conversation.changed"

// ConversationChanged is the full thread and its latest turn (open, or the most recently closed) after a committed
// change, so consumers can upsert idempotently. It carries no message bodies or customer contact data.
type ConversationChanged struct {
	ThreadID        string    `json:"thread_id"`
	ExternalKey     string    `json:"external_key"`
	DisplayName     string    `json:"display_name"`
	ConnectionID    string    `json:"connection_id,omitempty"`
	RecipientID     string    `json:"recipient_id,omitempty"`
	ThreadCreatedAt time.Time `json:"thread_created_at"`
	ThreadUpdatedAt time.Time `json:"thread_updated_at"`
	TurnID          string    `json:"turn_id"`
	Lifecycle       string    `json:"lifecycle"`
	Control         string    `json:"control"`
	AssigneeID      *string   `json:"assignee_id"`
	NeedsHuman      bool      `json:"needs_human"`
	// NeedsHumanReason is the handoff reason code behind NeedsHuman; omitted when none was given.
	NeedsHumanReason  string    `json:"needs_human_reason,omitempty"`
	ContextRevision   int64     `json:"context_revision"`
	ControlGeneration int64     `json:"control_generation"`
	TurnVersion       int64     `json:"turn_version"`
	LastActivityAt    time.Time `json:"last_activity_at"`
	// ReviewPending is true while a bot-handled turn's proposal waits for staff review; omitted when false, so
	// snapshots written before the field existed decode as false.
	ReviewPending bool `json:"review_pending,omitempty"`
}

// Key orders changes per thread.
func (e ConversationChanged) Key() string { return e.ThreadID }

func (e ConversationChanged) Validate() error {
	if e.ThreadID == "" || e.TurnID == "" {
		return errors.New("invalid conversation change")
	}
	return nil
}
