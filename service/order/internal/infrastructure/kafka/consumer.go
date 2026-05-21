package kafka

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/segmentio/kafka-go"
	"gitlab.com/ecommercehub1/order/config"
	"gitlab.com/ecommercehub1/order/internal/common/log"
	"gitlab.com/ecommercehub1/order/internal/core/model"
	"gitlab.com/ecommercehub1/order/internal/repository"
)

type Consumer interface {
	Start(ctx context.Context) error
	Close() error
}

type kafkaConsumer struct {
	reader    *kafka.Reader
	orderRepo repository.IOrderRepository
}

func NewConsumer(cfg *config.Config, orderRepo repository.IOrderRepository) Consumer {
	if len(cfg.Kafka.Brokers) == 0 {
		log.Error(context.Background(), "Kafka brokers not configured")
	}

	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  cfg.Kafka.Brokers,
		GroupID:  "order-service-group",
		Topic:    "orders",
		MaxBytes: 10e6, // 10MB
	})

	return &kafkaConsumer{
		reader:    r,
		orderRepo: orderRepo,
	}
}

func (c *kafkaConsumer) Start(ctx context.Context) error {
	log.Info(ctx, "Starting Kafka consumer for orders")
	for {
		m, err := c.reader.ReadMessage(ctx)
		if err != nil {
			log.Error(ctx, "failed to read message from kafka", "error", err)
			return err
		}

		var order model.Order
		if err := json.Unmarshal(m.Value, &order); err != nil {
			log.Error(ctx, "failed to unmarshal order message", "error", err)
			continue
		}

		// Save to Database
		_, dberr := c.orderRepo.Create(ctx, &order)
		if dberr != nil {
			log.Error(ctx, "failed to save order to db", "error", dberr)
			continue
		}

		log.Info(ctx, fmt.Sprintf("Successfully processed order %s", order.ID))
	}
}

func (c *kafkaConsumer) Close() error {
	return c.reader.Close()
}
