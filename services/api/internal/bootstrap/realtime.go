package bootstrap

import (
	"context"

	"github.com/TruongHoang2004/Hexta/services/api/internal/core/service"
	"github.com/TruongHoang2004/Hexta/services/api/internal/infrastructure/cache"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

// BuildRealtime owns the SSE fan-out hub and its Redis subscription for the API process lifetime.
func BuildRealtime() fx.Option {
	return fx.Options(
		fx.Provide(service.NewConversationStreamHub, cache.NewRealtimeBusFromRedis),
		fx.Invoke(registerConversationStreamHub),
	)
}

func registerConversationStreamHub(lc fx.Lifecycle, hub *service.ConversationStreamHub, bus *cache.RealtimeBus, logger *zap.SugaredLogger) {
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	lc.Append(fx.Hook{
		OnStart: func(context.Context) error {
			go func() {
				defer close(done)
				hub.Run(ctx, bus, func(err error) { logger.Warnw("realtime subscription lost; clients will resync", "error", err.Error()) })
			}()
			return nil
		},
		OnStop: func(stopCtx context.Context) error {
			cancel()
			select {
			case <-done:
			case <-stopCtx.Done():
				return stopCtx.Err()
			}
			return nil
		},
	})
}
