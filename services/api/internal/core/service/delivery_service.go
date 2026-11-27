package service

import (
	"context"
	"time"

	"github.com/TruongHoang2004/Hexta/packages/shared/events/connection"
	"github.com/TruongHoang2004/Hexta/packages/shared/events/delivery"
	apperrors "github.com/TruongHoang2004/Hexta/packages/shared/pkg/errors"
	"github.com/TruongHoang2004/Hexta/services/api/internal/core/types"
	"github.com/TruongHoang2004/Hexta/services/api/internal/repository"
)

// DeliveryService applies connector outcomes. Results are deduplicated by attempt state, so Kafka redelivery
// cannot regress a recorded outcome, and no result ever triggers an automatic resend.
type DeliveryService struct {
	store     repository.DeliveryStore
	publisher RealtimePublisher
}

func NewDeliveryService(store repository.DeliveryStore) *DeliveryService {
	return &DeliveryService{store: store}
}

// WithPublisher announces committed delivery changes (sent, failed) to connected clients.
func (s *DeliveryService) WithPublisher(publisher RealtimePublisher) *DeliveryService {
	s.publisher = publisher
	return s
}

func (s *DeliveryService) ConsumeResult(ctx context.Context, e delivery.SendCompleted) *apperrors.Error {
	if e.Validate() != nil {
		return apperrors.ErrBadRequest(ctx).SetDetail("invalid outbound result")
	}
	var thread, messageID string
	err := s.store.Within(ctx, func(r repository.DeliveryStore) *apperrors.Error {
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
		if attempt.Status != types.IntentPending && !(cancelled && e.Status != delivery.StatusFailed) {
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
		if s.publisher != nil {
			messageID = intent.MessageID
			if thread, err = r.ThreadOfMessage(ctx, e.TenantID, intent.MessageID); err != nil {
				return err
			}
		}
		// Only a confirmed send is turn activity, measured from the provider outcome time (never in the future).
		// Failed and unknown outcomes, and redelivered results (filtered above), never extend the turn.
		if e.Status == delivery.StatusSent {
			activity := at
			if now := time.Now().UTC(); activity.After(now) {
				activity = now
			}
			if err := r.RecordSendActivity(ctx, e.TenantID, intent.MessageID, activity); err != nil {
				return err
			}
		}
		if e.ErrorCategory == delivery.ErrorAuthExpired {
			return r.MarkAuthExpired(ctx, e.TenantID, e.ConnectionID, attempt.ConnectionVersion, e.ErrorCode, at)
		}
		return nil
	})
	if err == nil && thread != "" && s.publisher != nil {
		// Best effort after commit; clients resynchronize on reconnect.
		update := &types.MessagingMessage{ID: messageID, ThreadID: thread, DeliveryStatus: e.Status, DeliveryErrorReason: e.ErrorCategory}
		_ = s.publisher.Publish(ctx, e.TenantID, types.RealtimeHint{ThreadID: thread, Reason: types.RealtimeDeliveryChanged, Message: update})
	}
	return err
}

func (s *DeliveryService) ConsumeHealth(ctx context.Context, e connection.HealthChanged) *apperrors.Error {
	if e.Validate() != nil {
		return apperrors.ErrBadRequest(ctx).SetDetail("invalid connection health event")
	}
	return s.store.Within(ctx, func(r repository.DeliveryStore) *apperrors.Error {
		return r.MarkAuthExpired(ctx, e.TenantID, e.ConnectionID, e.ConnectionVersion, e.ErrorCode, e.OccurredAt.UTC())
	})
}
