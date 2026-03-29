package bootstrap

import (
	"gitlab.com/ecommercehub1/catalog/internal/core/service"
	"go.uber.org/fx"
)

func ServiceModule() fx.Option {
	return fx.Options(
		fx.Provide(service.NewProductService),
		fx.Provide(service.NewCategoryService),
	)
}
