package repository

import (
	"context"
	stderrors "errors"
	apperrors "github.com/TruongHoang2004/Hexta/packages/shared/pkg/errors"
	"github.com/TruongHoang2004/Hexta/services/api/internal/core/model"
	"github.com/TruongHoang2004/Hexta/services/api/internal/core/types"
	"github.com/TruongHoang2004/Hexta/services/api/internal/mapper/messaging/persistencemap"
	"github.com/TruongHoang2004/Hexta/services/api/internal/repository/messagingmodel"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type MessagingStore interface {
	Within(context.Context, func(MessagingStore) *apperrors.Error) *apperrors.Error
	Scope(context.Context, string, string) (types.MessagingScope, *apperrors.Error)
	MemberExists(context.Context, string, string) (bool, *apperrors.Error)
	Conversations(context.Context, string, string, types.MessagingQuery) (types.MessagingPage, *apperrors.Error)
	Thread(context.Context, string, string, bool) (*types.MessagingThread, *apperrors.Error)
	UpsertThread(context.Context, types.MessagingThread) (*types.MessagingThread, *apperrors.Error)
	CurrentTurn(context.Context, string, string) (*types.MessagingTurn, *apperrors.Error)
	Turn(context.Context, string, string) (*types.MessagingTurn, *apperrors.Error)
	SaveTurn(context.Context, types.MessagingTurn, bool) *apperrors.Error
	SaveThread(context.Context, types.MessagingThread) *apperrors.Error
	MessageByKey(context.Context, string, string) (*types.MessagingMessage, *apperrors.Error)
	InsertMessage(context.Context, types.MessagingMessage) *apperrors.Error
	Messages(context.Context, string, string, types.MessagingQuery) ([]types.MessagingMessage, string, *apperrors.Error)
	Audit(context.Context, string, string, string, string, string, any) *apperrors.Error
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
func (r *MessagingRepository) Scope(ctx context.Context, tenant, actor string) (types.MessagingScope, *apperrors.Error) {
	workspace, err := r.Tenant(ctx, tenant)
	if err != nil {
		return types.MessagingScope{}, err
	}
	member, err := r.Member(ctx, tenant, actor)
	if err != nil {
		return types.MessagingScope{}, err
	}
	return types.MessagingScope{Active: workspace.Status == model.TenantStatusActive, Member: member != nil, Owner: workspace.OwnerID == actor}, nil
}
func (r *MessagingRepository) MemberExists(ctx context.Context, tenant, actor string) (bool, *apperrors.Error) {
	member, err := r.Member(ctx, tenant, actor)
	return member != nil, err
}
func (r *MessagingRepository) Conversations(ctx context.Context, tenant, actor string, input types.MessagingQuery) (types.MessagingPage, *apperrors.Error) {
	rows := []messagingmodel.Thread{}
	q := r.db.WithContext(ctx).Table("messaging_threads AS th").Select("th.*").Joins("JOIN messaging_turns AS t ON t.tenant_id = th.tenant_id AND t.thread_id = th.id AND t.lifecycle = 'open'").Where("th.tenant_id = ?", tenant)
	switch input.Filter {
	case "unassigned":
		q = q.Where("t.assignee_id IS NULL")
	case "mine":
		q = q.Where("t.assignee_id = ?", actor)
	case "needs_human":
		q = q.Where("(t.needs_human = TRUE OR (t.control = ? AND t.assignee_id IS NULL))", "sales")
	}
	if input.Cursor != "" {
		q = q.Where("th.id < ?", input.Cursor)
	}
	if err := q.Order("th.id DESC").Limit(input.Limit + 1).Scan(&rows).Error; err != nil {
		return types.MessagingPage{}, salesDBError(ctx, err)
	}
	result := types.MessagingPage{Items: []types.MessagingConversation{}}
	if len(rows) > input.Limit {
		rows = rows[:input.Limit]
		result.NextCursor = rows[len(rows)-1].ID
	}
	if len(rows) == 0 {
		return result, nil
	}
	ids := make([]string, len(rows))
	for i, row := range rows {
		ids[i] = row.ID
	}
	turns := []messagingmodel.Turn{}
	if err := r.db.WithContext(ctx).Where("tenant_id = ? AND thread_id IN ? AND lifecycle = 'open'", tenant, ids).Find(&turns).Error; err != nil {
		return result, salesDBError(ctx, err)
	}
	byThread := map[string]types.MessagingTurn{}
	for _, row := range turns {
		byThread[row.ThreadID] = persistencemap.Turn(row)
	}
	for _, row := range rows {
		turn := byThread[row.ID]
		result.Items = append(result.Items, types.MessagingConversation{Thread: persistencemap.Thread(row), Turn: &turn})
	}
	return result, nil
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
func (r *MessagingRepository) Messages(ctx context.Context, tenant, thread string, input types.MessagingQuery) ([]types.MessagingMessage, string, *apperrors.Error) {
	rows := []messagingmodel.Message{}
	q := r.db.WithContext(ctx).Where("tenant_id = ? AND thread_id = ?", tenant, thread)
	if input.Cursor != "" {
		q = q.Where("id < ?", input.Cursor)
	}
	if err := q.Order("id DESC").Limit(input.Limit + 1).Find(&rows).Error; err != nil {
		return nil, "", salesDBError(ctx, err)
	}
	next := ""
	if len(rows) > input.Limit {
		rows = rows[:input.Limit]
		next = rows[len(rows)-1].ID
	}
	result := make([]types.MessagingMessage, len(rows))
	for i, row := range rows {
		result[i] = persistencemap.Message(row)
	}
	return result, next, nil
}
