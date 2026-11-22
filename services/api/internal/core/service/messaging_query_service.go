package service

import (
	"context"

	apperrors "github.com/TruongHoang2004/Hexta/packages/shared/pkg/errors"
	"github.com/TruongHoang2004/Hexta/services/api/internal/core/types"
	"github.com/TruongHoang2004/Hexta/services/api/internal/repository"
	"github.com/google/uuid"
)

// MessagingQueryService is the read side of messaging: it never mutates state.
type MessagingQueryService struct {
	repo repository.MessagingQueryStore
}

func NewMessagingQueryService(repo repository.MessagingQueryStore) *MessagingQueryService {
	return &MessagingQueryService{repo: repo}
}
func (s *MessagingQueryService) authorized(ctx context.Context, tenant, actor string, fn func(repository.MessagingQueryStore) *apperrors.Error) *apperrors.Error {
	if actor == "" {
		return apperrors.ErrUnauthorized(ctx)
	}
	if _, err := uuid.Parse(tenant); err != nil {
		return apperrors.ErrBadRequest(ctx).SetDetail("valid tenant ID required")
	}
	return s.repo.ReadWithin(ctx, func(r repository.MessagingQueryStore) *apperrors.Error {
		scope, err := r.ReadScope(ctx, tenant, actor)
		if err != nil {
			return err
		}
		if !scope.Active || !scope.Member {
			return apperrors.ErrForbidden(ctx).SetDetail("active tenant membership required")
		}
		return fn(r)
	})
}
func messagingQuery(ctx context.Context, q types.MessagingQuery) *apperrors.Error {
	if q.Limit < 1 || q.Limit > 100 {
		return apperrors.ErrBadRequest(ctx).SetDetail("limit must be between 1 and 100")
	}
	if q.Cursor != "" {
		if _, err := uuid.Parse(q.Cursor); err != nil {
			return apperrors.ErrBadRequest(ctx).SetDetail("invalid cursor")
		}
	}
	switch q.Filter {
	case "", "all", "mine", "unassigned", "needs_human", "bot":
		return nil
	}
	return apperrors.ErrBadRequest(ctx).SetDetail("invalid conversation filter")
}
func (s *MessagingQueryService) Conversations(ctx context.Context, tenant, actor string, input types.MessagingQuery) (types.MessagingPage, *apperrors.Error) {
	var result types.MessagingPage
	if err := messagingQuery(ctx, input); err != nil {
		return result, err
	}
	err := s.authorized(ctx, tenant, actor, func(r repository.MessagingQueryStore) *apperrors.Error {
		var err *apperrors.Error
		result, err = r.Conversations(ctx, tenant, actor, input)
		return err
	})
	return result, err
}
func (s *MessagingQueryService) Detail(ctx context.Context, tenant, actor, id string, input types.MessagingQuery) (types.MessagingDetail, *apperrors.Error) {
	var result types.MessagingDetail
	if err := messagingQuery(ctx, input); err != nil {
		return result, err
	}
	err := s.authorized(ctx, tenant, actor, func(r repository.MessagingQueryStore) *apperrors.Error {
		thread, err := r.Thread(ctx, tenant, id, false)
		if err != nil {
			return err
		}
		turn, err := r.CurrentTurn(ctx, tenant, id)
		if err != nil {
			return err
		}
		result.Conversation = types.MessagingConversation{Thread: *thread, Turn: turn}
		result.Messages, result.NextCursor, err = r.Messages(ctx, tenant, id, input)
		return err
	})
	return result, err
}
