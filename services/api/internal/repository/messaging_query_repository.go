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
	ReadScope(context.Context, string, string) (types.MessagingScope, *apperrors.Error)
	Conversations(context.Context, string, string, types.MessagingQuery) (types.MessagingPage, *apperrors.Error)
	Thread(context.Context, string, string, bool) (*types.MessagingThread, *apperrors.Error)
	CurrentTurn(context.Context, string, string) (*types.MessagingTurn, *apperrors.Error)
	LatestTurn(context.Context, string, string) (*types.MessagingTurn, *apperrors.Error)
	WaitingStaffCount(context.Context, string) (int64, *apperrors.Error)
	Messages(context.Context, string, string, types.MessagingQuery) ([]types.MessagingMessage, string, *apperrors.Error)
	// LiveProposals returns the turn's routing or pending bot reply proposals with their script source label.
	LiveProposals(context.Context, string, string) ([]types.ProposalView, *apperrors.Error)
	// Sessions lists a thread's turns newest first with message statistics and linked orders.
	Sessions(context.Context, string, string, types.MessagingQuery) (types.MessagingSessionPage, *apperrors.Error)
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

// ReadScope resolves membership without row locks, which a read-only transaction cannot take.
func (r *MessagingRepository) ReadScope(ctx context.Context, tenant, actor string) (types.MessagingScope, *apperrors.Error) {
	var workspace model.Tenant
	if err := r.db.WithContext(ctx).Where("id = ?", tenant).First(&workspace).Error; err != nil {
		return types.MessagingScope{}, salesDBError(ctx, err)
	}
	var members int64
	if err := r.db.WithContext(ctx).Model(&model.TenantMember{}).Where("tenant_id = ? AND user_id = ?", tenant, actor).Count(&members).Error; err != nil {
		return types.MessagingScope{}, salesDBError(ctx, err)
	}
	return types.MessagingScope{Active: workspace.Status == model.TenantStatusActive, Member: members > 0, Owner: workspace.OwnerID == actor}, nil
}

// Conversations reads the conversation read model; it never touches the write-side thread/turn tables.
func (r *MessagingRepository) Conversations(ctx context.Context, tenant, actor string, input types.MessagingQuery) (types.MessagingPage, *apperrors.Error) {
	rows := []messagingmodel.ConversationView{}
	q := r.db.WithContext(ctx).Where("tenant_id = ?", tenant)
	// Every filter except "all" lists open turns only; closed conversations stay visible under "all".
	if input.Filter != "" && input.Filter != "all" {
		q = q.Where("lifecycle = ?", types.TurnOpen)
	}
	switch input.Filter {
	case "unassigned":
		q = q.Where("assignee_id IS NULL")
	case "mine":
		q = q.Where("assignee_id = ?", actor)
	case "needs_human":
		q = q.Where("(needs_human = TRUE OR (control = ? AND assignee_id IS NULL))", "sales")
	case "waiting":
		// Mirrors MessagingTurn.HandlingState() == WAITING_STAFF.
		q = q.Where("control = ? AND assignee_id IS NULL", "sales")
	case "bot":
		// Mirrors MessagingTurn.HandlingState() == BOT_ACTIVE.
		q = q.Where("control = ? AND assignee_id IS NULL AND needs_human = FALSE", "system")
	}
	if input.Cursor != "" {
		q = q.Where("thread_id < ?", input.Cursor)
	}
	if err := q.Order("thread_id DESC").Limit(input.Limit + 1).Find(&rows).Error; err != nil {
		return types.MessagingPage{}, salesDBError(ctx, err)
	}
	result := types.MessagingPage{Items: []types.MessagingConversation{}}
	if len(rows) > input.Limit {
		rows = rows[:input.Limit]
		result.NextCursor = rows[len(rows)-1].ThreadID
	}
	for _, row := range rows {
		thread, turn := persistencemap.ConversationView(row)
		result.Items = append(result.Items, types.MessagingConversation{Thread: thread, Turn: &turn})
	}
	return result, nil
}

// WaitingStaffCount counts the tenant's open conversations waiting for staff, from the read model.
func (r *MessagingRepository) WaitingStaffCount(ctx context.Context, tenant string) (int64, *apperrors.Error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&messagingmodel.ConversationView{}).
		Where("tenant_id = ? AND lifecycle = ? AND control = ? AND assignee_id IS NULL", tenant, types.TurnOpen, "sales").
		Count(&count).Error; err != nil {
		return 0, salesDBError(ctx, err)
	}
	return count, nil
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
