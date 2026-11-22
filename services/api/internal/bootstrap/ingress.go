package bootstrap

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/TruongHoang2004/Hexta/packages/shared/events"
	"github.com/TruongHoang2004/Hexta/packages/shared/events/connection"
	"github.com/TruongHoang2004/Hexta/packages/shared/events/delivery"
	"github.com/TruongHoang2004/Hexta/packages/shared/events/messaging"
	"github.com/TruongHoang2004/Hexta/packages/shared/kafkatransport"
	apperrors "github.com/TruongHoang2004/Hexta/packages/shared/pkg/errors"
	"github.com/TruongHoang2004/Hexta/services/api/internal/core/service"
	"github.com/TruongHoang2004/Hexta/services/api/internal/repository"
	"github.com/segmentio/kafka-go"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"os"
	"strings"
	"time"
)

// RunIngress is a dedicated API-owned consumer process, sharing only the API database. It consumes inbound
// messages plus connector delivery results and connection health; any consumer failure stops the process.
func RunIngress(ctx context.Context, db *gorm.DB) error {
	if os.Getenv("KAFKA_BROKERS") == "" {
		return errors.New("KAFKA_BROKERS required")
	}
	brokers := strings.Split(os.Getenv("KAFKA_BROKERS"), ",")
	store := repository.NewIngressRepository(db)
	consumer := service.NewIngressService(store)
	deliveryStore := repository.NewDeliveryRepository(db)
	deliveries := service.NewDeliveryService(deliveryStore)
	views := repository.NewConversationViewRepository(db)
	projector := service.NewConversationProjector(views)
	consumers := []struct {
		group, topic string
		quarantine   func(context.Context, string, string) error
		apply        func([]byte) (bool, func() *apperrors.Error)
	}{
		{"hexta-api-ingress-v1", messaging.MessageReceivedTopic, store.Quarantine, func(value []byte) (bool, func() *apperrors.Error) {
			var e messaging.MessageReceived
			if json.Unmarshal(value, &e) != nil || e.Validate() != nil {
				return false, nil
			}
			return true, func() *apperrors.Error { return consumer.Consume(ctx, e) }
		}},
		{"hexta-api-outbound-result-v1", delivery.SendCompletedTopic, deliveryStore.Quarantine, func(value []byte) (bool, func() *apperrors.Error) {
			var e delivery.SendCompleted
			if json.Unmarshal(value, &e) != nil || e.Validate() != nil {
				return false, nil
			}
			return true, func() *apperrors.Error { return deliveries.ConsumeResult(ctx, e) }
		}},
		{"hexta-api-conversation-view-v1", messaging.ConversationChangedTopic, views.Quarantine, func(value []byte) (bool, func() *apperrors.Error) {
			tenant, change, ok := projector.Decode(value)
			if !ok {
				return false, nil
			}
			if change == nil {
				return true, func() *apperrors.Error { return nil }
			}
			return true, func() *apperrors.Error { return projector.Project(ctx, tenant, *change) }
		}},
		{"hexta-api-connector-health-v1", connection.HealthChangedTopic, deliveryStore.Quarantine, func(value []byte) (bool, func() *apperrors.Error) {
			var e connection.HealthChanged
			if json.Unmarshal(value, &e) != nil || e.Validate() != nil {
				return false, nil
			}
			return true, func() *apperrors.Error { return deliveries.ConsumeHealth(ctx, e) }
		}},
	}
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	errs := make(chan error, len(consumers))
	for _, c := range consumers {
		go func() {
			reader := kafkatransport.NewTopicReader(brokers, c.group, c.topic)
			defer reader.Close()
			errs <- consumeTopic(ctx, reader, c.quarantine, c.apply)
		}()
	}
	err := <-errs
	cancel()
	for i := 1; i < len(consumers); i++ {
		<-errs
	}
	return err
}

// consumeTopic commits each offset only after its durable mutation; malformed envelopes are quarantined by hash.
func consumeTopic(ctx context.Context, reader *kafka.Reader, quarantine func(context.Context, string, string) error, apply func([]byte) (bool, func() *apperrors.Error)) error {
	for {
		m, err := reader.FetchMessage(ctx)
		if err != nil {
			return err
		}
		valid, run := apply(m.Value)
		if !valid {
			if err := quarantine(ctx, events.Hash(string(m.Value)), "invalid_envelope"); err != nil {
				return errors.New("quarantine unavailable")
			}
			if err := reader.CommitMessages(ctx, m); err != nil {
				return err
			}
			continue
		}
		for {
			if run() == nil {
				break
			}
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(time.Second):
			}
		}
		if err = reader.CommitMessages(ctx, m); err != nil {
			return err
		}
	}
}

func RunIngressProcess(ctx context.Context) error {
	dsn := os.Getenv("API_DATABASE_URL")
	if dsn == "" {
		return errors.New("API_DATABASE_URL required")
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		return errors.New("database unavailable")
	}
	sqlDB, err := db.DB()
	if err != nil {
		return errors.New("database unavailable")
	}
	defer sqlDB.Close()
	return RunIngress(ctx, db)
}
