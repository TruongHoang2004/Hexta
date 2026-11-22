package bootstrap

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/TruongHoang2004/Hexta/packages/shared/events/connection"
	"github.com/TruongHoang2004/Hexta/packages/shared/events/delivery"
	"github.com/TruongHoang2004/Hexta/packages/shared/events/messaging"
	"github.com/TruongHoang2004/Hexta/packages/shared/kafkatransport"
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

// mockTelegram is a local Telegram Bot API stand-in. Chat IDs select send outcomes; it never sees real tokens.
type mockTelegram struct {
	mu      sync.Mutex
	bots    map[string]int64 // token -> bot ID
	webhook map[string]string
	sends   map[string]int
}

func (m *mockTelegram) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	m.mu.Lock()
	defer m.mu.Unlock()
	token, method, _ := strings.Cut(strings.TrimPrefix(r.URL.Path, "/bot"), "/")
	id, ok := m.bots[token]
	if !ok {
		w.WriteHeader(401)
		io.WriteString(w, `{"ok":false,"error_code":401,"description":"Unauthorized"}`)
		return
	}
	var body map[string]any
	data, _ := io.ReadAll(r.Body)
	_ = json.Unmarshal(data, &body)
	switch method {
	case "getMe":
		json.NewEncoder(w).Encode(map[string]any{"ok": true, "result": map[string]any{"id": id, "is_bot": true, "first_name": "Fixture Bot", "username": "fixture_bot"}})
	case "setWebhook":
		m.webhook[token], _ = body["secret_token"].(string)
		io.WriteString(w, `{"ok":true,"result":true}`)
	case "deleteWebhook":
		delete(m.webhook, token)
		io.WriteString(w, `{"ok":true,"result":true}`)
	case "sendMessage":
		chat, _ := body["chat_id"].(string)
		m.sends[chat]++
		switch chat {
		case "9001":
			io.WriteString(w, `{"ok":true,"result":{"message_id":55}}`)
		case "9002":
			w.WriteHeader(403)
			io.WriteString(w, `{"ok":false,"error_code":403,"description":"Forbidden: bot was blocked by the user"}`)
		case "9003":
			w.WriteHeader(400)
			io.WriteString(w, `{"ok":false,"error_code":400,"description":"Bad Request: chat not found"}`)
		case "9004":
			// The request reached Telegram but the response is lost: the outcome is unknown.
			conn, _, _ := w.(http.Hijacker).Hijack()
			conn.Close()
		case "9005":
			w.WriteHeader(401)
			io.WriteString(w, `{"ok":false,"error_code":401,"description":"Unauthorized"}`)
		}
	default:
		w.WriteHeader(404)
	}
}

func (m *mockTelegram) secret(token string) string {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.webhook[token]
}

