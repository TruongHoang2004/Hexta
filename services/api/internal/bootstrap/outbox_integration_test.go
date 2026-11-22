package bootstrap

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/TruongHoang2004/Hexta/packages/shared/events"
	"os"
	"testing"
	"time"

	"github.com/TruongHoang2004/Hexta/packages/shared/events/order"
	"github.com/TruongHoang2004/Hexta/packages/shared/kafkatransport"
	"github.com/TruongHoang2004/Hexta/services/api/internal/core/model"
	"github.com/TruongHoang2004/Hexta/services/api/internal/core/service"
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
	topic := order.OrderConfirmedTopic
	conn, err := kafka.DialContext(ctx, "tcp", broker)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	if err = conn.CreateTopics(kafka.TopicConfig{Topic: topic, NumPartitions: 1, ReplicationFactor: 1}); err != nil && !errors.Is(err, kafka.TopicAlreadyExists) {
		t.Fatal(err)
	}

	event := orderEvent(t, uuid.NewString())
	if err = db.Create(event).Error; err != nil {
		t.Fatal(err)
	}
	publisher := kafkatransport.NewMultiTopicPublisher([]string{broker})
	// The shared default transport caches topic metadata for seconds; a writer from an earlier test in this
	// process would not yet know a newly created topic and the single relay attempt would fail as unknown topic.
	publisher.Writer.Transport = &kafka.Transport{}
	defer publisher.Writer.Close()
	claimed, failure := service.NewOutboxRelay(repository.NewOutboxRepository(db), publisher).RelayOne(ctx)
	if failure != nil || !claimed {
		t.Fatal("relay", claimed, failure)
	}

	// The topic is shared across runs: read until this run's envelope, keyed by the order aggregate.
	reader := kafka.NewReader(kafka.ReaderConfig{Brokers: []string{broker}, Topic: topic, MaxBytes: 1 << 20})
	defer reader.Close()
	for {
		message, err := reader.ReadMessage(ctx)
		if err != nil {
			t.Fatal(err)
		}
		var envelope events.Envelope
		if json.Unmarshal(message.Value, &envelope) != nil || envelope.EventID != event.ID {
			continue
		}
		var got order.OrderConfirmed
		if string(message.Key) != "order" || envelope.SchemaVersion != 1 || envelope.EventType != model.DomainEventOrderConfirmed || json.Unmarshal(envelope.Data, &got) != nil || got.Validate() != nil {
			t.Fatalf("unexpected message key=%s envelope=%+v", message.Key, envelope)
		}
		break
	}
	var stored model.DomainEvent
	if err = db.First(&stored, "id = ?", event.ID).Error; err != nil || stored.State != model.DomainEventStatePublished {
		t.Fatal("event not marked published", stored.State, err)
	}
}
