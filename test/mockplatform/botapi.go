package mockplatform

import (
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// botToken matches the BotFather shape <bot id>:<secret>. Any well-formed token is accepted and creates its bot.
var botToken = regexp.MustCompile(`^bot([0-9]{1,20}):[A-Za-z0-9_-]{30,80}$`)

func botError(w http.ResponseWriter, status int, description string) {
	writeJSON(w, status, map[string]any{"ok": false, "error_code": status, "description": description})
}

func botOK(w http.ResponseWriter, result any) {
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "result": result})
}

// botAPI serves POST /bot<token>/<method> for the methods the connector calls.
func (s *Server) botAPI(w http.ResponseWriter, r *http.Request) {
	match := botToken.FindStringSubmatch(r.PathValue("bot"))
	if match == nil {
		botError(w, http.StatusNotFound, "Not Found")
		return
	}
	id := match[1]
	s.mu.Lock()
	b := s.bots[id]
	if b == nil {
		b = &Bot{ID: id, Username: "mock_" + id + "_bot"}
		s.bots[id] = b
	}
	s.mu.Unlock()
	switch r.PathValue("method") {
	case "getMe":
		n, _ := strconv.ParseInt(id, 10, 64)
		botOK(w, map[string]any{"id": n, "is_bot": true, "first_name": "Mock Bot " + id, "username": b.Username})
	case "setWebhook":
		var in struct {
			URL    string `json:"url"`
			Secret string `json:"secret_token"`
		}
		if !decodeJSON(r, &in) || !strings.HasPrefix(in.URL, "http") {
			botError(w, http.StatusBadRequest, "Bad Request: bad webhook: invalid URL")
			return
		}
		s.mu.Lock()
		b.webhookURL, b.secret, b.WebhookSet = in.URL, in.Secret, true
		s.mu.Unlock()
		botOK(w, true)
	case "deleteWebhook":
		s.mu.Lock()
		b.webhookURL, b.secret, b.WebhookSet = "", "", false
		s.mu.Unlock()
		botOK(w, true)
	case "sendMessage":
		s.sendTelegram(w, r, id)
	default:
		botError(w, http.StatusNotFound, "Not Found: method not found")
	}
}

// telegramWebhook returns a bot's registered webhook URL and secret so the load generator can post updates directly.
// It exposes a mock-only secret on the mock's private port; real Telegram never reveals it.
func (s *Server) telegramWebhook(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	b := s.bots[r.URL.Query().Get("bot")]
	var target, secret string
	if b != nil {
		target, secret = b.webhookURL, b.secret
	}
	s.mu.Unlock()
	if target == "" {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "bot is unknown or has no webhook"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"url": target, "secret_token": secret})
}

func (s *Server) sendTelegram(w http.ResponseWriter, r *http.Request, botID string) {
	var in struct {
		ChatID string `json:"chat_id"`
		Text   string `json:"text"`
	}
	if !decodeJSON(r, &in) || strings.TrimSpace(in.Text) == "" {
		botError(w, http.StatusBadRequest, "Bad Request: message text is empty")
		return
	}
	chat, err := strconv.ParseInt(in.ChatID, 10, 64)
	if err != nil || chat == 0 {
		botError(w, http.StatusBadRequest, "Bad Request: chat not found")
		return
	}
	if fault := s.inject(r.Context(), botID); fault != "" {
		s.count("telegram.send." + fault)
		switch fault {
		case FaultRateLimit:
			botError(w, http.StatusTooManyRequests, "Too Many Requests: retry after 1")
		case FaultBlocked:
			botError(w, http.StatusForbidden, "Forbidden: bot was blocked by the user")
		case FaultAuthExpired:
			botError(w, http.StatusUnauthorized, "Unauthorized")
		case FaultChatNotFound:
			botError(w, http.StatusBadRequest, "Bad Request: chat not found")
		default:
			botError(w, http.StatusInternalServerError, "Internal Server Error")
		}
		return
	}
	m := s.record(Message{Platform: "telegram", AccountID: botID, Customer: in.ChatID, Direction: "outbound", Text: in.Text, Status: http.StatusOK, At: time.Now()})
	s.count("telegram.send.ok")
	botOK(w, map[string]any{"message_id": m.Seq, "date": m.At.Unix(), "chat": map[string]any{"id": chat, "type": "private"}, "text": in.Text})
}
