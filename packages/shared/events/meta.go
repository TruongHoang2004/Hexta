// Package events owns Hexta's asynchronous contracts. Each event lives in one file of its domain package and owns
// exactly one Kafka topic; no topic carries two event types. Most payloads are flat JSON embedding Meta; the order
// and conversation topics keep their versioned Envelope.
package events

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
)

// Meta is embedded by every event. EventID is the consumer deduplication key; CorrelationID links an event to the
// request or event that caused it and is optional unless an event documents otherwise.
type Meta struct {
	EventID       string    `json:"event_id"`
	CorrelationID string    `json:"correlation_id,omitempty"`
	TenantID      string    `json:"tenant_id"`
	OccurredAt    time.Time `json:"occurred_at"`
}

// NewMeta returns metadata with a fresh event ID.
func NewMeta(tenantID, correlationID string, at time.Time) Meta {
	return Meta{EventID: uuid.NewString(), CorrelationID: correlationID, TenantID: tenantID, OccurredAt: at.UTC()}
}

// Validate checks the fields every event requires.
func (m Meta) Validate() error {
	if !ValidUUID(m.EventID) || m.TenantID == "" || len(m.TenantID) > 36 || m.OccurredAt.IsZero() {
		return errors.New("invalid event metadata")
	}
	if m.CorrelationID != "" && !ValidUUID(m.CorrelationID) {
		return errors.New("invalid event correlation id")
	}
	return nil
}

// ValidUUID reports whether every value is a UUID.
func ValidUUID(values ...string) bool {
	for _, v := range values {
		if _, err := uuid.Parse(v); err != nil {
			return false
		}
	}
	return true
}

// Hash is a stable SHA-256 over ordered parts, used for partition keys and idempotency identities.
func Hash(parts ...string) string {
	b, _ := json.Marshal(parts)
	h := sha256.Sum256(b)
	return hex.EncodeToString(h[:])
}
