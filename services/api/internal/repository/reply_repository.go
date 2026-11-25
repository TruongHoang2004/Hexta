package repository

import (
	"context"
	stderrors "errors"
	"time"

	apperrors "github.com/TruongHoang2004/Hexta/packages/shared/pkg/errors"
	"github.com/TruongHoang2004/Hexta/services/api/internal/core/model"
	"github.com/TruongHoang2004/Hexta/services/api/internal/core/types"
	"github.com/TruongHoang2004/Hexta/services/api/internal/mapper/messaging/persistencemap"
	replymap "github.com/TruongHoang2004/Hexta/services/api/internal/mapper/reply/persistencemap"
	"github.com/TruongHoang2004/Hexta/services/api/internal/repository/messagingmodel"
	"github.com/TruongHoang2004/Hexta/services/api/internal/repository/replymodel"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// ReplySettingsStore persists the tenant's bot reply mode. Authorization stays in the service.
type ReplySettingsStore interface {
	WithinReplySettings(context.Context, func(ReplySettingsStore) *apperrors.Error) *apperrors.Error
	Scope(context.Context, string, string) (types.MessagingScope, *apperrors.Error)
	ReplySettings(context.Context, string, bool) (types.ReplySettings, *apperrors.Error)
	// SaveReplySettings inserts the first row (create) or updates the row whose version is value.Version-1.
	SaveReplySettings(context.Context, types.ReplySettings, bool) *apperrors.Error
	Audit(context.Context, string, string, string, string, string, any) *apperrors.Error
}

// ReplyRouterStore is the persistence used by the reply router job. Reads outside Within take no locks; every
// decision is rechecked under the thread lock inside Within.
type ReplyRouterStore interface {
	MessagingStore
	// RoutingDue lists due routing jobs, oldest first, bounded by limit.
	RoutingDue(context.Context, time.Time, int) ([]types.ProposalRef, *apperrors.Error)
}

func (r *MessagingRepository) WithinReplySettings(ctx context.Context, fn func(ReplySettingsStore) *apperrors.Error) *apperrors.Error {
	return r.Within(ctx, func(tx MessagingStore) *apperrors.Error { return fn(tx.(*MessagingRepository)) })
}

// ReplySettings reads the tenant's reply mode; a missing row is the unconfigured suggest default.
func (r *MessagingRepository) ReplySettings(ctx context.Context, tenant string, lock bool) (types.ReplySettings, *apperrors.Error) {
	var row replymodel.Settings
	q := r.db.WithContext(ctx).Where("tenant_id = ?", tenant)
	if lock {
		q = q.Clauses(clause.Locking{Strength: "UPDATE"})
	}
	err := q.First(&row).Error
	if stderrors.Is(err, gorm.ErrRecordNotFound) {
		return types.ReplySettings{TenantID: tenant, Mode: types.ReplyModeSuggest}, nil
	}
	if err != nil {
		return types.ReplySettings{}, salesDBError(ctx, err)
	}
	value, mapErr := replymap.Settings(row)
	if mapErr != nil {
		return types.ReplySettings{}, apperrors.ErrSystemError(ctx, "stored reply settings are invalid")
	}
	return value, nil
}

func (r *MessagingRepository) SaveReplySettings(ctx context.Context, value types.ReplySettings, create bool) *apperrors.Error {
	row := replymap.SettingsRow(value)
	if create {
		result := r.db.WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).Create(&row)
		if result.Error != nil {
			return salesDBError(ctx, result.Error)
		}
		if result.RowsAffected != 1 {
			return apperrors.ErrConflict(ctx, "reply settings", "changed; refresh before retrying")
		}
		return nil
	}
	result := r.db.WithContext(ctx).Model(&replymodel.Settings{}).Where("tenant_id = ? AND version = ?", value.TenantID, value.Version-1).
		Select("reply_mode", "version", "updated_by", "updated_at").Updates(&row)
	if result.Error != nil {
		return salesDBError(ctx, result.Error)
	}
	if result.RowsAffected != 1 {
		return apperrors.ErrConflict(ctx, "reply settings", "changed; refresh before retrying")
	}
	return nil
}

// InsertRouting supersedes the turn's live (routing or pending) proposals and records a new routing job, so a
// turn has at most one live proposal and it always answers the newest message.
func (r *MessagingRepository) InsertRouting(ctx context.Context, value types.ReplyProposal) *apperrors.Error {
	db := r.db.WithContext(ctx)
	if err := db.Model(&replymodel.Proposal{}).
		Where("tenant_id = ? AND turn_id = ? AND status IN ?", value.TenantID, value.TurnID, []string{types.ProposalRouting, types.ProposalPending}).
		Updates(map[string]any{"status": types.ProposalSuperseded, "version": gorm.Expr("version + 1"), "updated_at": value.CreatedAt}).Error; err != nil {
		return salesDBError(ctx, err)
	}
	row := replymap.ProposalRow(value)
	if err := db.Create(&row).Error; err != nil {
		return salesDBError(ctx, err)
	}
	return nil
}

