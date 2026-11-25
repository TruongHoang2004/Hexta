package repository

import (
	"context"
	stderrors "errors"
	"time"

	apperrors "github.com/TruongHoang2004/Hexta/packages/shared/pkg/errors"
	"github.com/TruongHoang2004/Hexta/services/api/internal/core/types"
	channelmap "github.com/TruongHoang2004/Hexta/services/api/internal/mapper/channel/persistencemap"
	"github.com/TruongHoang2004/Hexta/services/api/internal/mapper/messaging/persistencemap"
	"github.com/TruongHoang2004/Hexta/services/api/internal/repository/messagingmodel"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// TurnExpiryStore is the persistence used by the turn expiry job. Lock order matches every other messaging writer:
// thread, then pending outbound intents (the delivery consumer locks those first), then the turn.
type TurnExpiryStore interface {
	// ExpiryCandidates lists open turns that are neither flagged for a person nor holding a bot proposal waiting for
	// review, whose recorded activity is at or before cutoff, oldest first.
	// It takes no locks; every candidate is rechecked under the thread lock.
	ExpiryCandidates(context.Context, time.Time, int) ([]types.TurnRef, *apperrors.Error)
	WithinExpiry(context.Context, func(TurnExpiryStore) *apperrors.Error) *apperrors.Error
	// LockThreadSkipLocked locks the thread unless another transaction holds it; false means skip this round.
	LockThreadSkipLocked(context.Context, string, string) (bool, *apperrors.Error)
	Turn(context.Context, string, string) (*types.MessagingTurn, *apperrors.Error)
	// LockPendingIntents locks the turn's outbound intents that still have no outcome.
	LockPendingIntents(context.Context, string, string) ([]types.OutboundIntent, *apperrors.Error)
	// CloseTurn closes the turn only if it is still exactly the observed open, unflagged state (no needs_human, no
	// proposal waiting for review); false otherwise.
	CloseTurn(context.Context, types.MessagingTurn, time.Time) (bool, *apperrors.Error)
	// CancelIntents cancels pending intents and attempts; their still-sending messages become unknown.
	CancelIntents(context.Context, string, []string, time.Time) *apperrors.Error
	Audit(context.Context, string, string, string, string, string, any) *apperrors.Error
}

func (r *MessagingRepository) ExpiryCandidates(ctx context.Context, cutoff time.Time, limit int) ([]types.TurnRef, *apperrors.Error) {
	rows := []messagingmodel.Turn{}
	if err := r.db.WithContext(ctx).Select("tenant_id", "thread_id", "id").
		Where("lifecycle = ? AND needs_human = FALSE AND review_pending = FALSE AND last_activity_at <= ?", types.TurnOpen, cutoff).
		Order("last_activity_at").Limit(limit).Find(&rows).Error; err != nil {
		return nil, salesDBError(ctx, err)
	}
	refs := make([]types.TurnRef, len(rows))
	for i, row := range rows {
		refs[i] = types.TurnRef{TenantID: row.TenantID, ThreadID: row.ThreadID, TurnID: row.ID}
	}
	return refs, nil
}

func (r *MessagingRepository) WithinExpiry(ctx context.Context, fn func(TurnExpiryStore) *apperrors.Error) *apperrors.Error {
	return r.Within(ctx, func(tx MessagingStore) *apperrors.Error { return fn(tx.(*MessagingRepository)) })
}

func (r *MessagingRepository) LockThreadSkipLocked(ctx context.Context, tenant, thread string) (bool, *apperrors.Error) {
	var row messagingmodel.Thread
	err := r.db.WithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE", Options: "SKIP LOCKED"}).
		Where("tenant_id = ? AND id = ?", tenant, thread).First(&row).Error
	if stderrors.Is(err, gorm.ErrRecordNotFound) {
		return false, nil
	}
	if err != nil {
		return false, salesDBError(ctx, err)
	}
	return true, nil
}

func (r *MessagingRepository) LockPendingIntents(ctx context.Context, tenant, turn string) ([]types.OutboundIntent, *apperrors.Error) {
	rows := []messagingmodel.OutboundIntent{}
	if err := r.db.WithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE", Table: clause.Table{Name: "outbound_intents"}}).
		Joins("JOIN messaging_messages m ON m.tenant_id = outbound_intents.tenant_id AND m.id = outbound_intents.message_id").
		Where("outbound_intents.tenant_id = ? AND m.turn_id = ? AND outbound_intents.status = ?", tenant, turn, types.IntentPending).
		Order("outbound_intents.id").Find(&rows).Error; err != nil {
		return nil, salesDBError(ctx, err)
	}
	result := make([]types.OutboundIntent, len(rows))
	for i, row := range rows {
		result[i] = channelmap.Intent(row)
	}
	return result, nil
}

func (r *MessagingRepository) CloseTurn(ctx context.Context, turn types.MessagingTurn, now time.Time) (bool, *apperrors.Error) {
	closed := turn
	closed.Lifecycle, closed.ClosedAt = types.TurnClosed, &now
	closed.Version++
	closed.ControlGeneration++
	row := persistencemap.TurnRow(closed)
	result := r.db.WithContext(ctx).Model(&messagingmodel.Turn{}).
		Where("tenant_id = ? AND id = ? AND version = ? AND lifecycle = ? AND needs_human = FALSE AND review_pending = FALSE AND last_activity_at = ?", turn.TenantID, turn.ID, turn.Version, types.TurnOpen, turn.LastActivityAt).
		Select("lifecycle", "closed_at", "version", "control_generation").Updates(&row)
	if result.Error != nil {
		return false, salesDBError(ctx, result.Error)
	}
	if result.RowsAffected != 1 {
		return false, nil
	}
	r.markConversation(turn.TenantID, turn.ThreadID)
	return true, nil
}

func (r *MessagingRepository) CancelIntents(ctx context.Context, tenant string, ids []string, now time.Time) *apperrors.Error {
	if len(ids) == 0 {
		return nil
	}
	db := r.db.WithContext(ctx)
	if err := db.Model(&messagingmodel.OutboundAttempt{}).Where("tenant_id = ? AND intent_id IN ? AND status = ?", tenant, ids, types.IntentPending).
		Updates(map[string]any{"status": types.IntentCancelled, "completed_at": now}).Error; err != nil {
		return salesDBError(ctx, err)
	}
	if err := db.Model(&messagingmodel.OutboundIntent{}).Where("tenant_id = ? AND id IN ? AND status = ?", tenant, ids, types.IntentPending).
		Updates(map[string]any{"status": types.IntentCancelled, "updated_at": now}).Error; err != nil {
		return salesDBError(ctx, err)
	}
	if err := db.Model(&messagingmodel.Message{}).
		Where("tenant_id = ? AND delivery_status = ? AND id IN (?)", tenant, types.DeliverySending,
			db.Model(&messagingmodel.OutboundIntent{}).Select("message_id").Where("tenant_id = ? AND id IN ?", tenant, ids)).
		Update("delivery_status", types.DeliveryUnknown).Error; err != nil {
		return salesDBError(ctx, err)
	}
	return nil
}

var _ TurnExpiryStore = (*MessagingRepository)(nil)
