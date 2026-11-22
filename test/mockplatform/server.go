// Package mockplatform simulates the Meta Graph API and the Telegram Bot API for local end-to-end runs and load
// tests. It speaks the provider wire formats the connector uses, records sends, delivers simulated customer messages
// to the connector's webhooks and injects latency or provider errors. State is in memory only. Tokens, secrets and
// message text are never logged.
package mockplatform

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Config holds the identity the connector is configured with. AppSecret must equal the connector's Meta app secret
// so simulated webhooks verify.
type Config struct {
	AppID        string
	AppSecret    []byte
	ConnectorURL string
	Pages        int
	// Echo delivers a message_echoes webhook after each successful Facebook send, as Meta does for subscribed Pages.
	Echo bool
}

// Page is a simulated Facebook Page returned by me/accounts.
type Page struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Subscribed bool   `json:"subscribed"`
	token      string
}

// Bot is a simulated Telegram bot, created on the first call with its token.
type Bot struct {
	ID         string `json:"id"`
	Username   string `json:"username"`
	WebhookSet bool   `json:"webhook_set"`
	webhookURL string
	secret     string
}

// Message is one simulated customer message or one provider send received from the connector.
type Message struct {
	Seq       int64     `json:"seq"`
	Platform  string    `json:"platform"`
	AccountID string    `json:"account_id"`
	Customer  string    `json:"customer"`
	Direction string    `json:"direction"`
	Text      string    `json:"text"`
	Status    int       `json:"status"`
	At        time.Time `json:"at"`
}

// maxMessages bounds the in-memory message log; older entries are dropped.
const maxMessages = 20000

type Server struct {
	cfg    Config
	client *http.Client

	mu        sync.Mutex
	pages     []*Page
	bots      map[string]*Bot
	codes     map[string]string // OAuth code → redirect_uri, single use
	userToks  map[string]bool   // issued short- and long-lived user tokens
	messages  []Message
	seq       int64
	updateID  int64
	faults    Faults
	counters  map[string]int64
	startedAt time.Time
}

func New(cfg Config) *Server {
	if cfg.Pages <= 0 {
		cfg.Pages = 3
	}
	cfg.ConnectorURL = strings.TrimRight(cfg.ConnectorURL, "/")
	s := &Server{cfg: cfg, client: &http.Client{Timeout: 15 * time.Second}, bots: map[string]*Bot{}, codes: map[string]string{}, userToks: map[string]bool{}, counters: map[string]int64{}, startedAt: time.Now()}
	for i := 1; i <= cfg.Pages; i++ {
		id := PageID(i)
		// Deterministic Page tokens keep connector-stored credentials valid across mock restarts.
		s.pages = append(s.pages, &Page{ID: id, Name: "Mock Shop " + strconv.Itoa(i), token: s.derive("page:" + id)})
	}
	return s
}

// PageID returns the ID of the i-th simulated Page (1-based).
func PageID(i int) string { return strconv.Itoa(1000000000000000 + i) }

func (s *Server) derive(label string) string {
	mac := hmac.New(sha256.New, s.cfg.AppSecret)
	mac.Write([]byte(label))
	return "mock-" + hex.EncodeToString(mac.Sum(nil))[:40]
}

func (s *Server) Handler() http.Handler {
	m := http.NewServeMux()
	m.HandleFunc("GET /{$}", s.ui)
	m.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusNoContent) })
	m.HandleFunc("GET /mock/state", s.state)
	m.HandleFunc("GET /mock/messages", s.listMessages)
	m.HandleFunc("GET /mock/stats", s.stats)
	m.HandleFunc("GET /mock/faults", s.getFaults)
	m.HandleFunc("PUT /mock/faults", s.putFaults)
	m.HandleFunc("POST /mock/customers/facebook", s.customerFacebook)
	m.HandleFunc("POST /mock/customers/telegram", s.customerTelegram)
	m.HandleFunc("GET /mock/telegram/webhook", s.telegramWebhook)
	m.HandleFunc("GET /{version}/dialog/oauth", s.dialog)
	m.HandleFunc("GET /{version}/oauth/access_token", s.accessToken)
	m.HandleFunc("GET /{version}/me/accounts", s.accounts)
	m.HandleFunc("POST /{version}/{page}/subscribed_apps", s.subscribe)
	m.HandleFunc("DELETE /{version}/{page}/subscribed_apps", s.unsubscribe)
	m.HandleFunc("POST /{version}/{page}/messages", s.sendFacebook)
	m.HandleFunc("POST /{bot}/{method}", s.botAPI)
	return m
}

func (s *Server) record(m Message) Message {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.seq++
	m.Seq = s.seq
	s.messages = append(s.messages, m)
	if len(s.messages) > maxMessages {
		s.messages = append([]Message(nil), s.messages[len(s.messages)-maxMessages:]...)
	}
	return m
}

func (s *Server) count(key string) {
	s.mu.Lock()
	s.counters[key]++
	s.mu.Unlock()
}

func (s *Server) state(w http.ResponseWriter, _ *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()
	pages := make([]Page, 0, len(s.pages))
	for _, p := range s.pages {
		pages = append(pages, *p)
	}
	bots := make([]Bot, 0, len(s.bots))
	for _, b := range s.bots {
		bots = append(bots, *b)
	}
	writeJSON(w, http.StatusOK, map[string]any{"app_id": s.cfg.AppID, "pages": pages, "bots": bots})
}

// listMessages returns messages with seq greater than ?after, optionally filtered by ?platform, ?account and ?customer.
func (s *Server) listMessages(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	after, _ := strconv.ParseInt(q.Get("after"), 10, 64)
	limit, err := strconv.Atoi(q.Get("limit"))
	if err != nil || limit <= 0 || limit > maxMessages {
		limit = 500
	}
	s.mu.Lock()
	out := []Message{}
	for _, m := range s.messages {
		if m.Seq <= after || (q.Get("platform") != "" && m.Platform != q.Get("platform")) || (q.Get("account") != "" && m.AccountID != q.Get("account")) || (q.Get("customer") != "" && m.Customer != q.Get("customer")) {
			continue
		}
		out = append(out, m)
		if len(out) == limit {
			break
		}
	}
	s.mu.Unlock()
	writeJSON(w, http.StatusOK, map[string]any{"messages": out})
}

func (s *Server) stats(w http.ResponseWriter, _ *http.Request) {
	s.mu.Lock()
	counters := make(map[string]int64, len(s.counters))
	for k, v := range s.counters {
		counters[k] = v
	}
	s.mu.Unlock()
	writeJSON(w, http.StatusOK, map[string]any{"uptime_seconds": int64(time.Since(s.startedAt).Seconds()), "counters": counters})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func decodeJSON(r *http.Request, v any) bool {
	return json.NewDecoder(http.MaxBytesReader(nil, r.Body, 1<<20)).Decode(v) == nil
}

func bearer(r *http.Request) string {
	return strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
}
