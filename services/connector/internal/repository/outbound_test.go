package repository

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/TruongHoang2004/Hexta/packages/shared/events"
	"github.com/TruongHoang2004/Hexta/packages/shared/events/connection"
	"github.com/TruongHoang2004/Hexta/packages/shared/events/delivery"
	"github.com/TruongHoang2004/Hexta/services/connector/internal/plugin"
	"github.com/TruongHoang2004/Hexta/services/connector/internal/secret"
	"github.com/TruongHoang2004/Hexta/services/connector/internal/service"
	"github.com/google/uuid"
)

type fakeProvider struct {
	pages        []plugin.PageCredential
	subscribed   []string
	unsubscribed []string
	sendErr      error
	sends        int
}

func (f *fakeProvider) ExchangeCode(context.Context, string, string) ([]plugin.PageCredential, error) {
	return f.pages, nil
}
func (f *fakeProvider) Subscribe(_ context.Context, page, _ string) error {
	f.subscribed = append(f.subscribed, page)
	return nil
}
func (f *fakeProvider) Unsubscribe(_ context.Context, page, _ string) error {
	f.unsubscribed = append(f.unsubscribed, page)
	return nil
}
func (f *fakeProvider) SendText(_ context.Context, _, token, _, _ string) (string, error) {
	f.sends++
	if token != "page-token" {
		return "", errors.New("wrong token")
	}
	return "mid.1", f.sendErr
}

func channelFixture(t *testing.T) (*Store, *service.ChannelService, *fakeProvider, service.Connection) {
	s, c := fixtureStore(t)
	sealer, err := secret.NewSealer(strings.Repeat("0f", 32), 1)
	if err != nil {
		t.Fatal(err)
	}
	provider := &fakeProvider{pages: []plugin.PageCredential{{ID: c.AccountID, Name: "Shop", Token: "page-token"}}}
	return s, &service.ChannelService{Store: s, Provider: provider, Sealer: sealer, KeyVer: 1, AppID: c.AppID}, provider, c
}

func TestCredentialLifecycleReplayAndPageHandover(t *testing.T) {
	s, channels, provider, c := channelFixture(t)
	ctx := context.Background()
	session, err := channels.Exchange(ctx, "code", "https://app.example/cb")
	if err != nil || len(session.Pages) != 1 || session.Pages[0].PageName != "Shop" {
		t.Fatal("exchange", err)
	}
	var raw []byte
	if s.Pool.QueryRow(ctx, "SELECT payload_ciphertext FROM connector.oauth_sessions WHERE id=$1", session.OAuthSessionID).Scan(&raw) != nil || strings.Contains(string(raw), "page-token") {
		t.Fatal("session token stored in plaintext")
	}
	c.Version++
	c.OAuthSessionID = session.OAuthSessionID
	result, err := channels.Upsert(ctx, c)
	if err != nil || !result.Subscribed || len(provider.subscribed) != 1 {
		t.Fatal("connect", err)
	}
	if _, err := channels.Upsert(ctx, c); err != nil {
		t.Fatal("same-version replay must be idempotent", err)
	}
	_, sealed, err := s.ConnectionCredential(ctx, c.ID)
	if err != nil || sealed == nil || strings.Contains(string(sealed.Ciphertext), "page-token") {
		t.Fatal("credential not sealed", err)
	}
	other := c
	other.ID = uuid.NewString()
	other.TenantID = uuid.NewString()
	other.Version = 1
	other.OAuthSessionID = ""
	if _, err := channels.Upsert(ctx, other); !errors.Is(err, service.ErrConflict) {
		t.Fatal("paused/active Page reassigned to another tenant", err)
	}
	c.Version++
	c.Enabled = false
	c.OAuthSessionID = ""
	c.UnsubscribeOnDisable = true
	if _, err := channels.Upsert(ctx, c); err != nil || len(provider.unsubscribed) != 1 {
		t.Fatal("disconnect", err)
	}
	if _, sealed, _ := s.ConnectionCredential(ctx, c.ID); sealed != nil {
		t.Fatal("credential survived disconnect")
	}
	if _, err := channels.Upsert(ctx, other); err != nil {
		t.Fatal("disconnected Page cannot be connected by another tenant", err)
	}
	expired := uuid.NewString()
	if err := s.SaveOAuthSession(ctx, expired, c.AppID, service.Sealed{Ciphertext: []byte("x"), KeyVersion: 1}, time.Now().Add(-time.Second), time.Now()); err != nil {
		t.Fatal(err)
	}
	if _, err := s.OAuthSession(ctx, expired, c.AppID, time.Now()); !errors.Is(err, service.ErrSessionNotFound) {
		t.Fatal("expired session usable")
	}
}

