package service

import (
	"context"
	"encoding/json"
	"github.com/TruongHoang2004/Hexta/packages/shared/events"
	"strings"
	"time"

	"github.com/TruongHoang2004/Hexta/packages/shared/events/delivery"
	apperrors "github.com/TruongHoang2004/Hexta/packages/shared/pkg/errors"
	"github.com/TruongHoang2004/Hexta/services/api/internal/core/model"
	"github.com/TruongHoang2004/Hexta/services/api/internal/core/types"
	"github.com/TruongHoang2004/Hexta/services/api/internal/repository"
	"github.com/google/uuid"
)

type MessagingCommandService struct{ repo repository.MessagingStore }

func NewMessagingCommandService(repo repository.MessagingStore) *MessagingCommandService {
	return &MessagingCommandService{repo: repo}
}
func (s *MessagingCommandService) authorized(ctx context.Context, tenant, actor string, fn func(repository.MessagingStore, bool) *apperrors.Error) *apperrors.Error {
	if actor == "" {
		return apperrors.ErrUnauthorized(ctx)
	}
	if _, err := uuid.Parse(tenant); err != nil {
		return apperrors.ErrBadRequest(ctx).SetDetail("valid tenant ID required")
	}
	return s.repo.Within(ctx, func(r repository.MessagingStore) *apperrors.Error {
		scope, err := r.Scope(ctx, tenant, actor)
		if err != nil {
			return err
		}
		if !scope.Active || !scope.Member {
			return apperrors.ErrForbidden(ctx).SetDetail("active tenant membership required")
		}
		return fn(r, scope.Owner)
	})
}
func (s *MessagingCommandService) Inbound(ctx context.Context, tenant, actor string, input types.MessagingInbound) (*types.MessagingMessage, *apperrors.Error) {
	input.ExternalKey = strings.TrimSpace(input.ExternalKey)
	input.DisplayName = strings.TrimSpace(input.DisplayName)
	input.Body = strings.TrimSpace(input.Body)
	if types.ChannelThreadPlatform(input.ExternalKey) != "" || !messagingText(input.ExternalKey, 120) || !messagingText(input.DisplayName, 120) || !messagingText(input.Body, 4000) {
		return nil, apperrors.ErrBadRequest(ctx).SetDetail("external_key/display_name (1–120) and body (1–4000) required")
	}
	if err := messagingKey(ctx, input.RequestKey); err != nil {
		return nil, err
	}
	hash := messagingHash(struct {
		Actor string
		Input types.MessagingInbound
	}{actor, input})
	var result *types.MessagingMessage
	err := s.authorized(ctx, tenant, actor, func(r repository.MessagingStore, _ bool) *apperrors.Error {
		now := time.Now().UTC()
		thread, err := r.UpsertThread(ctx, types.MessagingThread{ID: messagingID(), TenantID: tenant, ExternalKey: input.ExternalKey, DisplayName: input.DisplayName, CreatedAt: now, UpdatedAt: now})
		if err != nil {
			return err
		}
		old, err := r.MessageByKey(ctx, tenant, input.RequestKey)
		if err != nil {
			return err
		}
		if old != nil {
			if old.RequestHash != hash {
				return apperrors.ErrConflict(ctx, "request key", "already used for different input")
			}
			result = old
			return nil
		}
		turn, err := r.LatestTurn(ctx, tenant, thread.ID)
		if err != nil {
			return err
		}
		// Simulated input is always new at processing time, so a closed latest turn opens a new unassigned turn.
		create := types.InboundPlacement(turn, now, now) != types.InboundAppend
		if create {
			settings, err := r.ReplySettings(ctx, tenant, false)
			if err != nil {
				return err
			}
			turn = &types.MessagingTurn{ID: messagingID(), TenantID: tenant, ThreadID: thread.ID, Lifecycle: types.TurnOpen, ControlGeneration: 1}
			newTurnHandling(turn, settings)
		}
		turn.ContextRevision++
		turn.Version++
		turn.LastActivityAt = now
		if err := r.SaveTurn(ctx, *turn, create); err != nil {
			return err
		}
		result = &types.MessagingMessage{ID: messagingID(), TenantID: tenant, ThreadID: thread.ID, TurnID: turn.ID, Direction: "inbound", ActorID: actor, Body: input.Body, RequestKey: input.RequestKey, RequestHash: hash, Status: "simulated", CreatedAt: now}
		if err := r.InsertMessage(ctx, *result); err != nil {
			return err
		}
		if _, err := enqueueBotRouting(ctx, r, *turn, *result, "", now); err != nil {
			return err
		}
		thread.UpdatedAt = now
		if err := r.SaveThread(ctx, *thread); err != nil {
			return err
		}
		return r.Audit(ctx, tenant, actor, "messaging.simulated_inbound", "messaging_thread", thread.ID, map[string]string{"message_id": result.ID})
	})
	return result, err
}

