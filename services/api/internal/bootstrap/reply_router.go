package bootstrap

import (
	"context"
	"time"

	"github.com/TruongHoang2004/Hexta/services/api/internal/core/service"
	"github.com/TruongHoang2004/Hexta/services/api/internal/infrastructure/ai"
	"github.com/TruongHoang2004/Hexta/services/api/internal/repository"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

// replyRouterInterval is the idle wait between scans of due routing jobs.
const replyRouterInterval = time.Second

// BuildReplyRouter runs the bot reply router (MVP 5) in the messaging command process, which owns conversation
// writes. Several instances may run it: every decision is applied under the thread lock with a conditional
// proposal update, so a job is decided once.
func BuildReplyRouter() fx.Option {
	return fx.Options(
		fx.Provide(
			func(r *repository.MessagingRepository) repository.ReplyRouterStore { return r },
			func(s *service.ScriptService) service.ReplyScriptSource { return s },
			ai.NewRunClient,
			func(c *ai.RunClient) service.IReplyAIClient { return c },
			service.NewReplyRouterService,
		),
		fx.Invoke(registerReplyRouter),
	)
}

func registerReplyRouter(lc fx.Lifecycle, router *service.ReplyRouterService, client *ai.RunClient, logger *zap.SugaredLogger) {
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	lc.Append(fx.Hook{
		OnStart: func(context.Context) error {
			go func() {
				defer close(done)
				runReplyRouter(ctx, router, logger)
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
			return client.Close()
		},
	})
}

// runReplyRouter drains due jobs, then waits for the next interval. Failures are logged without payloads and retried.
func runReplyRouter(ctx context.Context, router *service.ReplyRouterService, logger *zap.SugaredLogger) {
	for {
		for ctx.Err() == nil {
			seen, decided, err := router.RouteDue(ctx)
			if err != nil && ctx.Err() == nil {
				logger.Warnw("Reply router run failed", "detail", err.GetDetail())
			}
			if decided > 0 {
				logger.Infow("Routed customer messages", "decided", decided)
			}
			if seen < service.ReplyRouterBatch || decided == 0 {
				break
			}
		}
		select {
		case <-ctx.Done():
			return
		case <-time.After(replyRouterInterval):
		}
	}
}
