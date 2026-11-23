package repository

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/TruongHoang2004/Hexta/packages/shared/events/delivery"
	"github.com/TruongHoang2004/Hexta/services/connector/internal/plugin"
	"github.com/TruongHoang2004/Hexta/services/connector/internal/plugin/telegram"
	"github.com/TruongHoang2004/Hexta/services/connector/internal/secret"
	"github.com/TruongHoang2004/Hexta/services/connector/internal/service"
	"github.com/google/uuid"
)

// fakeBots stands in for the Telegram Bot API; tokens map to bot IDs.
type fakeBots struct {
	bots    map[string]string
	webhook map[string]string // token -> secret currently registered
	deleted []string
	sends   int
	sendErr error
}

func (f *fakeBots) GetMe(_ context.Context, token string) (plugin.BotIdentity, error) {
	id, ok := f.bots[token]
	if !ok {
		return plugin.BotIdentity{}, plugin.ErrInvalidCredential
	}
	return plugin.BotIdentity{ID: id, DisplayName: "Shop Bot", Username: "shop_bot"}, nil
}
func (f *fakeBots) SetWebhook(_ context.Context, token, callback, secret string) error {
	if !strings.Contains(callback, "/webhooks/telegram/") {
		return errors.New("callback")
	}
	f.webhook[token] = secret
	return nil
}
func (f *fakeBots) DeleteWebhook(_ context.Context, token string) error {
	delete(f.webhook, token)
	f.deleted = append(f.deleted, token)
	return nil
}
func (f *fakeBots) SendText(context.Context, string, string, string, string) (string, error) {
	f.sends++
	return "77", f.sendErr
}

func botToken(id, suffix string) string {
	return id + ":AAFixtureSecretPart_" + suffix + strings.Repeat("x", 20)
}

