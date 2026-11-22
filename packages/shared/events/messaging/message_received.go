// Package messaging holds conversation events.
package messaging

import (
	"errors"
	"time"
	"unicode/utf8"

	"github.com/TruongHoang2004/Hexta/packages/shared/events"
)

// MessageReceivedTopic carries MessageReceived. Producer: connector. Consumer: API ingress (inbox).
const MessageReceivedTopic = "hexta.ingress.received.v1"

// Directions. Outbound marks a provider echo of a message the account itself sent; it never opens a turn.
const (
	DirectionInbound  = "inbound"
	DirectionOutbound = "outbound"
)

const (
	MaxTextRunes        = 4000
	MaxUnsupportedMedia = 10
)

// MediaKinds are the attachment kinds a message may report. Media binaries and URLs are never fetched or stored.
var MediaKinds = map[string]bool{"image": true, "video": true, "audio": true, "file": true, "sticker": true}

// MessageReceived is one provider message observed on one connection. Messages are text; a media-only message
// carries no text and lists its media kinds so the inbox can show a placeholder.
type MessageReceived struct {
	events.Meta
	Platform          string `json:"platform"`
	ConnectionID      string `json:"connection_id"`
	ConnectionVersion int64  `json:"connection_version"`
	AccountID         string `json:"external_account_id"`
	ConversationKey   string `json:"external_conversation_key"`
	MessageID         string `json:"external_message_id,omitempty"`
	SourceIdentity    string `json:"source_identity"`
	SourceRevision    string `json:"source_revision,omitempty"`
	Sender            string `json:"sender"`
	Recipient         string `json:"recipient"`
	Direction         string `json:"direction"`
	Text              string `json:"text,omitempty"`
	ReplyTo           string `json:"reply_to,omitempty"`
	// UnsupportedMedia lists media kinds of a message without text; empty for a text message.
	UnsupportedMedia []string  `json:"unsupported_media_kinds,omitempty"`
	SourceAt         time.Time `json:"source_timestamp"`
	Acquisition      string    `json:"acquisition"`
	Purpose          string    `json:"purpose"`
}

// Key orders messages per conversation of one connection.
func (e MessageReceived) Key() string {
	return events.Hash(e.TenantID, e.ConnectionID, e.ConversationKey)
}

// Identity is the idempotency key of the provider message revision.
func (e MessageReceived) Identity() string {
	return events.Hash(e.ConnectionID, e.SourceIdentity, e.SourceRevision)
}

func (e MessageReceived) Validate() error {
	if err := e.Meta.Validate(); err != nil {
		return err
	}
	if !events.ValidUUID(e.ConnectionID) || e.ConnectionVersion < 1 || !events.Platforms[e.Platform] {
		return errors.New("invalid message connection")
	}
	if e.AccountID == "" || e.ConversationKey == "" || e.SourceIdentity == "" || e.Sender == "" || e.Recipient == "" || e.SourceAt.IsZero() {
		return errors.New("invalid message source")
	}
	if e.Direction != DirectionInbound && e.Direction != DirectionOutbound {
		return errors.New("invalid message direction")
	}
	if e.Acquisition != "webhook" && e.Acquisition != "poll" {
		return errors.New("invalid message acquisition")
	}
	if e.Purpose != "live" && e.Purpose != "history" && e.Purpose != "reconciliation" {
		return errors.New("invalid message purpose")
	}
	if len(e.UnsupportedMedia) > 0 {
		if e.Text != "" || len(e.UnsupportedMedia) > MaxUnsupportedMedia {
			return errors.New("invalid message media")
		}
		for _, kind := range e.UnsupportedMedia {
			if !MediaKinds[kind] {
				return errors.New("invalid message media")
			}
		}
		return nil
	}
	if e.Text == "" || !utf8.ValidString(e.Text) || utf8.RuneCountInString(e.Text) > MaxTextRunes {
		return errors.New("invalid message text")
	}
	return nil
}
