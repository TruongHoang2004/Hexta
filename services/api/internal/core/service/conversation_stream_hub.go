package service

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/TruongHoang2004/Hexta/services/api/internal/core/types"
)

// MaxStreamsPerUser bounds open streams per user and tenant so one account cannot exhaust connections.
const MaxStreamsPerUser = 5

// ErrTooManyStreams is returned when a user already holds MaxStreamsPerUser streams.
var ErrTooManyStreams = errors.New("too many open streams")

// RealtimePublisher publishes best-effort conversation hints; failures never affect the committed change.
type RealtimePublisher interface {
	Publish(ctx context.Context, tenant string, hint types.RealtimeHint) error
}

// RealtimeSource delivers hints from every tenant until ctx ends.
type RealtimeSource interface {
	Receive(ctx context.Context, handle func(tenant string, hint types.RealtimeHint)) error
}

// ConversationStreamHub fans hints out to the SSE streams of this API process. Slow streams drop hints rather
// than block, because clients resynchronize on reconnect.
type ConversationStreamHub struct {
	mu     sync.Mutex
	byTent map[string]map[*StreamSubscription]struct{}
	perUse map[string]int
}

// StreamSubscription is one open client stream. Hints arrive on C; Close must be called exactly when done.
type StreamSubscription struct {
	C      chan types.RealtimeHint
	tenant string
	user   string
	hub    *ConversationStreamHub
	once   sync.Once
}

func NewConversationStreamHub() *ConversationStreamHub {
	return &ConversationStreamHub{byTent: map[string]map[*StreamSubscription]struct{}{}, perUse: map[string]int{}}
}

// Subscribe registers a stream for a verified tenant member.
func (h *ConversationStreamHub) Subscribe(tenant, user string) (*StreamSubscription, error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	key := tenant + "/" + user
	if h.perUse[key] >= MaxStreamsPerUser {
		return nil, ErrTooManyStreams
	}
	sub := &StreamSubscription{C: make(chan types.RealtimeHint, 16), tenant: tenant, user: user, hub: h}
	if h.byTent[tenant] == nil {
		h.byTent[tenant] = map[*StreamSubscription]struct{}{}
	}
	h.byTent[tenant][sub] = struct{}{}
	h.perUse[key]++
	return sub, nil
}

func (s *StreamSubscription) Close() {
	s.once.Do(func() {
		h := s.hub
		h.mu.Lock()
		defer h.mu.Unlock()
		delete(h.byTent[s.tenant], s)
		if len(h.byTent[s.tenant]) == 0 {
			delete(h.byTent, s.tenant)
		}
		key := s.tenant + "/" + s.user
		if h.perUse[key]--; h.perUse[key] <= 0 {
			delete(h.perUse, key)
		}
	})
}

// Dispatch delivers a hint to every stream of the tenant without blocking.
func (h *ConversationStreamHub) Dispatch(tenant string, hint types.RealtimeHint) {
	h.mu.Lock()
	defer h.mu.Unlock()
	for sub := range h.byTent[tenant] {
		select {
		case sub.C <- hint:
		default:
		}
	}
}

// Resync asks every open stream to refetch, used when hints may have been lost.
func (h *ConversationStreamHub) Resync() {
	h.mu.Lock()
	defer h.mu.Unlock()
	for _, subs := range h.byTent {
		for sub := range subs {
			select {
			case sub.C <- types.RealtimeHint{Reason: types.RealtimeResync}:
			default:
			}
		}
	}
}

// Run feeds the hub from source until ctx ends, retrying after source failures.
func (h *ConversationStreamHub) Run(ctx context.Context, source RealtimeSource, onError func(error)) {
	for ctx.Err() == nil {
		if err := source.Receive(ctx, h.Dispatch); err != nil && ctx.Err() == nil {
			h.Resync()
			if onError != nil {
				onError(err)
			}
		}
		select {
		case <-ctx.Done():
		case <-time.After(2 * time.Second):
		}
	}
}
