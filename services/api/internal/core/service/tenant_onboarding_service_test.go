package service

import (
	"context"
	"os"
	"sync"
	"testing"
	"time"

	apperrors "github.com/TruongHoang2004/Hexta/packages/shared/pkg/errors"
	"github.com/TruongHoang2004/Hexta/services/api/internal/core/model"
	"github.com/TruongHoang2004/Hexta/services/api/internal/repository"
)

func onboardingFixture(t *testing.T) (salesFixture, *TenantOnboardingService) {
	t.Helper()
	f := salesTestFixture(t)
	for _, user := range []string{f.owner, f.sale, f.peer} {
		if err := f.db.Create(&model.AuthIdentities{UserID: user, Identifier: user + "@example.test", Provider: model.ProviderLocal}).Error; err != nil {
			t.Fatal(err)
		}
	}
	if err := f.db.Create(&model.PlatformOperator{UserID: "operator"}).Error; err != nil {
		t.Fatal(err)
	}
	return f, NewTenantOnboardingService(repository.NewTenantOnboardingRepository(f.db))
}
func TestOnboardingApprovalAndAuthority(t *testing.T) {
	f, s := onboardingFixture(t)
	ctx := context.Background()
	request, err := s.Apply(ctx, f.sale, "Applied Business", "applied-business")
	if err != nil {
		t.Fatal(err)
	}
	var count int64
	f.db.Model(&model.Tenant{}).Where("slug = ?", request.Slug).Count(&count)
	if count != 0 {
		t.Fatal("application granted tenant access before approval")
	}
	_, err = s.Review(ctx, f.owner, request.ID, "approved", "")
	requireSalesStatus(t, err, 403)
	approved, err := s.Review(ctx, "operator", request.ID, "approved", "Verified")
	if err != nil {
		t.Fatal(err)
	}
	if approved.TenantID == nil {
		t.Fatal("approval did not provision tenant")
	}
	repeated, err := s.Review(ctx, "operator", request.ID, "approved", "")
	if err != nil {
		t.Fatal(err)
	}
	if *approved.TenantID != *repeated.TenantID {
		t.Fatal("repeat approval created another tenant")
	}
	_, err = s.Review(ctx, "operator", request.ID, "rejected", "Different decision")
	requireSalesStatus(t, err, 409)
	rows, err := s.Requests(ctx, f.peer, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 0 {
		t.Fatal("applicant saw another account's request")
	}
	_, err = s.Requests(ctx, f.owner, true)
	requireSalesStatus(t, err, 403)
	member, err := repository.NewSalesRepository(f.db).Member(ctx, *approved.TenantID, f.sale)
	if err != nil || member == nil || member.Role != model.RoleOwner {
		t.Fatal("initial owner missing")
	}
	plan := "enterprise"
	_, err = NewTenantService(repository.NewTenantDBRepository(f.db)).UpdateTenant(ctx, f.sale, *approved.TenantID, nil, &plan, nil)
	requireSalesStatus(t, err, 403)
}
func TestOnboardingInvitationActivationAndRecipient(t *testing.T) {
	f, s := onboardingFixture(t)
	ctx := context.Background()
	// Remove the fixture's existing employee membership; an invitation must not restore it until acceptance.
	f.db.Where("tenant_id = ? AND user_id = ?", f.tenant, f.sale).Delete(&model.TenantMember{})
	invitation, token, err := s.Invite(ctx, f.owner, f.tenant, "SALE@example.test")
	if err != nil {
		t.Fatal(err)
	}
	if invitation.TokenHash == token || len(token) != 64 {
		t.Fatal("raw token was not protected")
	}
	_, _, err = s.UseInvitation(ctx, f.peer, token, "accept")
	requireSalesStatus(t, err, 403)
	_, _, err = s.UseInvitation(ctx, f.sale, token, "preview")
	if err != nil {
		t.Fatal(err)
	}
	_, _, err = f.service.Orders(ctx, f.tenant, f.sale, 1, 20)
	requireSalesStatus(t, err, 403)
	accepted, _, err := s.UseInvitation(ctx, f.sale, token, "accept")
	if err != nil {
		t.Fatal(err)
	}
	if accepted.Status != "accepted" {
		t.Fatal("not accepted")
	}
	if _, _, err := f.service.Orders(ctx, f.tenant, f.sale, 1, 20); err != nil {
		t.Fatal("accepted member cannot access tenant", err)
	}
	if _, _, err := s.UseInvitation(ctx, f.sale, token, "accept"); err != nil {
		t.Fatal("repeat accept must be idempotent", err)
	}
	var count int64
	f.db.Model(&model.TenantMember{}).Where("tenant_id = ? AND user_id = ?", f.tenant, f.sale).Count(&count)
	if count != 1 {
		t.Fatal("duplicate membership")
	}
	// A consumed invitation cannot restore access after a subsequent owner revocation.
	f.db.Where("tenant_id = ? AND user_id = ?", f.tenant, f.sale).Delete(&model.TenantMember{})
	if _, _, err := s.UseInvitation(ctx, f.sale, token, "accept"); err != nil {
		t.Fatal(err)
	}
	_, _, err = f.service.Orders(ctx, f.tenant, f.sale, 1, 20)
	requireSalesStatus(t, err, 403)
}
func TestOnboardingRevokedExpiredAndReissuedInvitations(t *testing.T) {
	f, s := onboardingFixture(t)
	ctx := context.Background()
	_, _, err := s.Invite(ctx, f.sale, f.tenant, "peer@example.test")
	requireSalesStatus(t, err, 403)
	first, old, err := s.Invite(ctx, f.owner, f.tenant, "sale@example.test")
	if err != nil {
		t.Fatal(err)
	}
	next, token, err := s.Invite(ctx, f.owner, f.tenant, "sale@example.test")
	if err != nil {
		t.Fatal(err)
	}
	_, _, err = s.UseInvitation(ctx, f.sale, old, "preview")
	requireSalesStatus(t, err, 409)
	if err := s.Revoke(ctx, f.owner, f.other, next.ID); err == nil {
		t.Fatal("cross-tenant revoke allowed")
	}
	if err := s.Revoke(ctx, f.owner, f.tenant, next.ID); err != nil {
		t.Fatal(err)
	}
	_, _, err = s.UseInvitation(ctx, f.sale, token, "accept")
	requireSalesStatus(t, err, 409)
	_, token, err = s.Invite(ctx, f.owner, f.tenant, "sale@example.test")
	if err != nil {
		t.Fatal(err)
	}
	s.now = func() time.Time { return first.ExpiresAt.Add(24 * time.Hour) }
	_, _, err = s.UseInvitation(ctx, f.sale, token, "accept")
	requireSalesStatus(t, err, 409)
}
func TestOnboardingConcurrentApprovalPostgres(t *testing.T) {
	if os.Getenv("SALES_TEST_DATABASE_URL") == "" && os.Getenv("API_TEST_DATABASE_URL") == "" {
		t.Skip("requires PostgreSQL row locks")
	}
	f, s := onboardingFixture(t)
	ctx := context.Background()
	request, err := s.Apply(ctx, f.sale, "Race Business", "race-business")
	if err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	results := make(chan *apperrors.Error, 2)
	start := make(chan struct{})
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			_, err := s.Review(ctx, "operator", request.ID, "approved", "")
			results <- err
		}()
	}
	close(start)
	wg.Wait()
	close(results)
	for err := range results {
		if err != nil {
			t.Fatal(err)
		}
	}
	var count int64
	f.db.Model(&model.Tenant{}).Where("slug = ?", request.Slug).Count(&count)
	if count != 1 {
		t.Fatal("approval provisioned multiple tenants")
	}
}
func TestOnboardingConcurrentAcceptPostgres(t *testing.T) {
	if os.Getenv("SALES_TEST_DATABASE_URL") == "" && os.Getenv("API_TEST_DATABASE_URL") == "" {
		t.Skip("requires PostgreSQL row locks")
	}
	f, s := onboardingFixture(t)
	ctx := context.Background()
	f.db.Where("tenant_id = ? AND user_id = ?", f.tenant, f.sale).Delete(&model.TenantMember{})
	_, token, err := s.Invite(ctx, f.owner, f.tenant, "sale@example.test")
	if err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	results := make(chan *apperrors.Error, 2)
	start := make(chan struct{})
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			_, _, err := s.UseInvitation(ctx, f.sale, token, "accept")
			results <- err
		}()
	}
	close(start)
	wg.Wait()
	close(results)
	for err := range results {
		if err != nil {
			t.Fatal(err)
		}
	}
	var count int64
	f.db.Model(&model.TenantMember{}).Where("tenant_id = ? AND user_id = ?", f.tenant, f.sale).Count(&count)
	if count != 1 {
		t.Fatal("accept provisioned multiple memberships")
	}
}
