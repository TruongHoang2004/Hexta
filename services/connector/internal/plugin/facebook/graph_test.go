package facebook

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/TruongHoang2004/Hexta/packages/shared/outbound"
	"github.com/TruongHoang2004/Hexta/services/connector/internal/plugin"
)

func TestClassify(t *testing.T) {
	for _, c := range []struct {
		status, code, sub int
		category          string
		unknown           bool
	}{
		{400, 10, 2018278, outbound.ErrorWindow24h, false},
		{400, 2022, 0, outbound.ErrorWindow24h, false},
		{400, 190, 460, outbound.ErrorAuthExpired, false},
		{403, 200, 0, outbound.ErrorAuthExpired, false},
		{400, 613, 0, outbound.ErrorRateLimit, false},
		{400, 100, 2018001, outbound.ErrorInvalidRecipient, false},
		{500, 2, 0, outbound.ErrorTransport, true},
		{400, 100, 0, outbound.ErrorTransport, false},
	} {
		got := Classify(c.status, c.code, c.sub)
		if got.Category != c.category || got.Unknown != c.unknown {
			t.Fatal("classification", c, got)
		}
	}
}

func TestGraphKeepsSecretsOutOfErrorsAndURLs(t *testing.T) {
	var sawToken bool
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasSuffix(r.URL.Path, "/oauth/access_token"):
			json.NewEncoder(w).Encode(map[string]string{"access_token": "user-token-" + r.URL.Query().Get("grant_type")})
		case strings.HasSuffix(r.URL.Path, "/me/accounts"):
			sawToken = r.Header.Get("Authorization") == "Bearer user-token-fb_exchange_token"
			json.NewEncoder(w).Encode(map[string]any{"data": []map[string]string{{"id": "p1", "name": "Shop", "access_token": "page-token"}, {"id": "p2", "name": "No token"}}})
		case strings.HasSuffix(r.URL.Path, "/p1/messages"):
			body, _ := io.ReadAll(r.Body)
			if strings.Contains(r.URL.RawQuery, "page-token") || !strings.Contains(string(body), `"id":"psid"`) {
				w.WriteHeader(400)
				return
			}
			w.WriteHeader(400)
			io.WriteString(w, `{"error":{"message":"outside window page-token","code":10,"error_subcode":2018278}}`)
		default:
			w.WriteHeader(404)
		}
	}))
	defer srv.Close()
	g := NewGraph(srv.URL, "v21.0", "app", []byte("app-secret-value"))
	pages, err := g.ExchangeCode(context.Background(), "code", "https://app.example/cb")
	if err != nil || !sawToken || len(pages) != 1 || pages[0].Token != "page-token" {
		t.Fatal("exchange", err, pages)
	}
	_, err = g.SendText(context.Background(), "p1", "page-token", "psid", "hi")
	if ge, ok := err.(*plugin.ProviderError); !ok || ge.Category != outbound.ErrorWindow24h || strings.Contains(err.Error(), "page-token") {
		t.Fatal("send classification", err)
	}
	g.BaseURL = "http://127.0.0.1:1"
	_, err = g.SendText(context.Background(), "p1", "page-token", "psid", "hi")
	if ge, ok := err.(*plugin.ProviderError); !ok || !ge.Unknown || strings.Contains(err.Error(), "127.0.0.1") {
		t.Fatal("network failure must be unknown and sanitized", err)
	}
}
