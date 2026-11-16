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
		if e.Direction != "inbound" || e.Purpose != "live" {
			return nil
		}
		message := types.MessagingMessage{ContentType: types.ContentText, Body: e.Text}
		switch e.Type {
		case ingress.TypeMessage:
			if e.Text == "" || len([]rune(e.Text)) > 4000 || len(e.Attachments) > 0 {
				return nil
			}
		case ingress.TypeUnsupportedMedia:
			// Placeholder only: kinds are recorded, media is never fetched or stored.
			if e.Text != "" || len(e.Attachments) == 0 || len(e.Attachments) > 10 {
				return nil
			}
			for _, a := range e.Attachments {
				if !ingress.MediaKinds[a.Type] {
					return nil
				}
				message.MediaKinds = append(message.MediaKinds, a.Type)
			}
			message.ContentType = types.ContentUnsupportedMedia
		default:
			return nil
		}
		name, ok := types.ChannelThreadNames[e.Platform]
		if !ok {
			return nil
		}
		thread, err := tx.UpsertThread(ctx, types.MessagingThread{ID: messagingID(), TenantID: e.TenantID, ExternalKey: e.Platform + ":" + e.Key(), DisplayName: name, CreatedAt: e.ReceivedAt, UpdatedAt: e.ReceivedAt, ConnectionID: e.ConnectionID, RecipientID: e.ConversationKey})
		if err != nil {
			return err
		}
		// Threads created before channel binding existed are bound on their next inbound message.
		if thread.ConnectionID == "" {
			if err = tx.BindThreadChannel(ctx, e.TenantID, thread.ID, e.ConnectionID, e.ConversationKey); err != nil {
				return err
			}
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
		message.ID, message.TenantID, message.ThreadID, message.TurnID, message.Direction, message.ActorID = messagingID(), e.TenantID, thread.ID, turn.ID, "inbound", e.Sender
		message.RequestKey, message.RequestHash, message.Status, message.DeliveryStatus, message.CreatedAt = uuid.NewSHA1(uuid.NameSpaceOID, []byte(e.Identity())).String(), ingress.Hash(e.Text), "received", types.DeliveryReceived, e.SourceAt
		if err = tx.InsertMessage(ctx, message); err != nil {
			return err
		}
		if e.ReceivedAt.After(thread.UpdatedAt) {
			thread.UpdatedAt = e.ReceivedAt
		}
		return tx.SaveThread(ctx, *thread)
	})
}
