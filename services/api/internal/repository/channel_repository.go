package repository

import (
	"context"
	stderrors "errors"
	"time"

	apperrors "github.com/TruongHoang2004/Hexta/packages/shared/pkg/errors"
	"github.com/TruongHoang2004/Hexta/services/api/internal/core/types"
	"github.com/TruongHoang2004/Hexta/services/api/internal/mapper/channel/persistencemap"
	"github.com/TruongHoang2004/Hexta/services/api/internal/repository/messagingmodel"
	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ChannelStore interface {
	Within(context.Context, func(ChannelStore) *apperrors.Error) *apperrors.Error
	Scope(context.Context, string, string) (types.MessagingScope, *apperrors.Error)
	SaveOAuthState(context.Context, types.ChannelOAuthState) *apperrors.Error
	ConsumeOAuthState(context.Context, string, time.Time) (*types.ChannelOAuthState, *apperrors.Error)
	AttachOAuthSession(context.Context, string, string, time.Time, []types.ChannelPage) *apperrors.Error
	OAuthSessionState(context.Context, string, string, string, time.Time) (*types.ChannelOAuthState, *apperrors.Error)
	Connections(context.Context, string, string) ([]types.ChannelConnection, *apperrors.Error)
	BoundAccounts(context.Context, string, []string) ([]types.ChannelConnection, *apperrors.Error)
	Connection(context.Context, string, string, bool) (*types.ChannelConnection, *apperrors.Error)
	TenantAccount(context.Context, string, string, string) (*types.ChannelConnection, *apperrors.Error)
	InsertConnection(context.Context, types.ChannelConnection) *apperrors.Error
	SaveConnection(context.Context, types.ChannelConnection, int64) *apperrors.Error
	RecordConnectionError(context.Context, string, string, int64, string, string) *apperrors.Error
	CancelPendingIntents(context.Context, string, string, string, time.Time) *apperrors.Error
	Audit(context.Context, string, string, string, string, string, any) *apperrors.Error
}

type ChannelRepository struct{ *MessagingRepository }

func NewChannelRepository(db *gorm.DB) *ChannelRepository {
	return &ChannelRepository{NewMessagingRepository(db)}
}

func (r *ChannelRepository) Within(ctx context.Context, fn func(ChannelStore) *apperrors.Error) *apperrors.Error {
	var failure *apperrors.Error
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if failure = fn(NewChannelRepository(tx)); failure != nil {
			return failure
		}
		return nil
	})
	if failure != nil {
		return failure
	}
	if err != nil {
		return channelDBError(ctx, err)
	}
	return nil
}

// channelDBError maps the 1-account-1-tenant partial unique index (Facebook Page, Telegram Bot) to a stable
// conflict without SQL detail.
func channelDBError(ctx context.Context, err error) *apperrors.Error {
	var pg *pgconn.PgError
	if stderrors.As(err, &pg) && pg.Code == "23505" && pg.ConstraintName == "uq_messaging_connections_active_page" {
		return apperrors.ErrConflict(ctx, "channel account", "is already connected to another workspace")
	}
	return salesDBError(ctx, err)
}

func (r *ChannelRepository) SaveOAuthState(ctx context.Context, value types.ChannelOAuthState) *apperrors.Error {
	row := persistencemap.OAuthStateRow(value)
	if err := r.db.WithContext(ctx).Create(&row).Error; err != nil {
		return channelDBError(ctx, err)
	}
	return nil
}

// ConsumeOAuthState atomically marks an unexpired, unconsumed state used; a replay or expired state returns nil.
func (r *ChannelRepository) ConsumeOAuthState(ctx context.Context, hash string, now time.Time) (*types.ChannelOAuthState, *apperrors.Error) {
	var rows []messagingmodel.OAuthState
	err := r.db.WithContext(ctx).Raw(`UPDATE channel_oauth_states SET consumed_at=? WHERE state_hash=? AND consumed_at IS NULL AND expires_at>? RETURNING *`, now, hash, now).Scan(&rows).Error
	if err != nil {
		return nil, channelDBError(ctx, err)
	}
	if len(rows) == 0 {
		return nil, nil
	}
	value := persistencemap.OAuthState(rows[0])
	return &value, nil
}

func (r *ChannelRepository) AttachOAuthSession(ctx context.Context, id, session string, expires time.Time, pages []types.ChannelPage) *apperrors.Error {
	err := r.db.WithContext(ctx).Model(&messagingmodel.OAuthState{}).Where("id = ?", id).Updates(map[string]any{"oauth_session_id": session, "oauth_session_expires_at": expires, "pages": gorm.Expr("?::jsonb", *persistencemap.Pages(pages))}).Error
	if err != nil {
		return channelDBError(ctx, err)
	}
	return nil
}

// OAuthSessionState resolves a staged session only for the tenant and owner that completed its OAuth flow.
func (r *ChannelRepository) OAuthSessionState(ctx context.Context, tenant, actor, session string, now time.Time) (*types.ChannelOAuthState, *apperrors.Error) {
	var row messagingmodel.OAuthState
	err := r.db.WithContext(ctx).Where("tenant_id = ? AND actor_id = ? AND oauth_session_id = ? AND oauth_session_expires_at > ?", tenant, actor, session, now).First(&row).Error
	if stderrors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, channelDBError(ctx, err)
	}
	value := persistencemap.OAuthState(row)
	return &value, nil
}

func connectionList(rows []messagingmodel.Connection) []types.ChannelConnection {
	result := make([]types.ChannelConnection, len(rows))
	for i, row := range rows {
		result[i] = persistencemap.Connection(row)
	}
	return result
}

