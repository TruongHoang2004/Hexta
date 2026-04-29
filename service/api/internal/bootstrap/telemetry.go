package bootstrap

import (
	"context"

	"gitlab.com/ecommercehub1/api/config"
	"gitlab.com/ecommercehub1/lib/pkg/telemetry"
	"go.uber.org/fx"
)

var TelemetryModule = fx.Module("telemetry",
	fx.Provide(func() (context.Context, context.CancelFunc) {
		return context.WithCancel(context.Background())
	}),
	fx.Invoke(func(lc fx.Lifecycle) {
		lc.Append(fx.Hook{
			OnStart: func(ctx context.Context) error {
				tp, err := telemetry.InitTracer(ctx, config.AppConfig.Server.Name)
				if err != nil {
					return err
				}
				lc.Append(fx.Hook{
					OnStop: func(ctx context.Context) error {
						return tp.Shutdown(ctx)
					},
				})
				return nil
			},
		})
	}),
)
