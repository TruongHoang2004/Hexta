package bootstrap

import (
	"context"

	"gitlab.com/ecommercehub1/lib/pkg/telemetry"
	"gitlab.com/ecommercehub1/order/config"
	"go.uber.org/fx"
)

func TelemetryModule() fx.Option {
	return fx.Module("telemetry",
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
}
