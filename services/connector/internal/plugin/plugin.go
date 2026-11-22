package plugin

import (
	"context"
	"errors"
	"github.com/TruongHoang2004/Hexta/packages/shared/events/messaging"
	"net/url"
	"time"
)

var ErrUnsupported = errors.New("unsupported capability")

type Capabilities struct{ Webhook, Fetch, Settings, Outbound bool }
type Item struct {
	AccountID string
	Event     messaging.MessageReceived
	Reason    string
}
type Webhook interface {
	Verify([]byte, string, []byte) error
	Challenge(url.Values, string) (string, error)
	Parse([]byte, time.Time) ([]Item, error)
}
type FetchRequest struct {
	AccountID, Cursor, Mode string
	Limit                   int
}
type Page struct {
	Items      []Item
	NextCursor string
	RetryAfter time.Duration
}
type Fetcher interface {
	Fetch(context.Context, FetchRequest) (Page, error)
}
type Settings interface {
	Subscribe(context.Context, string) error
}

// Outbound requires a current, single-attempt API permit. The generic Sender capability stays unwired;
// Facebook text delivery uses TextSender through the durable outbound attempt flow.
type SendRequest struct{ IntentID, AttemptID, Permit, Text string }
type Sender interface {
	Send(context.Context, SendRequest) error
}

// PageCredential is a provider account and its access token, held only in memory or sealed at rest.
type PageCredential struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Token string `json:"access_token"`
}

// Provisioner exchanges owner consent and manages the provider webhook subscription.
type Provisioner interface {
	ExchangeCode(ctx context.Context, code, redirectURI string) ([]PageCredential, error)
	Subscribe(ctx context.Context, accountID, token string) error
	Unsubscribe(ctx context.Context, accountID, token string) error
}

// BotIdentity is a provider bot account verified with its own token.
type BotIdentity struct{ ID, DisplayName, Username string }

// ErrInvalidCredential means the provider rejected a credential or it is malformed.
var ErrInvalidCredential = errors.New("invalid credential")

// BotProvisioner verifies bot tokens and manages a per-connection, secret-authenticated webhook (Telegram).
type BotProvisioner interface {
	GetMe(ctx context.Context, token string) (BotIdentity, error)
	SetWebhook(ctx context.Context, token, callback, secret string) error
	DeleteWebhook(ctx context.Context, token string) error
}

// TextSender performs exactly one provider call; callers must not retry an Unknown ProviderError.
type TextSender interface {
	SendText(ctx context.Context, accountID, token, recipient, text string) (string, error)
}

// ProviderError is a sanitized provider failure. Unknown means the request may have reached the provider.
type ProviderError struct {
	Category string
	Code     string
	Unknown  bool
}

func (e *ProviderError) Error() string { return "provider request failed: " + e.Category }

type Plugin interface {
	Platform() string
	Capabilities() Capabilities
}
type BasePlugin struct{}

func (BasePlugin) Platform() string                                  { return "base" }
func (BasePlugin) Capabilities() Capabilities                        { return Capabilities{} }
func (BasePlugin) Verify([]byte, string, []byte) error               { return ErrUnsupported }
func (BasePlugin) Challenge(url.Values, string) (string, error)      { return "", ErrUnsupported }
func (BasePlugin) Parse([]byte, time.Time) ([]Item, error)           { return nil, ErrUnsupported }
func (BasePlugin) Fetch(context.Context, FetchRequest) (Page, error) { return Page{}, ErrUnsupported }
func (BasePlugin) Subscribe(context.Context, string) error           { return ErrUnsupported }
func (BasePlugin) Send(context.Context, SendRequest) error           { return ErrUnsupported }

type Registry struct{ plugins map[string]Plugin }

func NewRegistry(ps ...Plugin) (*Registry, error) {
	r := &Registry{map[string]Plugin{}}
	for _, p := range ps {
		if p == nil || p.Platform() == "" || r.plugins[p.Platform()] != nil {
			return nil, errors.New("invalid plugin registration")
		}
		c := p.Capabilities()
		if c.Webhook {
			if _, ok := p.(Webhook); !ok {
				return nil, ErrUnsupported
			}
		}
		if c.Fetch {
			if _, ok := p.(Fetcher); !ok {
				return nil, ErrUnsupported
			}
		}
		if c.Settings {
			if _, ok := p.(Settings); !ok {
				return nil, ErrUnsupported
			}
		}
		if c.Outbound {
			if _, ok := p.(Sender); !ok {
				return nil, ErrUnsupported
			}
		}
		r.plugins[p.Platform()] = p
	}
	return r, nil
}
func (r *Registry) Webhook(platform string) (Webhook, error) {
	p := r.plugins[platform]
	if p == nil || !p.Capabilities().Webhook {
		return nil, ErrUnsupported
	}
	return p.(Webhook), nil
}

var _ Webhook = BasePlugin{}
var _ Fetcher = BasePlugin{}
var _ Settings = BasePlugin{}
var _ Sender = BasePlugin{}

func (r *Registry) Fetcher(platform string) (Fetcher, error) {
	p := r.plugins[platform]
	if p == nil || !p.Capabilities().Fetch {
		return nil, ErrUnsupported
	}
	return p.(Fetcher), nil
}
