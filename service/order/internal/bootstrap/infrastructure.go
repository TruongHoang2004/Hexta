package bootstrap

import (
	"context"

	"gitlab.com/ecommercehub1/order/config"
	"gitlab.com/ecommercehub1/order/internal/infrastructure/kafka"
	"go.uber.org/fx"
)

func InfrastructureModule() fx.Option {
	return fx.Options(
		fx.Provide(func() *config.Config {
			return config.AppConfig
		}),
		fx.Provide(kafka.NewProducer),
		fx.Provide(kafka.NewConsumer),
		fx.Invoke(startKafkaConsumer),
	)
}

func startKafkaConsumer(lc fx.Lifecycle, consumer kafka.Consumer) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go func() {
				_ = consumer.Start(context.Background())
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			return consumer.Close()
		},
	})
}
