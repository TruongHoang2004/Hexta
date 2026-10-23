package bootstrap

import (
	"github.com/TruongHoang2004/Hexta/services/api/internal/repository"
	"go.uber.org/fx"
)

func BuildRepository() fx.Option {
	return fx.Provide(
		fx.Annotate(repository.NewAICatalogRepository, fx.As(new(repository.IAICatalogRepository))),
		repository.NewTenantOnboardingRepository,
		repository.NewWorkflowRepository,
		repository.NewBaseRepository,
		fx.Annotate(repository.NewSalesRepository, fx.As(new(repository.ISalesRepository))),
		repository.NewIdentityRepository,
		repository.NewSessionDBRepository,
		repository.NewSessionCacheWrapper,
		repository.NewTenantDBRepository,
		repository.NewTenantCacheWrapper,
		fx.Annotate(
			repository.NewInventoryDBRepository,
			fx.As(new(repository.IInventoryRepository)),
		),
		fx.Annotate(
			repository.NewOrderDBRepository,
			fx.As(new(repository.IOrderRepository)),
		),
	)
}
