package bootstrap

import (
	"gitlab.com/ecommercehub1/api/internal/present/http/controller"
	"gitlab.com/ecommercehub1/api/internal/present/http/validator"

	"go.uber.org/fx"
)

func BuildController() fx.Option {
	return fx.Options(
		fx.Provide(controller.NewBaseController),
		fx.Provide(controller.NewAuthController),
		fx.Provide(controller.NewUserController),
		fx.Provide(controller.NewHealthController),
	)
}

func BuildValidator() fx.Option {
	return fx.Options(
		fx.Provide(validator.NewValidator),
		fx.Invoke(validator.RegisterDecimalTypeFunc),
		fx.Invoke(validator.RegisterValidations),
	)
}
