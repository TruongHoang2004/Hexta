package service

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/TruongHoang2004/Hexta/packages/shared/ingress"
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
	e := ingress.Event{SchemaVersion: 1, EventID: uuid.NewString(), CorrelationID: "correlation", Platform: "fixture", ConnectionID: uuid.NewString(), ConnectionVersion: 1, TenantID: "tenant", AccountID: "page", ConversationKey: "thread", SourceIdentity: "message:1", Type: "message", Sender: "customer", Recipient: "page", Direction: "inbound", SourceAt: time.Now(), ReceivedAt: time.Now(), Acquisition: "webhook", Purpose: "live"}
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