func TestConnectorTelegramConnectIngressSendAndResults(t *testing.T) {
	broker := os.Getenv("KAFKA_TEST_BROKERS")
	connectorDB := os.Getenv("CONNECTOR_TEST_DATABASE_URL")
	if broker == "" || connectorDB == "" {
		t.Skip("isolated broker and migrated connector test database required")
	}
	if !strings.Contains(connectorDB, "/connector_fixture?") {
		t.Fatal("requires explicit connector_fixture disposable database")
	}
	db := testutil.Database(t)
	tenant, other, owner, sale := uuid.NewString(), uuid.NewString(), "fixture-owner", "fixture-sale"
	for _, id := range []string{tenant, other} {
		if err := db.Create(&model.Tenant{ID: id, Name: "Fixture", Slug: id, OwnerID: owner, Status: model.TenantStatusActive}).Error; err != nil {
			t.Fatal(err)
		}
		for _, user := range []string{owner, sale} {
			if err := db.Create(&model.TenantMember{TenantID: id, UserID: user}).Error; err != nil {
				t.Fatal(err)
			}
		}
	}
	ctx, cancel := context.WithTimeout(context.Background(), 150*time.Second)
	defer cancel()
	createTopics(t, ctx, broker, messaging.MessageReceivedTopic, delivery.SendRequestedTopic, delivery.SendCompletedTopic, connection.HealthChangedTopic)

	bot := time.Now().UnixNano() % 1_000_000_000_000
	botID := strconv.FormatInt(bot, 10)
	token := botID + ":AAFixtureSecretPart_" + strings.Repeat("a", 20)
	otherToken := strconv.FormatInt(bot+1, 10) + ":AAFixtureSecretPart_" + strings.Repeat("b", 20)
	telegram := &mockTelegram{bots: map[string]int64{token: bot, otherToken: bot + 1}, webhook: map[string]string{}, sends: map[string]int{}}
	telegramServer := httptest.NewServer(telegram)
	defer telegramServer.Close()

	_, file, _, _ := runtime.Caller(0)
	root := filepath.Clean(filepath.Join(filepath.Dir(file), "../../../.."))
	dir := t.TempDir()
	binary := filepath.Join(dir, "connector")
	build := exec.CommandContext(ctx, "go", "build", "-o", binary, "./services/connector/cmd/connector")
	build.Dir = root
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build: %s", output)
	}
	management := randomSecret(t, 24)
	for name, value := range map[string]string{"secret": randomSecret(t, 24), "verify": randomSecret(t, 24), "management": management, "credential": randomSecret(t, 32)} {
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
	base := "http://" + addr
	child := exec.CommandContext(ctx, binary)
	child.Env = append(os.Environ(), "CONNECTOR_DATABASE_URL="+connectorDB, "KAFKA_BROKERS="+broker, "META_APP_ID="+uuid.NewString(), "META_APP_SECRET_FILE="+filepath.Join(dir, "secret"), "META_VERIFY_TOKEN_FILE="+filepath.Join(dir, "verify"), "CONNECTOR_MANAGEMENT_TOKEN_FILE="+filepath.Join(dir, "management"), "CONNECTOR_CREDENTIAL_KEY_FILE="+filepath.Join(dir, "credential"), "CONNECTOR_LISTEN_ADDR="+addr, "TELEGRAM_API_URL="+telegramServer.URL, "TELEGRAM_WEBHOOK_BASE_URL="+base)
	if err := child.Start(); err != nil {
		t.Fatal(err)
	}
	defer func() { cancel(); _ = child.Wait() }()
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
	if _, appErr := channels.ConnectTelegramBot(ctx, tenant, owner, botID+":AAFixtureSecretPart_"+strings.Repeat("z", 20)); appErr == nil || appErr.HTTPStatus != 400 {
		t.Fatal("a token Telegram rejects must return 400", appErr)
	}
	conn, appErr := channels.ConnectTelegramBot(ctx, tenant, owner, token)
	if appErr != nil || conn.AccountID != botID || conn.AccountUsername != "fixture_bot" {
		t.Fatal("connect", appErr)
	}
	if _, appErr := channels.ConnectTelegramBot(ctx, other, owner, token); appErr == nil || appErr.HTTPStatus != 409 {
		t.Fatal("1 Bot = 1 tenant", appErr)
	}
	if _, appErr := channels.UpdateTelegramConnection(ctx, tenant, owner, conn.ID, "rotate_token", otherToken); appErr == nil || appErr.HTTPStatus != 409 {
		t.Fatal("rotate_token must reject a different bot", appErr)
	}
	secret := telegram.secret(token)
	if len(secret) < 32 {
		t.Fatal("webhook not registered with a secret")
	}

	t.Setenv("KAFKA_BROKERS", broker)
	consumerDone := make(chan error, 1)
	go func() { consumerDone <- RunIngress(ctx, db) }()
	defer func() { cancel(); <-consumerDone }()
	intents := kafkatransport.NewTopicPublisher([]string{broker}, delivery.SendRequestedTopic)
	defer intents.Writer.Close()
	relay := service.NewOutboxRelay(repository.NewOutboxRepository(db), intentsOnly{intents})

	update := 0
	post := func(chat int64, chatType, content, secretHeader string) int {
		update++
		body := `{"update_id":` + strconv.Itoa(update) + `,"message":{"message_id":` + strconv.Itoa(update) + `,"date":` + strconv.FormatInt(time.Now().Unix(), 10) + `,"from":{"id":` + strconv.FormatInt(chat, 10) + `},"chat":{"id":` + strconv.FormatInt(chat, 10) + `,"type":"` + chatType + `"},` + content + `}}`
		request, _ := http.NewRequestWithContext(ctx, "POST", base+"/webhooks/telegram/"+conn.ID, bytes.NewReader([]byte(body)))
		request.Header.Set("X-Telegram-Bot-Api-Secret-Token", secretHeader)
		resp, err := client.Do(request)
		if err != nil {
			t.Fatal("webhook", err)
		}
		resp.Body.Close()
		return resp.StatusCode
	}
	if post(9001, "private", `"text":"forged"`, "wrong-secret") != 403 {
		t.Fatal("wrong secret must be rejected")
	}
	if post(-100, "group", `"text":"group chatter"`, secret) != 200 || post(-101, "supergroup", `"text":"x"`, secret) != 200 {
		t.Fatal("group updates must be acknowledged and dropped")
	}
	if post(9001, "private", `"voice":{"file_id":"v"}`, secret) != 200 {
		t.Fatal("voice")
	}
	chats := []int64{9001, 9002, 9003, 9004, 9005}
	for _, chat := range chats {
		if post(chat, "private", `"text":"Xin chào shop"`, secret) != 200 {
			t.Fatal("private text acceptance")
		}
	}
	var threads []messagingmodel.Thread
	for len(threads) < len(chats) {
		db.Where("tenant_id = ? AND recipient_id IS NOT NULL", tenant).Find(&threads)
		select {
		case <-ctx.Done():
			t.Fatal("ingress timed out", len(threads))
		case <-time.After(100 * time.Millisecond):
		}
	}
	var messages int64
	for messages < int64(len(chats)+1) {
		db.Model(&messagingmodel.Message{}).Where("tenant_id = ? AND direction = 'inbound'", tenant).Count(&messages)
		select {
		case <-ctx.Done():
			t.Fatal("ingress messages timed out", messages)
		case <-time.After(100 * time.Millisecond):
		}
	}
	var placeholders int64
	db.Model(&messagingmodel.Message{}).Where("tenant_id = ? AND content_type = ?", tenant, types.ContentUnsupportedMedia).Count(&placeholders)
	if len(threads) != len(chats) || messages != int64(len(chats)+1) || placeholders != 1 {
		t.Fatal("group updates must not reach the inbox; voice must be a placeholder", len(threads), messages, placeholders)
	}

	inbox := newMessagingFacade(repository.NewMessagingRepository(db))
	want := map[string][2]string{
		"9001": {types.DeliverySent, ""},
		"9002": {types.DeliveryFailed, delivery.ErrorBotBlocked},
		"9003": {types.DeliveryFailed, delivery.ErrorChatNotFound},
		"9004": {types.DeliveryUnknown, delivery.ErrorTransport},
		"9005": {types.DeliveryFailed, delivery.ErrorAuthExpired},
	}
	expected := map[string][2]string{}
	for _, thread := range threads {
		if *thread.ConnectionID != conn.ID || !strings.HasPrefix(thread.ExternalKey, "telegram:") {
			t.Fatal("thread binding", thread.ExternalKey)
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
		expected[message.ID] = want[*thread.RecipientID]
	}
	for claimed := true; claimed; {
		claimed, _ = relay.RelayOne(ctx)
	}
	for id, outcome := range expected {
		var message messagingmodel.Message
		for message.DeliveryStatus != outcome[0] {
			db.Where("tenant_id = ? AND id = ?", tenant, id).First(&message)
			select {
			case <-ctx.Done():
				t.Fatalf("delivery result timed out: want %v, got %s", outcome, message.DeliveryStatus)
			case <-time.After(100 * time.Millisecond):
			}
		}
		reason := ""
		if message.DeliveryErrorReason != nil {
			reason = *message.DeliveryErrorReason
		}
		if reason != outcome[1] {
			t.Fatal("classification", outcome, reason)
		}
	}
	telegram.mu.Lock()
	for _, chat := range chats {
		if telegram.sends[strconv.FormatInt(chat, 10)] != 1 {
			telegram.mu.Unlock()
			t.Fatal("each intent must reach Telegram exactly once, with no retry", telegram.sends)
		}
	}
	telegram.mu.Unlock()
	var status string
	for status != types.ChannelStatusAuthExpired {
		list, appErr := channels.ListTelegramConnections(ctx, tenant, sale)
		if appErr != nil {
			t.Fatal(appErr)
		}
		status = list[0].Status
		select {
		case <-ctx.Done():
			t.Fatal("401 must expire the connection", status)
		case <-time.After(100 * time.Millisecond):
		}
	}
	// Rotating the token for the same bot recovers the connection and re-registers the webhook.
	rotated, appErr := channels.UpdateTelegramConnection(ctx, tenant, owner, conn.ID, "rotate_token", token)
	if appErr != nil || rotated.Status != types.ChannelStatusActive || telegram.secret(token) == secret {
		t.Fatal("rotate", appErr)
	}
	// Disconnect deletes the webhook; the old secret no longer verifies.
	if _, appErr := channels.UpdateTelegramConnection(ctx, tenant, owner, conn.ID, "disconnect", ""); appErr != nil || telegram.secret(token) != "" {
		t.Fatal("disconnect", appErr)
	}
	if post(9001, "private", `"text":"after disconnect"`, telegram.secret(token)) != 403 {
		t.Fatal("a disconnected bot must not accept updates")
	}
}
