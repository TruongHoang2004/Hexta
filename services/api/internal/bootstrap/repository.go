package bootstrap

import (
	"github.com/TruongHoang2004/Hexta/services/api/internal/repository"
	"go.uber.org/fx"
)

func BuildRepository() fx.Option {
	return fx.Provide(
		repository.NewBaseRepository,
		repository.NewIdentityRepository,
		repository.NewSessionDBRepository,
		repository.NewSessionCacheWrapper,
		repository.NewTenantDBRepository,
		repository.NewTenantCacheWrapper,
	)
}
