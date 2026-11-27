package repository

import (
	"context"
	"github.com/TruongHoang2004/Hexta/packages/shared/events/messaging"
	apperrors "github.com/TruongHoang2004/Hexta/packages/shared/pkg/errors"
	"gorm.io/gorm"
	"time"
)

type IngressBinding struct {
	ID, TenantID, Platform, AccountID string
	Version                           int64
	Enabled, Active                   bool
}
type IngressStore interface {
	Apply(context.Context, messaging.MessageReceived, func(MessagingStore, IngressBinding, bool) *apperrors.Error) *apperrors.Error
}
type IngressRepository struct{ db *gorm.DB }

func NewIngressRepository(db *gorm.DB) *IngressRepository { return &IngressRepository{db} }

// The binding lock, consumer dedup and business mutation share a single transaction.
func (r *IngressRepository) Apply(ctx context.Context, e messaging.MessageReceived, fn func(MessagingStore, IngressBinding, bool) *apperrors.Error) *apperrors.Error {
	var failure *apperrors.Error
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var b IngressBinding
		if err := tx.Raw(`SELECT c.*, t.status='active' AS active FROM messaging_connections c JOIN tenants t ON t.id=c.tenant_id WHERE c.id=? AND c.tenant_id=? FOR SHARE OF c,t`, e.ConnectionID, e.TenantID).Scan(&b).Error; err != nil {
			return err
		}
		if b.ID == "" {
			return tx.Exec("INSERT INTO messaging_ingress_quarantine(identity,reason,received_at) VALUES(?,?,?) ON CONFLICT(identity) DO NOTHING", e.Identity(), "untrusted_binding", time.Now().UTC()).Error
		}
		result := tx.Exec(`INSERT INTO messaging_ingress(identity,event_id,tenant_id,connection_id,message_id,external_conversation_key,source_identity,source_revision,reply_to,source_timestamp,acquisition,purpose,received_at,outcome) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?) ON CONFLICT(identity) DO NOTHING`, e.Identity(), e.EventID, e.TenantID, e.ConnectionID, e.MessageID, e.ConversationKey, e.SourceIdentity, e.SourceRevision, e.ReplyTo, e.SourceAt, e.Acquisition, e.Purpose, e.OccurredAt, "handled")
		if result.Error != nil {
			return result.Error
		}
		duplicate := result.RowsAffected == 0
		repo := newTxMessagingRepository(tx)
		if failure = fn(repo, b, duplicate); failure == nil {
			failure = repo.flushConversationEvents(ctx)
		}
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

// Quarantine stores only a hash and safe reason, never arbitrary broker payloads.
func (r *IngressRepository) Quarantine(ctx context.Context, identity, reason string) error {
	return r.db.WithContext(ctx).Exec("INSERT INTO messaging_ingress_quarantine(identity,reason,received_at) VALUES(?,?,?) ON CONFLICT(identity) DO NOTHING", identity, reason, time.Now().UTC()).Error
}
