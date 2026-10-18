package bootstrap

import (
	"github.com/TruongHoang2004/Hexta/services/api/config"
	"github.com/TruongHoang2004/Hexta/services/api/internal/core/service"
	"github.com/TruongHoang2004/Hexta/services/api/internal/infrastructure/gemini"
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
		fx.Annotate(
			service.NewOrderService,
			fx.As(new(service.IOrderService)),
		),
		fx.Annotate(
			gemini.NewGeminiClient,
			fx.As(new(gemini.IGeminiClient)),
		),
		fx.Annotate(
			service.NewAIAgentService,
			fx.As(new(service.IAIAgentService)),
		),
	)
}
