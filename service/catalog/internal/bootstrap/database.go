package bootstrap

import (
	"gitlab.com/ecommercehub1/catalog/internal/infrastructure/database"
	"go.uber.org/fx"
)

func DatabaseModule() fx.Option {
	return fx.Options(
		fx.Provide(database.NewDatabase),
	)
}
