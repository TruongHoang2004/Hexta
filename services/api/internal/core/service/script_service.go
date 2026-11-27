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

const (
	scriptHistoryLimit      = 100
	scriptActiveLimit       = 500
	scriptPermissionMembers = 500
)

// ScriptMemberDirectory resolves member display emails for the permission screen; it grants no authority.
type ScriptMemberDirectory interface {
	TenantMemberEmails(ctx context.Context, tenantID string) (map[string]string, *apperrors.Error)
}

// ScriptService owns the script lifecycle (inactive → active vN → archived) and the four delegated script
// permissions. Every command rechecks current membership and the one permission it needs inside its transaction.
type ScriptService struct {
	store   repository.ScriptStore
	members ScriptMemberDirectory
	now     func() time.Time
}

func NewScriptService(store repository.ScriptStore, members ScriptMemberDirectory) *ScriptService {
	return &ScriptService{store: store, members: members, now: func() time.Time { return time.Now().UTC() }}
}

// authorized opens the command transaction and checks membership, plus permission when one is named.
func (s *ScriptService) authorized(ctx context.Context, tenant, actor, permission string, fn func(repository.ScriptStore, types.ScriptAccess) *apperrors.Error) *apperrors.Error {
	if actor == "" {
		return apperrors.ErrUnauthorized(ctx)
	}
	if _, err := uuid.Parse(tenant); err != nil {
		return apperrors.ErrBadRequest(ctx).SetDetail("valid tenant ID required")
	}
	return s.store.Within(ctx, func(r repository.ScriptStore) *apperrors.Error {
		access, err := r.Access(ctx, tenant, actor)
		if err != nil {
			return err
		}
		if !access.ActiveTenant || !access.Member {
			return apperrors.ErrForbidden(ctx).SetDetail("active tenant membership required")
		}
		if permission != "" && !access.Allows(permission) {
			return apperrors.ErrForbidden(ctx).SetDetail(permission + " permission required")
		}
		return fn(r, access)
	})
}

func scriptID(ctx context.Context, id string) *apperrors.Error {
	if _, err := uuid.Parse(id); err != nil {
		return apperrors.ErrBadRequest(ctx).SetDetail("invalid script ID")
	}
	return nil
}

// normalizeScript trims and validates a name and content; trigger phrases are de-duplicated case-insensitively.
func normalizeScript(ctx context.Context, name *string, content *types.ScriptContent) *apperrors.Error {
	*name = strings.TrimSpace(*name)
	if *name == "" || utf8.RuneCountInString(*name) > types.ScriptMaxNameLength {
		return apperrors.ErrBadRequest(ctx).SetDetail("script name must be 1–120 characters")
	}
	content.Answer = strings.TrimSpace(content.Answer)
	if content.Answer == "" || utf8.RuneCountInString(content.Answer) > types.ScriptMaxAnswerLength {
		return apperrors.ErrBadRequest(ctx).SetDetail("answer must be 1–2000 characters")
	}
	seen := map[string]bool{}
	phrases := []string{}
	for _, raw := range content.Triggers {
		phrase := strings.Join(strings.Fields(raw), " ")
		if phrase == "" {
			continue
		}
		if utf8.RuneCountInString(phrase) > types.ScriptMaxTriggerLength {
			return apperrors.ErrBadRequest(ctx).SetDetail("each trigger phrase must be at most 120 characters")
		}
		key := strings.ToLower(phrase)
		if !seen[key] {
			seen[key] = true
			phrases = append(phrases, phrase)
		}
	}
	if len(phrases) == 0 || len(phrases) > types.ScriptMaxTriggers {
		return apperrors.ErrBadRequest(ctx).SetDetail("1–20 trigger phrases required")
	}
	content.Triggers = phrases
	return nil
}

