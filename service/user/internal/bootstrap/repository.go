package bootstrap

import (
	"gitlab.com/ecommercehub1/user/internal/repository"
	"go.uber.org/fx"
)

func RepositoryModule() fx.Option {
	return fx.Options(
		fx.Provide(repository.NewUserDBRepository),
		fx.Provide(repository.NewUserCacheWrapper),
		fx.Provide(repository.NewAddressRepository),
	)
}
