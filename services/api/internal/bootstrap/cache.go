package bootstrap

import (
	"github.com/TruongHoang2004/Hexta/services/api/internal/infrastructure/cache"
	"go.uber.org/fx"
)

func BuildCache() fx.Option {
	return fx.Provide(
		cache.NewRedisClient,
		cache.NewSessionCache,
		cache.NewTenantCache,
	)
}
