package controller

import (
	"bufio"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/TruongHoang2004/Hexta/services/api/common"
	applog "github.com/TruongHoang2004/Hexta/services/api/common/log"
	"github.com/TruongHoang2004/Hexta/services/api/config"
	"github.com/TruongHoang2004/Hexta/services/api/internal/core/service"
	"github.com/TruongHoang2004/Hexta/services/api/internal/core/types"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

func streamServer(hub *service.ConversationStreamHub) *httptest.Server {
	gin.SetMode(gin.TestMode)
	if config.AppConfig == nil {
		config.AppConfig = &config.Config{}
	}
	applog.GetLogger()
	ctrl := NewConversationStreamController(validator.New(), hub)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		if user := c.GetHeader("Test-Actor"); user != "" {
			c.Request = c.Request.WithContext(common.SetAuthInfo(c.Request.Context(), &types.AuthInfo{UserID: user}))
		}
	})
	r.GET("/tenants/:tenantID/conversations/stream", ctrl.Stream)
	return httptest.NewServer(r)
}

func openStream(t *testing.T, srv *httptest.Server, tenant, user string) *http.Response {
	t.Helper()
	req, _ := http.NewRequest("GET", srv.URL+"/tenants/"+tenant+"/conversations/stream", nil)
	req.Header.Set("Test-Actor", user)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	return resp
}

func readEvent(t *testing.T, r *bufio.Reader) string {
	t.Helper()
	var lines []string
	deadline := time.AfterFunc(3*time.Second, func() { t.Error("timed out waiting for an SSE event") })
	defer deadline.Stop()
	for {
		line, err := r.ReadString('\n')
		if err != nil {
			t.Fatal(err)
		}
		if line == "\n" {
			return strings.Join(lines, "")
		}
		lines = append(lines, line)
	}
}

func TestStreamSendsReadyThenTenantHints(t *testing.T) {
	hub := service.NewConversationStreamHub()
	srv := streamServer(hub)
	defer srv.Close()
	resp := openStream(t, srv, "tenant-a", "u1")
	defer resp.Body.Close()
	if ct := resp.Header.Get("Content-Type"); ct != "text/event-stream" {
		t.Fatalf("content type %q", ct)
	}
	reader := bufio.NewReader(resp.Body)
	if got := readEvent(t, reader); !strings.HasPrefix(got, "event: ready") {
		t.Fatalf("first event must be ready: %q", got)
	}
	hub.Dispatch("tenant-b", types.RealtimeHint{ThreadID: "other", Reason: types.RealtimeConversationChanged})
	hub.Dispatch("tenant-a", types.RealtimeHint{ThreadID: "mine", Reason: types.RealtimeConversationChanged})
	got := readEvent(t, reader)
	if !strings.Contains(got, "event: changed") || !strings.Contains(got, `"thread_id":"mine"`) || strings.Contains(got, "other") {
		t.Fatalf("stream must carry only its tenant's hint: %q", got)
	}
}

func TestStreamRejectsUnauthenticatedAndExcessStreams(t *testing.T) {
	hub := service.NewConversationStreamHub()
	srv := streamServer(hub)
	defer srv.Close()
	if resp := openStream(t, srv, "t", ""); resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("missing identity must be 401, got %d", resp.StatusCode)
	}
	var open []*http.Response
	for i := 0; i < service.MaxStreamsPerUser; i++ {
		resp := openStream(t, srv, "t", "u")
		if resp.StatusCode != 200 {
			t.Fatalf("stream %d status %d", i, resp.StatusCode)
		}
		readEvent(t, bufio.NewReader(resp.Body))
		open = append(open, resp)
	}
	if resp := openStream(t, srv, "t", "u"); resp.StatusCode != http.StatusTooManyRequests {
		t.Fatalf("excess stream must be 429, got %d", resp.StatusCode)
	}
	for _, r := range open {
		r.Body.Close()
	}
}
