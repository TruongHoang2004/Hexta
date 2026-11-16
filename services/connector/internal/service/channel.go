package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/TruongHoang2004/Hexta/services/connector/internal/plugin"
	"github.com/google/uuid"
)

// OAuthSessionTTL bounds how long exchanged Page tokens stay staged before an owner must re-consent.
const OAuthSessionTTL = 10 * time.Minute

var (
	ErrConflict          = errors.New("configuration conflict")
	ErrCredential        = errors.New("credential unavailable")
	ErrProvider          = errors.New("provider unavailable")
	ErrSessionNotFound   = errors.New("oauth session not found")
	ErrConnectionMissing = errors.New("connection not found")
	ErrUnavailable       = errors.New("platform not configured")
)

// Sealed is an encrypted credential and the key version that sealed it.
type Sealed struct {
	Ciphertext []byte
	KeyVersion int
}

// CredentialChange describes the credential effect of a connection upsert: keep, replace or wipe.
// WebhookSecret is sealed with Set's key version and is replaced or wiped together with it.
type CredentialChange struct {
	Set           *Sealed
	WebhookSecret []byte
	Wipe          bool
}

type Sealer interface {
	Seal(plain, associated []byte) ([]byte, error)
	Open(sealed []byte, version int, associated []byte) ([]byte, error)
}

type ChannelStore interface {
	SaveOAuthSession(ctx context.Context, id, app string, sealed Sealed, expires, now time.Time) error
	OAuthSession(ctx context.Context, id, app string, now time.Time) (Sealed, error)
	PutConnection(ctx context.Context, c Connection, credential CredentialChange) error
	ConnectionCredential(ctx context.Context, id string) (Connection, *Sealed, error)
	// ConnectionSecret returns the sealed webhook secret, or nil when none is stored.
	ConnectionSecret(ctx context.Context, id string) (Connection, *Sealed, error)
}

type ExchangedPage struct {
	PageID   string `json:"page_id"`
	PageName string `json:"page_name"`
}
type OAuthSession struct {
	OAuthSessionID string          `json:"oauth_session_id"`
	ExpiresAt      time.Time       `json:"expires_at"`
	Pages          []ExchangedPage `json:"pages"`
}
type UpsertResult struct {
	ID         string `json:"id"`
	Version    int64  `json:"version"`
	Enabled    bool   `json:"enabled"`
	Subscribed bool   `json:"subscribed"`
}

// ChannelService owns provider credentials. It never decides tenant authority: the API sends trusted,
// already-authorized connection configuration through the machine endpoint.
type ChannelService struct {
	Store    ChannelStore
	Provider plugin.Provisioner
	Sealer   Sealer
	KeyVer   int
	AppID    string
	Now      func() time.Time
	// Bots and WebhookBase enable Telegram; WebhookBase is the public HTTPS origin Telegram calls back.
	Bots        plugin.BotProvisioner
	WebhookBase string
}

func (s *ChannelService) now() time.Time {
	if s.Now != nil {
		return s.Now().UTC()
	}
	return time.Now().UTC()
}

func sessionAD(id string) []byte    { return []byte("oauth_session:" + id) }
func connectionAD(id string) []byte { return []byte("connection:" + id) }
func webhookAD(id string) []byte    { return []byte("webhook_secret:" + id) }

// Exchange stages Page tokens encrypted for a short TTL and returns only Page identities.
func (s *ChannelService) Exchange(ctx context.Context, code, redirectURI string) (OAuthSession, error) {
	pages, err := s.Provider.ExchangeCode(ctx, code, redirectURI)
	if err != nil {
		return OAuthSession{}, ErrProvider
	}
	plain, err := json.Marshal(pages)
	if err != nil {
		return OAuthSession{}, ErrCredential
	}
	id := uuid.NewString()
	sealed, err := s.Sealer.Seal(plain, sessionAD(id))
	if err != nil {
		return OAuthSession{}, ErrCredential
	}
	now := s.now()
	result := OAuthSession{OAuthSessionID: id, ExpiresAt: now.Add(OAuthSessionTTL), Pages: make([]ExchangedPage, len(pages))}
	if err := s.Store.SaveOAuthSession(ctx, id, s.AppID, Sealed{sealed, s.KeyVer}, result.ExpiresAt, now); err != nil {
		return OAuthSession{}, err
	}
	for i, p := range pages {
		result.Pages[i] = ExchangedPage{PageID: p.ID, PageName: p.Name}
	}
	return result, nil
}

func (s *ChannelService) sessionToken(ctx context.Context, session, account string) (string, error) {
	sealed, err := s.Store.OAuthSession(ctx, session, s.AppID, s.now())
	if err != nil {
		return "", err
	}
	plain, err := s.Sealer.Open(sealed.Ciphertext, sealed.KeyVersion, sessionAD(session))
	if err != nil {
		return "", ErrCredential
	}
	var pages []plugin.PageCredential
	if json.Unmarshal(plain, &pages) != nil {
		return "", ErrCredential
	}
	for _, p := range pages {
		if p.ID == account && p.Token != "" {
			return p.Token, nil
		}
	}
	return "", ErrCredential
}

