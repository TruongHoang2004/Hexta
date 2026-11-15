package transport

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"github.com/TruongHoang2004/Hexta/services/connector/internal/plugin"
	"github.com/TruongHoang2004/Hexta/services/connector/internal/plugin/facebook"
	"github.com/TruongHoang2004/Hexta/services/connector/internal/service"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

type unavailable struct{}

func (unavailable) Accept(context.Context, string, string, string, []plugin.Item, time.Time) error {
	return errors.New("private storage failure must not leak")
}
func TestNoACKOnStorageFailure(t *testing.T) {
	p := facebook.FacebookPlugin{}
	registry, _ := plugin.NewRegistry(p)
	h := &Handler{Plugin: p, Secret: []byte("fixture"), Inbound: &service.InboundService{Registry: registry, Store: unavailable{}}}
	body := `{"object":"page","entry":[{"id":"p","messaging":[{}]}]}`
	mac := hmac.New(sha256.New, h.Secret)
	mac.Write([]byte(body))
	r := httptest.NewRequest("POST", "/webhooks/facebook", strings.NewReader(body))
	r.Header.Set("X-Hub-Signature-256", "sha256="+hex.EncodeToString(mac.Sum(nil)))
	w := httptest.NewRecorder()
	h.Routes().ServeHTTP(w, r)
	if w.Code != http.StatusServiceUnavailable || strings.Contains(w.Body.String(), "private") {
		t.Fatal("false ACK or leaked error", w.Code)
	}
	w = httptest.NewRecorder()
	h.Routes().ServeHTTP(w, httptest.NewRequest("POST", "/webhooks/facebook", strings.NewReader(body)))
	if w.Code != 403 {
		t.Fatal("signature")
	}
}

type channelsStub struct{ err error }

func (c channelsStub) Exchange(context.Context, string, string) (service.OAuthSession, error) {
	return service.OAuthSession{OAuthSessionID: "s", Pages: []service.ExchangedPage{{PageID: "p", PageName: "Shop"}}}, c.err
}
func (c channelsStub) Upsert(_ context.Context, conn service.Connection) (service.UpsertResult, error) {
	return service.UpsertResult{ID: conn.ID, Version: conn.Version, Enabled: conn.Enabled, Subscribed: true}, c.err
}

func TestManagementEndpointsRequireTokenAndHideTokens(t *testing.T) {
	h := &Handler{AppID: "app", ManagementToken: "management-token-123456", Channels: channelsStub{}}
	call := func(method, path, body, token string) *httptest.ResponseRecorder {
		r := httptest.NewRequest(method, path, strings.NewReader(body))
		if token != "" {
			r.Header.Set("Authorization", "Bearer "+token)
		}
		w := httptest.NewRecorder()
		h.Routes().ServeHTTP(w, r)
		return w
	}
	exchange := `{"code":"c","redirect_uri":"https://app.example/channels/facebook/callback"}`
	if call("POST", "/internal/facebook/oauth/exchange", exchange, "").Code != 401 || call("POST", "/internal/facebook/oauth/exchange", exchange, "wrong").Code != 401 {
		t.Fatal("exchange without management token")
	}
	w := call("POST", "/internal/facebook/oauth/exchange", exchange, h.ManagementToken)
	if w.Code != 200 || !strings.Contains(w.Body.String(), `"page_id":"p"`) || strings.Contains(w.Body.String(), "access_token") {
		t.Fatal("exchange response", w.Code, w.Body.String())
	}
	if call("POST", "/internal/facebook/oauth/exchange", `{"code":"c","redirect_uri":"javascript:x"}`, h.ManagementToken).Code != 400 {
		t.Fatal("redirect validation")
	}
	conn := `{"id":"6f1c2b52-8c86-4a39-9a51-1b7f6f8e1c11","version":2,"tenant_id":"t","platform":"facebook","account_id":"p","app_id":"app","enabled":true,"oauth_session_id":"s"}`
	if w := call("PUT", "/internal/connections", conn, h.ManagementToken); w.Code != 200 || !strings.Contains(w.Body.String(), `"subscribed":true`) {
		t.Fatal("upsert response", w.Code)
	}
	h.Channels = channelsStub{err: service.ErrProvider}
	if w := call("PUT", "/internal/connections", conn, h.ManagementToken); w.Code != 502 {
		t.Fatal("provider failure status", w.Code)
	}
	h.Channels = channelsStub{err: service.ErrSessionNotFound}
	if w := call("PUT", "/internal/connections", conn, h.ManagementToken); w.Code != 422 {
		t.Fatal("expired session status", w.Code)
	}
}
