package bootstrap

import (
	"context"
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/TruongHoang2004/Hexta/packages/shared/kafkatransport"
	"github.com/TruongHoang2004/Hexta/services/api/internal/core/model"
	"github.com/TruongHoang2004/Hexta/services/api/internal/core/service"
	"github.com/TruongHoang2004/Hexta/services/api/internal/core/types"
	"github.com/TruongHoang2004/Hexta/services/api/internal/repository"
	"github.com/TruongHoang2004/Hexta/services/api/internal/testutil"
	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
)

func TestOutboxRelayPublishesCommittedEventToKafka(t *testing.T) {
	broker := os.Getenv("KAFKA_TEST_BROKERS")
	if broker == "" {
		t.Skip("isolated broker required")
	}
	db := testutil.Database(t)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	// A unique topic keeps this test independent of other runs against the same broker.
	topic := "hexta.orders.events.test." + uuid.NewString()
	conn, err := kafka.DialContext(ctx, "tcp", broker)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	if err = conn.CreateTopics(kafka.TopicConfig{Topic: topic, NumPartitions: 1, ReplicationFactor: 1}); err != nil {
		t.Fatal(err)
	}

	event := model.DomainEvent{ID: uuid.NewString(), TenantID: "tenant", EventType: model.DomainEventOrderConfirmed, AggregateID: "order", Payload: `{"order_id":"order","version":2}`}
	if err = db.Create(&event).Error; err != nil {
		t.Fatal(err)
	}
	publisher := kafkatransport.NewTopicPublisher([]string{broker}, topic)
	defer publisher.Writer.Close()
	claimed, failure := service.NewOutboxRelay(repository.NewOutboxRepository(db), publisher).RelayOne(ctx)
	if failure != nil || !claimed {
		t.Fatal("relay", claimed, failure)
	}

	reader := kafka.NewReader(kafka.ReaderConfig{Brokers: []string{broker}, Topic: topic, MaxBytes: 1 << 20})
	defer reader.Close()
	message, err := reader.ReadMessage(ctx)
	if err != nil {
		t.Fatal(err)
	}
	var envelope types.DomainEventEnvelope
	if err = json.Unmarshal(message.Value, &envelope); err != nil || envelope.EventID != event.ID || string(message.Key) != "order" {
		t.Fatalf("unexpected message key=%s envelope=%+v err=%v", message.Key, envelope, err)
	}
	var stored model.DomainEvent
	if err = db.First(&stored, "id = ?", event.ID).Error; err != nil || stored.State != model.DomainEventStatePublished {
		t.Fatal("event not marked published", stored.State, err)
	}
}
