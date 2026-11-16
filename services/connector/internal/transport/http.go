package transport

import (
	"context"
	"crypto/hmac"
	"encoding/json"
	"errors"
	"github.com/TruongHoang2004/Hexta/services/connector/internal/plugin"
	"github.com/TruongHoang2004/Hexta/services/connector/internal/plugin/telegram"
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
	VerifyBot(ctx context.Context, token string) (plugin.BotIdentity, error)
	WebhookSecret(ctx context.Context, id string) (service.Connection, []byte, error)
}
type Handler struct {
	Inbound                             *service.InboundService
	Plugin                              plugin.Webhook
	AppID, VerifyToken, ManagementToken string
	Secret                              []byte
	Channels                            Channels
	// Telegram verifies per-connection webhook secrets.
	Telegram plugin.Webhook
}

func (h *Handler) Routes() http.Handler {
	m := http.NewServeMux()
	m.HandleFunc("/webhooks/facebook", h.Webhook)
	m.HandleFunc("/internal/connections", h.Manage)
	m.HandleFunc("/internal/facebook/oauth/exchange", h.Exchange)
	m.HandleFunc("/internal/telegram/bots/verify", h.VerifyBot)
	m.HandleFunc("/webhooks/telegram/{connectionID}", h.TelegramWebhook)
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
	case errors.Is(err, service.ErrUnavailable):
		http.Error(w, "platform not configured", 503)
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
	if d.Decode(&c) != nil || c.Version < 1 || c.TenantID == "" || c.AccountID == "" || !validConnection(c, h.AppID) {
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

// validConnection enforces per-platform identity and credential fields. A Bot Token is accepted only to enable a
// Telegram connection whose account is that token's bot.
func validConnection(c service.Connection, facebookApp string) bool {
	switch c.Platform {
	case "facebook":
		return c.AppID == facebookApp && c.BotToken == "" && !(c.OAuthSessionID != "" && c.UnsubscribeOnDisable)
	case "telegram":
		if c.AppID != telegram.AppID || c.OAuthSessionID != "" || c.CredentialRef != "" || !telegram.ValidBotID(c.AccountID) {
			return false
		}
		return c.BotToken == "" || (c.Enabled && telegram.ValidToken(c.BotToken) && telegram.TokenBotID(c.BotToken) == c.AccountID)
	}
	return false
}

// VerifyBot checks a Telegram Bot Token with getMe and returns only the bot identity.
func (h *Handler) VerifyBot(w http.ResponseWriter, r *http.Request) {
	if !h.authorized(w, r, http.MethodPost) {
		return
	}
	var input struct {
		BotToken string `json:"bot_token"`
	}
	d := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4096))
	d.DisallowUnknownFields()
	if d.Decode(&input) != nil || !telegram.ValidToken(input.BotToken) {
		http.Error(w, "invalid bot token", 400)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()
	bot, err := h.Channels.VerifyBot(ctx, input.BotToken)
	if err != nil {
		channelError(w, err)
		return
	}
	writeJSON(w, map[string]string{"bot_id": bot.ID, "display_name": bot.DisplayName, "username": bot.Username})
}

// TelegramWebhook accepts one Update for one connection. The secret header is checked before parsing; dropped
// updates (group/channel chats, non-message updates) return 200 so Telegram does not retry them. Accepted updates
// return 200 only after the receipt and outbox commit.
func (h *Handler) TelegramWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(405)
		return
	}
	id := r.PathValue("connectionID")
	if _, err := uuid.Parse(id); err != nil || h.Telegram == nil {
		http.NotFound(w, r)
		return
	}
	body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, 1<<20))
	if err != nil {
		http.Error(w, "body limit", 413)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()
	c, secret, err := h.Channels.WebhookSecret(ctx, id)
	switch {
	case errors.Is(err, service.ErrConnectionMissing), errors.Is(err, service.ErrCredential):
		http.Error(w, "verification failed", 403)
		return
	case err != nil:
		http.Error(w, "acceptance unavailable", 503)
		return
	}
	if c.Platform != "telegram" || h.Telegram.Verify(body, r.Header.Get(telegram.SecretHeader), secret) != nil {
		http.Error(w, "verification failed", 403)
		return
	}
	err = h.Inbound.ReceiveConnection(ctx, c, body)
	switch {
	case errors.Is(err, service.ErrInvalidPayload):
		http.Error(w, "invalid update", 400)
	case err != nil:
		http.Error(w, "acceptance unavailable", 503)
	default:
		w.WriteHeader(200)
	}
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