func (r *MessagingRepository) Proposal(ctx context.Context, tenant, id string, lock bool) (*types.ReplyProposal, *apperrors.Error) {
	var row replymodel.Proposal
	q := r.db.WithContext(ctx).Where("tenant_id = ? AND id = ?", tenant, id)
	if lock {
		q = q.Clauses(clause.Locking{Strength: "UPDATE"})
	}
	if err := q.First(&row).Error; err != nil {
		return nil, salesDBError(ctx, err)
	}
	value := replymap.Proposal(row)
	return &value, nil
}

// SaveProposal updates the proposal only if it is still at value.Version-1.
func (r *MessagingRepository) SaveProposal(ctx context.Context, value types.ReplyProposal) *apperrors.Error {
	row := replymap.ProposalRow(value)
	result := r.db.WithContext(ctx).Model(&replymodel.Proposal{}).Where("tenant_id = ? AND id = ? AND version = ?", value.TenantID, value.ID, value.Version-1).
		Select("status", "route", "source", "script_version_id", "ai_run_id", "text", "reason", "decided_by", "sent_message_id", "attempts", "available_at", "version", "updated_at").
		Updates(&row)
	if result.Error != nil {
		return salesDBError(ctx, result.Error)
	}
	if result.RowsAffected != 1 {
		return apperrors.ErrConflict(ctx, "proposal", "changed; refresh before retrying")
	}
	return nil
}

// Message reads one message of the tenant.
func (r *MessagingRepository) Message(ctx context.Context, tenant, id string) (*types.MessagingMessage, *apperrors.Error) {
	var row messagingmodel.Message
	if err := r.db.WithContext(ctx).Where("tenant_id = ? AND id = ?", tenant, id).First(&row).Error; err != nil {
		return nil, salesDBError(ctx, err)
	}
	value := persistencemap.Message(row)
	return &value, nil
}

// LatestInbound returns the newest customer message of a turn; nil when it has none.
func (r *MessagingRepository) LatestInbound(ctx context.Context, tenant, turn string) (*types.MessagingMessage, *apperrors.Error) {
	var row messagingmodel.Message
	err := r.db.WithContext(ctx).Where("tenant_id = ? AND turn_id = ? AND direction = ?", tenant, turn, "inbound").Order("id DESC").First(&row).Error
	if stderrors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, salesDBError(ctx, err)
	}
	value := persistencemap.Message(row)
	return &value, nil
}

// RecentMessages returns up to limit messages of a turn created before the given message, oldest first.
func (r *MessagingRepository) RecentMessages(ctx context.Context, tenant, turn, before string, limit int) ([]types.MessagingMessage, *apperrors.Error) {
	rows := []messagingmodel.Message{}
	if err := r.db.WithContext(ctx).Where("tenant_id = ? AND turn_id = ? AND id < ?", tenant, turn, before).Order("id DESC").Limit(limit).Find(&rows).Error; err != nil {
		return nil, salesDBError(ctx, err)
	}
	result := make([]types.MessagingMessage, len(rows))
	for i, row := range rows {
		result[len(rows)-1-i] = persistencemap.Message(row)
	}
	return result, nil
}

// CancelPendingBotSends cancels the turn's automatic (bot-authored) sends that have not been dispatched: the intent
// is still pending and its outbox event was never handed to the relay. Such a send provably never reached the
// connector: its outbox event becomes cancelled (distinct from a relay give-up, which is failed) and its message
// failed/cancelled. A send whose event is already leased or published is dispatched and is left to be reconciled
// from the connector's outcome. Lock order: thread (held by the caller), outbound intents, outbox events; the
// caller saves the turn afterwards.
func (r *MessagingRepository) CancelPendingBotSends(ctx context.Context, tenant, turn string, now time.Time) (int, *apperrors.Error) {
	return r.cancelUndispatched(ctx, tenant, turn, true, now)
}

// CancelUndispatchedSends is CancelPendingBotSends for every author: a manual session close cancels staff and bot
// sends that never reached the connector. Dispatched sends are reconciled from the connector's outcome.
func (r *MessagingRepository) CancelUndispatchedSends(ctx context.Context, tenant, turn string, now time.Time) (int, *apperrors.Error) {
	return r.cancelUndispatched(ctx, tenant, turn, false, now)
}

