package service

import (
	"context"
	"testing"

	"gitlab.com/ecommercehub1/api/internal/core/model"
	"gitlab.com/ecommercehub1/shared/pkg/errors"
)

type mockTenantRepo struct {
	tenants  map[string]*model.Tenant
	slugs    map[string]*model.Tenant
	members  map[string]*model.TenantMember
	memberList map[string][]*model.TenantMember
}

func newMockTenantRepo() *mockTenantRepo {
	return &mockTenantRepo{
		tenants:    make(map[string]*model.Tenant),
		slugs:      make(map[string]*model.Tenant),
		members:    make(map[string]*model.TenantMember),
		memberList: make(map[string][]*model.TenantMember),
	}
}

func (m *mockTenantRepo) CreateTenant(ctx context.Context, tenant *model.Tenant, ownerID string) (*model.Tenant, *errors.Error) {
	m.tenants[tenant.ID] = tenant
	m.slugs[tenant.Slug] = tenant
	ownerMember := &model.TenantMember{
		ID:       int64(len(m.members) + 1),
		TenantID: tenant.ID,
		UserID:   ownerID,
		Role:     model.RoleOwner,
	}
	key := tenant.ID + ":" + ownerID
	m.members[key] = ownerMember
	m.memberList[tenant.ID] = append(m.memberList[tenant.ID], ownerMember)
	return tenant, nil
}

func (m *mockTenantRepo) GetTenantByID(ctx context.Context, id string) (*model.Tenant, *errors.Error) {
	if t, ok := m.tenants[id]; ok {
		return t, nil
	}
	return nil, nil
}

func (m *mockTenantRepo) GetTenantBySlug(ctx context.Context, slug string) (*model.Tenant, *errors.Error) {
	if t, ok := m.slugs[slug]; ok {
		return t, nil
	}
	return nil, nil
}

func (m *mockTenantRepo) UpdateTenant(ctx context.Context, tenant *model.Tenant) (*model.Tenant, *errors.Error) {
	m.tenants[tenant.ID] = tenant
	m.slugs[tenant.Slug] = tenant
	return tenant, nil
}

func (m *mockTenantRepo) GetUserTenants(ctx context.Context, userID string) ([]*model.Tenant, *errors.Error) {
	var result []*model.Tenant
	for _, member := range m.members {
		if member.UserID == userID {
			if t, ok := m.tenants[member.TenantID]; ok {
				result = append(result, t)
			}
		}
	}
	return result, nil
}

func (m *mockTenantRepo) GetTenantMember(ctx context.Context, tenantID, userID string) (*model.TenantMember, *errors.Error) {
	key := tenantID + ":" + userID
	if member, ok := m.members[key]; ok {
		return member, nil
	}
	return nil, nil
}

func (m *mockTenantRepo) ListTenantMembers(ctx context.Context, tenantID string) ([]*model.TenantMember, *errors.Error) {
	return m.memberList[tenantID], nil
}

func (m *mockTenantRepo) AddTenantMember(ctx context.Context, member *model.TenantMember) (*model.TenantMember, *errors.Error) {
	member.ID = int64(len(m.members) + 1)
	key := member.TenantID + ":" + member.UserID
	m.members[key] = member
	m.memberList[member.TenantID] = append(m.memberList[member.TenantID], member)
	return member, nil
}

func TestTenantService_CreateTenant(t *testing.T) {
	repo := newMockTenantRepo()
	svc := NewTenantService(repo)
	ctx := context.Background()

	// 1. Successful creation
	tenant, err := svc.CreateTenant(ctx, "user-1", "Acme Inc", "acme-inc", model.PlanPro)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if tenant.Name != "Acme Inc" || tenant.Slug != "acme-inc" || tenant.Plan != model.PlanPro {
		t.Fatalf("unexpected tenant data: %+v", tenant)
	}

	// 2. Duplicate slug should return conflict
	_, err = svc.CreateTenant(ctx, "user-2", "Another Acme", "acme-inc", model.PlanFree)
	if err == nil {
		t.Fatal("expected conflict error on duplicate slug, got nil")
	}

	// 3. Empty name should return bad request
	_, err = svc.CreateTenant(ctx, "user-1", "", "unique-slug", model.PlanFree)
	if err == nil {
		t.Fatal("expected bad request error on empty name, got nil")
	}
}

func TestTenantService_GetTenant_MembershipCheck(t *testing.T) {
	repo := newMockTenantRepo()
	svc := NewTenantService(repo)
	ctx := context.Background()

	tenant, err := svc.CreateTenant(ctx, "owner-user", "My Workspace", "my-workspace", model.PlanFree)
	if err != nil {
		t.Fatalf("failed to create tenant: %v", err)
	}

	// Owner can access
	fetched, err := svc.GetTenant(ctx, "owner-user", tenant.ID)
	if err != nil {
		t.Fatalf("expected owner to access tenant, got error: %v", err)
	}
	if fetched.ID != tenant.ID {
		t.Fatalf("expected tenant ID %s, got %s", tenant.ID, fetched.ID)
	}

	// Non-member cannot access
	_, err = svc.GetTenant(ctx, "stranger-user", tenant.ID)
	if err == nil {
		t.Fatal("expected forbidden error for non-member, got nil")
	}
}

func TestTenantService_InviteMember(t *testing.T) {
	repo := newMockTenantRepo()
	svc := NewTenantService(repo)
	ctx := context.Background()

	tenant, _ := svc.CreateTenant(ctx, "owner-user", "Dev Org", "dev-org", model.PlanPro)

	// Non-admin cannot invite
	_, err := svc.InviteMember(ctx, "random-user", tenant.ID, "target-user", model.RoleMember)
	if err == nil {
		t.Fatal("expected forbidden error when non-admin invites member, got nil")
	}

	// Owner can invite
	newMember, err := svc.InviteMember(ctx, "owner-user", tenant.ID, "invited-user", model.RoleMember)
	if err != nil {
		t.Fatalf("expected owner to invite member, got: %v", err)
	}
	if newMember.Role != model.RoleMember || newMember.UserID != "invited-user" {
		t.Fatalf("unexpected member data: %+v", newMember)
	}

	// Cannot invite same user twice
	_, err = svc.InviteMember(ctx, "owner-user", tenant.ID, "invited-user", model.RoleMember)
	if err == nil {
		t.Fatal("expected conflict error on duplicate invite, got nil")
	}
}