func scriptHash(input types.ScriptCreateInput) string {
	data, _ := json.Marshal(input)
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

func scriptWritable(ctx context.Context, script *types.Script, expected int64) *apperrors.Error {
	if script.Version != expected {
		return apperrors.ErrConflict(ctx, "script", "changed; refresh before retrying")
	}
	if script.Status == types.ScriptStatusArchived {
		return apperrors.ErrConflict(ctx, "script", "is archived")
	}
	return nil
}

func (s *ScriptService) detail(ctx context.Context, r repository.ScriptStore, script *types.Script) (*types.ScriptDetail, *apperrors.Error) {
	versions, err := r.Versions(ctx, script.TenantID, script.ID, scriptHistoryLimit)
	if err != nil {
		return nil, err
	}
	return &types.ScriptDetail{Script: *script, Versions: versions}, nil
}

// List returns a page of scripts; any active member may read them.
func (s *ScriptService) List(ctx context.Context, tenant, actor string, filter types.ScriptListFilter, page, size int) ([]types.ScriptSummary, int64, *apperrors.Error) {
	switch filter.Status {
	case "", types.ScriptStatusInactive, types.ScriptStatusActive, types.ScriptStatusArchived:
	default:
		return nil, 0, apperrors.ErrBadRequest(ctx).SetDetail("status must be inactive, active or archived")
	}
	var rows []types.ScriptSummary
	var total int64
	err := s.authorized(ctx, tenant, actor, "", func(r repository.ScriptStore, _ types.ScriptAccess) *apperrors.Error {
		limit, offset := pageBounds(page, size)
		var err *apperrors.Error
		rows, total, err = r.Scripts(ctx, tenant, filter, limit, offset)
		return err
	})
	return rows, total, err
}

// Get returns a script with its version history; any active member may read it.
func (s *ScriptService) Get(ctx context.Context, tenant, actor, id string) (*types.ScriptDetail, *apperrors.Error) {
	if err := scriptID(ctx, id); err != nil {
		return nil, err
	}
	var result *types.ScriptDetail
	err := s.authorized(ctx, tenant, actor, "", func(r repository.ScriptStore, _ types.ScriptAccess) *apperrors.Error {
		script, err := r.Script(ctx, tenant, id, false)
		if err != nil {
			return err
		}
		result, err = s.detail(ctx, r, script)
		return err
	})
	return result, err
}

// Create needs script.create. The script starts inactive with draft v1. The UUID idempotency key becomes the
// script ID: replaying the same payload returns the original script; a different payload conflicts.
func (s *ScriptService) Create(ctx context.Context, tenant, actor, key string, input types.ScriptCreateInput) (*types.ScriptDetail, *apperrors.Error) {
	parsed, parseErr := uuid.Parse(key)
	if parseErr != nil {
		return nil, apperrors.ErrBadRequest(ctx).SetDetail("Idempotency-Key must be a UUID")
	}
	if err := normalizeScript(ctx, &input.Name, &input.Content); err != nil {
		return nil, err
	}
	id, hash, now := parsed.String(), scriptHash(input), s.now()
	var result *types.ScriptDetail
	err := s.authorized(ctx, tenant, actor, types.ScriptPermissionCreate, func(r repository.ScriptStore, _ types.ScriptAccess) *apperrors.Error {
		script := types.Script{ID: id, TenantID: tenant, Name: input.Name, Status: types.ScriptStatusInactive, Version: 1, CreateHash: hash, CreatedBy: actor, CreatedAt: now, UpdatedAt: now}
		inserted, err := r.InsertScript(ctx, script)
		if err != nil {
			return err
		}
		if !inserted {
			existing, err := r.Script(ctx, tenant, id, false)
			if err != nil {
				// The key exists in another tenant: never reveal or reuse it.
				return apperrors.ErrConflict(ctx, "idempotency key", "already used")
			}
			if existing.CreateHash != hash || existing.CreatedBy != actor {
				return apperrors.ErrConflict(ctx, "idempotency key", "already used for different script data")
			}
			result, err = s.detail(ctx, r, existing)
			return err
		}
		draft := types.ScriptVersion{ID: uuid.NewString(), TenantID: tenant, ScriptID: id, Number: 1, Content: input.Content, CreatedBy: actor, UpdatedBy: actor, CreatedAt: now, UpdatedAt: now}
		if err := r.InsertVersion(ctx, draft); err != nil {
			return err
		}
		if err := r.Audit(ctx, tenant, actor, "script.create", "script", id, map[string]any{"version": 1, "draft": 1}); err != nil {
			return err
		}
		result = &types.ScriptDetail{Script: script, Versions: []types.ScriptVersion{draft}}
		return nil
	})
	return result, err
}

// SaveDraft needs script.edit. It rewrites the open draft, or opens draft vN+1 when every version is published
// (for example after editing an active script). Published content is never modified.
func (s *ScriptService) SaveDraft(ctx context.Context, tenant, actor, id string, input types.ScriptDraftInput) (*types.ScriptDetail, *apperrors.Error) {
	if err := scriptID(ctx, id); err != nil {
		return nil, err
	}
	if err := normalizeScript(ctx, &input.Name, &input.Content); err != nil {
		return nil, err
	}
	now := s.now()
	var result *types.ScriptDetail
	err := s.authorized(ctx, tenant, actor, types.ScriptPermissionEdit, func(r repository.ScriptStore, _ types.ScriptAccess) *apperrors.Error {
		script, err := r.Script(ctx, tenant, id, true)
		if err != nil {
			return err
		}
		if err := scriptWritable(ctx, script, input.ExpectedVersion); err != nil {
			return err
		}
		draft, err := r.Draft(ctx, tenant, id)
		if err != nil {
			return err
		}
		if draft != nil {
			draft.Content, draft.UpdatedBy, draft.UpdatedAt = input.Content, actor, now
			if err := r.UpdateDraft(ctx, *draft); err != nil {
				return err
			}
		} else {
			number, err := r.NextNumber(ctx, tenant, id)
			if err != nil {
				return err
			}
			draft = &types.ScriptVersion{ID: uuid.NewString(), TenantID: tenant, ScriptID: id, Number: number, Content: input.Content, CreatedBy: actor, UpdatedBy: actor, CreatedAt: now, UpdatedAt: now}
			if err := r.InsertVersion(ctx, *draft); err != nil {
				return err
			}
		}
		updated := *script
		updated.Name, updated.Version, updated.UpdatedAt = input.Name, script.Version+1, now
		if err := r.SaveScript(ctx, updated, script.Version); err != nil {
			return err
		}
		// An active script's name is its knowledge title; its active content is unchanged by a draft.
		if updated.Status == types.ScriptStatusActive {
			if err := r.EnqueueKnowledgeSync(ctx, tenant, types.KnowledgeSourceScript, id, now); err != nil {
				return err
			}
		}
		if err := r.Audit(ctx, tenant, actor, "script.edit", "script", id, map[string]any{"version": updated.Version, "draft": draft.Number}); err != nil {
			return err
		}
		result, err = s.detail(ctx, r, &updated)
		return err
	})
	return result, err
}

// Activate needs script.activate. It points the script at the selected version (the open draft or any earlier
// published version) and freezes that version's content on its first activation.
func (s *ScriptService) Activate(ctx context.Context, tenant, actor, id, versionID string, expected int64) (*types.ScriptDetail, *apperrors.Error) {
	if err := scriptID(ctx, id); err != nil {
		return nil, err
	}
	if _, err := uuid.Parse(versionID); err != nil {
		return nil, apperrors.ErrBadRequest(ctx).SetDetail("invalid script version ID")
	}
	now := s.now()
	var result *types.ScriptDetail
	err := s.authorized(ctx, tenant, actor, types.ScriptPermissionActivate, func(r repository.ScriptStore, _ types.ScriptAccess) *apperrors.Error {
		script, err := r.Script(ctx, tenant, id, true)
		if err != nil {
			return err
		}
		if err := scriptWritable(ctx, script, expected); err != nil {
			return err
		}
		version, err := r.Version(ctx, tenant, id, versionID)
		if err != nil {
			return err
		}
		if script.Status == types.ScriptStatusActive && script.ActiveVersionID == version.ID {
			return apperrors.ErrConflict(ctx, "script version", "is already active")
		}
		if version.Draft() {
			if err := r.PublishVersion(ctx, tenant, id, version.ID, actor, now); err != nil {
				return err
			}
		}
		updated := *script
		updated.Status, updated.ActiveVersionID, updated.Version, updated.UpdatedAt = types.ScriptStatusActive, version.ID, script.Version+1, now
		if err := r.SaveScript(ctx, updated, script.Version); err != nil {
			return err
		}
		// The active version becomes bot knowledge (with its version_id) through the durable knowledge sync job.
		if err := r.EnqueueKnowledgeSync(ctx, tenant, types.KnowledgeSourceScript, id, now); err != nil {
			return err
		}
		if err := r.Audit(ctx, tenant, actor, "script.activate", "script", id, map[string]any{"version": updated.Version, "active_number": version.Number}); err != nil {
			return err
		}
		result, err = s.detail(ctx, r, &updated)
		return err
	})
	return result, err
}

// Deactivate needs script.activate. The script returns to inactive; every version stays in history.
func (s *ScriptService) Deactivate(ctx context.Context, tenant, actor, id string, expected int64) (*types.ScriptDetail, *apperrors.Error) {
	return s.transition(ctx, tenant, actor, id, expected, types.ScriptPermissionActivate, "script.deactivate", func(ctx context.Context, script *types.Script, now time.Time) *apperrors.Error {
		if script.Status != types.ScriptStatusActive {
			return apperrors.ErrConflict(ctx, "script", "is not active")
		}
		script.Status, script.ActiveVersionID = types.ScriptStatusInactive, ""
		return nil
	})
}

// Archive needs script.archive. Archived is terminal; the script and all versions are retained, never deleted.
func (s *ScriptService) Archive(ctx context.Context, tenant, actor, id string, expected int64) (*types.ScriptDetail, *apperrors.Error) {
	return s.transition(ctx, tenant, actor, id, expected, types.ScriptPermissionArchive, "script.archive", func(_ context.Context, script *types.Script, now time.Time) *apperrors.Error {
		script.Status, script.ActiveVersionID, script.ArchivedAt = types.ScriptStatusArchived, "", &now
		return nil
	})
}

func (s *ScriptService) transition(ctx context.Context, tenant, actor, id string, expected int64, permission, action string, apply func(context.Context, *types.Script, time.Time) *apperrors.Error) (*types.ScriptDetail, *apperrors.Error) {
	if err := scriptID(ctx, id); err != nil {
		return nil, err
	}
	now := s.now()
	var result *types.ScriptDetail
	err := s.authorized(ctx, tenant, actor, permission, func(r repository.ScriptStore, _ types.ScriptAccess) *apperrors.Error {
		script, err := r.Script(ctx, tenant, id, true)
		if err != nil {
			return err
		}
		if err := scriptWritable(ctx, script, expected); err != nil {
			return err
		}
		updated := *script
		if err := apply(ctx, &updated, now); err != nil {
			return err
		}
		updated.Version, updated.UpdatedAt = script.Version+1, now
		if err := r.SaveScript(ctx, updated, script.Version); err != nil {
			return err
		}
		// A deactivated or archived script leaves the bot's knowledge through the knowledge sync job.
		if err := r.EnqueueKnowledgeSync(ctx, tenant, types.KnowledgeSourceScript, id, now); err != nil {
			return err
		}
		if err := r.Audit(ctx, tenant, actor, action, "script", id, map[string]any{"version": updated.Version, "status": updated.Status}); err != nil {
			return err
		}
		result, err = s.detail(ctx, r, &updated)
		return err
	})
	return result, err
}

// MyPermissions returns the caller's effective script permissions, for showing or hiding actions. The API still
// authorizes every command.
func (s *ScriptService) MyPermissions(ctx context.Context, tenant, actor string) (types.ScriptAccess, *apperrors.Error) {
	var result types.ScriptAccess
	err := s.authorized(ctx, tenant, actor, "", func(_ repository.ScriptStore, access types.ScriptAccess) *apperrors.Error {
		result = access
		return nil
	})
	return result, err
}

// MemberPermissions lists current members with their grants. Owner only.
func (s *ScriptService) MemberPermissions(ctx context.Context, tenant, actor string) ([]types.ScriptMemberPermissions, *apperrors.Error) {
	var result []types.ScriptMemberPermissions
	err := s.authorized(ctx, tenant, actor, "", func(r repository.ScriptStore, access types.ScriptAccess) *apperrors.Error {
		if !access.Owner {
			return apperrors.ErrForbidden(ctx).SetDetail("tenant owner required")
		}
		var err *apperrors.Error
		result, err = r.MemberPermissions(ctx, tenant, actor, scriptPermissionMembers)
		return err
	})
	if err != nil {
		return nil, err
	}
	if s.members != nil {
		emails, err := s.members.TenantMemberEmails(ctx, tenant)
		if err != nil {
			return nil, err
		}
		for i := range result {
			result[i].Email = emails[result[i].UserID]
		}
	}
	return result, nil
}

// Grant gives one permission to a current, non-owner member. Owner only; expected is the member's grant version.
func (s *ScriptService) Grant(ctx context.Context, tenant, actor, user, permission string, expected int64) (*types.ScriptMemberPermissions, *apperrors.Error) {
	return s.setPermission(ctx, tenant, actor, user, permission, expected, true)
}

// Revoke removes one permission from a member. Owner only; it takes effect on the member's next action.
func (s *ScriptService) Revoke(ctx context.Context, tenant, actor, user, permission string, expected int64) (*types.ScriptMemberPermissions, *apperrors.Error) {
	return s.setPermission(ctx, tenant, actor, user, permission, expected, false)
}

func (s *ScriptService) setPermission(ctx context.Context, tenant, actor, user, permission string, expected int64, granted bool) (*types.ScriptMemberPermissions, *apperrors.Error) {
	valid := false
	for _, p := range types.ScriptPermissions {
		valid = valid || p == permission
	}
	if !valid {
		return nil, apperrors.ErrBadRequest(ctx).SetDetail("permission must be script.create, script.edit, script.activate or script.archive")
	}
	if user == "" || len(user) > 255 || expected < 0 {
		return nil, apperrors.ErrBadRequest(ctx).SetDetail("member and expected_version required")
	}
	now := s.now()
	var result *types.ScriptMemberPermissions
	err := s.authorized(ctx, tenant, actor, "", func(r repository.ScriptStore, access types.ScriptAccess) *apperrors.Error {
		if !access.Owner {
			return apperrors.ErrForbidden(ctx).SetDetail("tenant owner required")
		}
		if user == actor {
			return apperrors.ErrBadRequest(ctx).SetDetail("the owner always holds every script permission")
		}
		member, err := r.IsMember(ctx, tenant, user)
		if err != nil {
			return err
		}
		if !member {
			return apperrors.ErrBadRequest(ctx).SetDetail("permissions can only be granted to current members")
		}
		current, err := r.PermissionSet(ctx, tenant, user)
		if err != nil {
			return err
		}
		if current.Version != expected {
			return apperrors.ErrConflict(ctx, "script permissions", "were updated elsewhere; reload to see them")
		}
		next := current
		if current.Has(permission) != granted {
			next = current.With(permission, granted)
			next.Version, next.UpdatedBy, next.UpdatedAt = current.Version+1, actor, now
			if err := r.SavePermissionSet(ctx, next, current.Version); err != nil {
				return err
			}
			action := "script.permission.revoke"
			if granted {
				action = "script.permission.grant"
			}
			if err := r.Audit(ctx, tenant, actor, action, "script_member_permissions", user, map[string]any{"permission": permission, "version": next.Version, "permissions": next.Granted()}); err != nil {
				return err
			}
		}
		result = &types.ScriptMemberPermissions{UserID: user, Grants: next}
		return nil
	})
	return result, err
}

// ActiveScriptVersions is an internal read for the reply router (MVP 5 FR-5.2). It returns the published content
// of every active script in an active tenant (none for a suspended tenant), bounded at 500, ordered by script
// name. It takes no actor and grants no
// business authority: callers must not expose it over HTTP, must apply their own send/handoff gates, and should
// record VersionID as the proposal source because that content never changes.
func (s *ScriptService) ActiveScriptVersions(ctx context.Context, tenant string) ([]types.ActiveScriptVersion, *apperrors.Error) {
	if _, err := uuid.Parse(tenant); err != nil {
		return nil, apperrors.ErrBadRequest(ctx).SetDetail("valid tenant ID required")
	}
	return s.store.ActiveVersions(ctx, tenant, scriptActiveLimit)
}