func intentFor(c service.Connection) delivery.SendRequested {
	now := time.Now().UTC()
	return delivery.SendRequested{Meta: events.Meta{EventID: uuid.NewString(), CorrelationID: uuid.NewString(), TenantID: c.TenantID, OccurredAt: now}, IntentID: uuid.NewString(), AttemptID: uuid.NewString(), ConnectionID: c.ID, ConnectionVersion: c.Version, Platform: "facebook", AccountID: c.AccountID, RecipientID: "psid", Text: "hello", ExpiresAt: now.Add(time.Minute)}
}

func results(t *testing.T, s *Store, intent string) []delivery.SendCompleted {
	t.Helper()
	rows, err := s.Pool.Query(context.Background(), "SELECT payload FROM connector.outbound_events WHERE event_key=$1 AND topic=$2 ORDER BY available_at", intent, delivery.SendCompletedTopic)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	var out []delivery.SendCompleted
	for rows.Next() {
		var payload []byte
		var r delivery.SendCompleted
		if rows.Scan(&payload) != nil || json.Unmarshal(payload, &r) != nil {
			t.Fatal("result decode")
		}
		out = append(out, r)
	}
	return out
}

func TestOutboundSingleAttemptAndClassification(t *testing.T) {
	s, channels, provider, c := channelFixture(t)
	ctx := context.Background()
	session, _ := channels.Exchange(ctx, "code", "https://app.example/cb")
	c.Version++
	c.OAuthSessionID = session.OAuthSessionID
	if _, err := channels.Upsert(ctx, c); err != nil {
		t.Fatal(err)
	}
	sender := &service.OutboundService{Store: s, Senders: map[string]plugin.TextSender{"facebook": provider}, Sealer: channels.Sealer}
	intent := intentFor(c)
	payload, _ := json.Marshal(intent)
	for i := 0; i < 2; i++ {
		if err := sender.Handle(ctx, payload); err != nil {
			t.Fatal(err)
		}
	}
	got := results(t, s, intent.IntentID)
	if provider.sends != 1 || len(got) != 1 || got[0].Status != delivery.StatusSent || got[0].ProviderMessageID != "mid.1" {
		t.Fatal("redelivery resent or missing result", provider.sends, got)
	}

	// A crash after the attempt claim leaves it dispatching; redelivery must report unknown, never resend.
	crashed := intentFor(c)
	if _, err := s.Pool.Exec(ctx, "INSERT INTO connector.outbound_attempts(attempt_id,intent_id,connection_id,status,created_at) VALUES($1,$2,$3,'dispatching',now())", crashed.AttemptID, crashed.IntentID, c.ID); err != nil {
		t.Fatal(err)
	}
	payload, _ = json.Marshal(crashed)
	if err := sender.Handle(ctx, payload); err != nil {
		t.Fatal(err)
	}
	if got := results(t, s, crashed.IntentID); provider.sends != 1 || len(got) != 1 || got[0].Status != delivery.StatusUnknown {
		t.Fatal("interrupted attempt", got)
	}

	provider.sendErr = &plugin.ProviderError{Category: delivery.ErrorAuthExpired, Code: "190"}
	revoked := intentFor(c)
	payload, _ = json.Marshal(revoked)
	if err := sender.Handle(ctx, payload); err != nil {
		t.Fatal(err)
	}
	if got := results(t, s, revoked.IntentID); len(got) != 1 || got[0].Status != delivery.StatusFailed || got[0].ErrorCategory != delivery.ErrorAuthExpired {
		t.Fatal("auth classification", got)
	}
	var health int
	s.Pool.QueryRow(ctx, "SELECT count(*) FROM connector.outbound_events WHERE topic=$1 AND event_key=$2", connection.HealthChangedTopic, c.ID).Scan(&health)
	if health != 1 {
		t.Fatal("auth expiry health event missing", health)
	}

	provider.sendErr = nil
	stale := intentFor(c)
	stale.ConnectionVersion--
	expired := intentFor(c)
	expired.OccurredAt = expired.OccurredAt.Add(-time.Hour)
	expired.ExpiresAt = time.Now().Add(-time.Second)
	for intent, category := range map[*delivery.SendRequested]string{&stale: delivery.ErrorConnectionInactive, &expired: delivery.ErrorExpired} {
		payload, _ = json.Marshal(intent)
		if err := sender.Handle(ctx, payload); err != nil {
			t.Fatal(err)
		}
		if got := results(t, s, intent.IntentID); len(got) != 1 || got[0].ErrorCategory != category {
			t.Fatal("precheck", category, got)
		}
	}
	if provider.sends != 2 {
		t.Fatal("rejected intents reached the provider", provider.sends)
	}
	leased, err := s.ClaimOutbound(ctx)
	if err != nil || s.FinishOutbound(ctx, leased, true) != nil {
		t.Fatal("outbound relay lease", err)
	}
}
