package service

import (
	"context"
	"time"

	apperrors "github.com/TruongHoang2004/Hexta/packages/shared/pkg/errors"
	"github.com/TruongHoang2004/Hexta/services/api/internal/core/types"
	"github.com/TruongHoang2004/Hexta/services/api/internal/repository"
	"github.com/google/uuid"
)

// ReplySettingsService owns the tenant's bot reply mode (MVP 5 FR-5.1). Every member may read it; only the tenant
// owner changes it, with an optimistic version. The mode is read again by the router and by proposal commands at
// the moment they act, so a change applies to the next routed message and never re-sends existing proposals.
type ReplySettingsService struct {
	store repository.ReplySettingsStore
	now   func() time.Time
}

func NewReplySettingsService(store repository.ReplySettingsStore) *ReplySettingsService {
	return &ReplySettingsService{store: store, now: func() time.Time { return time.Now().UTC() }}
}

func (s *ReplySettingsService) within(ctx context.Context, tenant, actor string, fn func(repository.ReplySettingsStore, types.MessagingScope) *apperrors.Error) *apperrors.Error {
	if actor == "" {
		return apperrors.ErrUnauthorized(ctx)
	}
	if _, err := uuid.Parse(tenant); err != nil {
		return apperrors.ErrBadRequest(ctx).SetDetail("valid tenant ID required")
	}
	return s.store.WithinReplySettings(ctx, func(r repository.ReplySettingsStore) *apperrors.Error {
		scope, err := r.Scope(ctx, tenant, actor)
		if err != nil {
			return err
		}
		if !scope.Active || !scope.Member {
			return apperrors.ErrForbidden(ctx).SetDetail("active tenant membership required")
		}
		return fn(r, scope)
	})
}

// Get returns the current reply mode; Configured=false means the owner never chose one.
func (s *ReplySettingsService) Get(ctx context.Context, tenant, actor string) (types.ReplySettings, *apperrors.Error) {
	var result types.ReplySettings
	err := s.within(ctx, tenant, actor, func(r repository.ReplySettingsStore, _ types.MessagingScope) *apperrors.Error {
		var err *apperrors.Error
		result, err = r.ReplySettings(ctx, tenant, false)
		return err
	})
	return result, err
}

// Update changes the reply mode as the tenant owner. ExpectedVersion is the version the owner saw (0 before the
// first save); a stale version is a conflict. Saving the current mode again is a no-op that keeps the version.
func (s *ReplySettingsService) Update(ctx context.Context, tenant, actor string, input types.ReplySettingsUpdate) (types.ReplySettings, *apperrors.Error) {
	if input.Mode != types.ReplyModeSuggest && input.Mode != types.ReplyModeAutoSend {
		return types.ReplySettings{}, apperrors.ErrBadRequest(ctx).SetDetail("reply_mode must be suggest or auto_send")
	}
	if input.ExpectedVersion < 0 {
		return types.ReplySettings{}, apperrors.ErrBadRequest(ctx).SetDetail("expected_version must not be negative")
	}
	var result types.ReplySettings
	err := s.within(ctx, tenant, actor, func(r repository.ReplySettingsStore, scope types.MessagingScope) *apperrors.Error {
		if !scope.Owner {
			return apperrors.ErrForbidden(ctx).SetDetail("only the tenant owner can change the reply mode")
		}
		current, err := r.ReplySettings(ctx, tenant, true)
		if err != nil {
			return err
		}
		if current.Version != input.ExpectedVersion {
			return apperrors.ErrConflict(ctx, "reply settings", "changed; refresh before retrying")
		}
		if current.Configured && current.Mode == input.Mode {
			result = current
			return nil
		}
		next := types.ReplySettings{TenantID: tenant, Mode: input.Mode, Configured: true, Version: current.Version + 1, UpdatedBy: actor, UpdatedAt: s.now()}
		if err := r.SaveReplySettings(ctx, next, !current.Configured); err != nil {
			return err
		}
		result = next
		return r.Audit(ctx, tenant, actor, "messaging.reply_mode_changed", "tenant_reply_settings", tenant, map[string]any{
			"before": current.Mode, "before_configured": current.Configured, "after": next.Mode, "version": next.Version,
		})
	})
	return result, err
}
