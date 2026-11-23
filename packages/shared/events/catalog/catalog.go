// Package catalog lists every event with its single topic, so producers, the outbox relay and infrastructure
// provisioning share one source.
package catalog

import (
	"encoding/json"

	"github.com/TruongHoang2004/Hexta/packages/shared/events/connection"
	"github.com/TruongHoang2004/Hexta/packages/shared/events/delivery"
	"github.com/TruongHoang2004/Hexta/packages/shared/events/messaging"
	"github.com/TruongHoang2004/Hexta/packages/shared/events/order"
)

// Event is implemented by every event payload.
type Event interface {
	Key() string
	Validate() error
}

// Entry describes one event. EventType is set for events written to the API outbox; Enveloped events are published
// inside events.Envelope, the others as their raw payload.
type Entry struct {
	Topic     string
	EventType string
	Enveloped bool
	Decode    func([]byte) (Event, error)
}

func decoder[T Event]() func([]byte) (Event, error) {
	return func(b []byte) (Event, error) {
		var e T
		if err := json.Unmarshal(b, &e); err != nil {
			return nil, err
		}
		return e, nil
	}
}

// Entries is the complete event list, one topic per event. Adding an event adds one file in its domain package and
// one line here.
var Entries = []Entry{
	{Topic: messaging.MessageReceivedTopic, Decode: decoder[messaging.MessageReceived]()},
	{Topic: messaging.ConversationChangedTopic, EventType: messaging.ConversationChangedEventType, Enveloped: true, Decode: decoder[messaging.ConversationChanged]()},
	{Topic: delivery.SendRequestedTopic, EventType: delivery.SendRequestedEventType, Decode: decoder[delivery.SendRequested]()},
	{Topic: delivery.SendCompletedTopic, Decode: decoder[delivery.SendCompleted]()},
	{Topic: connection.HealthChangedTopic, Decode: decoder[connection.HealthChanged]()},
	{Topic: order.OrderConfirmedTopic, EventType: order.OrderConfirmedEventType, Enveloped: true, Decode: decoder[order.OrderConfirmed]()},
}

// Lookup returns the entry for a topic.
func Lookup(topic string) (Entry, bool) {
	for _, e := range Entries {
		if e.Topic == topic {
			return e, true
		}
	}
	return Entry{}, false
}

// ByEventType returns the entry for an API outbox event type.
func ByEventType(eventType string) (Entry, bool) {
	for _, e := range Entries {
		if e.EventType != "" && e.EventType == eventType {
			return e, true
		}
	}
	return Entry{}, false
}

// Topics returns every topic name.
func Topics() []string {
	topics := make([]string, len(Entries))
	for i, e := range Entries {
		topics[i] = e.Topic
	}
	return topics
}
