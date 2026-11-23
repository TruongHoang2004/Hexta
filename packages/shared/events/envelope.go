package events

import (
	"encoding/json"
	"time"
)

// EnvelopeSchemaVersion is the only envelope version producers write; consumers ignore newer versions.
const EnvelopeSchemaVersion = 1

// Envelope wraps events published from the API outbox on the order and conversation topics. Data holds the event
// payload named by EventType; each enveloped topic still carries a single event type.
type Envelope struct {
	SchemaVersion int             `json:"schema_version"`
	EventID       string          `json:"event_id"`
	EventType     string          `json:"event_type"`
	TenantID      string          `json:"tenant_id"`
	AggregateID   string          `json:"aggregate_id"`
	OccurredAt    time.Time       `json:"occurred_at"`
	Data          json.RawMessage `json:"data"`
}
