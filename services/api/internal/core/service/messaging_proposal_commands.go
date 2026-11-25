package service

import (
	"context"
	"strings"
	"time"

	apperrors "github.com/TruongHoang2004/Hexta/packages/shared/pkg/errors"
	"github.com/TruongHoang2004/Hexta/services/api/internal/core/types"
	"github.com/TruongHoang2004/Hexta/services/api/internal/repository"
	"github.com/google/uuid"
)

// lockedProposal loads a proposal of the locked turn; a proposal of another turn is not found.
func lockedProposal(ctx context.Context, r repository.MessagingStore, tenant, turn, id string) (*types.ReplyProposal, *apperrors.Error) {
	if _, err := uuid.Parse(id); err != nil {
		return nil, apperrors.ErrBadRequest(ctx).SetDetail("invalid proposal ID")
	}
	p, err := r.Proposal(ctx, tenant, id, true)
	if err != nil {
		return nil, err
	}
	if p.TurnID != turn {
		return nil, apperrors.ErrNotFound(ctx, "proposal", "not found")
	}
	return p, nil
}

// SendProposal sends a pending bot proposal as-is (empty text) or edited (MVP 5 FR-5.3). Go authorizes it:
//   - the proposal must still be pending and not stale (same control generation, open turn);
//   - on a STAFF_ACTIVE turn only the assigned sale may send;
//   - on a BOT_ACTIVE turn the sender self-claims the unassigned turn atomically with the send, so the only sender
//     is again the assigned sale; the claim bumps the control generation and cancels undispatched bot sends.
//
// Retrying with the same client_message_id replays the original message.
func (s *MessagingCommandService) SendProposal(ctx context.Context, tenant, actor, turnID, proposalID string, input types.ProposalSendInput) (*types.MessagingMessage, *apperrors.Error) {
	input.Text = strings.TrimSpace(input.Text)
	if input.Text != "" && !messagingText(input.Text, replyTextLimit) {
		return nil, apperrors.ErrBadRequest(ctx).SetDetail("text must be 1–2000 characters")
	}
	if _, err := uuid.Parse(input.ClientMessageID); err != nil {
		return nil, apperrors.ErrBadRequest(ctx).SetDetail("client_message_id must be a UUID")
	}
	var result *types.MessagingMessage
	err := s.authorized(ctx, tenant, actor, func(r repository.MessagingStore, _ bool) *apperrors.Error {
		turn, err := messagingLockedTurn(ctx, r, tenant, turnID)
		if err != nil {
			return err
		}
		p, err := lockedProposal(ctx, r, tenant, turnID, proposalID)
		if err != nil {
			return err
		}
		text := input.Text
		if text == "" {
			text = p.Text
		}
		hash := messagingHash(struct{ Actor, Turn, Proposal, Text string }{actor, turnID, proposalID, text})
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
		if p.Status != types.ProposalPending {
			return apperrors.ErrConflict(ctx, "proposal", "is no longer pending; refresh the conversation")
		}
		if p.Stale(*turn) {
			return apperrors.ErrConflict(ctx, "proposal", "is stale: conversation control changed after the bot proposed it, so it cannot be sent")
		}
		now := time.Now().UTC()
		claimed, cancelled := false, 0
		switch turn.HandlingState() {
		case "STAFF_ACTIVE":
			if turn.AssigneeID == nil || *turn.AssigneeID != actor {
				return apperrors.ErrForbidden(ctx).SetDetail("only the assigned sale can send this proposal")
			}
		case "BOT_ACTIVE":
			if cancelled, err = r.CancelPendingBotSends(ctx, tenant, turnID, now); err != nil {
				return err
			}
			turn.AssigneeID, turn.Control = &actor, "sales"
			turn.ControlGeneration++
			turn.ReviewPending = false
			claimed = true
		default:
			return apperrors.ErrForbidden(ctx).SetDetail("only the assigned sale can send this proposal")
		}
		thread, err := r.Thread(ctx, tenant, turn.ThreadID, false)
		if err != nil {
			return err
		}
		message, intentID, reason, err := recordProposalReply(ctx, r, *thread, turn, actor, text, input.ClientMessageID, hash, now)
		if err != nil {
			return err
		}
		if message == nil {
			return apperrors.ErrConflict(ctx, "conversation channel", "is not active; sending is unavailable ("+reason+")")
		}
		p.Status, p.DecidedBy, p.SentMessageID = types.ProposalSent, actor, message.ID
		p.Version++
		p.UpdatedAt = now
		if err := r.SaveProposal(ctx, *p); err != nil {
			return err
		}
		result = message
		return r.Audit(ctx, tenant, actor, "messaging.proposal_sent", "messaging_turn", turnID, map[string]any{
			"proposal_id": p.ID, "message_id": message.ID, "intent_id": intentID, "edited": text != p.Text,
			"route": p.Route, "source": p.Source, "script_version_id": p.ScriptVersionID, "ai_run_id": p.AIRunID,
			"claimed": claimed, "control_generation": turn.ControlGeneration, "cancelled_bot_sends": cancelled,
		})
	})
	return result, err
}

