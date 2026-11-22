package mockplatform

import (
	"context"
	"math/rand/v2"
	"net/http"
	"time"
)

// Fault kinds a send can fail with. Each maps to the provider response the connector classifies.
const (
	FaultServerError      = "server_error"   // 500: outcome unknown, never resent
	FaultRateLimit        = "rate_limit"     // Meta code 4 / Telegram 429
	FaultBlocked          = "blocked"        // Meta code 551 / Telegram 403 bot blocked
	FaultAuthExpired      = "auth_expired"   // Meta code 190 / Telegram 401
	FaultChatNotFound     = "chat_not_found" // Telegram 400 chat not found / Meta invalid recipient
	FaultWindow24h        = "window_24h"     // Meta code 10 subcode 2018278; Telegram treats it as server_error
	maxInjectedLatencyMS  = 120000
	defaultFaultErrorKind = FaultServerError
)

// Faults applies to provider send calls only (Send API and sendMessage). Account latency overrides the default for a
// Page or bot ID, so one slow account can be isolated (T3).
type Faults struct {
	LatencyMS        int            `json:"latency_ms"`
	AccountLatencyMS map[string]int `json:"account_latency_ms,omitempty"`
	ErrorRate        float64        `json:"error_rate"`
	ErrorKind        string         `json:"error_kind,omitempty"`
}

func (f Faults) valid() bool {
	if f.LatencyMS < 0 || f.LatencyMS > maxInjectedLatencyMS || f.ErrorRate < 0 || f.ErrorRate > 1 {
		return false
	}
	for _, v := range f.AccountLatencyMS {
		if v < 0 || v > maxInjectedLatencyMS {
			return false
		}
	}
	switch f.ErrorKind {
	case "", FaultServerError, FaultRateLimit, FaultBlocked, FaultAuthExpired, FaultChatNotFound, FaultWindow24h:
		return true
	}
	return false
}

// inject waits for the configured latency and returns the fault kind to answer with, or "" for success.
func (s *Server) inject(ctx context.Context, account string) string {
	s.mu.Lock()
	f := s.faults
	s.mu.Unlock()
	latency := f.LatencyMS
	if v, ok := f.AccountLatencyMS[account]; ok {
		latency = v
	}
	if latency > 0 {
		select {
		case <-ctx.Done():
		case <-time.After(time.Duration(latency) * time.Millisecond):
		}
	}
	if f.ErrorRate > 0 && rand.Float64() < f.ErrorRate {
		if f.ErrorKind == "" {
			return defaultFaultErrorKind
		}
		return f.ErrorKind
	}
	return ""
}

func (s *Server) getFaults(w http.ResponseWriter, _ *http.Request) {
	s.mu.Lock()
	f := s.faults
	s.mu.Unlock()
	writeJSON(w, http.StatusOK, f)
}

// putFaults replaces the fault configuration; an empty object restores normal behavior.
func (s *Server) putFaults(w http.ResponseWriter, r *http.Request) {
	var f Faults
	if !decodeJSON(r, &f) || !f.valid() {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid faults"})
		return
	}
	s.mu.Lock()
	s.faults = f
	s.mu.Unlock()
	writeJSON(w, http.StatusOK, f)
}
