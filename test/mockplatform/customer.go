package mockplatform

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type customerInput struct {
	AccountID string `json:"account_id"`
	Customer  string `json:"customer"`
	Text      string `json:"text"`
	Media     string `json:"media"`
}

func (in customerInput) valid() bool {
	text := strings.TrimSpace(in.Text)
	return in.AccountID != "" && in.Customer != "" && len(in.Customer) <= 64 && (in.Media == MediaImage || (text != "" && len([]rune(in.Text)) <= 4000)) && (in.Media == "" || in.Media == MediaImage)
}

func (in customerInput) shown() string {
	if in.Media == MediaImage {
		return "[image]"
	}
	return in.Text
}

// customerFacebook sends a customer's message to a subscribed mock Page through the connector's signed webhook.
func (s *Server) customerFacebook(w http.ResponseWriter, r *http.Request) {
	var in customerInput
	if !decodeJSON(r, &in) || !in.valid() {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "account_id, customer and text or media=image are required"})
		return
	}
	s.mu.Lock()
	var page *Page
	for _, p := range s.pages {
		if p.ID == in.AccountID {
			page = p
		}
	}
	subscribed := page != nil && page.Subscribed
	s.mu.Unlock()
	mid := "m_customer_" + randomHex(12)
	if !subscribed {
		writeJSON(w, http.StatusConflict, map[string]string{"error": "page is unknown or not subscribed; connect it in Hexta first"})
		return
	}
	status := s.deliverFacebook(FacebookMessage(in.AccountID, in.Customer, mid, in.Text, in.Media, time.Now()), "facebook.customer")
	s.record(Message{Platform: "facebook", AccountID: in.AccountID, Customer: in.Customer, Direction: "inbound", Text: in.shown(), Status: status, At: time.Now()})
	writeJSON(w, http.StatusOK, map[string]any{"webhook_status": status, "message_id": mid})
}

// customerTelegram sends a private-chat update to the webhook a bot registered through setWebhook.
func (s *Server) customerTelegram(w http.ResponseWriter, r *http.Request) {
	var in customerInput
	if !decodeJSON(r, &in) || !in.valid() {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "account_id, customer and text or media=image are required"})
		return
	}
	chat, err := strconv.ParseInt(in.Customer, 10, 64)
	if err != nil || chat <= 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "customer must be a positive numeric chat ID"})
		return
	}
	s.mu.Lock()
	b := s.bots[in.AccountID]
	var target, secret string
	if b != nil {
		target, secret = b.webhookURL, b.secret
	}
	s.updateID++
	update := s.updateID
	s.mu.Unlock()
	if target == "" {
		writeJSON(w, http.StatusConflict, map[string]string{"error": "bot is unknown or has no webhook; connect it in Hexta first"})
		return
	}
	status := s.post(target, TelegramUpdate(update, chat, update, in.Text, in.Media, time.Now()), map[string]string{TelegramSecretHeader: secret}, "telegram.customer")
	s.record(Message{Platform: "telegram", AccountID: in.AccountID, Customer: in.Customer, Direction: "inbound", Text: in.shown(), Status: status, At: time.Now()})
	writeJSON(w, http.StatusOK, map[string]any{"webhook_status": status, "update_id": update})
}

func (s *Server) deliverFacebook(body []byte, counter string) int {
	return s.post(s.cfg.ConnectorURL+"/webhooks/facebook", body, map[string]string{"X-Hub-Signature-256": SignFacebook(s.cfg.AppSecret, body)}, counter)
}

// post delivers one webhook and returns the HTTP status, or 0 when the connector was unreachable.
func (s *Server) post(target string, body []byte, headers map[string]string, counter string) int {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, target, bytes.NewReader(body))
	if err != nil {
		s.count(counter + ".error")
		return 0
	}
	req.Header.Set("Content-Type", "application/json")
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	res, err := s.client.Do(req)
	if err != nil {
		s.count(counter + ".error")
		return 0
	}
	_, _ = io.Copy(io.Discard, res.Body)
	res.Body.Close()
	s.count(counter + "." + strconv.Itoa(res.StatusCode))
	return res.StatusCode
}