// lockedTurn uses the thread as the shared gate for input, ownership and reply commands.
func messagingLockedTurn(ctx context.Context, r repository.MessagingStore, tenant, id string) (*types.MessagingTurn, *apperrors.Error) {
	turn, err := r.Turn(ctx, tenant, id)
	if err != nil {
		return nil, err
	}
	if _, err := r.Thread(ctx, tenant, turn.ThreadID, true); err != nil {
		return nil, err
	}
	return r.Turn(ctx, tenant, id)
}
func (s *MessagingCommandService) Command(ctx context.Context, tenant, actor, id, action string, input types.MessagingCommand) (*types.MessagingTurn, *apperrors.Error) {
	if input.ExpectedVersion < 1 {
		return nil, apperrors.ErrBadRequest(ctx).SetDetail("positive expected_version required")
	}
	if action == "handoff" {
		switch input.Reason {
		case "customer_requested_staff", "insufficient_information", "ai_error", "business_review", "staff_handoff":
		default:
			return nil, apperrors.ErrBadRequest(ctx).SetDetail("valid handoff reason required")
		}
	}
	var result *types.MessagingTurn
	err := s.authorized(ctx, tenant, actor, func(r repository.MessagingStore, owner bool) *apperrors.Error {
		// Lock target membership before the thread, consistent with actor membership checks.
		if action == "assign" {
			if !owner {
				return apperrors.ErrForbidden(ctx).SetDetail("tenant owner required")
			}
			if input.AssigneeID != nil {
				exists, err := r.MemberExists(ctx, tenant, *input.AssigneeID)
				if err != nil {
					return err
				}
				if !exists {
					return apperrors.ErrBadRequest(ctx).SetDetail("assignee must be a current tenant member")
				}
			}
		}
		turn, err := messagingLockedTurn(ctx, r, tenant, id)
		if err != nil {
			return err
		}
		if turn.Lifecycle != "open" || turn.Version != input.ExpectedVersion {
			return apperrors.ErrConflict(ctx, "turn", "changed; refresh before retrying")
		}
		before := *turn
		// Control-changing commands fence stale AI/output by bumping the control generation; resolve does not
		// change control.
		fence := true
		switch action {
		case "handoff":
			if turn.AssigneeID != nil && *turn.AssigneeID != actor && !owner {
				return apperrors.ErrForbidden(ctx).SetDetail("current assignee or tenant owner required")
			}
			if turn.HandlingState() == "WAITING_STAFF" {
				return apperrors.ErrConflict(ctx, "turn", "already waiting for staff")
			}
			applyHandoff(turn, input.Reason)
		case "return-to-bot":
			if !owner && (turn.AssigneeID == nil || *turn.AssigneeID != actor) {
				return apperrors.ErrForbidden(ctx).SetDetail("current assignee or tenant owner required")
			}
			if turn.HandlingState() != "STAFF_ACTIVE" && turn.HandlingState() != "WAITING_STAFF" {
				return apperrors.ErrConflict(ctx, "turn", "not under staff control")
			}
			// Explicit, authorized return is the only path back to automation; it clears the escalation.
			turn.AssigneeID = nil
			turn.Control = "system"
			turn.NeedsHuman = false
			turn.NeedsHumanReason = ""
		case "claim":
			if turn.AssigneeID != nil && *turn.AssigneeID != actor {
				return apperrors.ErrConflict(ctx, "turn", "already assigned")
			}
			// needs_human survives the claim: the assigned sale resolves it explicitly.
			turn.AssigneeID = &actor
			turn.Control = "sales"
		case "assign":
			turn.AssigneeID = input.AssigneeID
			turn.Control = "sales"
			// Unassignment returns the work to the staff queue, which must not expire unattended.
			if input.AssigneeID == nil {
				turn.NeedsHuman = true
			}
		case "resolve":
			if turn.AssigneeID == nil || *turn.AssigneeID != actor || turn.HandlingState() != "STAFF_ACTIVE" {
				return apperrors.ErrForbidden(ctx).SetDetail("only the assigned sale can resolve needs-human")
			}
			if !turn.NeedsHuman {
				return apperrors.ErrConflict(ctx, "turn", "does not need a person")
			}
			// Resolving restarts the inactivity window but keeps staff control; nothing returns to the bot.
			turn.NeedsHuman = false
			turn.NeedsHumanReason = ""
			if now := time.Now().UTC(); now.After(turn.LastActivityAt) {
				turn.LastActivityAt = now
			}
			fence = false
		default:
			return apperrors.ErrBadRequest(ctx).SetDetail("unsupported conversation command")
		}
		cancelled := 0
		if fence {
			turn.ControlGeneration++
			// A control change stops automatic sending, including bot replies already queued: undispatched sends
			// are cancelled here (before the turn row is written, matching the delivery consumer's lock order);
			// pending proposals become stale through the generation bump.
			if cancelled, err = r.CancelPendingBotSends(ctx, tenant, id, time.Now().UTC()); err != nil {
				return err
			}
		}
		turn.Version++
		if err := r.SaveTurn(ctx, *turn, false); err != nil {
			return err
		}
		result = turn
		return r.Audit(ctx, tenant, actor, "messaging."+action, "messaging_turn", id, map[string]any{
			"version": turn.Version, "control_generation": turn.ControlGeneration,
			"before": before.HandlingState(), "after": turn.HandlingState(),
			"previous_assignee_id": before.AssigneeID, "assignee_id": turn.AssigneeID,
			"reason": input.Reason, "needs_human": turn.NeedsHuman, "previous_needs_human_reason": before.NeedsHumanReason,
			"cancelled_bot_sends": cancelled,
		})
	})
	return result, err
}
func (s *MessagingCommandService) SendSimulation(ctx context.Context, tenant, actor, id string, input types.MessagingCommand) (*types.MessagingMessage, *apperrors.Error) {
	input.Body = strings.TrimSpace(input.Body)
	if !messagingText(input.Body, 4000) || input.ExpectedVersion < 1 || input.ContextRevision < 1 {
		return nil, apperrors.ErrBadRequest(ctx).SetDetail("body (1–4000) and positive version/context required")
	}
	if err := messagingKey(ctx, input.RequestKey); err != nil {
		return nil, err
	}
	// Keep the original serialized command shape: adding handoff-only fields
	// must not invalidate request hashes for previously saved simulated replies.
	type replyFingerprint struct {
		ExpectedVersion, ContextRevision int64
		AssigneeID                       *string
		Body, RequestKey                 string
	}
	hash := messagingHash(struct {
		Actor, Turn string
		Input       replyFingerprint
	}{actor, id, replyFingerprint{input.ExpectedVersion, input.ContextRevision, input.AssigneeID, input.Body, input.RequestKey}})
	var result *types.MessagingMessage
	err := s.authorized(ctx, tenant, actor, func(r repository.MessagingStore, _ bool) *apperrors.Error {
		turn, err := messagingLockedTurn(ctx, r, tenant, id)
		if err != nil {
			return err
		}
		source, sourceErr := r.Thread(ctx, tenant, turn.ThreadID, false)
		if sourceErr != nil {
			return sourceErr
		}
		if types.ChannelThreadPlatform(source.ExternalKey) != "" {
			return apperrors.ErrForbidden(ctx).SetDetail("real channel conversations must use channel send")
		}
		if turn.AssigneeID == nil || *turn.AssigneeID != actor || turn.HandlingState() != "STAFF_ACTIVE" {
			return apperrors.ErrForbidden(ctx).SetDetail("only the assigned sale under sales control may reply")
		}
		old, err := r.MessageByKey(ctx, tenant, input.RequestKey)
		if err != nil {
			return err
		}
		if old != nil {
			if old.RequestHash != hash {
				return apperrors.ErrConflict(ctx, "request key", "already used for different input")
			}
			result = old
			return nil
		}
		if turn.Lifecycle != "open" || turn.Version != input.ExpectedVersion || turn.ContextRevision != input.ContextRevision {
			return apperrors.ErrConflict(ctx, "turn", "new activity or ownership change; refresh and review before replying")
		}
		now := time.Now().UTC()
		result = &types.MessagingMessage{ID: messagingID(), TenantID: tenant, ThreadID: turn.ThreadID, TurnID: id, Direction: "outbound", ActorID: actor, Body: input.Body, RequestKey: input.RequestKey, RequestHash: hash, Status: "simulated", CreatedAt: now}
		if err := r.InsertMessage(ctx, *result); err != nil {
			return err
		}
		// A simulated reply has no delivery step, so it is a confirmed send and counts as activity.
		turn.Version++
		turn.ContextRevision++
		turn.LastActivityAt = now
		if err := r.SaveTurn(ctx, *turn, false); err != nil {
			return err
		}
		thread, err := r.Thread(ctx, tenant, turn.ThreadID, false)
		if err != nil {
			return err
		}
		thread.UpdatedAt = now
		if err := r.SaveThread(ctx, *thread); err != nil {
			return err
		}
		return r.Audit(ctx, tenant, actor, "messaging.simulated_reply", "messaging_turn", id, map[string]string{"message_id": result.ID})
	})
	return result, err
}

