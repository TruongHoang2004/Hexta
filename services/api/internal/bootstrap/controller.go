package bootstrap

import (
	"github.com/TruongHoang2004/Hexta/services/api/internal/present/http/controller"
	"github.com/TruongHoang2004/Hexta/services/api/internal/present/http/validator"

	"go.uber.org/fx"
)

func BuildController() fx.Option {
	return fx.Options(
		fx.Provide(controller.NewBaseController),
		fx.Provide(controller.NewHealthController),
		fx.Provide(controller.NewAuthController),
		fx.Provide(controller.NewTenantController),
	)
}

func BuildValidator() fx.Option {
	return fx.Options(
		fx.Provide(validator.NewValidator),
		fx.Invoke(validator.RegisterDecimalTypeFunc),
		fx.Invoke(validator.RegisterValidations),
	)
}
