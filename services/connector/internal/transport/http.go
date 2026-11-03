package transport

import (
	"context"
	"crypto/hmac"
	"encoding/json"
	"github.com/TruongHoang2004/Hexta/services/connector/internal/plugin"
	"github.com/TruongHoang2004/Hexta/services/connector/internal/service"
	"github.com/google/uuid"
	"io"
	"net/http"
	"time"
)

type Management interface {
	PutConnection(*http.Request, service.Connection) error
}
type Handler struct {
	Inbound                             *service.InboundService
	Plugin                              plugin.Webhook
	AppID, VerifyToken, ManagementToken string
	Secret                              []byte
	Put                                 func(*http.Request, service.Connection) error
}

func (h *Handler) Routes() http.Handler {
	m := http.NewServeMux()
	m.HandleFunc("/webhooks/facebook", h.Webhook)
	m.HandleFunc("/internal/connections", h.Manage)
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
func (h *Handler) Manage(w http.ResponseWriter, r *http.Request) {
	if h.ManagementToken == "" || !hmac.Equal([]byte(r.Header.Get("Authorization")), []byte("Bearer "+h.ManagementToken)) {
		http.Error(w, "unauthorized", 401)
		return
	}
	if r.Method != "PUT" {
		w.WriteHeader(405)
		return
	}
	var c service.Connection
	d := json.NewDecoder(http.MaxBytesReader(w, r.Body, 8192))
	d.DisallowUnknownFields()
	if d.Decode(&c) != nil || c.Version < 1 || c.TenantID == "" || c.Platform != "facebook" || c.AccountID == "" || c.AppID != h.AppID {
		http.Error(w, "invalid connection", 400)
		return
	}
	if _, err := uuid.Parse(c.ID); err != nil {
		http.Error(w, "invalid connection", 400)
		return
	}
	if h.Put(r, c) != nil {
		http.Error(w, "configuration conflict", 409)
		return
	}
	w.WriteHeader(204)
}
