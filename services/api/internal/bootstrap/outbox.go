package bootstrap

import (
	"context"
	"strings"
	"time"

	"github.com/TruongHoang2004/Hexta/packages/shared/kafkatransport"
	"github.com/TruongHoang2004/Hexta/services/api/config"
	"github.com/TruongHoang2004/Hexta/services/api/internal/core/service"
	"github.com/TruongHoang2004/Hexta/services/api/internal/core/types"
	"github.com/TruongHoang2004/Hexta/services/api/internal/repository"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

const outboxPollInterval = time.Second

func BuildOutbox() fx.Option {
	return fx.Options(
		fx.Provide(fx.Annotate(repository.NewOutboxRepository, fx.As(new(repository.IOutboxRepository)))),
		fx.Invoke(registerOutboxRelay),
	)
}

// registerOutboxRelay runs the relay for the process lifetime. Without brokers it stays off and events remain pending.
func registerOutboxRelay(lc fx.Lifecycle, cfg *config.Config, store repository.IOutboxRepository, logger *zap.SugaredLogger) {
	brokers := strings.TrimSpace(cfg.Kafka.Brokers)
	if brokers == "" {
		logger.Info("Outbox relay disabled: kafka.brokers is not configured; domain events stay pending")
		return
	}
	publisher := kafkatransport.NewTopicPublisher(strings.Split(brokers, ","), types.OrderEventsTopic)
	relay := service.NewOutboxRelay(store, publisher)
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	lc.Append(fx.Hook{
		OnStart: func(context.Context) error {
			go func() {
				defer close(done)
				runOutboxRelay(ctx, relay, logger)
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
			return publisher.Writer.Close()
		},
	})
}

// runOutboxRelay drains due events, then polls. Errors are logged without payloads and retried by lease expiry/backoff.
func runOutboxRelay(ctx context.Context, relay *service.OutboxRelay, logger *zap.SugaredLogger) {
	for {
		claimed, err := relay.RelayOne(ctx)
		if err != nil && ctx.Err() == nil {
			logger.Warnw("Outbox relay attempt failed", "detail", err.GetDetail())
		}
		if claimed && err == nil {
			continue
		}
		select {
		case <-ctx.Done():
			return
		case <-time.After(outboxPollInterval):
		}
	}
}
