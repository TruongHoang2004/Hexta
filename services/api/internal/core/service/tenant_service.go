package service

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"gitlab.com/ecommercehub1/api/internal/core/model"
	"gitlab.com/ecommercehub1/api/internal/repository"
	"gitlab.com/ecommercehub1/shared/pkg/errors"
)

type ITenantService interface {
	CreateTenant(ctx context.Context, userID, name, slug, plan string) (*model.Tenant, *errors.Error)
	GetTenant(ctx context.Context, userID, tenantID string) (*model.Tenant, *errors.Error)
	GetUserTenants(ctx context.Context, userID string) ([]*model.Tenant, *errors.Error)
	ListMembers(ctx context.Context, userID, tenantID string) ([]*model.TenantMember, *errors.Error)
	InviteMember(ctx context.Context, currentUserID, tenantID, targetUserID, role string) (*model.TenantMember, *errors.Error)
	UpdateTenant(ctx context.Context, currentUserID, tenantID string, name *string, plan *string, status *string) (*model.Tenant, *errors.Error)
}

type TenantService struct {
	*baseService
	tenantRepo repository.ITenantRepository
}

func NewTenantService(tenantRepo repository.ITenantRepository) *TenantService {
	return &TenantService{
		baseService: NewBaseService(),
		tenantRepo:  tenantRepo,
	}
}

func (s *TenantService) CreateTenant(ctx context.Context, userID, name, slug, plan string) (*model.Tenant, *errors.Error) {
	name = strings.TrimSpace(name)
	slug = strings.TrimSpace(slug)

	if name == "" {
		return nil, errors.ErrBadRequest(ctx).SetDetail("tenant name is required")
	}
	if slug == "" {
		return nil, errors.ErrBadRequest(ctx).SetDetail("tenant slug is required")
	}

	existingSlug, err := s.tenantRepo.GetTenantBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}
	if existingSlug != nil {
		return nil, errors.ErrConflict(ctx, "Slug", "already exists")
	}

	if plan == "" {
		plan = model.PlanFree
	}

	tenant := &model.Tenant{
		ID:      uuid.New().String(),
		Name:    name,
		Slug:    slug,
		Plan:    plan,
		Status:  model.TenantStatusActive,
		OwnerID: userID,
	}

	return s.tenantRepo.CreateTenant(ctx, tenant, userID)
}

func (s *TenantService) GetTenant(ctx context.Context, userID, tenantID string) (*model.Tenant, *errors.Error) {
	tenant, err := s.tenantRepo.GetTenantByID(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	if tenant == nil {
		return nil, errors.ErrNotFound(ctx, "Tenant", "not found")
	}

	member, err := s.tenantRepo.GetTenantMember(ctx, tenantID, userID)
	if err != nil {
		return nil, err
	}
	if member == nil {
		return nil, errors.ErrForbidden(ctx).SetDetail("user does not belong to this workspace")
	}

	return tenant, nil
}

func (s *TenantService) GetUserTenants(ctx context.Context, userID string) ([]*model.Tenant, *errors.Error) {
	return s.tenantRepo.GetUserTenants(ctx, userID)
}

func (s *TenantService) ListMembers(ctx context.Context, userID, tenantID string) ([]*model.TenantMember, *errors.Error) {
	member, err := s.tenantRepo.GetTenantMember(ctx, tenantID, userID)
	if err != nil {
		return nil, err
	}
	if member == nil {
		return nil, errors.ErrForbidden(ctx).SetDetail("user does not belong to this workspace")
	}

	return s.tenantRepo.ListTenantMembers(ctx, tenantID)
}

func (s *TenantService) InviteMember(ctx context.Context, currentUserID, tenantID, targetUserID, role string) (*model.TenantMember, *errors.Error) {
	callerMember, err := s.tenantRepo.GetTenantMember(ctx, tenantID, currentUserID)
	if err != nil {
		return nil, err
	}
	if callerMember == nil || (callerMember.Role != model.RoleOwner && callerMember.Role != model.RoleAdmin) {
		return nil, errors.ErrForbidden(ctx).SetDetail("only owner or admin can invite members")
	}

	if targetUserID == "" {
		return nil, errors.ErrBadRequest(ctx).SetDetail("target user_id is required")
	}

	if role == "" {
		role = model.RoleMember
	}
	if role != model.RoleAdmin && role != model.RoleMember {
		return nil, errors.ErrBadRequest(ctx).SetDetail("role must be admin or member")
	}

	existing, err := s.tenantRepo.GetTenantMember(ctx, tenantID, targetUserID)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, errors.ErrConflict(ctx, "User", "is already a member of this workspace")
	}

	newMember := &model.TenantMember{
		TenantID: tenantID,
		UserID:   targetUserID,
		Role:     role,
	}

	return s.tenantRepo.AddTenantMember(ctx, newMember)
}

func (s *TenantService) UpdateTenant(ctx context.Context, currentUserID, tenantID string, name *string, plan *string, status *string) (*model.Tenant, *errors.Error) {
	callerMember, err := s.tenantRepo.GetTenantMember(ctx, tenantID, currentUserID)
	if err != nil {
		return nil, err
	}
	if callerMember == nil || (callerMember.Role != model.RoleOwner && callerMember.Role != model.RoleAdmin) {
		return nil, errors.ErrForbidden(ctx).SetDetail("only owner or admin can update workspace")
	}

	tenant, err := s.tenantRepo.GetTenantByID(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	if tenant == nil {
		return nil, errors.ErrNotFound(ctx, "Tenant", "not found")
	}

	if name != nil && strings.TrimSpace(*name) != "" {
		tenant.Name = strings.TrimSpace(*name)
	}
	if plan != nil && strings.TrimSpace(*plan) != "" {
		tenant.Plan = strings.TrimSpace(*plan)
	}
	if status != nil && strings.TrimSpace(*status) != "" {
		tenant.Status = strings.TrimSpace(*status)
	}

	return s.tenantRepo.UpdateTenant(ctx, tenant)
}
