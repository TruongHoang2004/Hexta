// Command mockplatform serves the simulated Meta Graph API and Telegram Bot API for local runs and load tests.
//
// Environment:
//
//	MOCK_APP_SECRET_FILE  file with the Meta app secret shared with the connector (required)
//	MOCK_APP_ID           Meta app ID the connector and API use (default mock-app)
//	MOCK_CONNECTOR_URL    connector base URL for Facebook webhooks (default http://127.0.0.1:8091)
//	MOCK_LISTEN_ADDR      listen address (default 127.0.0.1:8090)
//	MOCK_PAGES            number of simulated Pages (default 3)
//	MOCK_ECHO             deliver message_echoes after Facebook sends (default true)
package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/TruongHoang2004/Hexta/test/mockplatform"
)

func env(key, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return fallback
}

func main() {
	secret, err := os.ReadFile(os.Getenv("MOCK_APP_SECRET_FILE"))
	secret = []byte(strings.TrimSpace(string(secret)))
	if err != nil || len(secret) < 16 {
		log.Fatal("MOCK_APP_SECRET_FILE must name a file with at least 16 bytes")
	}
	pages, err := strconv.Atoi(env("MOCK_PAGES", "3"))
	if err != nil || pages < 1 || pages > 1000 {
		log.Fatal("MOCK_PAGES must be between 1 and 1000")
	}
	server := mockplatform.New(mockplatform.Config{AppID: env("MOCK_APP_ID", "mock-app"), AppSecret: secret, ConnectorURL: env("MOCK_CONNECTOR_URL", "http://127.0.0.1:8091"), Pages: pages, Echo: env("MOCK_ECHO", "true") == "true"})
	addr := env("MOCK_LISTEN_ADDR", "127.0.0.1:8090")
	// No write timeout: injected latency can hold a send open past the connector's 15 s client timeout.
	srv := &http.Server{Addr: addr, Handler: server.Handler(), ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 15 * time.Second, IdleTimeout: 60 * time.Second}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	go func() {
		<-ctx.Done()
		shutdown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = srv.Shutdown(shutdown)
	}()
	log.Printf("mock platform listening on %s with %d pages", addr, pages)
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatal("mock platform stopped: ", err)
	}
}
