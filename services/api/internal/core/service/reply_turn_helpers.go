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
}

// newTurnHandling sets a new turn's starting control. Once the owner has configured a reply mode the bot handles
// new turns (suggest mode never sends without a sale, so this is safe by default); before that, a new turn keeps
// the pre-bot default of waiting for staff with needs_human raised and no reason.
func newTurnHandling(turn *types.MessagingTurn, settings types.ReplySettings) {
	if settings.Configured {
		turn.Control, turn.NeedsHuman = "system", false
		return
	}
	turn.Control, turn.NeedsHuman = "sales", true
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
