package bootstrap

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
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
)

// randomSecret returns a fresh hex secret so no test credential is hardcoded.
func randomSecret(t *testing.T, bytes int) string {
	t.Helper()
	raw := make([]byte, bytes)
	if _, err := rand.Read(raw); err != nil {
		t.Fatal(err)
	}
	return hex.EncodeToString(raw)
}

// TestLiveTelegramWebhookAcceptance runs the real connector against the real Telegram Bot API. It is skipped unless
// the operator provides, through the environment only:
//
//	TELEGRAM_LIVE_BOT_TOKEN         a test bot's @BotFather token (never logged or written to disk)
//	TELEGRAM_LIVE_WEBHOOK_BASE_URL  public HTTPS origin forwarding /webhooks/telegram/* to TELEGRAM_LIVE_CONNECTOR_ADDR
//	TELEGRAM_LIVE_CONNECTOR_ADDR    optional, default 127.0.0.1:18091
//	TELEGRAM_LIVE_WAIT              optional Go duration to wait for the human messages, default 5m
//
// plus the isolated KAFKA_TEST_BROKERS, CONNECTOR_TEST_DATABASE_URL and API_TEST_DATABASE_URL. While it runs, a person
// sends the bot a text and a photo in a private chat; the test replies once and finally disconnects the bot, which
// deletes the webhook.
func TestLiveTelegramWebhookAcceptance(t *testing.T) {
	token := strings.TrimSpace(os.Getenv("TELEGRAM_LIVE_BOT_TOKEN"))
	public := strings.TrimRight(strings.TrimSpace(os.Getenv("TELEGRAM_LIVE_WEBHOOK_BASE_URL")), "/")
	broker := os.Getenv("KAFKA_TEST_BROKERS")
	connectorDB := os.Getenv("CONNECTOR_TEST_DATABASE_URL")
	if token == "" || public == "" || broker == "" || connectorDB == "" {
		t.Skip("live Telegram acceptance requires TELEGRAM_LIVE_BOT_TOKEN, TELEGRAM_LIVE_WEBHOOK_BASE_URL and isolated test infrastructure")
	}
	if !strings.HasPrefix(public, "https://") {
		t.Fatal("TELEGRAM_LIVE_WEBHOOK_BASE_URL must be HTTPS")
	}
	if !strings.Contains(connectorDB, "/connector_fixture?") {
		t.Fatal("requires explicit connector_fixture disposable database")
	}
	addr := os.Getenv("TELEGRAM_LIVE_CONNECTOR_ADDR")
	if addr == "" {
		addr = "127.0.0.1:18091"
	}
	wait := 5 * time.Minute
	if v, err := time.ParseDuration(os.Getenv("TELEGRAM_LIVE_WAIT")); err == nil && v > 0 {
		wait = v
	}
	db := testutil.Database(t)
	tenant, owner, sale := uuid.NewString(), "live-owner", "live-sale"
	if err := db.Create(&model.Tenant{ID: tenant, Name: "Live", Slug: tenant, OwnerID: owner, Status: model.TenantStatusActive}).Error; err != nil {
		t.Fatal(err)
	}
	for _, user := range []string{owner, sale} {
		if err := db.Create(&model.TenantMember{TenantID: tenant, UserID: user}).Error; err != nil {
			t.Fatal(err)
		}
	}
	ctx, cancel := context.WithTimeout(context.Background(), wait+2*time.Minute)
	defer cancel()
	createTopics(t, ctx, broker, ingress.Topic, outbound.IntentTopic, outbound.ResultTopic, outbound.HealthTopic)

	_, file, _, _ := runtime.Caller(0)
	root := filepath.Clean(filepath.Join(filepath.Dir(file), "../../../.."))
	dir := t.TempDir()
	binary := filepath.Join(dir, "connector")
	build := exec.CommandContext(ctx, "go", "build", "-o", binary, "./services/connector/cmd/connector")
	build.Dir = root
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build: %s", output)
	}
	// Connector secrets are generated per run; the connector reads them from private temp files as in production.
	management := randomSecret(t, 24)
	for name, value := range map[string]string{"secret": randomSecret(t, 24), "verify": randomSecret(t, 24), "management": management, "credential": randomSecret(t, 32)} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(value), 0600); err != nil {
			t.Fatal(err)
		}
	}
	child := exec.CommandContext(ctx, binary)
	child.Env = append(os.Environ(), "CONNECTOR_DATABASE_URL="+connectorDB, "KAFKA_BROKERS="+broker, "META_APP_ID="+uuid.NewString(), "META_APP_SECRET_FILE="+filepath.Join(dir, "secret"), "META_VERIFY_TOKEN_FILE="+filepath.Join(dir, "verify"), "CONNECTOR_MANAGEMENT_TOKEN_FILE="+filepath.Join(dir, "management"), "CONNECTOR_CREDENTIAL_KEY_FILE="+filepath.Join(dir, "credential"), "CONNECTOR_LISTEN_ADDR="+addr, "TELEGRAM_WEBHOOK_BASE_URL="+public)
	if err := child.Start(); err != nil {
		t.Fatal(err)
	}
	defer func() { cancel(); _ = child.Wait() }()
	base := "http://" + addr
	client := &http.Client{Timeout: 5 * time.Second}
	for i := 0; ; i++ {
		resp, err := client.Post(base+"/webhooks/telegram/"+uuid.NewString(), "application/json", strings.NewReader(`{}`))
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == 403 {
				break
			}
		}
		if i > 100 {
			t.Fatal("connector did not start")
		}
		time.Sleep(50 * time.Millisecond)
	}

	cfg := &config.Config{}
	cfg.Connector.URL, cfg.Connector.ManagementToken = base, management
	connector, err := connectorclient.NewClient(cfg)
	if err != nil {
		t.Fatal(err)
	}
	channels := service.NewChannelService(repository.NewChannelRepository(db), connector, service.ChannelSettings{})
	conn, appErr := channels.ConnectTelegramBot(ctx, tenant, owner, token)
	if appErr != nil {
		t.Fatal("live connect", appErr.Detail)
	}
	// Always remove the webhook, even when a later step fails.
	defer func() {
		cleanup, stop := context.WithTimeout(context.Background(), 30*time.Second)
		defer stop()
		_, _ = channels.UpdateTelegramConnection(cleanup, tenant, owner, conn.ID, "disconnect", "")
	}()
	t.Logf("connected @%s; send it a text and then a photo in a private chat within %s", conn.AccountUsername, wait)

	t.Setenv("KAFKA_BROKERS", broker)
	consumerDone := make(chan error, 1)
	go func() { consumerDone <- RunIngress(ctx, db) }()
	defer func() { cancel(); <-consumerDone }()

	deadline := time.Now().Add(wait)
	var text, placeholder int64
	for text == 0 || placeholder == 0 {
		db.Model(&messagingmodel.Message{}).Where("tenant_id = ? AND direction = 'inbound' AND content_type = ?", tenant, types.ContentText).Count(&text)
		db.Model(&messagingmodel.Message{}).Where("tenant_id = ? AND direction = 'inbound' AND content_type = ?", tenant, types.ContentUnsupportedMedia).Count(&placeholder)
		if time.Now().After(deadline) {
			t.Fatalf("live inbound not received in time: text=%d media=%d", text, placeholder)
		}
		time.Sleep(time.Second)
	}
	t.Log("live inbound: text message and media placeholder reached the inbox")

	var thread messagingmodel.Thread
	if err := db.Where("tenant_id = ? AND connection_id = ?", tenant, conn.ID).First(&thread).Error; err != nil || !strings.HasPrefix(thread.ExternalKey, "telegram:") {
		t.Fatal("telegram thread", err)
	}
	inbox := newMessagingFacade(repository.NewMessagingRepository(db))
	turn, appErr := repository.NewMessagingRepository(db).CurrentTurn(ctx, tenant, thread.ID)
	if appErr != nil {
		t.Fatal(appErr)
	}
	turn, appErr = inbox.Command(ctx, tenant, sale, turn.ID, "claim", types.MessagingCommand{ExpectedVersion: turn.Version})
	if appErr != nil {
		t.Fatal(appErr)
	}
	message, appErr := inbox.SendChannel(ctx, tenant, sale, turn.ID, types.ChannelSendInput{ClientMessageID: uuid.NewString(), Text: "Hexta live acceptance: reply received.", ExpectedVersion: turn.Version, ContextRevision: turn.ContextRevision})
	if appErr != nil {
		t.Fatal("live send", appErr)
	}
	intents := kafkatransport.NewTopicPublisher([]string{broker}, outbound.IntentTopic)
	defer intents.Writer.Close()
	relay := service.NewOutboxRelay(repository.NewOutboxRepository(db), &fakeKafka{}).WithIntentPublisher(intents)
	for claimed := true; claimed; {
		claimed, _ = relay.RelayOne(ctx)
	}
	var sent messagingmodel.Message
	for sent.DeliveryStatus != types.DeliverySent {
		db.Where("tenant_id = ? AND id = ?", tenant, message.ID).First(&sent)
		if sent.DeliveryStatus == types.DeliveryFailed || sent.DeliveryStatus == types.DeliveryUnknown {
			t.Fatalf("live send outcome %s", sent.DeliveryStatus)
		}
		select {
		case <-ctx.Done():
			t.Fatal("live send result timed out")
		case <-time.After(500 * time.Millisecond):
		}
	}
	t.Log("live outbound: reply delivered (sent)")

	if _, appErr := channels.UpdateTelegramConnection(ctx, tenant, owner, conn.ID, "disconnect", ""); appErr != nil {
		t.Fatal("live disconnect", appErr)
	}
	// The token travels only in this request path to Telegram; the response is checked, never printed.
	resp, err := client.Get("https://api.telegram.org/bot" + token + "/getWebhookInfo")
	if err != nil {
		t.Fatal("getWebhookInfo request failed")
	}
	defer resp.Body.Close()
	var info struct {
		Result struct {
			URL string `json:"url"`
		} `json:"result"`
	}
	if json.NewDecoder(resp.Body).Decode(&info) != nil || info.Result.URL != "" {
		t.Fatal("disconnect must delete the webhook")
	}
	t.Log("live disconnect: webhook removed")
}
