package bootstrap

import (
	"context"

	apperrors "github.com/TruongHoang2004/Hexta/packages/shared/pkg/errors"
	"github.com/TruongHoang2004/Hexta/services/api/internal/core/service"
	"github.com/TruongHoang2004/Hexta/services/api/internal/core/types"
	"github.com/TruongHoang2004/Hexta/services/api/internal/repository"
)

// messagingFacade composes the read and write services so scenario tests can drive both sides.
// Conversations first rebuilds the read model, standing in for the asynchronous projector.
type messagingFacade struct {
	*service.MessagingCommandService
	*service.MessagingQueryService
	views repository.ConversationViewStore
}

func newMessagingFacade(repo *repository.MessagingRepository) *messagingFacade {
	return &messagingFacade{service.NewMessagingCommandService(repo), service.NewMessagingQueryService(repo), repo.ConversationViews()}
}

func (f *messagingFacade) Conversations(ctx context.Context, tenant, actor string, input types.MessagingQuery) (types.MessagingPage, *apperrors.Error) {
	if err := f.views.Rebuild(ctx, tenant); err != nil {
		return types.MessagingPage{}, err
	}
	return f.MessagingQueryService.Conversations(ctx, tenant, actor, input)
}
