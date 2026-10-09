package repository

import (
	"context"
	"testing"

	"github.com/TruongHoang2004/Hexta/services/api/internal/core/model"
	"github.com/google/uuid"
)

func TestTenantRepository_CRUD(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}

	tx := db.Begin()
	defer tx.Rollback()

	repo := NewTenantDBRepository(tx)
	ctx := context.Background()

	testOwnerID := uuid.New().String()
	testSlug := "test-org-" + uuid.New().String()[:8]
	testName := "Test Organization"

	// 1. Create Tenant
	tenant := &model.Tenant{
		ID:   uuid.New().String(),
		Name: testName,
		Slug: testSlug,
	}

	created, err := repo.CreateTenant(ctx, tenant, testOwnerID)
	if err != nil {
		t.Fatalf("failed to create tenant: %v", err)
	}
	if created == nil || created.ID == "" {
		t.Fatalf("expected created tenant with ID, got %+v", created)
	}

	// 2. Get Tenant By ID
	byID, err := repo.GetTenantByID(ctx, created.ID)
	if err != nil {
		t.Fatalf("failed to get tenant by ID: %v", err)
	}
	if byID == nil || byID.Slug != testSlug {
		t.Fatalf("expected tenant slug %s, got %+v", testSlug, byID)
	}

	// 3. Get Tenant By Slug
	bySlug, err := repo.GetTenantBySlug(ctx, testSlug)
	if err != nil {
		t.Fatalf("failed to get tenant by slug: %v", err)
	}
	if bySlug == nil || bySlug.ID != created.ID {
		t.Fatalf("expected tenant ID %s, got %+v", created.ID, bySlug)
	}

	// 4. Update Tenant
	byID.Name = "Updated Organization Name"
	updated, err := repo.UpdateTenant(ctx, byID)
	if err != nil {
		t.Fatalf("failed to update tenant: %v", err)
	}
	if updated.Name != "Updated Organization Name" {
		t.Fatalf("expected updated name, got %s", updated.Name)
	}

	// 5. Get Tenant Member
	member, err := repo.GetTenantMember(ctx, created.ID, testOwnerID)
	if err != nil {
		t.Fatalf("failed to get tenant member: %v", err)
	}
	if member == nil || member.Role != model.RoleOwner {
		t.Fatalf("expected owner member role, got %+v", member)
	}

	// 6. List Tenant Members
	members, err := repo.ListTenantMembers(ctx, created.ID)
	if err != nil {
		t.Fatalf("failed to list tenant members: %v", err)
	}
	if len(members) == 0 {
		t.Fatalf("expected at least 1 member, got %d", len(members))
	}
}
