package service

import (
	"context"
	"github.com/TruongHoang2004/Hexta/packages/shared/ingress"
	apperrors "github.com/TruongHoang2004/Hexta/packages/shared/pkg/errors"
	"github.com/TruongHoang2004/Hexta/services/api/internal/core/types"
	"github.com/TruongHoang2004/Hexta/services/api/internal/repository"
	"github.com/google/uuid"
)

type IngressService struct{ store repository.IngressStore }

func NewIngressService(store repository.IngressStore) *IngressService { return &IngressService{store} }
func (s *IngressService) Consume(ctx context.Context, e ingress.Event) *apperrors.Error {
	if e.Validate() != nil {
		return apperrors.ErrForbidden(ctx).SetDetail("invalid ingress")
	}
	return s.store.Apply(ctx, e, func(tx repository.MessagingStore, b repository.IngressBinding, duplicate bool) *apperrors.Error {
		// Disabled/version-stale events are durably ignored, never reassigned to another tenant.
		if duplicate || !b.Enabled || !b.Active || b.Version != e.ConnectionVersion || b.Platform != e.Platform || b.AccountID != e.AccountID {
			return nil
		}
		// History, echo and receipts cannot create active turns or schedule AI.
		if e.Type != "message" || e.Direction != "inbound" || e.Purpose != "live" || e.Text == "" || len([]rune(e.Text)) > 4000 || len(e.Attachments) > 0 {
			return nil
		}
		thread, err := tx.UpsertThread(ctx, types.MessagingThread{ID: messagingID(), TenantID: e.TenantID, ExternalKey: "facebook:" + e.Key(), DisplayName: "Facebook Messenger", CreatedAt: e.ReceivedAt, UpdatedAt: e.ReceivedAt})
		if err != nil {
			return err
		}
		turn, err := tx.CurrentTurn(ctx, e.TenantID, thread.ID)
		if err != nil {
			return err
		}
		if turn == nil {
			turn = &types.MessagingTurn{ID: messagingID(), TenantID: e.TenantID, ThreadID: thread.ID, Lifecycle: "open", Control: "sales", NeedsHuman: true, ContextRevision: 1, ControlGeneration: 1, Version: 1, LastActivityAt: e.SourceAt}
			if err = tx.SaveTurn(ctx, *turn, true); err != nil {
				return err
			}
		} else {
			turn.Version++
			turn.ContextRevision++
			if e.SourceAt.After(turn.LastActivityAt) {
				turn.LastActivityAt = e.SourceAt
			}
			if err = tx.SaveTurn(ctx, *turn, false); err != nil {
				return err
			}
		}
		if err = tx.InsertMessage(ctx, types.MessagingMessage{ID: messagingID(), TenantID: e.TenantID, ThreadID: thread.ID, TurnID: turn.ID, Direction: "inbound", ActorID: e.Sender, Body: e.Text, RequestKey: uuid.NewSHA1(uuid.NameSpaceOID, []byte(e.Identity())).String(), RequestHash: ingress.Hash(e.Text), Status: "received", CreatedAt: e.SourceAt}); err != nil {
			return err
		}
		if e.ReceivedAt.After(thread.UpdatedAt) {
			thread.UpdatedAt = e.ReceivedAt
		}
		return tx.SaveThread(ctx, *thread)
	})
}
