package bootstrap

import (
	"gitlab.com/ecommercehub1/api/internal/repository"
	"go.uber.org/fx"
)

func BuildRepository() fx.Option {
	return fx.Provide(
		repository.NewBaseRepository,
		repository.NewIdentityRepository,
		repository.NewSessionDBRepository,
		repository.NewSessionCacheWrapper,
	)
}
