package kafka

import (
	"context"
	"encoding/json"

	"github.com/segmentio/kafka-go"
	"gitlab.com/ecommercehub1/order/config"
	"gitlab.com/ecommercehub1/order/internal/common/log"
)

type Producer interface {
	Publish(ctx context.Context, topic string, key string, message interface{}) error
	Close() error
}

type kafkaProducer struct {
	writer *kafka.Writer
}

func NewProducer(cfg *config.Config) Producer {
	if len(cfg.Kafka.Brokers) == 0 {
		log.Error(context.Background(), "Kafka brokers not configured")
	}

	w := &kafka.Writer{
		Addr:                   kafka.TCP(cfg.Kafka.Brokers...),
		Balancer:               &kafka.LeastBytes{},
		AllowAutoTopicCreation: true,
	}

	return &kafkaProducer{writer: w}
}

func (p *kafkaProducer) Publish(ctx context.Context, topic string, key string, message interface{}) error {
	bytes, err := json.Marshal(message)
	if err != nil {
		return err
	}

	err = p.writer.WriteMessages(ctx,
		kafka.Message{
			Topic: topic,
			Key:   []byte(key),
			Value: bytes,
		},
	)
	if err != nil {
		log.Error(ctx, "failed to write message to kafka", "error", err, "topic", topic)
		return err
	}

	return nil
}

func (p *kafkaProducer) Close() error {
	return p.writer.Close()
}
