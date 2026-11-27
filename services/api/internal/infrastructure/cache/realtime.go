package cache

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/TruongHoang2004/Hexta/services/api/internal/core/types"
	"github.com/redis/go-redis/v9"
)

// RealtimeBus carries best-effort conversation hints over Redis pub/sub. Redis is never the authority for these
// events: a lost hint is repaired by the client refetching on reconnect.
type RealtimeBus struct{ client *redis.Client }

func NewRealtimeBus(client *redis.Client) *RealtimeBus { return &RealtimeBus{client: client} }

// NewRealtimeBusFromRedis shares the API's existing Redis connection.
func NewRealtimeBusFromRedis(r *RedisClient) *RealtimeBus { return NewRealtimeBus(r.Client) }

// NewRealtimeBusForAddr dials Redis directly for processes that do not load the API configuration.
func NewRealtimeBusForAddr(addr string) *RealtimeBus {
	return NewRealtimeBus(redis.NewClient(&redis.Options{Addr: addr, DB: 0, PoolSize: 4}))
}

func (b *RealtimeBus) Publish(ctx context.Context, tenant string, hint types.RealtimeHint) error {
	payload, err := json.Marshal(hint)
	if err != nil {
		return err
	}
	return b.client.Publish(ctx, types.RealtimeChannel(tenant), payload).Err()
}

// Receive blocks until ctx ends, calling handle for every hint of every tenant. The go-redis subscription
// reconnects by itself after a connection loss.
func (b *RealtimeBus) Receive(ctx context.Context, handle func(tenant string, hint types.RealtimeHint)) error {
	sub := b.client.PSubscribe(ctx, types.RealtimeChannelPrefix+"*")
	defer sub.Close()
	if _, err := sub.Receive(ctx); err != nil {
		return err
	}
	ch := sub.Channel()
	for {
		select {
		case <-ctx.Done():
			return nil
		case m, ok := <-ch:
			if !ok {
				return nil
			}
			var hint types.RealtimeHint
			if json.Unmarshal([]byte(m.Payload), &hint) != nil || hint.ThreadID == "" {
				continue
			}
			handle(strings.TrimPrefix(m.Channel, types.RealtimeChannelPrefix), hint)
		}
	}
}

func (b *RealtimeBus) Close() error { return b.client.Close() }
