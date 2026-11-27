package bootstrap

import (
	"context"
	"time"

	"github.com/TruongHoang2004/Hexta/services/api/internal/core/service"
	"github.com/TruongHoang2004/Hexta/services/api/internal/repository"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

const turnExpiryInterval = time.Minute

// BuildTurnExpiry runs the 24-hour turn expiry job in the messaging command process, which owns conversation
// writes. Several instances may run it at once; the service serializes per thread with SKIP LOCKED.
func BuildTurnExpiry() fx.Option {
	return fx.Options(
		fx.Provide(
			func(r *repository.MessagingRepository) repository.TurnExpiryStore { return r },
			service.NewTurnExpiryService,
		),
		fx.Invoke(registerTurnExpiry),
	)
}

func registerTurnExpiry(lc fx.Lifecycle, expiry *service.TurnExpiryService, logger *zap.SugaredLogger) {
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	lc.Append(fx.Hook{
		OnStart: func(context.Context) error {
			go func() {
				defer close(done)
				runTurnExpiry(ctx, expiry, logger)
			}()
			return nil
		},
		OnStop: func(stopCtx context.Context) error {
			cancel()
			select {
			case <-done:
				return nil
			case <-stopCtx.Done():
				return stopCtx.Err()
			}
		},
	})
}

// runTurnExpiry drains due turns, then waits for the next interval. Failures are logged without payloads and retried.
func runTurnExpiry(ctx context.Context, expiry *service.TurnExpiryService, logger *zap.SugaredLogger) {
	for {
		for ctx.Err() == nil {
			seen, closed, err := expiry.ExpireDue(ctx)
			if err != nil && ctx.Err() == nil {
				logger.Warnw("Turn expiry run failed", "detail", err.GetDetail())
			}
			if closed > 0 {
				logger.Infow("Closed inactive conversation turns", "closed", closed)
			}
			// Keep draining only while a full batch made progress; skipped (locked/in-flight) turns wait.
			if seen < service.TurnExpiryBatch || closed == 0 {
				break
			}
		}
		select {
		case <-ctx.Done():
			return
		case <-time.After(turnExpiryInterval):
		}
	}
}
