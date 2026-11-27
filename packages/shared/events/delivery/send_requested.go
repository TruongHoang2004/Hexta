// Package delivery holds outbound message delivery events between the API and the connector.
package delivery

import (
	"errors"
	"time"
	"unicode/utf8"

	"github.com/TruongHoang2004/Hexta/packages/shared/events"
)

// SendRequestedTopic carries SendRequested. Producer: API messaging command side (outbox). Consumer: connector.
const SendRequestedTopic = "hexta.outbound.intent.v1"

// SendRequestedEventType names SendRequested rows in the API outbox.
const SendRequestedEventType = "outbound.intent"

const MaxTextRunes = 2000

// SendRequested is a single-attempt permit to send one text message through one connection. CorrelationID is
// required and identifies the originating API message.
type SendRequested struct {
	events.Meta
	IntentID          string    `json:"intent_id"`
	AttemptID         string    `json:"attempt_id"`
	ConnectionID      string    `json:"connection_id"`
	ConnectionVersion int64     `json:"connection_version"`
	Platform          string    `json:"platform"`
	AccountID         string    `json:"account_id"`
	RecipientID       string    `json:"recipient_id"`
	Text              string    `json:"text"`
	ExpiresAt         time.Time `json:"expires_at"`
}

// Key orders deliveries per provider conversation.
func (e SendRequested) Key() string { return e.ConnectionID + ":" + e.RecipientID }

func (e SendRequested) Validate() error {
	if err := e.Meta.Validate(); err != nil {
		return err
	}
	if e.CorrelationID == "" || !events.ValidUUID(e.IntentID, e.AttemptID, e.TenantID, e.ConnectionID) {
		return errors.New("invalid send request identity")
	}
	if e.ConnectionVersion < 1 || !events.Platforms[e.Platform] || e.AccountID == "" || e.RecipientID == "" || len(e.RecipientID) > 255 {
		return errors.New("invalid send request target")
	}
	if e.Text == "" || !utf8.ValidString(e.Text) || utf8.RuneCountInString(e.Text) > MaxTextRunes {
		return errors.New("invalid send request text")
	}
	if !e.ExpiresAt.After(e.OccurredAt) {
		return errors.New("invalid send request lifetime")
	}
	return nil
}
