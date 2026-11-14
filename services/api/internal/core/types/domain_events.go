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
