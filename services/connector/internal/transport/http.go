package transport

import (
	"context"
	"crypto/hmac"
	"encoding/json"
	"errors"
	"github.com/TruongHoang2004/Hexta/services/connector/internal/plugin"
	"github.com/TruongHoang2004/Hexta/services/connector/internal/service"
	"github.com/google/uuid"
	"io"
	"net/http"
	"net/url"
	"time"
)

// Channels is the machine-only control plane used by the API after it has authorized the tenant owner.
type Channels interface {
	Exchange(ctx context.Context, code, redirectURI string) (service.OAuthSession, error)
	Upsert(ctx context.Context, c service.Connection) (service.UpsertResult, error)
}
type Handler struct {
	Inbound                             *service.InboundService
	Plugin                              plugin.Webhook
	AppID, VerifyToken, ManagementToken string
	Secret                              []byte
	Channels                            Channels
}

func (h *Handler) Routes() http.Handler {
	m := http.NewServeMux()
	m.HandleFunc("/webhooks/facebook", h.Webhook)
	m.HandleFunc("/internal/connections", h.Manage)
	m.HandleFunc("/internal/facebook/oauth/exchange", h.Exchange)
	return m
}
func (h *Handler) Webhook(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		challenge, err := h.Plugin.Challenge(r.URL.Query(), h.VerifyToken)
		if err != nil {
			http.Error(w, "verification failed", 403)
			return
		}
		w.Header().Set("Content-Type", "text/plain")
		io.WriteString(w, challenge)
		return
	}
	if r.Method != http.MethodPost {
		w.WriteHeader(405)
		return
	}
	body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, 1<<20))
	if err != nil {
		http.Error(w, "body limit", 413)
		return
	}
	if h.Plugin.Verify(body, r.Header.Get("X-Hub-Signature-256"), h.Secret) != nil {
		http.Error(w, "verification failed", 403)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()
	if h.Inbound.Receive(ctx, h.AppID, "facebook", body) != nil {
		http.Error(w, "acceptance unavailable", 503)
		return
	}
	w.WriteHeader(200)
}
func (h *Handler) authorized(w http.ResponseWriter, r *http.Request, method string) bool {
	if h.ManagementToken == "" || !hmac.Equal([]byte(r.Header.Get("Authorization")), []byte("Bearer "+h.ManagementToken)) {
		http.Error(w, "unauthorized", 401)
		return false
	}
	if r.Method != method {
		w.WriteHeader(405)
		return false
	}
	return true
}

func writeJSON(w http.ResponseWriter, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	_ = json.NewEncoder(w).Encode(value)
}

// channelError maps service failures to stable, non-sensitive statuses.
func channelError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrSessionNotFound), errors.Is(err, service.ErrCredential):
		http.Error(w, "credential unavailable", 422)
	case errors.Is(err, service.ErrProvider):
		http.Error(w, "provider unavailable", 502)
	case errors.Is(err, service.ErrConflict):
		http.Error(w, "configuration conflict", 409)
	default:
		http.Error(w, "connector unavailable", 503)
	}
}

func (h *Handler) Manage(w http.ResponseWriter, r *http.Request) {
	if !h.authorized(w, r, http.MethodPut) {
		return
	}
	var c service.Connection
	d := json.NewDecoder(http.MaxBytesReader(w, r.Body, 8192))
	d.DisallowUnknownFields()
	if d.Decode(&c) != nil || c.Version < 1 || c.TenantID == "" || c.Platform != "facebook" || c.AccountID == "" || c.AppID != h.AppID || (c.OAuthSessionID != "" && c.UnsubscribeOnDisable) {
		http.Error(w, "invalid connection", 400)
		return
	}
	if _, err := uuid.Parse(c.ID); err != nil {
		http.Error(w, "invalid connection", 400)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()
	result, err := h.Channels.Upsert(ctx, c)
	if err != nil {
		channelError(w, err)
		return
	}
	writeJSON(w, result)
}

func (h *Handler) Exchange(w http.ResponseWriter, r *http.Request) {
	if !h.authorized(w, r, http.MethodPost) {
		return
	}
	var input struct {
		Code        string `json:"code"`
		RedirectURI string `json:"redirect_uri"`
	}
	d := json.NewDecoder(http.MaxBytesReader(w, r.Body, 8192))
	d.DisallowUnknownFields()
	if d.Decode(&input) != nil || input.Code == "" || len(input.Code) > 2048 || input.RedirectURI == "" || len(input.RedirectURI) > 2048 {
		http.Error(w, "invalid exchange", 400)
		return
	}
	if u, err := url.Parse(input.RedirectURI); err != nil || (u.Scheme != "https" && u.Scheme != "http") || u.Host == "" {
		http.Error(w, "invalid exchange", 400)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()
	session, err := h.Channels.Exchange(ctx, input.Code, input.RedirectURI)
	if err != nil {
		channelError(w, err)
		return
	}
	writeJSON(w, session)
}
