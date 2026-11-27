package service

import (
	"context"
	"strings"
	"time"
	"unicode/utf8"

	apperrors "github.com/TruongHoang2004/Hexta/packages/shared/pkg/errors"
	"github.com/TruongHoang2004/Hexta/services/api/internal/core/types"
	"github.com/TruongHoang2004/Hexta/services/api/internal/repository"
)

// sessionInput validates a session outcome and bounded note; the outcome is required.
func sessionInput(ctx context.Context, input *types.SessionInput) *apperrors.Error {
	input.Outcome = strings.TrimSpace(input.Outcome)
	input.Note = strings.TrimSpace(input.Note)
	if input.ExpectedVersion < 1 {
		return apperrors.ErrBadRequest(ctx).SetDetail("positive expected_version required")
	}
	if !types.SessionOutcomes[input.Outcome] {
		return apperrors.ErrBadRequest(ctx).SetDetail("valid session outcome required")
	}
	if !utf8.ValidString(input.Note) || utf8.RuneCountInString(input.Note) > types.MaxSessionNoteRunes {
		return apperrors.ErrBadRequest(ctx).SetDetail("note must be at most 1000 characters")
	}
	return nil
}

// CloseSession closes the current open session (turn) before the 24-hour expiry. The assigned sale under staff
// control or the tenant owner may close it; nobody else, and a non-owner never closes an unassigned (bot or
// waiting) turn. Closing uses the expiry's closing write: version and control generation are bumped (stale
// proposals cannot be sent and router jobs are dropped), unstarted sends are cancelled, needs_human is cleared and
// history is kept. The next genuinely new customer message opens a new unassigned turn under the current reply mode.
func (s *MessagingCommandService) CloseSession(ctx context.Context, tenant, actor, id string, input types.SessionInput) (*types.MessagingTurn, *apperrors.Error) {
	if err := sessionInput(ctx, &input); err != nil {
		return nil, err
	}
	var result *types.MessagingTurn
	err := s.authorized(ctx, tenant, actor, func(r repository.MessagingStore, owner bool) *apperrors.Error {
		turn, err := messagingLockedTurn(ctx, r, tenant, id)
		if err != nil {
			return err
		}
		if turn.Lifecycle != types.TurnOpen || turn.Version != input.ExpectedVersion {
			return apperrors.ErrConflict(ctx, "session", "changed or already closed; refresh before retrying")
		}
		assigned := turn.AssigneeID != nil && *turn.AssigneeID == actor && turn.HandlingState() == "STAFF_ACTIVE"
		if !owner && !assigned {
			return apperrors.ErrForbidden(ctx).SetDetail("only the assigned sale under staff control or the tenant owner can close the session")
		}
		now := time.Now().UTC()
		// Lock order matches every messaging writer: thread (held), outbound intents, then the turn.
		cancelled, err := r.CancelUndispatchedSends(ctx, tenant, id, now)
		if err != nil {
			return err
		}
		ok, err := r.CloseTurn(ctx, *turn, types.TurnClosure{Cause: types.TurnCloseManual, ClosedBy: actor, Outcome: input.Outcome, Note: input.Note}, now)
		if err != nil {
			return err
		}
		if !ok {
			return apperrors.ErrConflict(ctx, "session", "changed; refresh before retrying")
		}
		if result, err = r.Turn(ctx, tenant, id); err != nil {
			return err
		}
		return r.Audit(ctx, tenant, actor, "messaging.turn_closed_manually", "messaging_turn", id, map[string]any{
			"version": result.Version, "control_generation": result.ControlGeneration, "closed_at": now,
			"handling_state": turn.HandlingState(), "assignee_id": turn.AssigneeID, "outcome": input.Outcome, "has_note": input.Note != "",
			"closed_while_needs_human": turn.NeedsHuman, "needs_human_reason": turn.NeedsHumanReason, "cancelled_sends": cancelled,
		})
	})
	return result, err
}

// UpdateSessionOutcome records or edits a closed session's outcome and note. Expired sessions start without an
// outcome. Only the tenant owner or the session's last handler may edit; the session stays closed.
func (s *MessagingCommandService) UpdateSessionOutcome(ctx context.Context, tenant, actor, id string, input types.SessionInput) (*types.MessagingTurn, *apperrors.Error) {
	if err := sessionInput(ctx, &input); err != nil {
		return nil, err
	}
	var result *types.MessagingTurn
	err := s.authorized(ctx, tenant, actor, func(r repository.MessagingStore, owner bool) *apperrors.Error {
		turn, err := messagingLockedTurn(ctx, r, tenant, id)
		if err != nil {
			return err
		}
		if turn.Lifecycle != types.TurnClosed {
			return apperrors.ErrConflict(ctx, "session", "is still open; close it to record its outcome")
		}
		handler := turn.LastHandlerID != nil && *turn.LastHandlerID == actor
		if !owner && !handler {
			return apperrors.ErrForbidden(ctx).SetDetail("only the session's last handler or the tenant owner can edit its outcome")
		}
		if turn.Version != input.ExpectedVersion {
			return apperrors.ErrConflict(ctx, "session", "changed; refresh before retrying")
		}
		previous := turn.Outcome
		turn.Outcome, turn.Note = input.Outcome, input.Note
		turn.Version++
		if err := r.SaveSessionOutcome(ctx, *turn); err != nil {
			return err
		}
		result = turn
		return r.Audit(ctx, tenant, actor, "messaging.turn_outcome_updated", "messaging_turn", id, map[string]any{
			"version": turn.Version, "previous_outcome": previous, "outcome": turn.Outcome, "has_note": turn.Note != "", "close_cause": turn.CloseCause,
		})
	})
	return result, err
}
