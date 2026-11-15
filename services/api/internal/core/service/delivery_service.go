package service

import (
	"context"

	"github.com/TruongHoang2004/Hexta/packages/shared/outbound"
	apperrors "github.com/TruongHoang2004/Hexta/packages/shared/pkg/errors"
	"github.com/TruongHoang2004/Hexta/services/api/internal/core/types"
	"github.com/TruongHoang2004/Hexta/services/api/internal/repository"
)

// DeliveryService applies connector outcomes. Results are deduplicated by attempt state, so Kafka redelivery
// cannot regress a recorded outcome, and no result ever triggers an automatic resend.
type DeliveryService struct{ store repository.DeliveryStore }

func NewDeliveryService(store repository.DeliveryStore) *DeliveryService {
	return &DeliveryService{store}
}

func (s *DeliveryService) ConsumeResult(ctx context.Context, e outbound.ResultEvent) *apperrors.Error {
	if e.Validate() != nil {
		return apperrors.ErrBadRequest(ctx).SetDetail("invalid outbound result")
	}
	return s.store.Within(ctx, func(r repository.DeliveryStore) *apperrors.Error {
		intent, err := r.Intent(ctx, e.TenantID, e.IntentID)
		if err != nil || intent == nil || intent.ConnectionID != e.ConnectionID {
			return err
		}
		attempt, err := r.Attempt(ctx, e.TenantID, e.AttemptID)
		if err != nil || attempt == nil || attempt.IntentID != intent.ID {
			return err
		}
		// A cancelled attempt may still have reached the provider; record sent/unknown truthfully.
		cancelled := attempt.Status == types.IntentCancelled
		if attempt.Status != types.IntentPending && !(cancelled && e.Status != outbound.StatusFailed) {
			return nil
		}
		at := e.OccurredAt.UTC()
		attempt.Status, attempt.ProviderMessageID, attempt.CompletedAt = e.Status, e.ProviderMessageID, &at
		if err := r.SaveAttempt(ctx, *attempt); err != nil {
			return err
		}
		intent.Status, intent.ErrorCategory, intent.ErrorCode, intent.UpdatedAt = e.Status, e.ErrorCategory, e.ErrorCode, at
		if err := r.SaveIntent(ctx, *intent); err != nil {
			return err
		}
		if err := r.SetMessageDelivery(ctx, e.TenantID, intent.MessageID, e.Status, e.ErrorCategory); err != nil {
			return err
		}
		if e.ErrorCategory == outbound.ErrorAuthExpired {
			return r.MarkAuthExpired(ctx, e.TenantID, e.ConnectionID, attempt.ConnectionVersion, e.ErrorCode, at)
		}
		return nil
	})
}

func (s *DeliveryService) ConsumeHealth(ctx context.Context, e outbound.ConnectionHealthEvent) *apperrors.Error {
	if e.Validate() != nil {
		return apperrors.ErrBadRequest(ctx).SetDetail("invalid connection health event")
	}
	return s.store.Within(ctx, func(r repository.DeliveryStore) *apperrors.Error {
		return r.MarkAuthExpired(ctx, e.TenantID, e.ConnectionID, e.ConnectionVersion, e.ErrorCode, e.OccurredAt.UTC())
	})
}
