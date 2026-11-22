// Package catalog lists every event topic so producers, relays and infrastructure provisioning share one source.
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

// Entry describes one topic and decodes its payload.
type Entry struct {
	Topic  string
	Decode func([]byte) (Event, error)
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

// Entries is the complete topic list. Adding an event adds one file in its domain package and one line here.
var Entries = []Entry{
	{messaging.MessageReceivedTopic, decoder[messaging.MessageReceived]()},
	{messaging.ConversationChangedTopic, decoder[messaging.ConversationChanged]()},
	{delivery.SendRequestedTopic, decoder[delivery.SendRequested]()},
	{delivery.SendCompletedTopic, decoder[delivery.SendCompleted]()},
	{connection.HealthChangedTopic, decoder[connection.HealthChanged]()},
	{order.OrderConfirmedTopic, decoder[order.OrderConfirmed]()},
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

// Topics returns every topic name.
func Topics() []string {
	topics := make([]string, len(Entries))
	for i, e := range Entries {
		topics[i] = e.Topic
	}
	return topics
}
