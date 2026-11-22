package repository

import (
	"context"
	stderrors "errors"
	apperrors "github.com/TruongHoang2004/Hexta/packages/shared/pkg/errors"
	"github.com/TruongHoang2004/Hexta/services/api/internal/core/model"
	"github.com/TruongHoang2004/Hexta/services/api/internal/core/types"
	channelmap "github.com/TruongHoang2004/Hexta/services/api/internal/mapper/channel/persistencemap"
	"github.com/TruongHoang2004/Hexta/services/api/internal/mapper/messaging/persistencemap"
	"github.com/TruongHoang2004/Hexta/services/api/internal/repository/messagingmodel"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type MessagingStore interface {
	Within(context.Context, func(MessagingStore) *apperrors.Error) *apperrors.Error
	Scope(context.Context, string, string) (types.MessagingScope, *apperrors.Error)
	MemberExists(context.Context, string, string) (bool, *apperrors.Error)
	Thread(context.Context, string, string, bool) (*types.MessagingThread, *apperrors.Error)
	UpsertThread(context.Context, types.MessagingThread) (*types.MessagingThread, *apperrors.Error)
	CurrentTurn(context.Context, string, string) (*types.MessagingTurn, *apperrors.Error)
	Turn(context.Context, string, string) (*types.MessagingTurn, *apperrors.Error)
	SaveTurn(context.Context, types.MessagingTurn, bool) *apperrors.Error
	SaveThread(context.Context, types.MessagingThread) *apperrors.Error
	MessageByKey(context.Context, string, string) (*types.MessagingMessage, *apperrors.Error)
	InsertMessage(context.Context, types.MessagingMessage) *apperrors.Error
	Audit(context.Context, string, string, string, string, string, any) *apperrors.Error
	BindThreadChannel(context.Context, string, string, string, string) *apperrors.Error
	SendConnection(context.Context, string, string) (*types.ChannelConnection, *apperrors.Error)
	InsertOutbound(context.Context, types.OutboundIntent, types.OutboundAttempt, *model.DomainEvent) *apperrors.Error
}
type MessagingRepository struct{ *SalesRepository }

func NewMessagingRepository(db *gorm.DB) *MessagingRepository {
	return &MessagingRepository{NewSalesRepository(db)}
}
func (r *MessagingRepository) Within(ctx context.Context, fn func(MessagingStore) *apperrors.Error) *apperrors.Error {
	var failure *apperrors.Error
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		failure = fn(NewMessagingRepository(tx))
		if failure != nil {
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
func (r *MessagingRepository) MemberExists(ctx context.Context, tenant, actor string) (bool, *apperrors.Error) {
	member, err := r.Member(ctx, tenant, actor)
	return member != nil, err
}
func (r *MessagingRepository) Thread(ctx context.Context, tenant, id string, lock bool) (*types.MessagingThread, *apperrors.Error) {
	var row messagingmodel.Thread
	q := r.db.WithContext(ctx).Where("tenant_id = ? AND id = ?", tenant, id)
	if lock {
		q = q.Clauses(clause.Locking{Strength: "UPDATE"})
	}
	if err := q.First(&row).Error; err != nil {
		return nil, salesDBError(ctx, err)
	}
	value := persistencemap.Thread(row)
	return &value, nil
}
func (r *MessagingRepository) UpsertThread(ctx context.Context, value types.MessagingThread) (*types.MessagingThread, *apperrors.Error) {
	row := persistencemap.ThreadRow(value)
	if err := r.db.WithContext(ctx).Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "tenant_id"}, {Name: "external_key"}}, DoNothing: true}).Create(&row).Error; err != nil {
		return nil, salesDBError(ctx, err)
	}
	row = messagingmodel.Thread{}
	if err := r.db.WithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).Where("tenant_id = ? AND external_key = ?", value.TenantID, value.ExternalKey).First(&row).Error; err != nil {
		return nil, salesDBError(ctx, err)
	}
	value = persistencemap.Thread(row)
	return &value, nil
}
func (r *MessagingRepository) CurrentTurn(ctx context.Context, tenant, thread string) (*types.MessagingTurn, *apperrors.Error) {
	var row messagingmodel.Turn
	err := r.db.WithContext(ctx).Where("tenant_id = ? AND thread_id = ? AND lifecycle = 'open'", tenant, thread).First(&row).Error
	if stderrors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, salesDBError(ctx, err)
	}
	value := persistencemap.Turn(row)
	return &value, nil
}
func (r *MessagingRepository) Turn(ctx context.Context, tenant, id string) (*types.MessagingTurn, *apperrors.Error) {
	var row messagingmodel.Turn
	if err := r.db.WithContext(ctx).Where("tenant_id = ? AND id = ?", tenant, id).First(&row).Error; err != nil {
		return nil, salesDBError(ctx, err)
	}
	value := persistencemap.Turn(row)
	return &value, nil
}
func (r *MessagingRepository) SaveTurn(ctx context.Context, value types.MessagingTurn, create bool) *apperrors.Error {
	row := persistencemap.TurnRow(value)
	if create {
		if err := r.db.WithContext(ctx).Create(&row).Error; err != nil {
			return salesDBError(ctx, err)
		}
		return nil
	}
	result := r.db.WithContext(ctx).Model(&messagingmodel.Turn{}).Where("tenant_id = ? AND id = ? AND version = ?", value.TenantID, value.ID, value.Version-1).Select("assignee_id", "control", "needs_human", "lifecycle", "context_revision", "control_generation", "version", "last_activity_at").Updates(&row)
	if result.Error != nil {
		return salesDBError(ctx, result.Error)
	}
	if result.RowsAffected != 1 {
		return apperrors.ErrConflict(ctx, "turn", "changed; refresh before retrying")
	}
	return nil
}
func (r *MessagingRepository) SaveThread(ctx context.Context, value types.MessagingThread) *apperrors.Error {
	if err := r.db.WithContext(ctx).Model(&messagingmodel.Thread{}).Where("tenant_id = ? AND id = ?", value.TenantID, value.ID).Update("updated_at", value.UpdatedAt).Error; err != nil {
		return salesDBError(ctx, err)
	}
	return nil
}
func (r *MessagingRepository) MessageByKey(ctx context.Context, tenant, key string) (*types.MessagingMessage, *apperrors.Error) {
	var row messagingmodel.Message
	err := r.db.WithContext(ctx).Where("tenant_id = ? AND request_key = ?", tenant, key).First(&row).Error
	if stderrors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, salesDBError(ctx, err)
	}
	value := persistencemap.Message(row)
	return &value, nil
}
func (r *MessagingRepository) InsertMessage(ctx context.Context, value types.MessagingMessage) *apperrors.Error {
	row := persistencemap.MessageRow(value)
	if err := r.db.WithContext(ctx).Create(&row).Error; err != nil {
		return salesDBError(ctx, err)
	}
	return nil
}