// Connections lists a tenant's current (non-disconnected) bindings of one platform, or of every platform when
// platform is empty; the per-tenant set is small and bounded.
func (r *ChannelRepository) Connections(ctx context.Context, tenant, platform string) ([]types.ChannelConnection, *apperrors.Error) {
	rows := []messagingmodel.Connection{}
	query := r.db.WithContext(ctx).Where("tenant_id = ? AND status <> ?", tenant, types.ChannelStatusDisconnected)
	if platform != "" {
		query = query.Where("platform = ?", platform)
	}
	if err := query.Order("updated_at DESC, id").Limit(200).Find(&rows).Error; err != nil {
		return nil, channelDBError(ctx, err)
	}
	return connectionList(rows), nil
}

// BoundAccounts is a platform-scoped lookup used only to compute Page availability, never to grant access.
func (r *ChannelRepository) BoundAccounts(ctx context.Context, platform string, accounts []string) ([]types.ChannelConnection, *apperrors.Error) {
	rows := []messagingmodel.Connection{}
	if len(accounts) == 0 {
		return nil, nil
	}
	if err := r.db.WithContext(ctx).Where("platform = ? AND account_id IN ? AND status <> ?", platform, accounts, types.ChannelStatusDisconnected).Find(&rows).Error; err != nil {
		return nil, channelDBError(ctx, err)
	}
	return connectionList(rows), nil
}

func (r *ChannelRepository) Connection(ctx context.Context, tenant, id string, lock bool) (*types.ChannelConnection, *apperrors.Error) {
	var row messagingmodel.Connection
	q := r.db.WithContext(ctx).Where("tenant_id = ? AND id = ?", tenant, id)
	if lock {
		q = q.Clauses(clause.Locking{Strength: "UPDATE"})
	}
	if err := q.First(&row).Error; err != nil {
		return nil, channelDBError(ctx, err)
	}
	value := persistencemap.Connection(row)
	return &value, nil
}

// TenantAccount locks the tenant's most recent binding for a Page so a reconnect reuses its identity.
func (r *ChannelRepository) TenantAccount(ctx context.Context, tenant, platform, account string) (*types.ChannelConnection, *apperrors.Error) {
	var row messagingmodel.Connection
	err := r.db.WithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).Where("tenant_id = ? AND platform = ? AND account_id = ?", tenant, platform, account).Order("(status <> 'disconnected') DESC, updated_at DESC").First(&row).Error
	if stderrors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, channelDBError(ctx, err)
	}
	value := persistencemap.Connection(row)
	return &value, nil
}

func (r *ChannelRepository) InsertConnection(ctx context.Context, value types.ChannelConnection) *apperrors.Error {
	row := persistencemap.ConnectionRow(value)
	if err := r.db.WithContext(ctx).Create(&row).Error; err != nil {
		return channelDBError(ctx, err)
	}
	return nil
}

// SaveConnection applies a versioned change; the expected version predicate rejects concurrent updates.
func (r *ChannelRepository) SaveConnection(ctx context.Context, value types.ChannelConnection, expected int64) *apperrors.Error {
	row := persistencemap.ConnectionRow(value)
	result := r.db.WithContext(ctx).Model(&messagingmodel.Connection{}).Where("tenant_id = ? AND id = ? AND version = ?", value.TenantID, value.ID, expected).
		Select("version", "enabled", "account_name", "account_username", "status", "last_error_code", "last_error_message", "updated_at").Updates(&row)
	if result.Error != nil {
		return channelDBError(ctx, result.Error)
	}
	if result.RowsAffected != 1 {
		return apperrors.ErrConflict(ctx, "connection", "changed; refresh before retrying")
	}
	return nil
}

// RecordConnectionError annotates a sync failure without changing authority or version.
func (r *ChannelRepository) RecordConnectionError(ctx context.Context, tenant, id string, version int64, code, message string) *apperrors.Error {
	if err := r.db.WithContext(ctx).Model(&messagingmodel.Connection{}).Where("tenant_id = ? AND id = ? AND version = ?", tenant, id, version).
		Updates(map[string]any{"last_error_code": code, "last_error_message": message, "updated_at": time.Now().UTC()}).Error; err != nil {
		return channelDBError(ctx, err)
	}
	return nil
}

// CancelPendingIntents cancels undelivered intents and marks their messages failed with the given reason.
func (r *ChannelRepository) CancelPendingIntents(ctx context.Context, tenant, connection, reason string, now time.Time) *apperrors.Error {
	db := r.db.WithContext(ctx)
	if err := db.Exec(`UPDATE messaging_messages m SET delivery_status=?, delivery_error_reason=? FROM outbound_intents i
		WHERE i.tenant_id=? AND i.connection_id=? AND i.status=? AND m.tenant_id=i.tenant_id AND m.id=i.message_id`, types.DeliveryFailed, reason, tenant, connection, types.IntentPending).Error; err != nil {
		return channelDBError(ctx, err)
	}
	if err := db.Exec(`UPDATE outbound_attempts a SET status=?, completed_at=? FROM outbound_intents i
		WHERE i.tenant_id=? AND i.connection_id=? AND i.status=? AND a.tenant_id=i.tenant_id AND a.intent_id=i.id AND a.status=?`, types.IntentCancelled, now, tenant, connection, types.IntentPending, types.IntentPending).Error; err != nil {
		return channelDBError(ctx, err)
	}
	if err := db.Exec(`UPDATE outbound_intents SET status=?, error_category=?, updated_at=? WHERE tenant_id=? AND connection_id=? AND status=?`, types.IntentCancelled, reason, now, tenant, connection, types.IntentPending).Error; err != nil {
		return channelDBError(ctx, err)
	}
	return nil
}

var _ ChannelStore = (*ChannelRepository)(nil)
