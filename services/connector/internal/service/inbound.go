package service

import (
	"context"
	"errors"
	"github.com/TruongHoang2004/Hexta/packages/shared/ingress"
	"github.com/TruongHoang2004/Hexta/services/connector/internal/plugin"
	"github.com/google/uuid"
	"time"
)

type Connection struct {
	ID            string `json:"id"`
	Version       int64  `json:"version"`
	TenantID      string `json:"tenant_id"`
	Platform      string `json:"platform"`
	AccountID     string `json:"account_id"`
	AppID         string `json:"app_id"`
	Enabled       bool   `json:"enabled"`
	CredentialRef string `json:"credential_ref"`
	// OAuthSessionID selects freshly exchanged credentials; UnsubscribeOnDisable wipes them (disconnect).
	OAuthSessionID       string `json:"oauth_session_id,omitempty"`
	UnsubscribeOnDisable bool   `json:"unsubscribe_on_disable,omitempty"`
	// BotToken is a raw Telegram Bot Token on connect/rotate only. It is sealed before storage and never returned.
	BotToken string `json:"bot_token,omitempty"`
}
type Accepted struct {
	Event  ingress.Event
	Reason string
}
type InboundStore interface {
	Accept(context.Context, string, string, string, []plugin.Item, time.Time) error
}
type InboundService struct {
	Registry *plugin.Registry
	Store    InboundStore
}

func (s *InboundService) Receive(ctx context.Context, app, platform string, body []byte) error {
	p, err := s.Registry.Webhook(platform)
	if err != nil {
		return err
	}
	now := time.Now().UTC()
	items, err := p.Parse(body, now)
	if err != nil {
		return err
	}
	if len(items) == 0 {
		return errors.New("empty batch")
	}
	return s.Store.Accept(ctx, app, platform, ingress.Hash(platform, string(body)), items, now)
}

// ErrInvalidPayload marks a verified callback whose body can never be parsed; retrying it cannot succeed.
var ErrInvalidPayload = errors.New("invalid webhook payload")

// ReceiveConnection accepts a callback addressed to one connection (Telegram's per-connection webhook URL). The
// update does not name the bot, so the connection supplies the account. Updates the plugin drops (non-message,
// non-private chats) produce no items and nothing is stored.
func (s *InboundService) ReceiveConnection(ctx context.Context, c Connection, body []byte) error {
	p, err := s.Registry.Webhook(c.Platform)
	if err != nil {
		return err
	}
	now := time.Now().UTC()
	items, err := p.Parse(body, now)
	if err != nil {
		return ErrInvalidPayload
	}
	if len(items) == 0 {
		return nil
	}
	for i := range items {
		items[i].AccountID = c.AccountID
		items[i].Event.AccountID = c.AccountID
		items[i].Event.Recipient = c.AccountID
	}
	return s.Store.Accept(ctx, c.AppID, c.Platform, ingress.Hash(c.Platform, c.ID, string(body)), items, now)
}
func Envelope(e ingress.Event, c Connection) ingress.Event {
	e.SchemaVersion = 1
	e.EventID = uuid.NewString()
	e.CorrelationID = e.EventID
	e.ConnectionID = c.ID
	e.ConnectionVersion = c.Version
	e.TenantID = c.TenantID
	return e
}
