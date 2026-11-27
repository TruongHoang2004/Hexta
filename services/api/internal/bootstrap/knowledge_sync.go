package bootstrap

import (
	"context"
	"time"

	"github.com/TruongHoang2004/Hexta/services/api/internal/core/service"
	"github.com/TruongHoang2004/Hexta/services/api/internal/core/types"
	"github.com/TruongHoang2004/Hexta/services/api/internal/infrastructure/ai"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

const (
	// knowledgeSyncInterval is the idle wait between scans of due knowledge sync jobs.
	knowledgeSyncInterval = 2 * time.Second
	// knowledgeSyncDrain bounds one drain so a burst of changes cannot starve shutdown.
	knowledgeSyncDrain = 50
	// knowledgeSyncReportInterval spaces the operator log line about failed or stuck knowledge sync jobs.
	knowledgeSyncReportInterval = 10 * time.Minute
)

// BuildKnowledgeSync runs the AI knowledge sync job in the messaging command process, next to the reply router
// whose AI run client (ai.run_addr) it shares; BuildReplyRouter must be in the same app. Business writes only
// enqueue jobs; several instances may run it, a job is leased with a committed fence.
func BuildKnowledgeSync() fx.Option {
	return fx.Options(
		fx.Provide(
			func(c *ai.RunClient) service.IKnowledgeAIClient { return c },
			service.NewKnowledgeSyncService,
		),
		fx.Invoke(registerKnowledgeSync),
	)
}

func registerKnowledgeSync(lc fx.Lifecycle, sync *service.KnowledgeSyncService, logger *zap.SugaredLogger) {
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	lc.Append(fx.Hook{
		OnStart: func(context.Context) error {
			go func() {
				defer close(done)
				runKnowledgeSync(ctx, sync, logger)
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

// runKnowledgeSync drains due jobs, then waits. Failures are logged without payloads; the job keeps its retry schedule.
func runKnowledgeSync(ctx context.Context, sync *service.KnowledgeSyncService, logger *zap.SugaredLogger) {
	var reported time.Time
	for {
		if time.Since(reported) >= knowledgeSyncReportInterval {
			reported = time.Now()
			reportKnowledgeSync(ctx, sync, logger)
		}
		for i := 0; i < knowledgeSyncDrain && ctx.Err() == nil; i++ {
			claimed, err := sync.SyncOne(ctx)
			if err != nil && ctx.Err() == nil {
				logger.Warnw("Knowledge sync attempt failed", "detail", err.GetDetail())
			}
			if !claimed {
				break
			}
		}
		select {
		case <-ctx.Done():
			return
		case <-time.After(knowledgeSyncInterval):
		}
	}
}

// reportKnowledgeSync logs counts only (never tenants, sources or payloads). Failed jobs need a source change or the
// knowledge-resync command; stuck jobs usually mean the AI service or its embedding model is unavailable.
func reportKnowledgeSync(ctx context.Context, sync *service.KnowledgeSyncService, logger *zap.SugaredLogger) {
	status, err := sync.Status(ctx)
	if err != nil {
		return
	}
	if status.Failed > 0 || status.Stuck > 0 {
		logger.Warnw("Knowledge sync jobs need attention", "failed", status.Failed, "stuck", status.Stuck,
			"pending", status.Pending, "done", status.Done, "stuck_attempts", types.KnowledgeSyncStuckAttempts)
	}
}
