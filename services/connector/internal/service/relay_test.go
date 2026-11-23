package service

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/TruongHoang2004/Hexta/packages/shared/events"
	"github.com/TruongHoang2004/Hexta/packages/shared/events/messaging"
	"github.com/google/uuid"
	"testing"
	"time"
)

type failingPublisher struct {
	failed bool
	key    string
}

func (p *failingPublisher) Publish(_ context.Context, key string, _ []byte) error {
	p.key = key
	if p.failed {
		return errors.New("broker unavailable")
	}
	return nil
}
func TestPublicationFailureAndRetryIdentity(t *testing.T) {
	e := messaging.MessageReceived{Meta: events.Meta{EventID: uuid.NewString(), CorrelationID: uuid.NewString(), TenantID: "tenant", OccurredAt: time.Now()}, Platform: "facebook", ConnectionID: uuid.NewString(), ConnectionVersion: 1, AccountID: "page", ConversationKey: "thread", SourceIdentity: "message:1", Text: "hello", Sender: "customer", Recipient: "page", Direction: "inbound", SourceAt: time.Now(), Acquisition: "webhook", Purpose: "live"}
	payload, _ := json.Marshal(e)
	p := &failingPublisher{failed: true}
	if Publish(context.Background(), p, payload) == nil {
		t.Fatal("broker failure hidden")
	}
	key := p.key
	p.failed = false
	if Publish(context.Background(), p, payload) != nil || p.key != key || key != e.Key() {
		t.Fatal("retry identity changed")
	}
}
