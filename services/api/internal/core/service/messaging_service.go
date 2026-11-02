package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"strings"
	"time"
	"unicode/utf8"

	apperrors "github.com/TruongHoang2004/Hexta/packages/shared/pkg/errors"
	"github.com/TruongHoang2004/Hexta/services/api/internal/core/types"
	"github.com/TruongHoang2004/Hexta/services/api/internal/repository"
	"github.com/google/uuid"
)

type MessagingService struct{ repo repository.MessagingStore }

func NewMessagingService(repo repository.MessagingStore) *MessagingService {
	return &MessagingService{repo: repo}
}
func (s *MessagingService) authorized(ctx context.Context, tenant, actor string, fn func(repository.MessagingStore, bool) *apperrors.Error) *apperrors.Error {
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
func messagingQuery(ctx context.Context, q types.MessagingQuery) *apperrors.Error {
	if q.Limit < 1 || q.Limit > 100 {
		return apperrors.ErrBadRequest(ctx).SetDetail("limit must be between 1 and 100")
	}
	if q.Cursor != "" {
		if _, err := uuid.Parse(q.Cursor); err != nil {
			return apperrors.ErrBadRequest(ctx).SetDetail("invalid cursor")
		}
	}
	switch q.Filter {
	case "", "all", "mine", "unassigned", "needs_human":
		return nil
	}
	return apperrors.ErrBadRequest(ctx).SetDetail("invalid conversation filter")
}
func messagingText(value string, max int) bool {
	return strings.TrimSpace(value) != "" && utf8.ValidString(value) && utf8.RuneCountInString(value) <= max
}
func messagingKey(ctx context.Context, key string) *apperrors.Error {
	if _, err := uuid.Parse(key); err != nil {
		return apperrors.ErrBadRequest(ctx).SetDetail("request_key must be a UUID")
	}
	return nil
}
func messagingID() string { return uuid.Must(uuid.NewV7()).String() }
func messagingHash(value any) string {
	data, _ := json.Marshal(value)
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}
func (s *MessagingService) Conversations(ctx context.Context, tenant, actor string, input types.MessagingQuery) (types.MessagingPage, *apperrors.Error) {
	var result types.MessagingPage
	if err := messagingQuery(ctx, input); err != nil {
		return result, err
	}
	err := s.authorized(ctx, tenant, actor, func(r repository.MessagingStore, _ bool) *apperrors.Error {
		var err *apperrors.Error
		result, err = r.Conversations(ctx, tenant, actor, input)
		return err
	})
	return result, err
}
func (s *MessagingService) Detail(ctx context.Context, tenant, actor, id string, input types.MessagingQuery) (types.MessagingDetail, *apperrors.Error) {
	var result types.MessagingDetail
	if err := messagingQuery(ctx, input); err != nil {
		return result, err
	}
	err := s.authorized(ctx, tenant, actor, func(r repository.MessagingStore, _ bool) *apperrors.Error {
		thread, err := r.Thread(ctx, tenant, id, false)
		if err != nil {
			return err
		}
		turn, err := r.CurrentTurn(ctx, tenant, id)
		if err != nil {
			return err
		}
		result.Conversation = types.MessagingConversation{Thread: *thread, Turn: turn}
		result.Messages, result.NextCursor, err = r.Messages(ctx, tenant, id, input)
		return err
	})
	return result, err
}
func (s *MessagingService) Inbound(ctx context.Context, tenant, actor string, input types.MessagingInbound) (*types.MessagingMessage, *apperrors.Error) {
	input.ExternalKey = strings.TrimSpace(input.ExternalKey)
	input.DisplayName = strings.TrimSpace(input.DisplayName)
	input.Body = strings.TrimSpace(input.Body)
	if !messagingText(input.ExternalKey, 120) || !messagingText(input.DisplayName, 120) || !messagingText(input.Body, 4000) {
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
		turn, err := r.CurrentTurn(ctx, tenant, thread.ID)
		if err != nil {
			return err
		}
		create := turn == nil
		if create {
			turn = &types.MessagingTurn{ID: messagingID(), TenantID: tenant, ThreadID: thread.ID, Lifecycle: "open", Control: "system", ControlGeneration: 1}
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
func (s *MessagingService) Command(ctx context.Context, tenant, actor, id, action string, input types.MessagingCommand) (*types.MessagingTurn, *apperrors.Error) {
	if input.ExpectedVersion < 1 {
		return nil, apperrors.ErrBadRequest(ctx).SetDetail("positive expected_version required")
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
		switch action {
		case "claim":
			if turn.AssigneeID != nil && *turn.AssigneeID != actor {
				return apperrors.ErrConflict(ctx, "turn", "already assigned")
			}
			turn.AssigneeID = &actor
		case "assign":
			turn.AssigneeID = input.AssigneeID
		default:
			return apperrors.ErrBadRequest(ctx).SetDetail("unsupported conversation command")
		}
		turn.Control = "sales"
		turn.ControlGeneration++
		turn.Version++
		if err := r.SaveTurn(ctx, *turn, false); err != nil {
			return err
		}
		result = turn
		return r.Audit(ctx, tenant, actor, "messaging."+action, "messaging_turn", id, map[string]any{"version": turn.Version, "assignee_id": turn.AssigneeID})
	})
	return result, err
}
func (s *MessagingService) SendSimulation(ctx context.Context, tenant, actor, id string, input types.MessagingCommand) (*types.MessagingMessage, *apperrors.Error) {
	input.Body = strings.TrimSpace(input.Body)
	if !messagingText(input.Body, 4000) || input.ExpectedVersion < 1 || input.ContextRevision < 1 {
		return nil, apperrors.ErrBadRequest(ctx).SetDetail("body (1–4000) and positive version/context required")
	}
	if err := messagingKey(ctx, input.RequestKey); err != nil {
		return nil, err
	}
	hash := messagingHash(struct {
		Actor, Turn string
		Input       types.MessagingCommand
	}{actor, id, input})
	var result *types.MessagingMessage
	err := s.authorized(ctx, tenant, actor, func(r repository.MessagingStore, _ bool) *apperrors.Error {
		turn, err := messagingLockedTurn(ctx, r, tenant, id)
		if err != nil {
			return err
		}
		if turn.AssigneeID == nil || *turn.AssigneeID != actor || turn.Control != "sales" {
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
