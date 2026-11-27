package repository

import (
	"context"
	stderrors "errors"
	"time"

	apperrors "github.com/TruongHoang2004/Hexta/packages/shared/pkg/errors"
	"github.com/TruongHoang2004/Hexta/services/api/internal/core/types"
	"github.com/TruongHoang2004/Hexta/services/api/internal/mapper/channel/persistencemap"
	"github.com/TruongHoang2004/Hexta/services/api/internal/repository/messagingmodel"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// DeliveryStore applies connector delivery results and connection health to API-owned state.
type DeliveryStore interface {
	Within(context.Context, func(DeliveryStore) *apperrors.Error) *apperrors.Error
	Intent(context.Context, string, string) (*types.OutboundIntent, *apperrors.Error)
	Attempt(context.Context, string, string) (*types.OutboundAttempt, *apperrors.Error)
	SaveAttempt(context.Context, types.OutboundAttempt) *apperrors.Error
	SaveIntent(context.Context, types.OutboundIntent) *apperrors.Error
	SetMessageDelivery(context.Context, string, string, string, string) *apperrors.Error
	ThreadOfMessage(context.Context, string, string) (string, *apperrors.Error)
	MarkAuthExpired(context.Context, string, string, int64, string, time.Time) *apperrors.Error
	// RecordSendActivity advances the open turn of a message to at, never backwards and never on a closed turn.
	RecordSendActivity(context.Context, string, string, time.Time) *apperrors.Error
	Quarantine(context.Context, string, string) error
}

type DeliveryRepository struct{ db *gorm.DB }

func NewDeliveryRepository(db *gorm.DB) *DeliveryRepository { return &DeliveryRepository{db} }

func (r *DeliveryRepository) Within(ctx context.Context, fn func(DeliveryStore) *apperrors.Error) *apperrors.Error {
	var failure *apperrors.Error
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if failure = fn(NewDeliveryRepository(tx)); failure != nil {
			return failure
		}
		return nil
	})
	if failure != nil {
		return failure
	}
	if err != nil {
		return salesDBError(ctx, err)
	}
	return nil
}

// Intent locks the intent row; nil means the event names an intent this tenant does not own.
func (r *DeliveryRepository) Intent(ctx context.Context, tenant, id string) (*types.OutboundIntent, *apperrors.Error) {
	var row messagingmodel.OutboundIntent
	err := r.db.WithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).Where("tenant_id = ? AND id = ?", tenant, id).First(&row).Error
	if stderrors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, salesDBError(ctx, err)
	}
	value := persistencemap.Intent(row)
	return &value, nil
}

func (r *DeliveryRepository) Attempt(ctx context.Context, tenant, id string) (*types.OutboundAttempt, *apperrors.Error) {
	var row messagingmodel.OutboundAttempt
	err := r.db.WithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).Where("tenant_id = ? AND id = ?", tenant, id).First(&row).Error
	if stderrors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, salesDBError(ctx, err)
	}
	value := persistencemap.Attempt(row)
	return &value, nil
}

func (r *DeliveryRepository) SaveAttempt(ctx context.Context, value types.OutboundAttempt) *apperrors.Error {
	row := persistencemap.AttemptRow(value)
	if err := r.db.WithContext(ctx).Model(&messagingmodel.OutboundAttempt{}).Where("tenant_id = ? AND id = ?", value.TenantID, value.ID).Select("status", "provider_message_id", "completed_at").Updates(&row).Error; err != nil {
		return salesDBError(ctx, err)
	}
	return nil
}

func (r *DeliveryRepository) SaveIntent(ctx context.Context, value types.OutboundIntent) *apperrors.Error {
	row := persistencemap.IntentRow(value)
	if err := r.db.WithContext(ctx).Model(&messagingmodel.OutboundIntent{}).Where("tenant_id = ? AND id = ?", value.TenantID, value.ID).Select("status", "error_category", "error_code", "updated_at").Updates(&row).Error; err != nil {
		return salesDBError(ctx, err)
	}
	return nil
}

func (r *DeliveryRepository) SetMessageDelivery(ctx context.Context, tenant, id, status, reason string) *apperrors.Error {
	var value *string
	if reason != "" {
		value = &reason
	}
	if err := r.db.WithContext(ctx).Model(&messagingmodel.Message{}).Where("tenant_id = ? AND id = ?", tenant, id).Updates(map[string]any{"delivery_status": status, "delivery_error_reason": value}).Error; err != nil {
		return salesDBError(ctx, err)
	}
	return nil
}

// ThreadOfMessage returns the thread holding a message, or "" when it is unknown.
func (r *DeliveryRepository) ThreadOfMessage(ctx context.Context, tenant, id string) (string, *apperrors.Error) {
	var thread string
	err := r.db.WithContext(ctx).Model(&messagingmodel.Message{}).Where("tenant_id = ? AND id = ?", tenant, id).Select("thread_id").Limit(1).Scan(&thread).Error
	if err != nil {
		return "", salesDBError(ctx, err)
	}
	return thread, nil
}

// MarkAuthExpired moves only an active binding at or below the reported version; paused/disconnected/newer
// (reconnected) bindings are left untouched.
func (r *DeliveryRepository) MarkAuthExpired(ctx context.Context, tenant, id string, version int64, code string, now time.Time) *apperrors.Error {
	var row messagingmodel.Connection
	err := r.db.WithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).Where("tenant_id = ? AND id = ? AND status = ? AND version <= ?", tenant, id, types.ChannelStatusActive, version).First(&row).Error
	if stderrors.Is(err, gorm.ErrRecordNotFound) {
		return nil
	}
	if err != nil {
		return salesDBError(ctx, err)
	}
	if err := r.db.WithContext(ctx).Model(&messagingmodel.Connection{}).Where("tenant_id = ? AND id = ? AND version = ?", tenant, id, row.Version).
		Updates(map[string]any{"status": types.ChannelStatusAuthExpired, "last_error_code": code, "last_error_message": types.ChannelAuthExpiredMessages[row.Platform], "updated_at": now}).Error; err != nil {
		return salesDBError(ctx, err)
	}
	return nil
}

// RecordSendActivity moves only last_activity_at (no turn version), so a confirmation never invalidates the sale's
// next reply. The conversation event is written in the same transaction.
func (r *DeliveryRepository) RecordSendActivity(ctx context.Context, tenant, message string, at time.Time) *apperrors.Error {
	var threads []string
	if err := r.db.WithContext(ctx).Raw(`
UPDATE messaging_turns t SET last_activity_at = ?
FROM messaging_messages m
WHERE m.tenant_id = ? AND m.id = ? AND t.tenant_id = m.tenant_id AND t.id = m.turn_id
  AND t.lifecycle = ? AND t.last_activity_at < ?
RETURNING t.thread_id`, at, tenant, message, types.TurnOpen, at).Scan(&threads).Error; err != nil {
		return salesDBError(ctx, err)
	}
	if len(threads) == 0 {
		return nil
	}
	events := newTxMessagingRepository(r.db)
	events.markConversation(tenant, threads[0])
	return events.flushConversationEvents(ctx)
}

// Quarantine stores only a payload hash and safe reason, reusing the ingress quarantine ledger.
func (r *DeliveryRepository) Quarantine(ctx context.Context, identity, reason string) error {
	return r.db.WithContext(ctx).Exec("INSERT INTO messaging_ingress_quarantine(identity,reason,received_at) VALUES(?,?,?) ON CONFLICT(identity) DO NOTHING", identity, reason, time.Now().UTC()).Error
}

var _ DeliveryStore = (*DeliveryRepository)(nil)
