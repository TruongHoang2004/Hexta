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
func Envelope(e ingress.Event, c Connection) ingress.Event {
	e.SchemaVersion = 1
	e.EventID = uuid.NewString()
	e.CorrelationID = e.EventID
	e.ConnectionID = c.ID
	e.ConnectionVersion = c.Version
	e.TenantID = c.TenantID
	return e
}
