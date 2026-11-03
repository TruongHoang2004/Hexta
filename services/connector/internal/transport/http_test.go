package transport

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"github.com/TruongHoang2004/Hexta/services/connector/internal/facebook"
	"github.com/TruongHoang2004/Hexta/services/connector/internal/plugin"
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
