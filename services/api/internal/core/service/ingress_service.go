package service

import (
	"context"
	"github.com/TruongHoang2004/Hexta/packages/shared/events"
	"github.com/TruongHoang2004/Hexta/packages/shared/events/messaging"
	apperrors "github.com/TruongHoang2004/Hexta/packages/shared/pkg/errors"
	"github.com/TruongHoang2004/Hexta/services/api/internal/core/types"
	"github.com/TruongHoang2004/Hexta/services/api/internal/repository"
	"github.com/google/uuid"
	"time"
)

type IngressService struct{ store repository.IngressStore }

func NewIngressService(store repository.IngressStore) *IngressService { return &IngressService{store} }
func (s *IngressService) Consume(ctx context.Context, e messaging.MessageReceived) *apperrors.Error {
	if e.Validate() != nil {
		return apperrors.ErrForbidden(ctx).SetDetail("invalid ingress")
	}
	return s.store.Apply(ctx, e, func(tx repository.MessagingStore, b repository.IngressBinding, duplicate bool) *apperrors.Error {
		// Disabled/version-stale events are durably ignored, never reassigned to another tenant.
		if duplicate || !b.Enabled || !b.Active || b.Version != e.ConnectionVersion || b.Platform != e.Platform || b.AccountID != e.AccountID {
			return nil
		}
		// History, echo and receipts cannot create active turns or schedule AI.
		if e.Direction != messaging.DirectionInbound || e.Purpose != "live" {
			return nil
		}
		// Validate guarantees text-only or media-only content. Media is a placeholder: kinds are recorded, media is
		// never fetched or stored.
		message := types.MessagingMessage{ContentType: types.ContentText, Body: e.Text}
		if len(e.UnsupportedMedia) > 0 {
			message.ContentType, message.MediaKinds = types.ContentUnsupportedMedia, append([]string(nil), e.UnsupportedMedia...)
		}
		name, ok := types.ChannelThreadNames[e.Platform]
		if !ok {
			return nil
		}
		thread, err := tx.UpsertThread(ctx, types.MessagingThread{ID: messagingID(), TenantID: e.TenantID, ExternalKey: e.Platform + ":" + e.Key(), DisplayName: name, CreatedAt: e.OccurredAt, UpdatedAt: e.OccurredAt, ConnectionID: e.ConnectionID, RecipientID: e.ConversationKey})
		if err != nil {
			return err
		}
		// Threads created before channel binding existed are bound on their next inbound message.
		if thread.ConnectionID == "" {
			if err = tx.BindThreadChannel(ctx, e.TenantID, thread.ID, e.ConnectionID, e.ConversationKey); err != nil {
				return err
			}
		}
		turn, err := tx.LatestTurn(ctx, e.TenantID, thread.ID)
		if err != nil {
			return err
		}
		// Placement uses the source event time: a delayed or replayed message older than the closed turn's
		// activity is kept as history on that turn and neither reopens it nor manufactures a new one.
		placement := types.InboundPlacement(turn, e.SourceAt, time.Now().UTC())
		switch placement {
		case types.InboundNewTurn:
			settings, err := tx.ReplySettings(ctx, e.TenantID, false)
			if err != nil {
				return err
			}
			turn = &types.MessagingTurn{ID: messagingID(), TenantID: e.TenantID, ThreadID: thread.ID, Lifecycle: types.TurnOpen, ContextRevision: 1, ControlGeneration: 1, Version: 1, LastActivityAt: e.SourceAt}
			newTurnHandling(turn, settings)
			if err = tx.SaveTurn(ctx, *turn, true); err != nil {
				return err
			}
		case types.InboundHistory:
		default:
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
		message.RequestKey, message.RequestHash, message.Status, message.DeliveryStatus, message.CreatedAt = uuid.NewSHA1(uuid.NameSpaceOID, []byte(e.Identity())).String(), events.Hash(e.Text), "received", types.DeliveryReceived, e.SourceAt
		if err = tx.InsertMessage(ctx, message); err != nil {
			return err
		}
		// A live customer message on a bot-handled turn is routed after commit; history never is.
		if placement != types.InboundHistory {
			if _, err = enqueueBotRouting(ctx, tx, *turn, message, "", time.Now().UTC()); err != nil {
				return err
			}
		}
		if e.OccurredAt.After(thread.UpdatedAt) {
			thread.UpdatedAt = e.OccurredAt
		}
		return tx.SaveThread(ctx, *thread)
	})
}
