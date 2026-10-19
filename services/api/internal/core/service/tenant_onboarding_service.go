package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"net/mail"
	"regexp"
	"strings"
	"time"

	apperrors "github.com/TruongHoang2004/Hexta/packages/shared/pkg/errors"
	"github.com/TruongHoang2004/Hexta/services/api/internal/core/model"
	"github.com/TruongHoang2004/Hexta/services/api/internal/repository"
	"github.com/google/uuid"
)

type TenantOnboardingService struct {
	repo *repository.TenantOnboardingRepository
	now  func() time.Time
}

func NewTenantOnboardingService(repo *repository.TenantOnboardingRepository) *TenantOnboardingService {
	return &TenantOnboardingService{repo, time.Now}
}

var tenantSlugPattern = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)

func (s *TenantOnboardingService) Operator(ctx context.Context, actor string) (bool, *apperrors.Error) {
	if actor == "" {
		return false, apperrors.ErrUnauthorized(ctx)
	}
	return s.repo.Operator(ctx, actor)
}
func (s *TenantOnboardingService) Requests(ctx context.Context, actor string, platform bool) ([]model.TenantRegistration, *apperrors.Error) {
	if actor == "" {
		return nil, apperrors.ErrUnauthorized(ctx)
	}
	if platform {
		allowed, err := s.repo.Operator(ctx, actor)
		if err != nil {
			return nil, err
		}
		if !allowed {
			return nil, apperrors.ErrForbidden(ctx).SetDetail("Hexta operator access required")
		}
		return s.repo.Requests(ctx, "")
	}
	return s.repo.Requests(ctx, actor)
}
func (s *TenantOnboardingService) Apply(ctx context.Context, actor, name, slug string) (*model.TenantRegistration, *apperrors.Error) {
	if actor == "" {
		return nil, apperrors.ErrUnauthorized(ctx)
	}
	name = strings.TrimSpace(name)
	slug = strings.ToLower(strings.TrimSpace(slug))
	if len(name) < 2 || len(name) > 100 || len(slug) < 2 || len(slug) > 100 || !tenantSlugPattern.MatchString(slug) {
		return nil, apperrors.ErrBadRequest(ctx).SetDetail("name and slug must be 2-100 characters; slug uses lowercase letters, digits and single hyphens")
	}
	row := &model.TenantRegistration{ID: uuid.NewString(), ApplicantID: actor, Name: name, Slug: slug, Status: "pending"}
	err := s.repo.Transaction(ctx, func(r *repository.TenantOnboardingRepository) *apperrors.Error {
		taken, err := r.SlugTaken(ctx, slug)
		if err != nil {
			return err
		}
		if taken {
			return apperrors.ErrConflict(ctx, "tenant slug", "already exists")
		}
		return r.SaveRequest(ctx, row, true)
	})
	return row, err
}
func (s *TenantOnboardingService) Review(ctx context.Context, actor, id, decision, reason string) (*model.TenantRegistration, *apperrors.Error) {
	if decision != "approved" && decision != "rejected" {
		return nil, apperrors.ErrBadRequest(ctx).SetDetail("decision must be approved or rejected")
	}
	reason = strings.TrimSpace(reason)
	if decision == "rejected" && reason == "" {
		return nil, apperrors.ErrBadRequest(ctx).SetDetail("a rejection reason is required")
	}
	var row *model.TenantRegistration
	err := s.repo.Transaction(ctx, func(r *repository.TenantOnboardingRepository) *apperrors.Error {
		allowed, err := r.Operator(ctx, actor)
		if err != nil {
			return err
		}
		if !allowed {
			return apperrors.ErrForbidden(ctx).SetDetail("Hexta operator access required")
		}
		row, err = r.Request(ctx, id)
		if err != nil {
			return err
		}
		if row.Status == decision {
			return nil
		}
		if row.Status != "pending" {
			return apperrors.ErrConflict(ctx, "application", "already reviewed")
		}
		now := s.now().UTC()
		row.Status = decision
		row.ReviewReason = reason
		row.ReviewedBy = actor
		row.ReviewedAt = &now
		if decision == "approved" {
			tenant := &model.Tenant{ID: uuid.NewString(), Name: row.Name, Slug: row.Slug, Plan: model.PlanFree, Status: model.TenantStatusActive, OwnerID: row.ApplicantID}
			if err := r.Provision(ctx, tenant); err != nil {
				return err
			}
			row.TenantID = &tenant.ID
			if err := r.Audit(ctx, tenant.ID, actor, "tenant.approved", tenant.ID, tenant); err != nil {
				return err
			}
		}
		return r.SaveRequest(ctx, row, false)
	})
	return row, err
}