const outboundIntentTTL = 10 * time.Minute

// SendChannel records an authorized real-channel reply and its single-attempt delivery intent in one transaction.
// The connector performs the provider call later; this request never calls Meta and never retries.
func (s *MessagingCommandService) SendChannel(ctx context.Context, tenant, actor, id string, input types.ChannelSendInput) (*types.MessagingMessage, *apperrors.Error) {
	input.Text = strings.TrimSpace(input.Text)
	if !messagingText(input.Text, delivery.MaxTextRunes) || input.ExpectedVersion < 1 || input.ContextRevision < 1 {
		return nil, apperrors.ErrBadRequest(ctx).SetDetail("text (1–2000) and positive version/context required")
	}
	if _, err := uuid.Parse(input.ClientMessageID); err != nil {
		return nil, apperrors.ErrBadRequest(ctx).SetDetail("client_message_id must be a UUID")
	}
	hash := messagingHash(struct{ Actor, Turn, Text string }{actor, id, input.Text})
	var result *types.MessagingMessage
	err := s.authorized(ctx, tenant, actor, func(r repository.MessagingStore, _ bool) *apperrors.Error {
		turn, err := messagingLockedTurn(ctx, r, tenant, id)
		if err != nil {
			return err
		}
		thread, err := r.Thread(ctx, tenant, turn.ThreadID, false)
		if err != nil {
			return err
		}
		if thread.ConnectionID == "" || thread.RecipientID == "" {
			return apperrors.ErrConflict(ctx, "conversation", "has no connected channel")
		}
		if turn.AssigneeID == nil || *turn.AssigneeID != actor || turn.HandlingState() != "STAFF_ACTIVE" {
			return apperrors.ErrForbidden(ctx).SetDetail("only the assigned sale under sales control may reply")
		}
		old, err := r.MessageByKey(ctx, tenant, input.ClientMessageID)
		if err != nil {
			return err
		}
		if old != nil {
			if old.RequestHash != hash {
				return apperrors.ErrConflict(ctx, "client message ID", "already used for different input")
			}
			result = old
			return nil
		}
		if turn.Lifecycle != "open" || turn.Version != input.ExpectedVersion || turn.ContextRevision != input.ContextRevision {
			return apperrors.ErrConflict(ctx, "turn", "new activity or ownership change; refresh and review before replying")
		}
		conn, err := r.SendConnection(ctx, tenant, thread.ConnectionID)
		if err != nil {
			return err
		}
		if conn.Status != types.ChannelStatusActive || !conn.Enabled {
			return apperrors.ErrConflict(ctx, conn.Platform+" connection", "is "+conn.Status+"; sending is unavailable")
		}
		now := time.Now().UTC()
		var intentID string
		result, intentID, err = enqueueChannelReply(ctx, r, *conn, *thread, id, actor, input.Text, input.ClientMessageID, hash, now)
		if err != nil {
			return err
		}
		// A pending channel send is not activity: the turn's inactivity clock advances only when the connector
		// confirms it was sent (DeliveryService). Failed and unknown outcomes never count.
		turn.Version++
		turn.ContextRevision++
		if err := r.SaveTurn(ctx, *turn, false); err != nil {
			return err
		}
		thread.UpdatedAt = now
		if err := r.SaveThread(ctx, *thread); err != nil {
			return err
		}
		return r.Audit(ctx, tenant, actor, "messaging.channel_reply", "messaging_turn", id, map[string]string{"message_id": result.ID, "intent_id": intentID})
	})
	return result, err
}