// DiscardProposal drops a pending proposal; nothing is sent. The assigned sale, the owner, or any member while the
// turn is unassigned may discard. Stale proposals can be discarded too.
func (s *MessagingCommandService) DiscardProposal(ctx context.Context, tenant, actor, turnID, proposalID string) (*types.ReplyProposal, *apperrors.Error) {
	var result *types.ReplyProposal
	err := s.authorized(ctx, tenant, actor, func(r repository.MessagingStore, owner bool) *apperrors.Error {
		turn, err := messagingLockedTurn(ctx, r, tenant, turnID)
		if err != nil {
			return err
		}
		p, err := lockedProposal(ctx, r, tenant, turnID, proposalID)
		if err != nil {
			return err
		}
		if !owner && turn.AssigneeID != nil && *turn.AssigneeID != actor {
			return apperrors.ErrForbidden(ctx).SetDetail("only the assigned sale or the tenant owner can discard this proposal")
		}
		if p.Status != types.ProposalPending {
			return apperrors.ErrConflict(ctx, "proposal", "is no longer pending; refresh the conversation")
		}
		now := time.Now().UTC()
		stale := p.Stale(*turn)
		p.Status, p.DecidedBy = types.ProposalDiscarded, actor
		p.Version++
		p.UpdatedAt = now
		if err := r.SaveProposal(ctx, *p); err != nil {
			return err
		}
		// The turn holds at most one live proposal, so a current one is the proposal waiting for review: the turn
		// leaves the waiting queue and normal expiry resumes. Control does not change.
		if !stale && turn.ReviewPending {
			turn.ReviewPending = false
			turn.Version++
			if err := r.SaveTurn(ctx, *turn, false); err != nil {
				return err
			}
		}
		result = p
		return r.Audit(ctx, tenant, actor, "messaging.proposal_discarded", "messaging_turn", turnID, map[string]any{
			"proposal_id": p.ID, "stale": stale, "source": p.Source, "script_version_id": p.ScriptVersionID, "ai_run_id": p.AIRunID,
		})
	})
	return result, err
}

// RegenerateProposal asks the router for a fresh proposal at the turn's current control generation, answering its
// newest customer message (the Desk "regenerate" action for a stale proposal). It supersedes the turn's live
// proposal. A person-requested proposal is never sent automatically and never moves control. Allowed for any
// member on a BOT_ACTIVE turn and for the assigned sale on a STAFF_ACTIVE turn, and refused while the bot is off.
func (s *MessagingCommandService) RegenerateProposal(ctx context.Context, tenant, actor, turnID string) (*types.ReplyProposal, *apperrors.Error) {
	var result *types.ReplyProposal
	err := s.authorized(ctx, tenant, actor, func(r repository.MessagingStore, _ bool) *apperrors.Error {
		turn, err := messagingLockedTurn(ctx, r, tenant, turnID)
		if err != nil {
			return err
		}
		if turn.Lifecycle != types.TurnOpen {
			return apperrors.ErrConflict(ctx, "turn", "is closed")
		}
		switch turn.HandlingState() {
		case "BOT_ACTIVE":
		case "STAFF_ACTIVE":
			if turn.AssigneeID == nil || *turn.AssigneeID != actor {
				return apperrors.ErrForbidden(ctx).SetDetail("only the assigned sale can request a new proposal")
			}
		default:
			return apperrors.ErrConflict(ctx, "turn", "is waiting for staff; claim it before requesting a proposal")
		}
		settings, err := r.ReplySettings(ctx, tenant, false)
		if err != nil {
			return err
		}
		if !settings.BotOn() {
			return errBotOff(ctx, "requesting a bot proposal")
		}
		message, err := r.LatestInbound(ctx, tenant, turnID)
		if err != nil {
			return err
		}
		if message == nil {
			return apperrors.ErrConflict(ctx, "turn", "has no customer message to answer")
		}
		now := time.Now().UTC()
		// The new job supersedes the live proposal, so the turn no longer waits for its review.
		if turn.ReviewPending {
			turn.ReviewPending = false
			turn.Version++
			if err := r.SaveTurn(ctx, *turn, false); err != nil {
				return err
			}
		}
		if result, err = enqueueBotRouting(ctx, r, *turn, *message, actor, now); err != nil {
			return err
		}
		return r.Audit(ctx, tenant, actor, "messaging.proposal_requested", "messaging_turn", turnID, map[string]any{
			"proposal_id": result.ID, "control_generation": turn.ControlGeneration, "trigger_message_id": message.ID,
		})
	})
	return result, err
}
