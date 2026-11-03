// Package kafkatransport provides transport only, without business authority.
package kafkatransport

import (
	"context"
	"github.com/TruongHoang2004/Hexta/packages/shared/ingress"
	"github.com/segmentio/kafka-go"
)

type Publisher struct{ Writer *kafka.Writer }

func NewPublisher(brokers []string) *Publisher {
	return &Publisher{&kafka.Writer{Addr: kafka.TCP(brokers...), Topic: ingress.Topic, Balancer: &kafka.Hash{}, RequiredAcks: kafka.RequireAll, Async: false, AllowAutoTopicCreation: false}}
}
func (p *Publisher) Publish(ctx context.Context, key string, value []byte) error {
	return p.Writer.WriteMessages(ctx, kafka.Message{Key: []byte(key), Value: value})
}
func NewReader(brokers []string, group string) *kafka.Reader {
	return kafka.NewReader(kafka.ReaderConfig{Brokers: brokers, GroupID: group, Topic: ingress.Topic, CommitInterval: 0, MaxBytes: 2 << 20})
}
