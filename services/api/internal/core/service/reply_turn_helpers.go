package service

import (
	"context"
	"time"

	apperrors "github.com/TruongHoang2004/Hexta/packages/shared/pkg/errors"
	"github.com/TruongHoang2004/Hexta/services/api/internal/core/types"
	"github.com/TruongHoang2004/Hexta/services/api/internal/repository"
)

// applyHandoff is the single transition to WAITING_STAFF, used by staff handoff and by the reply router. The caller
// bumps the control generation and version and saves the turn.
func applyHandoff(turn *types.MessagingTurn, reason string) {
	turn.AssigneeID = nil
	turn.Control = "sales"
	turn.NeedsHuman = true
	turn.NeedsHumanReason = reason
	turn.ReviewPending = false
}

// newTurnHandling sets a new turn's starting control from the reply mode. With suggest or auto_send the bot handles
// new turns (suggest never sends without a sale). With off, the default for every tenant, a new turn waits for
// staff with needs_human raised and no reason, and no routing job is created.
func newTurnHandling(turn *types.MessagingTurn, settings types.ReplySettings) {
	if settings.BotOn() {
		turn.Control, turn.NeedsHuman = "system", false
		return
	}
	turn.Control, turn.NeedsHuman = "sales", true
}

// botHandoff moves a BOT_ACTIVE turn to WAITING_STAFF as the bot: undispatched bot sends are cancelled (before the
// turn row is written, matching the delivery consumer's lock order), the control generation fences every pending
// proposal and in-flight routing job, and the handoff is audited. The caller holds the thread lock.
func botHandoff(ctx context.Context, r repository.MessagingStore, turn *types.MessagingTurn, reason string, detail map[string]any, now time.Time) *apperrors.Error {
	cancelled, err := r.CancelPendingBotSends(ctx, turn.TenantID, turn.ID, now)
	if err != nil {
		return err
	}
	before := turn.HandlingState()
	applyHandoff(turn, reason)
	turn.ControlGeneration++
	turn.Version++
	if err := r.SaveTurn(ctx, *turn, false); err != nil {
		return err
	}
	payload := map[string]any{
		"version": turn.Version, "control_generation": turn.ControlGeneration, "before": before, "after": turn.HandlingState(),
		"reason": reason, "cancelled_bot_sends": cancelled,
	}
	for key, value := range detail {
		payload[key] = value
	}
	return r.Audit(ctx, turn.TenantID, types.BotActorID, "messaging.handoff", "messaging_turn", turn.ID, payload)
}

// botIntake applies the reply mode to a live customer message appended to an existing turn, before the caller
// records the message, bumps the turn and saves it. On a BOT_ACTIVE turn while the bot is off the turn is handed to
// staff (reason bot_off) instead of being routed; otherwise the message will supersede the turn's live proposal
// (enqueueBotRouting), so a proposal waiting for review no longer holds the turn in the waiting queue.
func botIntake(ctx context.Context, r repository.MessagingStore, turn *types.MessagingTurn, now time.Time) *apperrors.Error {
	if turn.HandlingState() != "BOT_ACTIVE" {
		return nil
	}
	settings, err := r.ReplySettings(ctx, turn.TenantID, false)
	if err != nil {
		return err
	}
	if !settings.BotOn() {
		return botHandoff(ctx, r, turn, types.HandoffReasonBotOff, map[string]any{"router_reason": types.HandoffReasonBotOff}, now)
	}
	turn.ReviewPending = false
	return nil
}

// errBotOff refuses an action that needs the bot while the tenant's reply mode is off.
func errBotOff(ctx context.Context, action string) *apperrors.Error {
	return apperrors.ErrConflict(ctx, "bot", "is off: choose a reply mode in Settings before "+action)
}

// enqueueBotRouting records a durable router job for a customer message on a BOT_ACTIVE turn (or, with requestedBy,
// a person's regenerate request). The router runs after commit, so no transaction is held across script lookup or
// AI calls. Media-only messages are routed too: the router hands them to staff because the bot cannot read them.
func enqueueBotRouting(ctx context.Context, r repository.MessagingStore, turn types.MessagingTurn, message types.MessagingMessage, requestedBy string, now time.Time) (*types.ReplyProposal, *apperrors.Error) {
	if requestedBy == "" && turn.HandlingState() != "BOT_ACTIVE" {
		return nil, nil
	}
	job := types.ReplyProposal{
		ID: messagingID(), TenantID: turn.TenantID, ThreadID: turn.ThreadID, TurnID: turn.ID, TriggerMessageID: message.ID,
		ControlGeneration: turn.ControlGeneration, ContextRevision: turn.ContextRevision, Status: types.ProposalRouting,
		RequestedBy: requestedBy, AvailableAt: now, Version: 1, CreatedAt: now, UpdatedAt: now,
	}
	if err := r.InsertRouting(ctx, job); err != nil {
		return nil, err
	}
	return &job, nil
}
