package bootstrap

import (
	"gitlab.com/ecommercehub1/merchant/internal/repository"
	"go.uber.org/fx"
)

func RepositoryModule() fx.Option {
	return fx.Options(
		fx.Provide(repository.NewMerchantDBRepository),
		fx.Provide(repository.NewMerchantCacheWrapper),
	)
}
