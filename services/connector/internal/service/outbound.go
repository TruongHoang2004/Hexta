package service

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/TruongHoang2004/Hexta/packages/shared/events"
	"time"

	"github.com/TruongHoang2004/Hexta/packages/shared/events/connection"
	"github.com/TruongHoang2004/Hexta/packages/shared/events/delivery"
	"github.com/TruongHoang2004/Hexta/services/connector/internal/plugin"
)

const sendTimeout = 15 * time.Second

// AttemptOutcome is a durable attempt result. Health is set when the provider reported revoked/expired credentials.
type AttemptOutcome struct {
	Result delivery.SendCompleted
	Health *connection.HealthChanged
}

// Dispatch is returned when an attempt was durably claimed and the provider call may proceed.
type Dispatch struct {
	Connection Connection
	Credential Sealed
}

type OutboundStore interface {
	// BeginAttempt claims attempt_id once. A nil Dispatch means nothing may be sent: the attempt was a
	// duplicate, or it was rejected (inactive connection, stale version, expired intent) with its result enqueued.
	BeginAttempt(ctx context.Context, intent delivery.SendRequested, now time.Time, reject func(Connection, *Sealed) *AttemptOutcome) (*Dispatch, error)
	FinishAttempt(ctx context.Context, outcome AttemptOutcome) error
}

type OutboundService struct {
	Store OutboundStore
	// Senders maps a platform to its provider; an intent for an unconfigured platform fails without a call.
	Senders map[string]plugin.TextSender
	Sealer  Sealer
	Now     func() time.Time
}

func (s *OutboundService) now() time.Time {
	if s.Now != nil {
		return s.Now().UTC()
	}
	return time.Now().UTC()
}

// ErrInvalidIntent marks a message that can never be processed; the consumer commits past it.
var ErrInvalidIntent = errors.New("invalid outbound intent")

func outcome(intent delivery.SendRequested, status, category, code, provider string, at time.Time) *AttemptOutcome {
	return &AttemptOutcome{Result: delivery.SendCompleted{Meta: events.NewMeta(intent.TenantID, intent.CorrelationID, at), IntentID: intent.IntentID, AttemptID: intent.AttemptID, ConnectionID: intent.ConnectionID, Status: status, ErrorCategory: category, ErrorCode: code, ProviderMessageID: provider}}
}

func authExpired(o *AttemptOutcome, c Connection, at time.Time) *AttemptOutcome {
	o.Health = &connection.HealthChanged{Meta: events.NewMeta(c.TenantID, "", at), ConnectionID: c.ID, ConnectionVersion: c.Version, Status: connection.StatusAuthExpired, ErrorCode: o.Result.ErrorCode}
	return o
}

// Handle executes at most one provider call per attempt. A returned error (other than ErrInvalidIntent) is a
// connector storage failure; the consumer retries the same message without advancing its offset.
func (s *OutboundService) Handle(ctx context.Context, payload []byte) error {
	var intent delivery.SendRequested
	if json.Unmarshal(payload, &intent) != nil || intent.Validate() != nil {
		return ErrInvalidIntent
	}
	now := s.now()
	dispatch, err := s.Store.BeginAttempt(ctx, intent, now, func(c Connection, credential *Sealed) *AttemptOutcome {
		switch {
		case !now.Before(intent.ExpiresAt):
			return outcome(intent, delivery.StatusFailed, delivery.ErrorExpired, "", "", now)
		case c.ID == "" || !c.Enabled || c.Version != intent.ConnectionVersion || c.TenantID != intent.TenantID || c.AccountID != intent.AccountID || c.Platform != intent.Platform:
			return outcome(intent, delivery.StatusFailed, delivery.ErrorConnectionInactive, "", "", now)
		case s.Senders[c.Platform] == nil:
			return outcome(intent, delivery.StatusFailed, delivery.ErrorConnectionInactive, "platform_unavailable", "", now)
		case credential == nil:
			return authExpired(outcome(intent, delivery.StatusFailed, delivery.ErrorAuthExpired, "no_credential", "", now), c, now)
		}
		return nil
	})
	if err != nil || dispatch == nil {
		return err
	}
	token, err := s.Sealer.Open(dispatch.Credential.Ciphertext, dispatch.Credential.KeyVersion, connectionAD(dispatch.Connection.ID))
	if err != nil {
		return s.Store.FinishAttempt(ctx, *outcome(intent, delivery.StatusFailed, delivery.ErrorTransport, "credential_unreadable", "", s.now()))
	}
	call, cancel := context.WithTimeout(ctx, sendTimeout)
	providerID, sendErr := s.Senders[dispatch.Connection.Platform].SendText(call, intent.AccountID, string(token), intent.RecipientID, intent.Text)
	cancel()
	at := s.now()
	result := outcome(intent, delivery.StatusSent, "", "", providerID, at)
	if sendErr != nil {
		var pe *plugin.ProviderError
		if !errors.As(sendErr, &pe) {
			pe = &plugin.ProviderError{Category: delivery.ErrorTransport, Code: "unclassified", Unknown: true}
		}
		status := delivery.StatusFailed
		if pe.Unknown {
			status = delivery.StatusUnknown
		}
		result = outcome(intent, status, pe.Category, pe.Code, "", at)
		if pe.Category == delivery.ErrorAuthExpired {
			result = authExpired(result, dispatch.Connection, at)
		}
	}
	// Use a fresh context so a shutdown after the provider call still records the outcome.
	finish, stop := context.WithTimeout(context.WithoutCancel(ctx), 10*time.Second)
	defer stop()
	return s.Store.FinishAttempt(finish, *result)
}
