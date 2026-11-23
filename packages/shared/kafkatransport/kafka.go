// Package kafkatransport provides transport only, without business authority. Topic names come from the event
// contracts in packages/shared/events.
package kafkatransport

import (
	"context"
	"time"

	"github.com/segmentio/kafka-go"
)

// publishBatchTimeout bounds how long a synchronous write waits to fill a batch. kafka-go defaults to 1 s, which
// caps every single-message relay at about one event per second.
const publishBatchTimeout = 5 * time.Millisecond

func newWriter(brokers []string, topic string) *kafka.Writer {
	return &kafka.Writer{Addr: kafka.TCP(brokers...), Topic: topic, Balancer: &kafka.Hash{}, RequiredAcks: kafka.RequireAll, Async: false, AllowAutoTopicCreation: false, BatchTimeout: publishBatchTimeout}
}

// Publisher writes to one topic.
type Publisher struct{ Writer *kafka.Writer }

// NewTopicPublisher writes synchronously with acks=all; topics must be provisioned, never auto-created.
func NewTopicPublisher(brokers []string, topic string) *Publisher {
	return &Publisher{newWriter(brokers, topic)}
}
func (p *Publisher) Publish(ctx context.Context, key string, value []byte) error {
	return p.Writer.WriteMessages(ctx, kafka.Message{Key: []byte(key), Value: value})
}

// MultiTopicPublisher writes each message to the topic named per call, with the same guarantees as Publisher.
type MultiTopicPublisher struct{ Writer *kafka.Writer }

func NewMultiTopicPublisher(brokers []string) *MultiTopicPublisher {
	return &MultiTopicPublisher{newWriter(brokers, "")}
}
func (p *MultiTopicPublisher) PublishTo(ctx context.Context, topic, key string, value []byte) error {
	return p.Writer.WriteMessages(ctx, kafka.Message{Topic: topic, Key: []byte(key), Value: value})
}

// NewTopicReader commits offsets explicitly (CommitInterval 0) after the consumer's durable mutation.
func NewTopicReader(brokers []string, group, topic string) *kafka.Reader {
	return kafka.NewReader(kafka.ReaderConfig{Brokers: brokers, GroupID: group, Topic: topic, CommitInterval: 0, MaxBytes: 2 << 20})
}
