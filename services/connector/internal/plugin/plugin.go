package plugin

import (
	"context"
	"errors"
	"github.com/TruongHoang2004/Hexta/packages/shared/ingress"
	"net/url"
	"time"
)

var ErrUnsupported = errors.New("unsupported capability")

type Capabilities struct{ Webhook, Fetch, Settings, Outbound bool }
type Item struct {
	AccountID string
	Event     ingress.Event
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

// Outbound requires a current, single-attempt API permit. No sender is wired in this slice.
type SendRequest struct{ IntentID, AttemptID, Permit, Text string }
type Sender interface {
	Send(context.Context, SendRequest) error
}
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
