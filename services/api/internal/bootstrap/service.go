package bootstrap

import (
	"github.com/TruongHoang2004/Hexta/services/api/config"
	"github.com/TruongHoang2004/Hexta/services/api/internal/core/service"
	"go.uber.org/fx"
)

func BuildService() fx.Option {
	return fx.Provide(
		func() *config.Config { return config.AppConfig },
		service.NewBaseService,
		service.NewHealthService,
		service.NewAuthService,
		service.NewTenantService,
		fx.Annotate(
			service.NewInventoryService,
			fx.As(new(service.IInventoryService)),
		),
	)
}