// BindThreadChannel records the provider recipient once; an existing binding is never reassigned.
func (r *MessagingRepository) BindThreadChannel(ctx context.Context, tenant, thread, connection, recipient string) *apperrors.Error {
	if err := r.db.WithContext(ctx).Model(&messagingmodel.Thread{}).Where("tenant_id = ? AND id = ? AND connection_id IS NULL", tenant, thread).
		Updates(map[string]any{"connection_id": connection, "recipient_id": recipient}).Error; err != nil {
		return salesDBError(ctx, err)
	}
	return nil
}

// SendConnection share-locks the binding so a concurrent pause/disconnect serializes with intent creation.
func (r *MessagingRepository) SendConnection(ctx context.Context, tenant, id string) (*types.ChannelConnection, *apperrors.Error) {
	var row messagingmodel.Connection
	if err := r.db.WithContext(ctx).Clauses(clause.Locking{Strength: "SHARE"}).Where("tenant_id = ? AND id = ?", tenant, id).First(&row).Error; err != nil {
		return nil, salesDBError(ctx, err)
	}
	value := channelmap.Connection(row)
	return &value, nil
}

// InsertOutbound commits the intent, its single attempt and the outbox event with the outbound message.
func (r *MessagingRepository) InsertOutbound(ctx context.Context, intent types.OutboundIntent, attempt types.OutboundAttempt, event *model.DomainEvent) *apperrors.Error {
	intentRow := channelmap.IntentRow(intent)
	attemptRow := channelmap.AttemptRow(attempt)
	db := r.db.WithContext(ctx)
	for _, row := range []any{&intentRow, &attemptRow, event} {
		if err := db.Create(row).Error; err != nil {
			return salesDBError(ctx, err)
		}
	}
	return nil
}