func TestTelegramConnectionLifecycle(t *testing.T) {
	s, _ := fixtureStore(t)
	ctx := context.Background()
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		t.Fatal(err)
	}
	sealer, err := secret.NewSealer(hex.EncodeToString(key), 1)
	if err != nil {
		t.Fatal(err)
	}
	bot := strconv.FormatInt(time.Now().UnixNano()%1_000_000_000_000, 10)
	first, rotated, other := botToken(bot, "a"), botToken(bot, "b"), botToken("1"+bot, "c")
	bots := &fakeBots{bots: map[string]string{first: bot, rotated: bot, other: "1" + bot}, webhook: map[string]string{}}
	channels := &service.ChannelService{Store: s, Sealer: sealer, KeyVer: 1, Bots: bots, WebhookBase: "https://hooks.example"}
	c := service.Connection{ID: uuid.NewString(), TenantID: uuid.NewString(), Version: 1, Platform: "telegram", AccountID: bot, AppID: telegram.AppID, Enabled: true, BotToken: first}
	if r, err := channels.Upsert(ctx, c); err != nil || !r.Subscribed {
		t.Fatal("connect", err)
	}
	var stored []byte
	if s.Pool.QueryRow(ctx, "SELECT credential_ciphertext FROM connector.connections WHERE id=$1", c.ID).Scan(&stored) != nil || bytes.Contains(stored, []byte(first)) {
		t.Fatal("bot token must be sealed at rest")
	}
	_, webhookSecret, err := channels.WebhookSecret(ctx, c.ID)
	if err != nil || string(webhookSecret) != bots.webhook[first] || len(webhookSecret) < 32 {
		t.Fatal("registered webhook secret must be the stored secret", err)
	}

	// Inbound: private text is accepted into the outbox; a group update is dropped.
	registry, _ := plugin.NewRegistry(telegram.TelegramPlugin{})
	inbound := &service.InboundService{Registry: registry, Store: s}
	conn, _, _ := channels.WebhookSecret(ctx, c.ID)
	private := []byte(`{"update_id":1,"message":{"message_id":3,"date":1700000000,"from":{"id":42},"chat":{"id":42,"type":"private"},"text":"hi"}}`)
	for i := 0; i < 2; i++ {
		if err := inbound.ReceiveConnection(ctx, conn, private); err != nil {
			t.Fatal("accept", err)
		}
	}
	if err := inbound.ReceiveConnection(ctx, conn, []byte(`{"update_id":2,"message":{"message_id":4,"date":1700000000,"chat":{"id":-1,"type":"supergroup"},"text":"x"}}`)); err != nil {
		t.Fatal("drop", err)
	}
	var pending int
	s.Pool.QueryRow(ctx, "SELECT count(*) FROM connector.events WHERE app_id=$1 AND state='pending' AND convert_from(payload,'UTF8') LIKE '%'||$2||'%'", telegram.AppID, c.ID).Scan(&pending)
	if pending != 1 {
		t.Fatal("private update must be stored exactly once", pending)
	}

	// Rotating to a different bot is rejected; the same bot gets a fresh secret.
	c.Version, c.BotToken = 2, other
	if _, err := channels.Upsert(ctx, c); !errors.Is(err, service.ErrConflict) {
		t.Fatal("bot mismatch must conflict", err)
	}
	c.Version, c.BotToken = 3, rotated
	if _, err := channels.Upsert(ctx, c); err != nil {
		t.Fatal("rotate", err)
	}
	_, rotatedSecret, _ := channels.WebhookSecret(ctx, c.ID)
	if bytes.Equal(rotatedSecret, webhookSecret) || bots.webhook[rotated] != string(rotatedSecret) {
		t.Fatal("rotation must re-register the webhook with a new secret")
	}

	// Outbound uses the rotated token through the shared single-attempt flow.
	sender := &service.OutboundService{Store: s, Senders: map[string]plugin.TextSender{"telegram": bots}, Sealer: sealer}
	for category, sendErr := range map[string]error{"": nil, delivery.ErrorBotBlocked: &plugin.ProviderError{Category: delivery.ErrorBotBlocked, Code: "403"}, delivery.ErrorChatNotFound: &plugin.ProviderError{Category: delivery.ErrorChatNotFound, Code: "400"}} {
		bots.sendErr = sendErr
		intent := intentFor(c)
		intent.Platform, intent.RecipientID = "telegram", "42"
		payload, _ := json.Marshal(intent)
		for i := 0; i < 2; i++ {
			if err := sender.Handle(ctx, payload); err != nil {
				t.Fatal(err)
			}
		}
		got := results(t, s, intent.IntentID)
		if len(got) != 1 || got[0].ErrorCategory != category {
			t.Fatal("telegram outcome", category, got)
		}
	}
	if bots.sends != 3 {
		t.Fatal("each attempt must reach Telegram exactly once", bots.sends)
	}

	// Pause deletes the webhook but keeps credentials; resume re-registers the same secret.
	c.Version, c.Enabled, c.BotToken = 4, false, ""
	if r, err := channels.Upsert(ctx, c); err != nil || r.Subscribed || bots.webhook[rotated] != "" {
		t.Fatal("pause", err)
	}
	c.Version, c.Enabled = 5, true
	if r, err := channels.Upsert(ctx, c); err != nil || !r.Subscribed || bots.webhook[rotated] != string(rotatedSecret) {
		t.Fatal("resume", err)
	}
	// Disconnect deletes the webhook and wipes token and secret.
	c.Version, c.Enabled, c.UnsubscribeOnDisable = 6, false, true
	if _, err := channels.Upsert(ctx, c); err != nil || bots.webhook[rotated] != "" {
		t.Fatal("disconnect", err)
	}
	if _, _, err := channels.WebhookSecret(ctx, c.ID); !errors.Is(err, service.ErrCredential) {
		t.Fatal("secret must be wiped", err)
	}
	if _, sealed, _ := s.ConnectionCredential(ctx, c.ID); sealed != nil {
		t.Fatal("token must be wiped")
	}
	c.Version, c.Enabled, c.UnsubscribeOnDisable = 7, true, false
	if _, err := channels.Upsert(ctx, c); !errors.Is(err, service.ErrCredential) {
		t.Fatal("resume without credentials", err)
	}
}
