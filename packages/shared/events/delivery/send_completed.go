package delivery

import (
	"errors"

	"github.com/TruongHoang2004/Hexta/packages/shared/events"
)

// SendCompletedTopic carries SendCompleted. Producer: connector (outbox). Consumer: API delivery status.
const SendCompletedTopic = "hexta.delivery.send-completed.v1"

// Delivery outcomes. Unknown means the provider may have accepted the message; it is never resent automatically.
const (
	StatusSent    = "sent"
	StatusFailed  = "failed"
	StatusUnknown = "unknown"
)

// Error categories. Consumers must treat an unrecognized category as a generic failure.
const (
	ErrorWindow24h          = "window_24h"
	ErrorAuthExpired        = "auth_expired"
	ErrorRateLimit          = "rate_limit"
	ErrorInvalidRecipient   = "invalid_recipient"
	ErrorTransport          = "transport"
	ErrorExpired            = "expired"
	ErrorConnectionInactive = "connection_inactive"
	// Telegram: the user blocked the bot or deactivated their account (403), or the chat no longer exists (400).
	ErrorBotBlocked   = "bot_blocked"
	ErrorChatNotFound = "chat_not_found"
)

// SendCompleted is the outcome of one SendRequested attempt.
type SendCompleted struct {
	events.Meta
	IntentID          string `json:"intent_id"`
	AttemptID         string `json:"attempt_id"`
	ConnectionID      string `json:"connection_id"`
	Status            string `json:"status"`
	ErrorCategory     string `json:"error_category,omitempty"`
	ErrorCode         string `json:"error_code,omitempty"`
	ProviderMessageID string `json:"provider_message_id,omitempty"`
}

// Key orders outcomes per intent.
func (e SendCompleted) Key() string { return e.IntentID }

func (e SendCompleted) Validate() error {
	if err := e.Meta.Validate(); err != nil {
		return err
	}
	if !events.ValidUUID(e.IntentID, e.AttemptID, e.TenantID, e.ConnectionID) {
		return errors.New("invalid send outcome")
	}
	switch e.Status {
	case StatusSent:
		if e.ErrorCategory != "" {
			return errors.New("invalid send outcome")
		}
	case StatusFailed, StatusUnknown:
		if e.ErrorCategory == "" {
			return errors.New("invalid send outcome")
		}
	default:
		return errors.New("invalid send outcome")
	}
	if len(e.ErrorCode) > 64 || len(e.ProviderMessageID) > 255 {
		return errors.New("invalid send outcome")
	}
	return nil
}
