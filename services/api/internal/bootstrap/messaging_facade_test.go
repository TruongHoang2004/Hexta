package bootstrap

import (
	"github.com/TruongHoang2004/Hexta/services/api/internal/core/service"
	"github.com/TruongHoang2004/Hexta/services/api/internal/repository"
)

// messagingFacade composes the read and write services so scenario tests can drive both sides.
type messagingFacade struct {
	*service.MessagingCommandService
	*service.MessagingQueryService
}

func newMessagingFacade(repo *repository.MessagingRepository) *messagingFacade {
	return &messagingFacade{service.NewMessagingCommandService(repo), service.NewMessagingQueryService(repo)}
}
