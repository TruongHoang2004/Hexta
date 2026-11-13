// Package ingress owns the additive v1 connector-to-business wire contract.
package ingress

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"github.com/google/uuid"
	"time"
)

const Topic = "hexta.ingress.received.v1"

type Event struct {
	SchemaVersion     int          `json:"schema_version"`
	EventID           string       `json:"event_id"`
	CorrelationID     string       `json:"correlation_id"`
	Platform          string       `json:"platform"`
	ConnectionID      string       `json:"connection_id"`
	ConnectionVersion int64        `json:"connection_version"`
	TenantID          string       `json:"tenant_id"`
	AccountID         string       `json:"external_account_id"`
	ConversationKey   string       `json:"external_conversation_key"`
	MessageID         string       `json:"external_message_id,omitempty"`
	SourceIdentity    string       `json:"source_identity"`
	SourceRevision    string       `json:"source_revision,omitempty"`
	Sender            string       `json:"sender"`
	Recipient         string       `json:"recipient"`
	Direction         string       `json:"direction"`
	Type              string       `json:"event_type"`
	Text              string       `json:"text,omitempty"`
	ReplyTo           string       `json:"reply_to,omitempty"`
	Attachments       []Attachment `json:"attachments,omitempty"`
	SourceAt          time.Time    `json:"source_timestamp"`
	ReceivedAt        time.Time    `json:"received_timestamp"`
	Acquisition       string       `json:"acquisition"`
	Purpose           string       `json:"purpose"`
}
type Attachment struct {
	Type      string `json:"type"`
	Reference string `json:"reference,omitempty"`
}

func Hash(parts ...string) string {
	b, _ := json.Marshal(parts)
	h := sha256.Sum256(b)
	return hex.EncodeToString(h[:])
}
func (e Event) Key() string      { return Hash(e.TenantID, e.ConnectionID, e.ConversationKey) }
func (e Event) Identity() string { return Hash(e.ConnectionID, e.SourceIdentity, e.SourceRevision) }
func (e Event) Validate() error {
	if _, err := uuid.Parse(e.EventID); err != nil {
		return errors.New("invalid event id")
	}
	if _, err := uuid.Parse(e.ConnectionID); err != nil {
		return errors.New("invalid connection id")
	}
	if e.Type == "" || e.Sender == "" || e.Recipient == "" {
		return errors.New("invalid event classification")
	}
	if e.SchemaVersion != 1 || e.EventID == "" || e.CorrelationID == "" || e.Platform == "" || e.ConnectionID == "" || e.ConnectionVersion < 1 || e.TenantID == "" || e.AccountID == "" || e.ConversationKey == "" || e.SourceIdentity == "" || e.SourceAt.IsZero() || e.ReceivedAt.IsZero() {
		return errors.New("invalid ingress envelope")
	}
	if e.Acquisition != "webhook" && e.Acquisition != "poll" {
		return errors.New("invalid acquisition")
	}
	if e.Purpose != "live" && e.Purpose != "history" && e.Purpose != "reconciliation" {
		return errors.New("invalid purpose")
	}
	if e.Direction != "inbound" && e.Direction != "outbound" {
		return errors.New("invalid direction")
	}
	return nil
}