// Raw tokens are returned only on creation; only their hash is persisted.
func invitationHash(token string) string {
	value := sha256.Sum256([]byte(token))
	return hex.EncodeToString(value[:])
}
func (s *TenantOnboardingService) Invite(ctx context.Context, actor, tenant, email string) (*model.TenantInvitation, string, *apperrors.Error) {
	email = strings.ToLower(strings.TrimSpace(email))
	address, parseErr := mail.ParseAddress(email)
	if parseErr != nil || address.Address != email || len(email) > 255 {
		return nil, "", apperrors.ErrBadRequest(ctx).SetDetail("valid recipient email required")
	}
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return nil, "", apperrors.ErrSystemError(ctx, "could not generate invitation")
	}
	token := hex.EncodeToString(tokenBytes)
	row := &model.TenantInvitation{ID: uuid.NewString(), TenantID: tenant, Email: email, Role: model.RoleMember, InvitedBy: actor, TokenHash: invitationHash(token), Status: "pending", ExpiresAt: s.now().UTC().Add(7 * 24 * time.Hour)}
	err := s.repo.Transaction(ctx, func(r *repository.TenantOnboardingRepository) *apperrors.Error {
		if _, err := r.Owner(ctx, tenant, actor); err != nil {
			return err
		}
		matches, err := r.MatchesEmail(ctx, actor, email)
		if err != nil {
			return err
		}
		if matches {
			return apperrors.ErrConflict(ctx, "invitation", "you are already the owner")
		}
		if err := r.ReserveInvitation(ctx, tenant, email); err != nil {
			return err
		}
		if err := r.SaveInvitation(ctx, row, true); err != nil {
			return err
		}
		return r.Audit(ctx, tenant, actor, "tenant.invitation_created", row.ID, row)
	})
	if err != nil {
		return nil, "", err
	}
	return row, token, nil
}
func (s *TenantOnboardingService) Invitations(ctx context.Context, actor, tenant string) ([]model.TenantInvitation, *apperrors.Error) {
	var rows []model.TenantInvitation
	err := s.repo.Transaction(ctx, func(r *repository.TenantOnboardingRepository) *apperrors.Error {
		if _, err := r.Owner(ctx, tenant, actor); err != nil {
			return err
		}
		var err *apperrors.Error
		rows, err = r.Invitations(ctx, tenant)
		return err
	})
	return rows, err
}
func (s *TenantOnboardingService) Revoke(ctx context.Context, actor, tenant, id string) *apperrors.Error {
	return s.repo.Transaction(ctx, func(r *repository.TenantOnboardingRepository) *apperrors.Error {
		if _, err := r.Owner(ctx, tenant, actor); err != nil {
			return err
		}
		row, err := r.Invitation(ctx, id, "", true)
		if err != nil {
			return err
		}
		if row.TenantID != tenant {
			return apperrors.ErrNotFound(ctx, "invitation", "not found")
		}
		if row.Status == "revoked" {
			return nil
		}
		if row.Status != "pending" {
			return apperrors.ErrConflict(ctx, "invitation", "is no longer pending")
		}
		row.Status = "revoked"
		if err := r.SaveInvitation(ctx, row, false); err != nil {
			return err
		}
		return r.Audit(ctx, tenant, actor, "tenant.invitation_revoked", id, row)
	})
}
func (s *TenantOnboardingService) UseInvitation(ctx context.Context, actor, token, action string) (*model.TenantInvitation, string, *apperrors.Error) {
	if actor == "" {
		return nil, "", apperrors.ErrUnauthorized(ctx)
	}
	if len(token) != 64 {
		return nil, "", apperrors.ErrNotFound(ctx, "invitation", "not found")
	}
	if action != "preview" && action != "accept" && action != "decline" {
		return nil, "", apperrors.ErrBadRequest(ctx)
	}
	var row *model.TenantInvitation
	var name string
	err := s.repo.Transaction(ctx, func(r *repository.TenantOnboardingRepository) *apperrors.Error {
		var err *apperrors.Error
		initial, err := r.Invitation(ctx, "", invitationHash(token), false)
		if err != nil {
			return err
		}
		// Always acquire tenant before invitation to match owner mutation lock order.
		tenant, err := r.Tenant(ctx, initial.TenantID)
		if err != nil {
			return err
		}
		if tenant.Status != model.TenantStatusActive {
			return apperrors.ErrForbidden(ctx).SetDetail("tenant is not active")
		}
		name = tenant.Name
		row, err = r.Invitation(ctx, "", invitationHash(token), true)
		if err != nil {
			return err
		}
		matches, err := r.MatchesEmail(ctx, actor, row.Email)
		if err != nil {
			return err
		}
		if !matches {
			return apperrors.ErrForbidden(ctx).SetDetail("sign in with the invited email account")
		}
		if row.Status == "accepted" && row.AcceptedBy == actor {
			if action == "accept" || action == "preview" {
				return nil
			}
		}
		if row.Status != "pending" || !row.ExpiresAt.After(s.now()) {
			return apperrors.ErrConflict(ctx, "invitation", "expired, revoked or already processed")
		}
		if action == "preview" {
			return nil
		}
		if action == "accept" {
			if err := r.AddMember(ctx, row.TenantID, actor); err != nil {
				return err
			}
			row.Status = "accepted"
			row.AcceptedBy = actor
		} else {
			row.Status = "declined"
		}
		if err := r.SaveInvitation(ctx, row, false); err != nil {
			return err
		}
		return r.Audit(ctx, row.TenantID, actor, "tenant.invitation_"+row.Status, row.ID, row)
	})
	return row, name, err
}
