package mockplatform

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"html"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// graphError writes a Graph API error body that the connector's facebook.Classify maps by code and subcode.
func graphError(w http.ResponseWriter, status, code, subcode int) {
	writeJSON(w, status, map[string]any{"error": map[string]any{"message": "mock graph error", "type": "OAuthException", "code": code, "error_subcode": subcode}})
}

func randomHex(n int) string {
	b := make([]byte, n)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

// dialog stands in for the Facebook login dialog: it shows a consent page whose links return to the API callback
// with a single-use code, or with error=access_denied when the owner cancels.
func (s *Server) dialog(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	redirect, err := url.Parse(q.Get("redirect_uri"))
	if q.Get("client_id") != s.cfg.AppID || err != nil || (redirect.Scheme != "http" && redirect.Scheme != "https") || redirect.Host == "" {
		http.Error(w, "invalid OAuth request", http.StatusBadRequest)
		return
	}
	code := "mock-code-" + randomHex(16)
	s.mu.Lock()
	s.codes[code] = redirect.String()
	s.mu.Unlock()
	allow, deny := *redirect, *redirect
	values := allow.Query()
	values.Set("code", code)
	values.Set("state", q.Get("state"))
	allow.RawQuery = values.Encode()
	values = deny.Query()
	values.Set("error", "access_denied")
	values.Set("state", q.Get("state"))
	deny.RawQuery = values.Encode()
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, `<!doctype html><html><head><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1"><title>Mock Facebook login</title>
<style>body{font:16px system-ui,sans-serif;max-width:28rem;margin:4rem auto;padding:0 1rem}a{display:inline-block;margin:.5rem .5rem 0 0;padding:.6rem 1rem;border-radius:.4rem;text-decoration:none}.ok{background:#1877f2;color:#fff}.no{border:1px solid #888;color:inherit}</style></head>
<body><h1>Mock Facebook</h1><p>Simulated consent for app <code>%s</code>. %d mock Pages will be offered.</p><a class="ok" href="%s">Continue</a><a class="no" href="%s">Cancel</a></body></html>`,
		html.EscapeString(s.cfg.AppID), len(s.pages), html.EscapeString(allow.String()), html.EscapeString(deny.String()))
}

// accessToken handles both the code exchange and the fb_exchange_token upgrade.
func (s *Server) accessToken(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	if q.Get("client_id") != s.cfg.AppID || q.Get("client_secret") != string(s.cfg.AppSecret) {
		graphError(w, http.StatusBadRequest, 101, 0)
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if q.Get("grant_type") == "fb_exchange_token" {
		if !s.userToks[q.Get("fb_exchange_token")] {
			graphError(w, http.StatusBadRequest, 190, 0)
			return
		}
		long := "mock-user-long-" + randomHex(16)
		s.userToks[long] = true
		writeJSON(w, http.StatusOK, map[string]any{"access_token": long, "token_type": "bearer"})
		return
	}
	redirect, ok := s.codes[q.Get("code")]
	delete(s.codes, q.Get("code"))
	if !ok || redirect != q.Get("redirect_uri") {
		graphError(w, http.StatusBadRequest, 100, 36009)
		return
	}
	short := "mock-user-short-" + randomHex(16)
	s.userToks[short] = true
	writeJSON(w, http.StatusOK, map[string]any{"access_token": short, "token_type": "bearer", "expires_in": 3600})
}

func (s *Server) accounts(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.userToks[bearer(r)] {
		graphError(w, http.StatusBadRequest, 190, 0)
		return
	}
	data := make([]map[string]string, 0, len(s.pages))
	for _, p := range s.pages {
		data = append(data, map[string]string{"id": p.ID, "name": p.Name, "access_token": p.token})
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": data})
}

// pageFor returns the Page addressed by the path if the bearer token is its Page token.
func (s *Server) pageFor(r *http.Request) *Page {
	id, token := r.PathValue("page"), bearer(r)
	for _, p := range s.pages {
		if p.ID == id && p.token == token {
			return p
		}
	}
	return nil
}

func (s *Server) subscribe(w http.ResponseWriter, r *http.Request) { s.setSubscribed(w, r, true) }

func (s *Server) unsubscribe(w http.ResponseWriter, r *http.Request) { s.setSubscribed(w, r, false) }

func (s *Server) setSubscribed(w http.ResponseWriter, r *http.Request, on bool) {
	s.mu.Lock()
	p := s.pageFor(r)
	if p != nil {
		p.Subscribed = on
	}
	s.mu.Unlock()
	if p == nil {
		graphError(w, http.StatusBadRequest, 190, 0)
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"success": true})
}

func (s *Server) sendFacebook(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	p := s.pageFor(r)
	s.mu.Unlock()
	if p == nil {
		s.count("facebook.send.auth_rejected")
		graphError(w, http.StatusBadRequest, 190, 0)
		return
	}
	var body struct {
		Recipient struct {
			ID string `json:"id"`
		} `json:"recipient"`
		Message struct {
			Text string `json:"text"`
		} `json:"message"`
	}
	if !decodeJSON(r, &body) || body.Recipient.ID == "" || strings.TrimSpace(body.Message.Text) == "" {
		graphError(w, http.StatusBadRequest, 100, 0)
		return
	}
	if fault := s.inject(r.Context(), p.ID); fault != "" {
		s.count("facebook.send." + fault)
		switch fault {
		case FaultRateLimit:
			graphError(w, http.StatusBadRequest, 4, 0)
		case FaultBlocked:
			graphError(w, http.StatusBadRequest, 551, 0)
		case FaultAuthExpired:
			graphError(w, http.StatusBadRequest, 190, 0)
		case FaultChatNotFound:
			graphError(w, http.StatusBadRequest, 100, 2018001)
		case FaultWindow24h:
			graphError(w, http.StatusBadRequest, 10, 2018278)
		default:
			graphError(w, http.StatusInternalServerError, 2, 0)
		}
		return
	}
	m := s.record(Message{Platform: "facebook", AccountID: p.ID, Customer: body.Recipient.ID, Direction: "outbound", Text: body.Message.Text, Status: http.StatusOK, At: time.Now()})
	s.count("facebook.send.ok")
	mid := fmt.Sprintf("m_mock_%d", m.Seq)
	writeJSON(w, http.StatusOK, map[string]string{"recipient_id": body.Recipient.ID, "message_id": mid})
	if s.cfg.Echo && p.Subscribed {
		go s.deliverFacebook(FacebookEcho(p.ID, body.Recipient.ID, mid, body.Message.Text, time.Now()), "facebook.echo")
	}
}
