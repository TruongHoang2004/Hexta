package controller

import (
	"bufio"
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/TruongHoang2004/Hexta/services/api/common"
	applog "github.com/TruongHoang2004/Hexta/services/api/common/log"
	"github.com/TruongHoang2004/Hexta/services/api/config"
	"github.com/TruongHoang2004/Hexta/services/api/internal/core/model"
	"github.com/TruongHoang2004/Hexta/services/api/internal/core/service"
	"github.com/TruongHoang2004/Hexta/services/api/internal/core/types"
	"github.com/TruongHoang2004/Hexta/services/api/internal/infrastructure/cache"
	"github.com/TruongHoang2004/Hexta/services/api/internal/repository"
	"github.com/TruongHoang2004/Hexta/services/api/internal/testutil"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
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

func TestStreamCarriesStoredMessagesAndDeliveryUpdates(t *testing.T) {
	hub := service.NewConversationStreamHub()
	srv := streamServer(hub)
	defer srv.Close()
	resp := openStream(t, srv, "tenant-a", "u1")
	defer resp.Body.Close()
	reader := bufio.NewReader(resp.Body)
	readEvent(t, reader)
	hub.Dispatch("tenant-a", types.RealtimeHint{ThreadID: "th", Reason: types.RealtimeMessageCreated, Message: &types.MessagingMessage{ID: "m1", ThreadID: "th", Direction: "inbound", Body: "Xin chào"}})
	got := readEvent(t, reader)
	if !strings.Contains(got, "event: message") || !strings.Contains(got, `"body":"Xin chào"`) || !strings.Contains(got, `"id":"m1"`) || !strings.Contains(got, `"direction":"inbound"`) {
		t.Fatalf("a new message must arrive whole, in the API message shape: %q", got)
	}
	hub.Dispatch("tenant-a", types.RealtimeHint{ThreadID: "th", Reason: types.RealtimeDeliveryChanged, Message: &types.MessagingMessage{ID: "m2", ThreadID: "th", DeliveryStatus: "failed", DeliveryErrorReason: "auth_expired"}})
	got = readEvent(t, reader)
	if !strings.Contains(got, "event: delivery") || !strings.Contains(got, `"message_id":"m2"`) || !strings.Contains(got, `"delivery_status":"failed"`) || strings.Contains(got, "body") {
		t.Fatalf("a delivery update carries only the new state: %q", got)
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

// A customer message stored by the command service must reach an open stream through Redis, whole.
func TestStoredMessageReachesOpenStreamThroughRedis(t *testing.T) {
	addr := os.Getenv("API_TEST_REDIS_ADDR")
	if addr == "" {
		t.Skip("API_TEST_REDIS_ADDR is required for the Redis end-to-end stream test")
	}
	db := testutil.Database(t)
	tenant := uuid.NewString()
	for _, row := range []any{&model.Tenant{ID: tenant, Name: "Live", Slug: tenant, OwnerID: "owner", Status: model.TenantStatusActive}, &model.TenantMember{TenantID: tenant, UserID: "owner"}, &model.TenantMember{TenantID: tenant, UserID: "sale"}} {
		if err := db.Create(row).Error; err != nil {
			t.Fatal(err)
		}
	}
	bus := cache.NewRealtimeBusForAddr(addr)
	defer bus.Close()
	hub := service.NewConversationStreamHub()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go hub.Run(ctx, bus, nil)
	commands := service.NewMessagingCommandService(repository.NewMessagingRepository(db)).WithPublisher(bus)

	srv := streamServer(hub)
	defer srv.Close()
	resp := openStream(t, srv, tenant, "sale")
	defer resp.Body.Close()
	reader := bufio.NewReader(resp.Body)
	readEvent(t, reader)
	time.Sleep(500 * time.Millisecond) // let the Redis pattern subscription become live

	if _, err := commands.Inbound(ctx, tenant, "sale", types.MessagingInbound{ExternalKey: "khach-1", DisplayName: "Khách", Body: "Shop còn size M không?", RequestKey: uuid.NewString()}); err != nil {
		t.Fatal(err)
	}
	got := readEvent(t, reader)
	if !strings.Contains(got, "event: message") || !strings.Contains(got, "Shop còn size M không?") || !strings.Contains(got, `"direction":"inbound"`) {
		t.Fatalf("the customer's message must arrive whole on the open stream: %q", got)
	}
}
