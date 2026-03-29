package bootstrap

import (
	"gitlab.com/ecommercehub1/merchant/internal/infrastructure/cache"
	"go.uber.org/fx"
)

func CacheModule() fx.Option {
	return fx.Options(
		fx.Provide(cache.NewRedisClient),
		fx.Provide(cache.NewMerchantCache),
	)
}
