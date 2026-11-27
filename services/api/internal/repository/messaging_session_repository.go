package repository

import (
	"context"
	stderrors "errors"

	apperrors "github.com/TruongHoang2004/Hexta/packages/shared/pkg/errors"
	"github.com/TruongHoang2004/Hexta/services/api/internal/core/types"
	"github.com/TruongHoang2004/Hexta/services/api/internal/mapper/messaging/persistencemap"
	"github.com/TruongHoang2004/Hexta/services/api/internal/repository/messagingmodel"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// SaveSessionOutcome writes a closed session's outcome and note at the next version. The caller holds the thread
// lock and has checked the expected version; value.Version is the new version.
func (r *MessagingRepository) SaveSessionOutcome(ctx context.Context, value types.MessagingTurn) *apperrors.Error {
	row := persistencemap.TurnRow(value)
	result := r.db.WithContext(ctx).Model(&messagingmodel.Turn{}).
		Where("tenant_id = ? AND id = ? AND version = ? AND lifecycle = ?", value.TenantID, value.ID, value.Version-1, types.TurnClosed).
		Updates(map[string]any{"outcome": row.Outcome, "note": row.Note, "version": row.Version})
	if result.Error != nil {
		return salesDBError(ctx, result.Error)
	}
	if result.RowsAffected != 1 {
		return apperrors.ErrConflict(ctx, "session", "changed; refresh before retrying")
	}
	r.markConversation(value.TenantID, value.ThreadID)
	return nil
}

// Sessions lists a thread's turns newest first (turn IDs are UUIDv7) with message statistics and linked orders.
// The caller has resolved the thread inside the tenant.
func (r *MessagingRepository) Sessions(ctx context.Context, tenant, thread string, input types.MessagingQuery) (types.MessagingSessionPage, *apperrors.Error) {
	result := types.MessagingSessionPage{Items: []types.MessagingSession{}}
	rows := []messagingmodel.Turn{}
	q := r.db.WithContext(ctx).Where("tenant_id = ? AND thread_id = ?", tenant, thread)
	if input.Cursor != "" {
		q = q.Where("id < ?", input.Cursor)
	}
	if err := q.Order("id DESC").Limit(input.Limit + 1).Find(&rows).Error; err != nil {
		return result, salesDBError(ctx, err)
	}
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
	stats := []messagingmodel.SessionStats{}
	if err := r.db.WithContext(ctx).Raw(`
SELECT m.turn_id, count(*) AS message_count, min(m.created_at) AS opened_at,
  (SELECT f.id FROM messaging_messages f WHERE f.tenant_id = m.tenant_id AND f.turn_id = m.turn_id ORDER BY f.id LIMIT 1) AS first_message_id
FROM messaging_messages m WHERE m.tenant_id = ? AND m.turn_id IN ? GROUP BY m.tenant_id, m.turn_id`, tenant, ids).Scan(&stats).Error; err != nil {
		return result, salesDBError(ctx, err)
	}
	orders := []messagingmodel.SessionOrder{}
	if err := r.db.WithContext(ctx).Table("orders").Select("id", "order_number", "status", "messaging_turn_id").
		Where("tenant_id = ? AND messaging_turn_id IN ? AND deleted_at IS NULL", tenant, ids).Order("created_at, id").Scan(&orders).Error; err != nil {
		return result, salesDBError(ctx, err)
	}
	byTurn := map[string]messagingmodel.SessionStats{}
	for _, stat := range stats {
		byTurn[stat.TurnID] = stat
	}
	linked := map[string][]messagingmodel.SessionOrder{}
	for _, order := range orders {
		linked[order.MessagingTurnID] = append(linked[order.MessagingTurnID], order)
	}
	for _, row := range rows {
		result.Items = append(result.Items, persistencemap.Session(row, byTurn[row.ID], linked[row.ID]))
	}
	return result, nil
}

// SessionTurnOpen share-locks a turn of the tenant and reports whether it is open, so a draft order is linked only to
// a live session and a concurrent close waits for the order to commit. A missing turn is not found.
func (r *SalesRepository) SessionTurnOpen(ctx context.Context, tenant, turn string) (bool, *apperrors.Error) {
	var row messagingmodel.Turn
	err := r.db.WithContext(ctx).Clauses(clause.Locking{Strength: "SHARE"}).Select("id", "lifecycle").
		Where("tenant_id = ? AND id = ?", tenant, turn).First(&row).Error
	if stderrors.Is(err, gorm.ErrRecordNotFound) {
		return false, apperrors.ErrNotFound(ctx, "conversation session", "not found")
	}
	if err != nil {
		return false, salesDBError(ctx, err)
	}
	return row.Lifecycle == types.TurnOpen, nil
}
