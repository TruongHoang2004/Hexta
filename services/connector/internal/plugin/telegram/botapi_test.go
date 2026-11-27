package telegram

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/TruongHoang2004/Hexta/packages/shared/events/delivery"
	"github.com/TruongHoang2004/Hexta/services/connector/internal/plugin"
)

const fixtureToken = "123456:AAFixtureSecretPart_abcdefghijklmnopq"

func TestTokenShape(t *testing.T) {
	if !ValidToken(fixtureToken) || TokenBotID(fixtureToken) != "123456" {
		t.Fatal("valid token")
	}
	for _, bad := range []string{"", "123456", "abc:AAFixtureSecretPart_abcdefghijklmnopq", "123456:short", "123456:AAFixture Secret_abcdefghijklmnopqrstu"} {
		if ValidToken(bad) {
			t.Fatal("accepted malformed token", bad)
		}
	}
}

func TestBotAPICallsAndClassification(t *testing.T) {
	var lastBody map[string]any
	status, reply := 200, `{"ok":true,"result":true}`
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.URL.Path, "/bot"+fixtureToken+"/") {
			w.WriteHeader(404)
			io.WriteString(w, `{"ok":false,"error_code":404,"description":"Not Found"}`)
			return
		}
		data, _ := io.ReadAll(r.Body)
		lastBody = map[string]any{}
		_ = json.Unmarshal(data, &lastBody)
		switch strings.TrimPrefix(r.URL.Path, "/bot"+fixtureToken+"/") {
		case "getMe":
			io.WriteString(w, `{"ok":true,"result":{"id":123456,"is_bot":true,"first_name":"Shop Bot","username":"shop_bot"}}`)
		default:
			w.WriteHeader(status)
			io.WriteString(w, reply)
		}
	}))
	defer server.Close()
	api := NewBotAPI(server.URL)
	ctx := context.Background()
	bot, err := api.GetMe(ctx, fixtureToken)
	if err != nil || bot.ID != "123456" || bot.DisplayName != "Shop Bot" || bot.Username != "shop_bot" {
		t.Fatal("getMe", err, bot)
	}
	if _, err := api.GetMe(ctx, "999:AAFixtureSecretPart_abcdefghijklmnopq"); !errors.Is(err, ErrInvalidToken) {
		t.Fatal("rejected token must be invalid", err)
	}
	if err := api.SetWebhook(ctx, fixtureToken, "https://hooks.example/webhooks/telegram/id", "secret"); err != nil || lastBody["secret_token"] != "secret" || lastBody["url"] != "https://hooks.example/webhooks/telegram/id" {
		t.Fatal("setWebhook", err, lastBody)
	}
	if updates, _ := lastBody["allowed_updates"].([]any); len(updates) != 1 || updates[0] != "message" {
		t.Fatal("allowed_updates", lastBody)
	}
	if err := api.DeleteWebhook(ctx, fixtureToken); err != nil || lastBody["drop_pending_updates"] != true {
		t.Fatal("deleteWebhook", err, lastBody)
	}
	reply = `{"ok":true,"result":{"message_id":77}}`
	id, err := api.SendText(ctx, "123456", fixtureToken, "42", "hello")
	if err != nil || id != "77" || lastBody["chat_id"] != "42" || lastBody["text"] != "hello" {
		t.Fatal("sendMessage", err, id, lastBody)
	}
	for _, c := range []struct {
		status   int
		desc     string
		category string
		unknown  bool
	}{
		{403, "Forbidden: bot was blocked by the user", delivery.ErrorBotBlocked, false},
		{403, "Forbidden: user is deactivated", delivery.ErrorBotBlocked, false},
		{400, "Bad Request: chat not found", delivery.ErrorChatNotFound, false},
		{401, "Unauthorized", delivery.ErrorAuthExpired, false},
		{429, "Too Many Requests: retry after 5", delivery.ErrorRateLimit, false},
		{400, "Bad Request: message is too long", delivery.ErrorTransport, false},
		{502, "Bad Gateway", delivery.ErrorTransport, true},
	} {
		status, reply = c.status, `{"ok":false,"error_code":0,"description":"`+c.desc+`"}`
		_, err := api.SendText(ctx, "123456", fixtureToken, "42", "hello")
		var pe *plugin.ProviderError
		if !errors.As(err, &pe) || pe.Category != c.category || pe.Unknown != c.unknown || strings.Contains(err.Error(), fixtureToken) {
			t.Fatal("classification", c, err)
		}
	}
	if _, err := api.SendText(ctx, "123456", "bad", "42", "hello"); !isCategory(err, delivery.ErrorAuthExpired) {
		t.Fatal("malformed stored token must surface as auth_expired", err)
	}
}

func isCategory(err error, category string) bool {
	var pe *plugin.ProviderError
	return errors.As(err, &pe) && pe.Category == category
}

func TestTimeoutIsUnknownAndHidesToken(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
	}))
	defer server.Close()
	api := NewBotAPI(server.URL)
	api.Client.Timeout = 20 * time.Millisecond
	_, err := api.SendText(context.Background(), "123456", fixtureToken, "42", "hello")
	var pe *plugin.ProviderError
	if !errors.As(err, &pe) || !pe.Unknown || pe.Category != delivery.ErrorTransport || strings.Contains(err.Error(), fixtureToken) {
		t.Fatal("timeout must be unknown and sanitized", err)
	}
}
