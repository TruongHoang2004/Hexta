package bootstrap

import (
	"gitlab.com/ecommercehub1/catalog/internal/repository"
	"go.uber.org/fx"
)

func RepositoryModule() fx.Option {
	return fx.Options(
		fx.Provide(repository.NewProductRepository),
		fx.Provide(repository.NewCategoryRepository),
	)
}
