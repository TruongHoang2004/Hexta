package service

import (
	"context"
	"time"

	apperrors "github.com/TruongHoang2004/Hexta/packages/shared/pkg/errors"
	"github.com/TruongHoang2004/Hexta/services/api/internal/core/types"
	"github.com/TruongHoang2004/Hexta/services/api/internal/repository"
)

const (
	// TurnExpiryBatch bounds one scan; the job repeats while a full batch was due.
	TurnExpiryBatch = 100
	// TurnExpiryActor is the audit actor for system-closed turns.
	TurnExpiryActor = "system:turn-expiry"
	// outboundSettleWindow keeps a turn open while a recent channel send may still be dispatched or confirmed. The
	// connector refuses an intent after outboundIntentTTL, so an older pending intent can no longer start.
	outboundSettleWindow = outboundIntentTTL + time.Minute
)

// TurnExpiryService closes open turns after the single 24-hour inactivity window. It is safe to run on several
// instances at once: each candidate is processed under its thread lock (SKIP LOCKED) and closed with a conditional
// update of the exact observed state, so a racing customer message, confirmed send, resolve or second instance
// wins and the turn stays open. Re-running is idempotent.
type TurnExpiryService struct {
	store repository.TurnExpiryStore
	now   func() time.Time
}

func NewTurnExpiryService(store repository.TurnExpiryStore) *TurnExpiryService {
	return &TurnExpiryService{store: store, now: time.Now}
}

// ExpireDue closes one batch of due turns and reports how many candidates were seen and how many closed.
func (s *TurnExpiryService) ExpireDue(ctx context.Context) (seen, closed int, failure *apperrors.Error) {
	now := s.now().UTC()
	refs, err := s.store.ExpiryCandidates(ctx, now.Add(-types.TurnInactivityWindow), TurnExpiryBatch)
	if err != nil {
		return 0, 0, err
	}
	for _, ref := range refs {
		if ctx.Err() != nil {
			break
		}
		ok, err := s.expire(ctx, ref, now)
		if err != nil {
			// One failed turn must not block the rest; it is retried on the next run.
			if failure == nil {
				failure = err
			}
			continue
		}
		if ok {
			closed++
		}
	}
	return len(refs), closed, failure
}

func (s *TurnExpiryService) expire(ctx context.Context, ref types.TurnRef, now time.Time) (bool, *apperrors.Error) {
	closed := false
	err := s.store.WithinExpiry(ctx, func(r repository.TurnExpiryStore) *apperrors.Error {
		locked, err := r.LockThreadSkipLocked(ctx, ref.TenantID, ref.ThreadID)
		if err != nil || !locked {
			return err
		}
		intents, err := r.LockPendingIntents(ctx, ref.TenantID, ref.TurnID)
		if err != nil {
			return err
		}
		ids := make([]string, 0, len(intents))
		for _, intent := range intents {
			if intent.CreatedAt.After(now.Add(-outboundSettleWindow)) {
				return nil
			}
			ids = append(ids, intent.ID)
		}
		// Reread after the locks: activity, flags and lifecycle may have changed since the candidate scan.
		turn, err := r.Turn(ctx, ref.TenantID, ref.TurnID)
		if err != nil {
			return err
		}
		if turn.ThreadID != ref.ThreadID || !turn.ExpiryDue(now) {
			return nil
		}
		ok, err := r.CloseTurn(ctx, *turn, now)
		if err != nil || !ok {
			return err
		}
		if err := r.CancelIntents(ctx, ref.TenantID, ids, now); err != nil {
			return err
		}
		closed = true
		return r.Audit(ctx, ref.TenantID, TurnExpiryActor, "messaging.turn_expired", "messaging_turn", turn.ID, map[string]any{
			"version": turn.Version + 1, "last_activity_at": turn.LastActivityAt, "closed_at": now,
			"handling_state": turn.HandlingState(), "assignee_id": turn.AssigneeID, "cancelled_intents": len(ids),
		})
	})
	if err != nil {
		return false, err
	}
	return closed, nil
}
