package repository

import (
	"context"
	"database/sql"

	apperrors "github.com/TruongHoang2004/Hexta/packages/shared/pkg/errors"
	"github.com/TruongHoang2004/Hexta/services/api/internal/core/model"
	"github.com/TruongHoang2004/Hexta/services/api/internal/core/types"
	"github.com/TruongHoang2004/Hexta/services/api/internal/mapper/messaging/persistencemap"
	"github.com/TruongHoang2004/Hexta/services/api/internal/repository/messagingmodel"
	"gorm.io/gorm"
)

// MessagingQueryStore is the read side of messaging. Implementations must not write.
type MessagingQueryStore interface {
	ReadWithin(context.Context, func(MessagingQueryStore) *apperrors.Error) *apperrors.Error
	Scope(context.Context, string, string) (types.MessagingScope, *apperrors.Error)
	Conversations(context.Context, string, string, types.MessagingQuery) (types.MessagingPage, *apperrors.Error)
	Thread(context.Context, string, string, bool) (*types.MessagingThread, *apperrors.Error)
	CurrentTurn(context.Context, string, string) (*types.MessagingTurn, *apperrors.Error)
	Messages(context.Context, string, string, types.MessagingQuery) ([]types.MessagingMessage, string, *apperrors.Error)
}

// ReadWithin runs fn in one read-only transaction so multi-query reads share a snapshot.
func (r *MessagingRepository) ReadWithin(ctx context.Context, fn func(MessagingQueryStore) *apperrors.Error) *apperrors.Error {
	var failure *apperrors.Error
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		failure = fn(NewMessagingRepository(tx))
		if failure != nil {
			return failure
		}
		return nil
	}, &sql.TxOptions{ReadOnly: true})
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
	case "bot":
		// Mirrors MessagingTurn.HandlingState() == BOT_ACTIVE.
		q = q.Where("t.control = ? AND t.assignee_id IS NULL AND t.needs_human = FALSE", "system")
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
