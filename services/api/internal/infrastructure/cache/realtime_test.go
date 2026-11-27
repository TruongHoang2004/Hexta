package cache

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/TruongHoang2004/Hexta/services/api/internal/core/types"
)

func TestRealtimeBusDeliversHintsToSubscribers(t *testing.T) {
	addr := os.Getenv("API_TEST_REDIS_ADDR")
	if addr == "" {
		t.Skip("API_TEST_REDIS_ADDR is required for Redis pub/sub tests")
	}
	bus := NewRealtimeBusForAddr(addr)
	defer bus.Close()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	type got struct {
		tenant string
		hint   types.RealtimeHint
	}
	received := make(chan got, 4)
	go bus.Receive(ctx, func(tenant string, hint types.RealtimeHint) { received <- got{tenant, hint} })
	want := types.RealtimeHint{ThreadID: "thread-1", Reason: types.RealtimeConversationChanged}
	deadline := time.After(5 * time.Second)
	for {
		// The subscription starts asynchronously; republish until it is live.
		if err := bus.Publish(ctx, "tenant-x", want); err != nil {
			t.Fatal(err)
		}
		select {
		case g := <-received:
			if g.tenant != "tenant-x" || g.hint != want {
				t.Fatalf("unexpected delivery: %+v", g)
			}
			return
		case <-time.After(200 * time.Millisecond):
		case <-deadline:
			t.Fatal("hint was not delivered")
		}
	}
}