// enqueueChannelReply records an outbound channel message with its single-attempt delivery intent and outbox event.
// It is the one outbound pipeline for staff replies, proposal sends and bot automatic sends; the caller has locked
// the thread, authorized the sender and checked that conn is active.
func enqueueChannelReply(ctx context.Context, r repository.MessagingStore, conn types.ChannelConnection, thread types.MessagingThread, turnID, actor, text, key, hash string, now time.Time) (*types.MessagingMessage, string, *apperrors.Error) {
	tenant := thread.TenantID
	message := &types.MessagingMessage{ID: messagingID(), TenantID: tenant, ThreadID: thread.ID, TurnID: turnID, Direction: "outbound", ActorID: actor, Body: text, RequestKey: key, RequestHash: hash, Status: "channel", CreatedAt: now, ContentType: types.ContentText, DeliveryStatus: types.DeliverySending}
	if err := r.InsertMessage(ctx, *message); err != nil {
		return nil, "", err
	}
	intent := types.OutboundIntent{ID: messagingID(), TenantID: tenant, ConnectionID: conn.ID, MessageID: message.ID, ClientMessageID: key, RequestHash: hash, RecipientID: thread.RecipientID, Status: types.IntentPending, CreatedAt: now, UpdatedAt: now}
	attempt := types.OutboundAttempt{ID: messagingID(), TenantID: tenant, IntentID: intent.ID, Status: types.IntentPending, ConnectionVersion: conn.Version, CreatedAt: now}
	meta := events.Meta{EventID: uuid.NewString(), CorrelationID: message.ID, TenantID: tenant, OccurredAt: now}
	payload, _ := json.Marshal(delivery.SendRequested{Meta: meta, IntentID: intent.ID, AttemptID: attempt.ID, ConnectionID: conn.ID, ConnectionVersion: conn.Version, Platform: conn.Platform, AccountID: conn.AccountID, RecipientID: thread.RecipientID, Text: text, ExpiresAt: now.Add(outboundIntentTTL)})
	event := &model.DomainEvent{ID: meta.EventID, TenantID: tenant, EventType: model.DomainEventOutboundIntent, AggregateID: intent.ID, Payload: string(payload), State: model.DomainEventStatePending}
	if err := r.InsertOutbound(ctx, intent, attempt, event); err != nil {
		return nil, "", err
	}
	return message, intent.ID, nil
}
