// Package outbound owns the additive v1 API-to-connector delivery contracts.
// API authorizes and records intents; the connector executes exactly one provider attempt per attempt_id.
package outbound

import (
	"errors"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
)

const (
	IntentTopic = "hexta.outbound.intent.v1"
	ResultTopic = "hexta.outbound.result.v1"
	HealthTopic = "hexta.connector.health.v1"
)

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

// Platforms that accept outbound intents.
var Platforms = map[string]bool{"facebook": true, "telegram": true}

const MaxTextRunes = 2000

type IntentEvent struct {
	SchemaVersion     int       `json:"schema_version"`
	IntentID          string    `json:"intent_id"`
	AttemptID         string    `json:"attempt_id"`
	CorrelationID     string    `json:"correlation_id"`
	TenantID          string    `json:"tenant_id"`
	ConnectionID      string    `json:"connection_id"`
	ConnectionVersion int64     `json:"connection_version"`
	Platform          string    `json:"platform"`
	AccountID         string    `json:"account_id"`
	RecipientID       string    `json:"recipient_id"`
	Text              string    `json:"text"`
	IssuedAt          time.Time `json:"issued_at"`
	ExpiresAt         time.Time `json:"expires_at"`
}

type ResultEvent struct {
	SchemaVersion     int       `json:"schema_version"`
	IntentID          string    `json:"intent_id"`
	AttemptID         string    `json:"attempt_id"`
	TenantID          string    `json:"tenant_id"`
	ConnectionID      string    `json:"connection_id"`
	Status            string    `json:"status"`
	ErrorCategory     string    `json:"error_category,omitempty"`
	ErrorCode         string    `json:"error_code,omitempty"`
	ProviderMessageID string    `json:"provider_message_id,omitempty"`
	OccurredAt        time.Time `json:"occurred_at"`
}

type ConnectionHealthEvent struct {
	SchemaVersion     int       `json:"schema_version"`
	EventID           string    `json:"event_id"`
	TenantID          string    `json:"tenant_id"`
	ConnectionID      string    `json:"connection_id"`
	ConnectionVersion int64     `json:"connection_version"`
	Status            string    `json:"status"`
	ErrorCode         string    `json:"error_code"`
	OccurredAt        time.Time `json:"occurred_at"`
}

func validUUID(values ...string) bool {
	for _, v := range values {
		if _, err := uuid.Parse(v); err != nil {
			return false
		}
	}
	return true
}

// Key orders deliveries per provider conversation.
func (e IntentEvent) Key() string { return e.ConnectionID + ":" + e.RecipientID }

func (e IntentEvent) Validate() error {
	if e.SchemaVersion != 1 || !validUUID(e.IntentID, e.AttemptID, e.CorrelationID, e.TenantID, e.ConnectionID) {
		return errors.New("invalid outbound intent identity")
	}
	if e.ConnectionVersion < 1 || !Platforms[e.Platform] || e.AccountID == "" || e.RecipientID == "" || len(e.RecipientID) > 255 {
		return errors.New("invalid outbound intent target")
	}
	if e.Text == "" || !utf8.ValidString(e.Text) || utf8.RuneCountInString(e.Text) > MaxTextRunes {
		return errors.New("invalid outbound intent text")
	}
	if e.IssuedAt.IsZero() || !e.ExpiresAt.After(e.IssuedAt) {
		return errors.New("invalid outbound intent lifetime")
	}
	return nil
}

func (e ResultEvent) Validate() error {
	if e.SchemaVersion != 1 || !validUUID(e.IntentID, e.AttemptID, e.TenantID, e.ConnectionID) || e.OccurredAt.IsZero() {
		return errors.New("invalid outbound result")
	}
	switch e.Status {
	case StatusSent:
		if e.ErrorCategory != "" {
			return errors.New("invalid outbound result")
		}
	case StatusFailed, StatusUnknown:
		if e.ErrorCategory == "" {
			return errors.New("invalid outbound result")
		}
	default:
		return errors.New("invalid outbound result")
	}
	if len(e.ErrorCode) > 64 || len(e.ProviderMessageID) > 255 {
		return errors.New("invalid outbound result")
	}
	return nil
}

func (e ConnectionHealthEvent) Validate() error {
	if e.SchemaVersion != 1 || !validUUID(e.EventID, e.TenantID, e.ConnectionID) || e.ConnectionVersion < 1 || e.Status != ErrorAuthExpired || len(e.ErrorCode) > 64 || e.OccurredAt.IsZero() {
		return errors.New("invalid connection health event")
	}
	return nil
}
