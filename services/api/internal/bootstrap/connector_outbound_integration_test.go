package bootstrap

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	stderrors "errors"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/TruongHoang2004/Hexta/packages/shared/ingress"
	"github.com/TruongHoang2004/Hexta/packages/shared/kafkatransport"
	"github.com/TruongHoang2004/Hexta/packages/shared/outbound"
	"github.com/TruongHoang2004/Hexta/services/api/config"
	"github.com/TruongHoang2004/Hexta/services/api/internal/core/model"
	"github.com/TruongHoang2004/Hexta/services/api/internal/core/service"
	"github.com/TruongHoang2004/Hexta/services/api/internal/core/types"
	connectorclient "github.com/TruongHoang2004/Hexta/services/api/internal/infrastructure/connector"
	"github.com/TruongHoang2004/Hexta/services/api/internal/repository"
	"github.com/TruongHoang2004/Hexta/services/api/internal/repository/messagingmodel"
	"github.com/TruongHoang2004/Hexta/services/api/internal/testutil"
	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
)

// mockGraph is a local Meta Graph stand-in. It never receives real credentials.
type mockGraph struct {
	mu    sync.Mutex
	sends map[string]int
	page  string
}

func (m *mockGraph) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	m.mu.Lock()
	defer m.mu.Unlock()
	switch {
	case strings.HasSuffix(r.URL.Path, "/oauth/access_token"):
		json.NewEncoder(w).Encode(map[string]string{"access_token": "fixture-user-token"})
	case strings.HasSuffix(r.URL.Path, "/me/accounts"):
		json.NewEncoder(w).Encode(map[string]any{"data": []map[string]string{{"id": m.page, "name": "Fixture Shop", "access_token": "fixture-page-token"}}})
	case strings.HasSuffix(r.URL.Path, "/subscribed_apps"):
		if r.Header.Get("Authorization") != "Bearer fixture-page-token" {
			w.WriteHeader(400)
			return
		}
		json.NewEncoder(w).Encode(map[string]bool{"success": true})
	case strings.HasSuffix(r.URL.Path, "/"+m.page+"/messages"):
		var body struct {
			Recipient struct {
				ID string `json:"id"`
			} `json:"recipient"`
		}
		data, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(data, &body)
		m.sends[body.Recipient.ID]++
		if r.Header.Get("Authorization") != "Bearer fixture-page-token" {
			w.WriteHeader(400)
			io.WriteString(w, `{"error":{"code":190}}`)
			return
		}
		if body.Recipient.ID == "psid-window" {
			w.WriteHeader(400)
			io.WriteString(w, `{"error":{"code":10,"error_subcode":2018278}}`)
			return
		}
		json.NewEncoder(w).Encode(map[string]string{"recipient_id": body.Recipient.ID, "message_id": "mid.fixture"})
	default:
		w.WriteHeader(404)
	}
}

func createTopics(t *testing.T, ctx context.Context, broker string, topics ...string) {
	t.Helper()
	conn, err := kafka.DialContext(ctx, "tcp", broker)
	if err != nil {
		t.Fatal("test broker unavailable")
	}
	defer conn.Close()
	for _, topic := range topics {
		if err := conn.CreateTopics(kafka.TopicConfig{Topic: topic, NumPartitions: 1, ReplicationFactor: 1}); err != nil && !stderrors.Is(err, kafka.TopicAlreadyExists) {
			t.Fatal(err)
		}
	}
}