// Upsert applies a versioned connection change. Provider calls happen before or after, never inside, a DB transaction.
func (s *ChannelService) Upsert(ctx context.Context, c Connection) (UpsertResult, error) {
	if c.Platform == "telegram" {
		return s.upsertBot(ctx, c)
	}
	change := CredentialChange{}
	subscribed := false
	if c.OAuthSessionID != "" {
		token, err := s.sessionToken(ctx, c.OAuthSessionID, c.AccountID)
		if err != nil {
			return UpsertResult{}, err
		}
		if c.Enabled {
			if err := s.Provider.Subscribe(ctx, c.AccountID, token); err != nil {
				return UpsertResult{}, ErrProvider
			}
			subscribed = true
		}
		sealed, err := s.Sealer.Seal([]byte(token), connectionAD(c.ID))
		if err != nil {
			return UpsertResult{}, ErrCredential
		}
		change.Set = &Sealed{sealed, s.KeyVer}
	} else if !c.Enabled && c.UnsubscribeOnDisable {
		change.Wipe = true
		if _, current, err := s.Store.ConnectionCredential(ctx, c.ID); err == nil && current != nil {
			// Best effort: a revoked token cannot unsubscribe, and the disabled binding quarantines webhooks anyway.
			if token, err := s.Sealer.Open(current.Ciphertext, current.KeyVersion, connectionAD(c.ID)); err == nil {
				subscribed = s.Provider.Unsubscribe(ctx, c.AccountID, string(token)) != nil
			}
		}
	} else if c.Enabled {
		if _, current, err := s.Store.ConnectionCredential(ctx, c.ID); err == nil {
			subscribed = current != nil
		}
	}
	if err := s.Store.PutConnection(ctx, c, change); err != nil {
		return UpsertResult{}, err
	}
	return UpsertResult{ID: c.ID, Version: c.Version, Enabled: c.Enabled, Subscribed: subscribed}, nil
}

func (s *ChannelService) botsConfigured() bool { return s.Bots != nil && s.WebhookBase != "" }

// VerifyBot checks a bot token with the provider and returns only the bot identity; the token is not stored.
func (s *ChannelService) VerifyBot(ctx context.Context, token string) (plugin.BotIdentity, error) {
	if !s.botsConfigured() {
		return plugin.BotIdentity{}, ErrUnavailable
	}
	bot, err := s.Bots.GetMe(ctx, token)
	if errors.Is(err, plugin.ErrInvalidCredential) {
		return plugin.BotIdentity{}, ErrCredential
	}
	if err != nil {
		return plugin.BotIdentity{}, ErrProvider
	}
	return bot, nil
}

// WebhookSecret returns a connection with its decrypted webhook secret for inbound verification.
func (s *ChannelService) WebhookSecret(ctx context.Context, id string) (Connection, []byte, error) {
	c, sealed, err := s.Store.ConnectionSecret(ctx, id)
	if err != nil {
		return c, nil, err
	}
	if sealed == nil {
		return c, nil, ErrCredential
	}
	secret, err := s.Sealer.Open(sealed.Ciphertext, sealed.KeyVersion, webhookAD(id))
	if err != nil {
		return c, nil, ErrCredential
	}
	return c, secret, nil
}

// storedBot decrypts the stored token and webhook secret; empty values mean none are usable.
func (s *ChannelService) storedBot(ctx context.Context, id string) (string, string) {
	_, token, err := s.Store.ConnectionCredential(ctx, id)
	if err != nil || token == nil {
		return "", ""
	}
	plainToken, err := s.Sealer.Open(token.Ciphertext, token.KeyVersion, connectionAD(id))
	if err != nil {
		return "", ""
	}
	_, secret, err := s.WebhookSecret(ctx, id)
	if err != nil {
		return string(plainToken), ""
	}
	return string(plainToken), string(secret)
}

func (s *ChannelService) callback(id string) string {
	return strings.TrimRight(s.WebhookBase, "/") + "/webhooks/telegram/" + id
}

// upsertBot applies a Telegram connection change. A new token is verified against the bound bot ID and sealed with
// a fresh webhook secret; both are persisted before setWebhook so Telegram's first callback is verifiable.
// Pause/disconnect delete the webhook; disconnect also wipes the token and secret.
func (s *ChannelService) upsertBot(ctx context.Context, c Connection) (UpsertResult, error) {
	if !s.botsConfigured() {
		return UpsertResult{}, ErrUnavailable
	}
	change := CredentialChange{Wipe: !c.Enabled && c.UnsubscribeOnDisable}
	token, secret := c.BotToken, ""
	if token != "" {
		bot, err := s.VerifyBot(ctx, token)
		if err != nil {
			return UpsertResult{}, err
		}
		if bot.ID != c.AccountID {
			return UpsertResult{}, ErrConflict
		}
		raw := make([]byte, 32)
		if _, err := rand.Read(raw); err != nil {
			return UpsertResult{}, ErrCredential
		}
		secret = base64.RawURLEncoding.EncodeToString(raw)
		sealedToken, err := s.Sealer.Seal([]byte(token), connectionAD(c.ID))
		if err != nil {
			return UpsertResult{}, ErrCredential
		}
		sealedSecret, err := s.Sealer.Seal([]byte(secret), webhookAD(c.ID))
		if err != nil {
			return UpsertResult{}, ErrCredential
		}
		change.Set, change.WebhookSecret = &Sealed{sealedToken, s.KeyVer}, sealedSecret
	} else {
		token, secret = s.storedBot(ctx, c.ID)
	}
	if c.Enabled && (token == "" || secret == "") {
		return UpsertResult{}, ErrCredential
	}
	if err := s.Store.PutConnection(ctx, c, change); err != nil {
		return UpsertResult{}, err
	}
	subscribed := false
	switch {
	case c.Enabled:
		if err := s.Bots.SetWebhook(ctx, token, s.callback(c.ID), secret); err != nil {
			return UpsertResult{}, ErrProvider
		}
		subscribed = true
	case token != "":
		// Best effort: a revoked token cannot delete its webhook, and a disabled connection quarantines updates anyway.
		subscribed = s.Bots.DeleteWebhook(ctx, token) != nil
	}
	return UpsertResult{ID: c.ID, Version: c.Version, Enabled: c.Enabled, Subscribed: subscribed}, nil
}
