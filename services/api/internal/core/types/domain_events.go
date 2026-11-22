package types

import (
	"encoding/json"
	"time"

	"github.com/shopspring/decimal"
)

// OrderEventsTopic carries order domain events. Consumers must ignore unknown event types and newer schema versions.
const OrderEventsTopic = "hexta.orders.events.v1"

// DomainEventEnvelope is the Kafka message value; Data holds the event-type-specific payload.
type DomainEventEnvelope struct {
	SchemaVersion int             `json:"schema_version"`
	EventID       string          `json:"event_id"`
	EventType     string          `json:"event_type"`
	TenantID      string          `json:"tenant_id"`
	AggregateID   string          `json:"aggregate_id"`
	OccurredAt    time.Time       `json:"occurred_at"`
	Data          json.RawMessage `json:"data"`
}

// OrderConfirmed is the order.confirmed payload. It carries no customer contact data.
type OrderConfirmed struct {
	OrderID     string          `json:"order_id"`
	OrderNumber string          `json:"order_number"`
	Version     int64           `json:"version"`
	TotalAmount decimal.Decimal `json:"total_amount"`
	ConfirmedBy string          `json:"confirmed_by"`
}

// ConversationEventsTopic carries messaging.conversation.changed snapshots that feed the conversation read model.
// Consumers must ignore unknown event types and newer schema versions.
const ConversationEventsTopic = "hexta.messaging.conversations.v1"

// ConversationChanged is the messaging.conversation.changed payload: the full thread and its open turn after a
// committed change, so consumers can upsert idempotently. It carries no message bodies or customer contact data.
type ConversationChanged struct {
	ThreadID          string    `json:"thread_id"`
	ExternalKey       string    `json:"external_key"`
	DisplayName       string    `json:"display_name"`
	ConnectionID      string    `json:"connection_id,omitempty"`
	RecipientID       string    `json:"recipient_id,omitempty"`
	ThreadCreatedAt   time.Time `json:"thread_created_at"`
	ThreadUpdatedAt   time.Time `json:"thread_updated_at"`
	TurnID            string    `json:"turn_id"`
	Lifecycle         string    `json:"lifecycle"`
	Control           string    `json:"control"`
	AssigneeID        *string   `json:"assignee_id"`
	NeedsHuman        bool      `json:"needs_human"`
	ContextRevision   int64     `json:"context_revision"`
	ControlGeneration int64     `json:"control_generation"`
	TurnVersion       int64     `json:"turn_version"`
	LastActivityAt    time.Time `json:"last_activity_at"`
}
