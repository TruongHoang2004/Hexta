package service

import (
	"context"

	apperrors "github.com/TruongHoang2004/Hexta/packages/shared/pkg/errors"
	"github.com/TruongHoang2004/Hexta/services/api/internal/core/types"
	"github.com/TruongHoang2004/Hexta/services/api/internal/repository"
)

// messagingFacade composes the read and write services so scenario tests can drive both sides.
// Conversations first rebuilds the read model, standing in for the asynchronous projector.
type messagingFacade struct {
	*MessagingCommandService
	*MessagingQueryService
	views repository.ConversationViewStore
}

// NewMessagingService builds both sides on one store; stores that wrap only the write side leave queries unusable.
func NewMessagingService(repo repository.MessagingStore) *messagingFacade {
	queries, _ := repo.(repository.MessagingQueryStore)
	facade := &messagingFacade{MessagingCommandService: NewMessagingCommandService(repo), MessagingQueryService: NewMessagingQueryService(queries)}
	if concrete, ok := repo.(*repository.MessagingRepository); ok {
		facade.views = concrete.ConversationViews()
	}
	return facade
}

func (f *messagingFacade) Conversations(ctx context.Context, tenant, actor string, input types.MessagingQuery) (types.MessagingPage, *apperrors.Error) {
	if f.views != nil {
		if err := f.views.Rebuild(ctx, tenant); err != nil {
			return types.MessagingPage{}, err
		}
	}
	return f.MessagingQueryService.Conversations(ctx, tenant, actor, input)
}
