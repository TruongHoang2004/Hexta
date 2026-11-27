package service

import (
	"context"
	"testing"

	"github.com/TruongHoang2004/Hexta/services/api/internal/core/types"
	"github.com/TruongHoang2004/Hexta/services/api/internal/repository"
	"github.com/google/uuid"
)

func TestCommandsPushStoredMessagesOnlyOnceAndOnlyWhenCommitted(t *testing.T) {
	f := salesTestFixture(t)
	ctx := context.Background()
	pub := &recordingPublisher{}
	s := NewMessagingCommandService(repository.NewMessagingRepository(f.db)).WithPublisher(pub)
	input := types.MessagingInbound{ExternalKey: "customer-1", DisplayName: "Customer", Body: "Hello", RequestKey: uuid.NewString()}
	message, err := s.Inbound(ctx, f.tenant, f.sale, input)
	if err != nil {
		t.Fatal(err)
	}
	if len(pub.hints) != 1 || pub.hints[0].Message == nil || pub.hints[0].Message.ID != message.ID || pub.hints[0].Message.Body != "Hello" || pub.hints[0].ThreadID != message.ThreadID {
		t.Fatalf("a stored message must be pushed with its body: %+v", pub.hints)
	}
	if _, err := s.Inbound(ctx, f.tenant, f.sale, input); err != nil {
		t.Fatal(err)
	}
	if len(pub.hints) != 1 {
		t.Fatalf("an idempotent replay must not push again: %+v", pub.hints)
	}
	changed := input
	changed.Body = "different"
	if _, err := s.Inbound(ctx, f.tenant, f.sale, changed); err == nil {
		t.Fatal("same key with different body must conflict")
	}
	if _, err := s.Inbound(ctx, f.tenant, uuid.NewString(), types.MessagingInbound{ExternalKey: "c2", DisplayName: "C", Body: "x", RequestKey: uuid.NewString()}); err == nil {
		t.Fatal("non-member must be rejected")
	}
	if len(pub.hints) != 1 {
		t.Fatalf("rejected or conflicting commands must not push: %+v", pub.hints)
	}
}
