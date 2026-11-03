package bootstrap

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"github.com/TruongHoang2004/Hexta/packages/shared/ingress"
	"github.com/TruongHoang2004/Hexta/services/api/internal/core/model"
	"github.com/TruongHoang2004/Hexta/services/api/internal/core/service"
	"github.com/TruongHoang2004/Hexta/services/api/internal/core/types"
	"github.com/TruongHoang2004/Hexta/services/api/internal/repository"
	"github.com/TruongHoang2004/Hexta/services/api/internal/repository/messagingmodel"
	"github.com/TruongHoang2004/Hexta/services/api/internal/testutil"
	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestConnectorFacebookToInbox(t *testing.T) {
	broker := os.Getenv("KAFKA_TEST_BROKERS")
	connectorDB := os.Getenv("CONNECTOR_TEST_DATABASE_URL")
	if broker == "" || connectorDB == "" {
		t.Skip("isolated broker and migrated connector test database required")
	}
	if !strings.Contains(connectorDB, "/connector_fixture?") {
		t.Fatal("requires explicit connector_fixture disposable database")
	}
	db := testutil.Database(t)
	tenant := uuid.NewString()
	connection := uuid.NewString()
	app := uuid.NewString()
	page := uuid.NewString()
	if err := db.Create(&model.Tenant{ID: tenant, Name: "Fixture", Slug: tenant, OwnerID: "operator", Status: model.TenantStatusActive}).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&model.TenantMember{TenantID: tenant, UserID: "operator"}).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Exec("INSERT INTO messaging_connections(id,tenant_id,version,platform,account_id,enabled) VALUES(?,?,1,'facebook',?,true)", connection, tenant, page).Error; err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	conn, err := kafka.DialContext(ctx, "tcp", broker)
	if err != nil {
		t.Fatal("test broker unavailable")
	}
	err = conn.CreateTopics(kafka.TopicConfig{Topic: ingress.Topic, NumPartitions: 1, ReplicationFactor: 1})
	conn.Close()
	if err != nil {
		t.Fatal(err)
	}
	_, file, _, _ := runtime.Caller(0)
	root := filepath.Clean(filepath.Join(filepath.Dir(file), "../../../.."))
	dir := t.TempDir()
	binary := filepath.Join(dir, "connector")
	build := exec.CommandContext(ctx, "go", "build", "-o", binary, "./services/connector/cmd/connector")
	build.Dir = root
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build: %s", output)
	}
	secret := "fixture-app-secret-123456"
	verify := "fixture-verify-token-123456"
	management := "fixture-management-123456"
	for name, value := range map[string]string{"secret": secret, "verify": verify, "management": management} {
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
	child.Env = append(os.Environ(), "CONNECTOR_DATABASE_URL="+connectorDB, "KAFKA_BROKERS=127.0.0.1:1", "META_APP_ID="+app, "META_APP_SECRET_FILE="+filepath.Join(dir, "secret"), "META_VERIFY_TOKEN_FILE="+filepath.Join(dir, "verify"), "CONNECTOR_MANAGEMENT_TOKEN_FILE="+filepath.Join(dir, "management"), "CONNECTOR_LISTEN_ADDR="+addr)
	if err := child.Start(); err != nil {
		t.Fatal(err)
	}
	defer func() { cancel(); _ = child.Wait() }()
	client := &http.Client{Timeout: time.Second}
	base := "http://" + addr
	ready := false
	for i := 0; i < 100; i++ {
		resp, err := client.Get(base + "/webhooks/facebook?hub.mode=subscribe&hub.verify_token=" + verify + "&hub.challenge=ok")
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == 200 {
				ready = true
				break
			}
		}
		time.Sleep(50 * time.Millisecond)
	}
	if !ready {
		t.Fatal("connector did not start")
	}
	payload, _ := json.Marshal(map[string]any{"id": connection, "version": 1, "tenant_id": tenant, "platform": "facebook", "account_id": page, "app_id": app, "enabled": true})
	req, _ := http.NewRequestWithContext(ctx, "PUT", base+"/internal/connections", bytes.NewReader(payload))
	req.Header.Set("Authorization", "Bearer "+management)
	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != 204 {
		t.Fatal("provisioning", resp.StatusCode)
	}
	t.Setenv("KAFKA_BROKERS", broker)
	consumerDone := make(chan error, 1)
	go func() { consumerDone <- RunIngress(ctx, db) }()
	defer func() { cancel(); <-consumerDone }()
	body, _ := json.Marshal(map[string]any{"object": "page", "entry": []any{map[string]any{"id": page, "messaging": []any{map[string]any{"sender": map[string]string{"id": "customer"}, "recipient": map[string]string{"id": page}, "timestamp": time.Now().UnixMilli(), "message": map[string]string{"mid": uuid.NewString(), "text": "Fixture from Facebook"}}}}}})
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	for i := 0; i < 2; i++ {
		request, _ := http.NewRequestWithContext(ctx, "POST", base+"/webhooks/facebook", bytes.NewReader(body))
		request.Header.Set("X-Hub-Signature-256", "sha256="+hex.EncodeToString(mac.Sum(nil)))
		resp, err := client.Do(request)
		if err != nil {
			t.Fatal(err)
		}
		resp.Body.Close()
		if resp.StatusCode != 200 {
			t.Fatal("acceptance", resp.StatusCode)
		}
	}
	// Kafka is unreachable during durable HTTP acceptance. Kill the connector
	// after ACK, then restart against the broker using the same PostgreSQL state.
	time.Sleep(1200 * time.Millisecond)
	if err := child.Process.Kill(); err != nil {
		t.Fatal(err)
	}
	_ = child.Wait()
	restartEnv := append(child.Env, "KAFKA_BROKERS="+broker)
	child = exec.CommandContext(ctx, binary)
	child.Env = restartEnv
	if err := child.Start(); err != nil {
		t.Fatal(err)
	}
	var message messagingmodel.Message
	for i := 0; i < 1000; i++ {
		db.Where("tenant_id=? AND status='received'", tenant).Limit(1).Find(&message)
		if message.ID != "" {
			break
		}
		select {
		case <-ctx.Done():
			t.Fatal("consumer timed out")
		case <-time.After(50 * time.Millisecond):
		}
	}
	if message.ID == "" {
		t.Fatal("webhook did not reach inbox")
	}
	var count int64
	db.Model(&messagingmodel.Message{}).Where("tenant_id=?", tenant).Count(&count)
	if count != 1 {
		t.Fatal("duplicate mutation", count)
	}
	inbox := service.NewMessagingService(repository.NewMessagingRepository(db))
	detail, appErr := inbox.Detail(ctx, tenant, "operator", message.ThreadID, types.MessagingQuery{Limit: 20})
	if appErr != nil || len(detail.Messages) != 1 || detail.Messages[0].Status != "received" || !detail.Conversation.Turn.NeedsHuman {
		t.Fatal("real inbox projection")
	}
	if _, err := inbox.Detail(ctx, tenant, "outsider", message.ThreadID, types.MessagingQuery{Limit: 20}); err == nil {
		t.Fatal("membership bypass")
	}
}
