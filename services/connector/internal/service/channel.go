package service

import (
	"context"
	"encoding/json"
	"errors"
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
)

// Sealed is an encrypted credential and the key version that sealed it.
type Sealed struct {
	Ciphertext []byte
	KeyVersion int
}

// CredentialChange describes the credential effect of a connection upsert: keep, replace or wipe.
type CredentialChange struct {
	Set  *Sealed
	Wipe bool
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
}

func (s *ChannelService) now() time.Time {
	if s.Now != nil {
		return s.Now().UTC()
	}
	return time.Now().UTC()
}

func sessionAD(id string) []byte    { return []byte("oauth_session:" + id) }
func connectionAD(id string) []byte { return []byte("connection:" + id) }

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
