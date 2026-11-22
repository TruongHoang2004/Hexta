package service

import (
	"context"
	"errors"
	"testing"
	"time"

	apperrors "github.com/TruongHoang2004/Hexta/packages/shared/pkg/errors"
	"github.com/TruongHoang2004/Hexta/services/api/internal/core/types"
)

func TestHubDeliversOnlyToTheSameTenant(t *testing.T) {
	hub := NewConversationStreamHub()
	a, _ := hub.Subscribe("tenant-a", "u1")
	b, _ := hub.Subscribe("tenant-b", "u2")
	defer a.Close()
	defer b.Close()
	hub.Dispatch("tenant-a", types.RealtimeHint{ThreadID: "t1", Reason: types.RealtimeConversationChanged})
	select {
	case hint := <-a.C:
		if hint.ThreadID != "t1" {
			t.Fatalf("unexpected hint %+v", hint)
		}
	default:
		t.Fatal("tenant-a stream must receive its hint")
	}
	select {
	case hint := <-b.C:
		t.Fatalf("tenant-b must not receive tenant-a hints: %+v", hint)
	default:
	}
}

func TestHubDropsHintsForSlowStreamsWithoutBlocking(t *testing.T) {
	hub := NewConversationStreamHub()
	slow, _ := hub.Subscribe("t", "u")
	defer slow.Close()
	done := make(chan struct{})
	go func() {
		for i := 0; i < 1000; i++ {
			hub.Dispatch("t", types.RealtimeHint{ThreadID: "x"})
		}
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("Dispatch must never block on a slow stream")
	}
}

func TestHubLimitsStreamsPerUserAndReleasesOnClose(t *testing.T) {
	hub := NewConversationStreamHub()
	var subs []*StreamSubscription
	for i := 0; i < MaxStreamsPerUser; i++ {
		sub, err := hub.Subscribe("t", "u")
		if err != nil {
			t.Fatal(err)
		}
		subs = append(subs, sub)
	}
	if _, err := hub.Subscribe("t", "u"); !errors.Is(err, ErrTooManyStreams) {
		t.Fatalf("sixth stream must be refused, got %v", err)
	}
	if _, err := hub.Subscribe("t", "other"); err != nil {
		t.Fatalf("another user is unaffected: %v", err)
	}
	subs[0].Close()
	subs[0].Close()
	if sub, err := hub.Subscribe("t", "u"); err != nil {
		t.Fatalf("closing frees a slot exactly once: %v", err)
	} else {
		sub.Close()
	}
}

type failingSource struct{ calls int }

func (s *failingSource) Receive(ctx context.Context, _ func(string, types.RealtimeHint)) error {
	s.calls++
	return errors.New("redis down")
}

func TestHubAsksStreamsToResyncWhenTheSourceFails(t *testing.T) {
	hub := NewConversationStreamHub()
	sub, _ := hub.Subscribe("t", "u")
	defer sub.Close()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go hub.Run(ctx, &failingSource{}, nil)
	select {
	case hint := <-sub.C:
		if hint.Reason != types.RealtimeResync {
			t.Fatalf("expected resync, got %+v", hint)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("a failed source must trigger a resync so lost hints are repaired")
	}
}

type recordingPublisher struct{ hints []types.RealtimeHint }

func (p *recordingPublisher) Publish(_ context.Context, _ string, hint types.RealtimeHint) error {
	p.hints = append(p.hints, hint)
	return nil
}

type fakeViews struct{ err *apperrors.Error }

func (v fakeViews) Apply(context.Context, string, types.ConversationChanged) *apperrors.Error {
	return v.err
}
func (fakeViews) Rebuild(context.Context, string) *apperrors.Error { return nil }
func (fakeViews) Quarantine(context.Context, string, string) error { return nil }

func TestProjectorPublishesOnlyAfterTheViewIsApplied(t *testing.T) {
	pub := &recordingPublisher{}
	ok := NewConversationProjector(fakeViews{}).WithPublisher(pub)
	if err := ok.Project(context.Background(), "t", types.ConversationChanged{ThreadID: "th"}); err != nil {
		t.Fatal(err)
	}
	if len(pub.hints) != 1 || pub.hints[0].ThreadID != "th" || pub.hints[0].Reason != types.RealtimeConversationChanged {
		t.Fatalf("applied change must publish one hint: %+v", pub.hints)
	}
	failing := NewConversationProjector(fakeViews{err: apperrors.ErrSystemError(context.Background(), "db")}).WithPublisher(pub)
	if err := failing.Project(context.Background(), "t", types.ConversationChanged{ThreadID: "th2"}); err == nil {
		t.Fatal("apply failure must surface for retry")
	}
	if len(pub.hints) != 1 {
		t.Fatalf("a failed apply must not publish: %+v", pub.hints)
	}
}
