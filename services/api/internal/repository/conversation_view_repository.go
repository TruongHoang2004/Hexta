package repository

import (
	"context"

	apperrors "github.com/TruongHoang2004/Hexta/packages/shared/pkg/errors"
	"github.com/TruongHoang2004/Hexta/services/api/internal/core/types"
	"gorm.io/gorm"
)

// ConversationViewStore maintains the conversation read model. Only the projector and repair tooling write it.
type ConversationViewStore interface {
	Apply(context.Context, string, types.ConversationChanged) *apperrors.Error
	Rebuild(context.Context, string) *apperrors.Error
	Quarantine(context.Context, string, string) error
}

type ConversationViewRepository struct{ db *gorm.DB }

func NewConversationViewRepository(db *gorm.DB) *ConversationViewRepository {
	return &ConversationViewRepository{db}
}

// Apply upserts one snapshot. Replays and out-of-order deliveries are harmless: an older (turn version, thread
// update time) never overwrites a newer row, while an identical snapshot is re-applied idempotently.
func (r *ConversationViewRepository) Apply(ctx context.Context, tenant string, e types.ConversationChanged) *apperrors.Error {
	err := r.db.WithContext(ctx).Exec(`
INSERT INTO messaging_conversation_views (tenant_id, thread_id, external_key, display_name, thread_created_at, thread_updated_at,
  connection_id, recipient_id, turn_id, lifecycle, control, assignee_id, needs_human, context_revision, control_generation, turn_version, last_activity_at)
VALUES (?, ?, ?, ?, ?, ?, NULLIF(?, '')::uuid, NULLIF(?, ''), ?, ?, ?, ?, ?, ?, ?, ?, ?)
ON CONFLICT (tenant_id, thread_id) DO UPDATE SET
  external_key = EXCLUDED.external_key, display_name = EXCLUDED.display_name, thread_created_at = EXCLUDED.thread_created_at,
  thread_updated_at = EXCLUDED.thread_updated_at, connection_id = EXCLUDED.connection_id, recipient_id = EXCLUDED.recipient_id,
  turn_id = EXCLUDED.turn_id, lifecycle = EXCLUDED.lifecycle, control = EXCLUDED.control, assignee_id = EXCLUDED.assignee_id,
  needs_human = EXCLUDED.needs_human, context_revision = EXCLUDED.context_revision, control_generation = EXCLUDED.control_generation,
  turn_version = EXCLUDED.turn_version, last_activity_at = EXCLUDED.last_activity_at
WHERE (messaging_conversation_views.turn_version, messaging_conversation_views.thread_updated_at) <= (EXCLUDED.turn_version, EXCLUDED.thread_updated_at)`,
		tenant, e.ThreadID, e.ExternalKey, e.DisplayName, e.ThreadCreatedAt, e.ThreadUpdatedAt,
		e.ConnectionID, e.RecipientID, e.TurnID, e.Lifecycle, e.Control, e.AssigneeID, e.NeedsHuman, e.ContextRevision, e.ControlGeneration, e.TurnVersion, e.LastActivityAt).Error
	if err != nil {
		return salesDBError(ctx, err)
	}
	return nil
}

// Rebuild recomputes a tenant's view rows from the authoritative tables (repair after a lost or poisoned event).
func (r *ConversationViewRepository) Rebuild(ctx context.Context, tenant string) *apperrors.Error {
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec(`DELETE FROM messaging_conversation_views WHERE tenant_id = ?`, tenant).Error; err != nil {
			return err
		}
		return tx.Exec(`
INSERT INTO messaging_conversation_views (tenant_id, thread_id, external_key, display_name, thread_created_at, thread_updated_at,
  connection_id, recipient_id, turn_id, lifecycle, control, assignee_id, needs_human, context_revision, control_generation, turn_version, last_activity_at)
SELECT th.tenant_id, th.id, th.external_key, th.display_name, th.created_at, th.updated_at, th.connection_id, th.recipient_id,
  t.id, t.lifecycle, t.control, t.assignee_id, t.needs_human, t.context_revision, t.control_generation, t.version, t.last_activity_at
FROM messaging_threads th
JOIN messaging_turns t ON t.tenant_id = th.tenant_id AND t.thread_id = th.id AND t.lifecycle = 'open'
WHERE th.tenant_id = ?`, tenant).Error
	})
	if err != nil {
		return salesDBError(ctx, err)
	}
	return nil
}

// Quarantine stores only a hash and safe reason for an event that could not be decoded.
func (r *ConversationViewRepository) Quarantine(ctx context.Context, identity, reason string) error {
	return NewDeliveryRepository(r.db).Quarantine(ctx, identity, reason)
}

// ConversationViews returns the read-model writer sharing this repository's connection.
func (r *MessagingRepository) ConversationViews() *ConversationViewRepository {
	return NewConversationViewRepository(r.db)
}
