package catalog

import (
	"regexp"
	"testing"
)

var topicName = regexp.MustCompile(`^hexta\.[a-z]+\.[a-z]+(-[a-z]+)*\.v[1-9][0-9]*$`)

func TestCatalogHasOneUniqueWellNamedTopicPerEvent(t *testing.T) {
	seen := map[string]bool{}
	for _, e := range Entries {
		if !topicName.MatchString(e.Topic) {
			t.Error("topic name", e.Topic)
		}
		if seen[e.Topic] {
			t.Error("duplicate topic", e.Topic)
		}
		seen[e.Topic] = true
		if _, err := e.Decode([]byte(`{}`)); err != nil {
			t.Error("decode", e.Topic, err)
		}
		if got, ok := Lookup(e.Topic); !ok || got.Topic != e.Topic {
			t.Error("lookup", e.Topic)
		}
	}
	if _, ok := Lookup("hexta.unknown.v1"); ok || len(Topics()) != len(Entries) {
		t.Fatal("lookup/topics")
	}
}

func TestDecodedEventsRejectEmptyPayloads(t *testing.T) {
	for _, e := range Entries {
		decoded, _ := e.Decode([]byte(`{}`))
		if decoded.Validate() == nil {
			t.Error("empty payload validated", e.Topic)
		}
	}
}