func TestConnectorFacebookOAuthSendAndDeliveryResults(t *testing.T) {
	broker := os.Getenv("KAFKA_TEST_BROKERS")
	connectorDB := os.Getenv("CONNECTOR_TEST_DATABASE_URL")
	if broker == "" || connectorDB == "" {
		t.Skip("isolated broker and migrated connector test database required")
	}
	if !strings.Contains(connectorDB, "/connector_fixture?") {
		t.Fatal("requires explicit connector_fixture disposable database")
	}
	db := testutil.Database(t)
	tenant, owner, sale := uuid.NewString(), "fixture-owner", "fixture-sale"
	page := "page-" + uuid.NewString()[:8]
	if err := db.Create(&model.Tenant{ID: tenant, Name: "Fixture", Slug: tenant, OwnerID: owner, Status: model.TenantStatusActive}).Error; err != nil {
		t.Fatal(err)
	}
	for _, user := range []string{owner, sale} {
		if err := db.Create(&model.TenantMember{TenantID: tenant, UserID: user}).Error; err != nil {
			t.Fatal(err)
		}
	}
	ctx, cancel := context.WithTimeout(context.Background(), 150*time.Second)
	defer cancel()
	createTopics(t, ctx, broker, ingress.Topic, outbound.IntentTopic, outbound.ResultTopic, outbound.HealthTopic)

	graph := &mockGraph{sends: map[string]int{}, page: page}
	graphServer := httptest.NewServer(graph)
	defer graphServer.Close()

	_, file, _, _ := runtime.Caller(0)
	root := filepath.Clean(filepath.Join(filepath.Dir(file), "../../../.."))
	dir := t.TempDir()
	binary := filepath.Join(dir, "connector")
	build := exec.CommandContext(ctx, "go", "build", "-o", binary, "./services/connector/cmd/connector")
	build.Dir = root
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build: %s", output)
	}
	app, secret, verify, management := uuid.NewString(), "fixture-app-secret-123456", "fixture-verify-token-123456", "fixture-management-123456"
	for name, value := range map[string]string{"secret": secret, "verify": verify, "management": management, "credential": strings.Repeat("b2", 32)} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(value), 0600); err != nil {
			t.Fatal(err)
		}
	}
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	addr := listener.Addr().String()
	listener.Close()
	child := exec.CommandContext(ctx, binary)
	child.Env = append(os.Environ(), "CONNECTOR_DATABASE_URL="+connectorDB, "KAFKA_BROKERS="+broker, "META_APP_ID="+app, "META_APP_SECRET_FILE="+filepath.Join(dir, "secret"), "META_VERIFY_TOKEN_FILE="+filepath.Join(dir, "verify"), "CONNECTOR_MANAGEMENT_TOKEN_FILE="+filepath.Join(dir, "management"), "CONNECTOR_CREDENTIAL_KEY_FILE="+filepath.Join(dir, "credential"), "CONNECTOR_LISTEN_ADDR="+addr, "META_GRAPH_URL="+graphServer.URL)
	if err := child.Start(); err != nil {
		t.Fatal(err)
	}
	defer func() { cancel(); _ = child.Wait() }()
	base := "http://" + addr
	client := &http.Client{Timeout: 5 * time.Second}
	for i := 0; ; i++ {
		resp, err := client.Get(base + "/webhooks/facebook?hub.mode=subscribe&hub.verify_token=" + verify + "&hub.challenge=ok")
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == 200 {
				break
			}
		}
		if i > 100 {
			t.Fatal("connector did not start")
		}
		time.Sleep(50 * time.Millisecond)
	}

	// Owner OAuth → Page picker → connect, through the real connector machine API.
	cfg := &config.Config{}
	cfg.Connector.URL, cfg.Connector.ManagementToken = base, management
	connector, err := connectorclient.NewClient(cfg)
	if err != nil {
		t.Fatal(err)
	}
	channels := service.NewChannelService(repository.NewChannelRepository(db), connector, service.ChannelSettings{AppID: app, AllowedOrigins: []string{"http://localhost:3000"}})
	start, appErr := channels.StartFacebookOAuth(ctx, tenant, owner, "http://localhost:3000/channels/facebook/callback")
	if appErr != nil {
		t.Fatal(appErr)
	}
	authURL, _ := url.Parse(start.AuthURL)
	pages, appErr := channels.CompleteFacebookOAuth(ctx, owner, "fixture-code", authURL.Query().Get("state"), "")
	if appErr != nil || len(pages.Pages) != 1 || pages.Pages[0].Availability != types.PageAvailable {
		t.Fatal("oauth callback", appErr)
	}
	connected, appErr := channels.ConnectFacebookPages(ctx, tenant, owner, pages.OAuthSessionID, []string{page})
	if appErr != nil {
		t.Fatal("connect", appErr)
	}

	t.Setenv("KAFKA_BROKERS", broker)
	consumerDone := make(chan error, 1)
	go func() { consumerDone <- RunIngress(ctx, db) }()
	defer func() { cancel(); <-consumerDone }()
	intents := kafkatransport.NewTopicPublisher([]string{broker}, outbound.IntentTopic)
	defer intents.Writer.Close()
	relay := service.NewOutboxRelay(repository.NewOutboxRepository(db), &fakeKafka{}).WithIntentPublisher(intents)

	inbox := newMessagingFacade(repository.NewMessagingRepository(db))
	for _, psid := range []string{"psid-ok", "psid-window"} {
		body, _ := json.Marshal(map[string]any{"object": "page", "entry": []any{map[string]any{"id": page, "messaging": []any{map[string]any{"sender": map[string]string{"id": psid}, "recipient": map[string]string{"id": page}, "timestamp": time.Now().UnixMilli(), "message": map[string]string{"mid": uuid.NewString(), "text": "Hello shop"}}}}}})
		mac := hmac.New(sha256.New, []byte(secret))
		mac.Write(body)
		request, _ := http.NewRequestWithContext(ctx, "POST", base+"/webhooks/facebook", bytes.NewReader(body))
		request.Header.Set("X-Hub-Signature-256", "sha256="+hex.EncodeToString(mac.Sum(nil)))
		resp, err := client.Do(request)
		if err != nil || resp.StatusCode != 200 {
			t.Fatal("webhook acceptance", err)
		}
		resp.Body.Close()
	}
	var threads []messagingmodel.Thread
	for len(threads) < 2 {
		db.Where("tenant_id = ? AND recipient_id IS NOT NULL", tenant).Find(&threads)
		select {
		case <-ctx.Done():
			t.Fatal("ingress timed out")
		case <-time.After(100 * time.Millisecond):
		}
	}
	want := map[string]string{}
	for _, thread := range threads {
		if *thread.ConnectionID != connected[0].ID {
			t.Fatal("thread bound to the wrong connection")
		}
		turn, appErr := repository.NewMessagingRepository(db).CurrentTurn(ctx, tenant, thread.ID)
		if appErr != nil {
			t.Fatal(appErr)
		}
		turn, appErr = inbox.Command(ctx, tenant, sale, turn.ID, "claim", types.MessagingCommand{ExpectedVersion: turn.Version})
		if appErr != nil {
			t.Fatal(appErr)
		}
		message, appErr := inbox.SendChannel(ctx, tenant, sale, turn.ID, types.ChannelSendInput{ClientMessageID: uuid.NewString(), Text: "Cảm ơn bạn!", ExpectedVersion: turn.Version, ContextRevision: turn.ContextRevision})
		if appErr != nil {
			t.Fatal("send", appErr)
		}
		want[message.ID] = map[string]string{"psid-ok": types.DeliverySent, "psid-window": types.DeliveryFailed}[*thread.RecipientID]
	}
	for claimed := true; claimed; {
		claimed, _ = relay.RelayOne(ctx)
	}
	for id, status := range want {
		var message messagingmodel.Message
		for message.DeliveryStatus != status {
			db.Where("tenant_id = ? AND id = ?", tenant, id).First(&message)
			select {
			case <-ctx.Done():
				t.Fatalf("delivery result timed out: want %s, got %s", status, message.DeliveryStatus)
			case <-time.After(100 * time.Millisecond):
			}
		}
		if status == types.DeliveryFailed && (message.DeliveryErrorReason == nil || *message.DeliveryErrorReason != outbound.ErrorWindow24h) {
			t.Fatal("window classification", message.DeliveryErrorReason)
		}
	}
	graph.mu.Lock()
	sends := map[string]int{"psid-ok": graph.sends["psid-ok"], "psid-window": graph.sends["psid-window"]}
	graph.mu.Unlock()
	if sends["psid-ok"] != 1 || sends["psid-window"] != 1 {
		t.Fatal("each intent must reach the provider exactly once, with no retry", sends)
	}
	list, appErr := channels.ListFacebookConnections(ctx, tenant, sale)
	if appErr != nil || list[0].Status != types.ChannelStatusActive {
		t.Fatal("window failure must not expire the connection")
	}
}

type fakeKafka struct{}

func (fakeKafka) Publish(context.Context, string, []byte) error {
	return stderrors.New("order topic unused")
}
