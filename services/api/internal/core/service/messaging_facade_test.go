package service

import "github.com/TruongHoang2004/Hexta/services/api/internal/repository"

// messagingFacade composes the read and write services so scenario tests can drive both sides.
type messagingFacade struct {
	*MessagingCommandService
	*MessagingQueryService
}

// NewMessagingService builds both sides on one store; stores that wrap only the write side leave queries unusable.
func NewMessagingService(repo repository.MessagingStore) *messagingFacade {
	queries, _ := repo.(repository.MessagingQueryStore)
	return &messagingFacade{NewMessagingCommandService(repo), NewMessagingQueryService(queries)}
}