func (r *MessagingRepository) cancelUndispatched(ctx context.Context, tenant, turn string, botOnly bool, now time.Time) (int, *apperrors.Error) {
	db := r.db.WithContext(ctx)
	var ids []string
	actor := ""
	if botOnly {
		actor = types.BotActorID
	}
	if err := db.Raw(`SELECT i.id FROM outbound_intents i JOIN messaging_messages m ON m.tenant_id = i.tenant_id AND m.id = i.message_id
WHERE i.tenant_id = ? AND m.turn_id = ? AND (CAST(? AS text) = '' OR m.actor_id = ?) AND i.status = ? ORDER BY i.id FOR UPDATE OF i`, tenant, turn, actor, actor, types.IntentPending).Scan(&ids).Error; err != nil {
		return 0, salesDBError(ctx, err)
	}
	if len(ids) == 0 {
		return 0, nil
	}
	var undispatched []string
	if err := db.Raw(`UPDATE domain_events SET state = ? WHERE event_type = ? AND aggregate_id IN ? AND state = ? RETURNING aggregate_id`,
		model.DomainEventStateCancelled, model.DomainEventOutboundIntent, ids, model.DomainEventStatePending).Scan(&undispatched).Error; err != nil {
		return 0, salesDBError(ctx, err)
	}
	if len(undispatched) == 0 {
		return 0, nil
	}
	if err := db.Model(&messagingmodel.OutboundAttempt{}).Where("tenant_id = ? AND intent_id IN ? AND status = ?", tenant, undispatched, types.IntentPending).
		Updates(map[string]any{"status": types.IntentCancelled, "completed_at": now}).Error; err != nil {
		return 0, salesDBError(ctx, err)
	}
	if err := db.Model(&messagingmodel.OutboundIntent{}).Where("tenant_id = ? AND id IN ? AND status = ?", tenant, undispatched, types.IntentPending).
		Updates(map[string]any{"status": types.IntentCancelled, "error_category": types.DeliveryCancelledReason, "updated_at": now}).Error; err != nil {
		return 0, salesDBError(ctx, err)
	}
	if err := db.Model(&messagingmodel.Message{}).
		Where("tenant_id = ? AND delivery_status = ? AND id IN (?)", tenant, types.DeliverySending,
			db.Model(&messagingmodel.OutboundIntent{}).Select("message_id").Where("tenant_id = ? AND id IN ?", tenant, undispatched)).
		Updates(map[string]any{"delivery_status": types.DeliveryFailed, "delivery_error_reason": types.DeliveryCancelledReason}).Error; err != nil {
		return 0, salesDBError(ctx, err)
	}
	return len(undispatched), nil
}

// ScriptVersionActive reports whether the version is still its script's active version, share-locking the script
// row so a concurrent deactivate/archive (which locks it for update) waits until this transaction ends.
func (r *MessagingRepository) ScriptVersionActive(ctx context.Context, tenant, version string) (bool, *apperrors.Error) {
	var ids []string
	if err := r.db.WithContext(ctx).Raw(`SELECT id FROM scripts WHERE tenant_id = ? AND active_version_id = ? AND status = ? FOR SHARE`,
		tenant, version, types.ScriptStatusActive).Scan(&ids).Error; err != nil {
		return false, salesDBError(ctx, err)
	}
	return len(ids) == 1, nil
}

func (r *MessagingRepository) RoutingDue(ctx context.Context, now time.Time, limit int) ([]types.ProposalRef, *apperrors.Error) {
	rows := []replymodel.Proposal{}
	if err := r.db.WithContext(ctx).Select("tenant_id", "thread_id", "id").Where("status = ? AND available_at <= ?", types.ProposalRouting, now).
		Order("available_at").Limit(limit).Find(&rows).Error; err != nil {
		return nil, salesDBError(ctx, err)
	}
	refs := make([]types.ProposalRef, len(rows))
	for i, row := range rows {
		refs[i] = types.ProposalRef{TenantID: row.TenantID, ThreadID: row.ThreadID, ProposalID: row.ID}
	}
	return refs, nil
}

// LiveProposals returns the turn's live proposals (routing or pending, at most one) with their script label.
func (r *MessagingRepository) LiveProposals(ctx context.Context, tenant, turn string) ([]types.ProposalView, *apperrors.Error) {
	rows := []replymodel.ProposalLabel{}
	if err := r.db.WithContext(ctx).Table("messaging_reply_proposals AS p").
		Select("p.*, s.name AS script_name, v.number AS script_version").
		Joins("LEFT JOIN script_versions v ON v.tenant_id = p.tenant_id AND v.id = p.script_version_id").
		Joins("LEFT JOIN scripts s ON s.tenant_id = v.tenant_id AND s.id = v.script_id").
		Where("p.tenant_id = ? AND p.turn_id = ? AND p.status IN ?", tenant, turn, []string{types.ProposalRouting, types.ProposalPending}).
		Order("p.id DESC").Limit(5).Scan(&rows).Error; err != nil {
		return nil, salesDBError(ctx, err)
	}
	result := make([]types.ProposalView, len(rows))
	for i, row := range rows {
		result[i] = replymap.ProposalView(row)
	}
	return result, nil
}

var (
	_ ReplySettingsStore = (*MessagingRepository)(nil)
	_ ReplyRouterStore   = (*MessagingRepository)(nil)
)
