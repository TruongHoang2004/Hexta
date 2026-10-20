package service

import (
	"context"
	"os"
	"testing"

	apperrors "github.com/TruongHoang2004/Hexta/packages/shared/pkg/errors"
	"github.com/TruongHoang2004/Hexta/services/api/internal/core/model"
	"github.com/TruongHoang2004/Hexta/services/api/internal/core/types"
	"github.com/TruongHoang2004/Hexta/services/api/internal/repository"
	"github.com/TruongHoang2004/Hexta/services/api/internal/testutil"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type workflowFixture struct {
	salesFixture
	workflow   *WorkflowService
	definition *repository.WorkflowDefinition
	order      *model.Order
}

func workflowTestFixture(t *testing.T, mode string) workflowFixture {
	t.Helper()
	f := salesTestFixture(t)
	s := NewWorkflowService(repository.NewWorkflowRepository(f.db))
	ctx := context.Background()
	d, err := s.Publish(ctx, f.tenant, f.owner, "", uuid.NewString(), types.WorkflowInput{Name: "Service", Steps: []types.WorkflowStepInput{{Name: "Prepare", HandoffMode: mode}, {Name: "Review", HandoffMode: "direct"}, {Name: "Finish", HandoffMode: "direct"}}})
	if err != nil {
		t.Fatal(err)
	}
	draft, err := f.service.SaveDraft(ctx, f.tenant, f.sale, "", f.completeInput(t))
	if err != nil {
		t.Fatal(err)
	}
	order, err := f.service.Confirm(ctx, f.tenant, f.sale, draft.ID, draft.Version)
	if err != nil {
		t.Fatal(err)
	}
	return workflowFixture{f, s, d, order}
}
func (f workflowFixture) input() types.WorkInput {
	return types.WorkInput{OrderID: f.order.ID, WorkflowVersionID: f.definition.Versions[0].ID, ResponsibleID: f.sale}
}
func (f workflowFixture) create(t *testing.T) *repository.WorkItem {
	t.Helper()
	row, err := f.workflow.Create(context.Background(), f.tenant, f.owner, uuid.NewString(), f.input())
	if err != nil {
		t.Fatal(err)
	}
	return row
}
func TestWorkflowAcceptanceAndPinnedVersions(t *testing.T) {
	f := workflowTestFixture(t, "acceptance")
	ctx := context.Background()
	row := f.create(t)
	_, err := f.workflow.Command(ctx, f.tenant, f.sale, row.ID, "complete", types.WorkCommand{ExpectedVersion: 1})
	requireSalesStatus(t, err, 409)
	row, err = f.workflow.Command(ctx, f.tenant, f.sale, row.ID, "handoff", types.WorkCommand{ExpectedVersion: 1, RecipientID: f.peer})
	if err != nil {
		t.Fatal(err)
	}
	if row.State != "waiting" || *row.ResponsibleID != f.sale || row.CurrentStep != 0 || *row.RecipientID != f.peer {
		t.Fatal("pending handoff lost current responsibility")
	}
	_, err = f.workflow.Command(ctx, f.tenant, f.owner, row.ID, "accept", types.WorkCommand{ExpectedVersion: 2})
	requireSalesStatus(t, err, 403)
	_, err = f.workflow.Command(ctx, f.tenant, f.sale, row.ID, "handoff", types.WorkCommand{ExpectedVersion: 2, RecipientID: f.owner})
	requireSalesStatus(t, err, 409)
	d, err := f.workflow.Publish(ctx, f.tenant, f.owner, f.definition.ID, "", types.WorkflowInput{Name: "Changed", ExpectedVersion: 1, Steps: []types.WorkflowStepInput{{Name: "New sole step", HandoffMode: "direct"}}})
	if err != nil {
		t.Fatal(err)
	}
	if d.Version != 2 {
		t.Fatal("new publication did not increment version")
	}
	row, err = f.workflow.Command(ctx, f.tenant, f.peer, row.ID, "accept", types.WorkCommand{ExpectedVersion: 2})
	if err != nil {
		t.Fatal(err)
	}
	if row.State != "active" || *row.ResponsibleID != f.peer || row.CurrentStep != 1 || row.Workflow.Number != 1 || len(row.Workflow.Steps) != 3 || row.Workflow.Steps[0].Name != "Prepare" {
		t.Fatal("acceptance or pinned workflow changed")
	}
	_, err = f.workflow.Command(ctx, f.tenant, f.peer, row.ID, "accept", types.WorkCommand{ExpectedVersion: 2})
	requireSalesStatus(t, err, 409)
	row, err = f.workflow.Command(ctx, f.tenant, f.peer, row.ID, "handoff", types.WorkCommand{ExpectedVersion: 3, RecipientID: f.sale})
	if err != nil {
		t.Fatal(err)
	}
	if row.State != "active" || row.CurrentStep != 2 || *row.ResponsibleID != f.sale {
		t.Fatal("direct handoff did not advance immediately")
	}
	row, err = f.workflow.Command(ctx, f.tenant, f.sale, row.ID, "complete", types.WorkCommand{ExpectedVersion: 4})
	if err != nil {
		t.Fatal(err)
	}
	stored, err := f.workflow.Item(ctx, f.tenant, f.owner, row.ID)
	if err != nil {
		t.Fatal(err)
	}
	if stored.State != "completed" || len(stored.Events) != 5 {
		t.Fatalf("completion/history not persisted: %+v", stored)
	}
	if stored.Events[2].Action != "accept" || *stored.Events[2].FromResponsibleID != f.sale || *stored.Events[2].ResponsibleID != f.peer {
		t.Fatal("history missing acceptance transfer")
	}
	orders, err := f.service.Order(ctx, f.tenant, f.owner, f.order.ID)
	if err != nil || orders.Status != model.OrderStatusConfirmed || orders.Version != f.order.Version {
		t.Fatal("work changed original order")
	}
}
func TestWorkflowOwnerReassignmentAndMembership(t *testing.T) {
	f := workflowTestFixture(t, "acceptance")
	ctx := context.Background()
	row := f.create(t)
	_, err := f.workflow.Command(ctx, f.tenant, f.peer, row.ID, "assign", types.WorkCommand{ExpectedVersion: 1, ResponsibleID: &f.peer})
	requireSalesStatus(t, err, 403)
	_, err = f.workflow.Command(ctx, f.tenant, f.peer, row.ID, "handoff", types.WorkCommand{ExpectedVersion: 1, RecipientID: f.sale})
	requireSalesStatus(t, err, 403)
	_, err = f.workflow.Command(ctx, f.tenant, f.sale, row.ID, "handoff", types.WorkCommand{ExpectedVersion: 1, RecipientID: "outsider"})
	requireSalesStatus(t, err, 400)
	row, err = f.workflow.Command(ctx, f.tenant, f.sale, row.ID, "handoff", types.WorkCommand{ExpectedVersion: 1, RecipientID: f.peer})
	if err != nil {
		t.Fatal(err)
	}
	row, err = f.workflow.Command(ctx, f.tenant, f.owner, row.ID, "assign", types.WorkCommand{ExpectedVersion: 2, ResponsibleID: &f.owner})
	if err != nil {
		t.Fatal(err)
	}
	if row.RecipientID != nil || row.CurrentStep != 0 || row.State != "active" {
		t.Fatal("reassign must cancel nomination without advancing")
	}
	_, err = f.workflow.Command(ctx, f.tenant, f.peer, row.ID, "accept", types.WorkCommand{ExpectedVersion: 3})
	requireSalesStatus(t, err, 403)
	row, err = f.workflow.Command(ctx, f.tenant, f.owner, row.ID, "assign", types.WorkCommand{ExpectedVersion: 3})
	if err != nil {
		t.Fatal(err)
	}
	if row.ResponsibleID != nil {
		t.Fatal("owner unassignment did not persist")
	}
	row, err = f.workflow.Command(ctx, f.tenant, f.owner, row.ID, "assign", types.WorkCommand{ExpectedVersion: 4, ResponsibleID: &f.sale})
	if err != nil {
		t.Fatal(err)
	}
	row, err = f.workflow.Command(ctx, f.tenant, f.sale, row.ID, "handoff", types.WorkCommand{ExpectedVersion: 5, RecipientID: f.peer})
	if err != nil {
		t.Fatal(err)
	}
	if err := f.db.Where("tenant_id = ? AND user_id = ?", f.tenant, f.peer).Delete(&model.TenantMember{}).Error; err != nil {
		t.Fatal(err)
	}
	_, err = f.workflow.Command(ctx, f.tenant, f.peer, row.ID, "accept", types.WorkCommand{ExpectedVersion: 6})
	requireSalesStatus(t, err, 403)
	stored, err := f.workflow.Item(ctx, f.tenant, f.owner, row.ID)
	if err != nil || stored.State != "waiting" || stored.Version != 6 {
		t.Fatal("rejected acceptance changed work")
	}
}
func TestWorkflowCreationReplayAndBoundaries(t *testing.T) {
	f := workflowTestFixture(t, "direct")
	ctx := context.Background()
	key := uuid.NewString()
	input := f.input()
	row, err := f.workflow.Create(ctx, f.tenant, f.owner, key, input)
	if err != nil {
		t.Fatal(err)
	}
	again, err := f.workflow.Create(ctx, f.tenant, f.owner, key, input)
	if err != nil || again.ID != row.ID || len(again.Events) != 1 {
		t.Fatal("creation replay duplicated work")
	}
	input.ResponsibleID = f.peer
	_, err = f.workflow.Create(ctx, f.tenant, f.owner, key, input)
	requireSalesStatus(t, err, 409)
	_, err = f.workflow.Create(ctx, f.tenant, f.sale, uuid.NewString(), input)
	requireSalesStatus(t, err, 403)
	_, err = f.workflow.Create(ctx, f.other, f.owner, uuid.NewString(), input)
	requireSalesStatus(t, err, 404)
	_, err = f.workflow.Item(ctx, f.other, f.owner, row.ID)
	requireSalesStatus(t, err, 404)
	_, _, err = f.workflow.Items(ctx, f.other, f.sale, 1, 20)
	requireSalesStatus(t, err, 403)
	_, err = f.workflow.Create(ctx, f.tenant, f.owner, "", input)
	requireSalesStatus(t, err, 400)
	draft, err := f.service.SaveDraft(ctx, f.tenant, f.sale, "", types.DraftInput{})
	if err != nil {
		t.Fatal(err)
	}
	input.OrderID = draft.ID
	_, err = f.workflow.Create(ctx, f.tenant, f.owner, uuid.NewString(), input)
	requireSalesStatus(t, err, 409)
	input = f.input()
	input.WorkflowVersionID = uuid.NewString()
	_, err = f.workflow.Create(ctx, f.tenant, f.owner, uuid.NewString(), input)
	requireSalesStatus(t, err, 404)
	var count int64
	f.db.Model(&repository.WorkItem{}).Count(&count)
	if count != 1 {
		t.Fatal("failed commands created work")
	}
	var events int64
	f.db.Model(&repository.WorkEvent{}).Count(&events)
	if events != 1 {
		t.Fatal("failed commands leaked events")
	}
}
func TestWorkflowPublicationValidationAndReplay(t *testing.T) {
	f := workflowTestFixture(t, "direct")
	ctx := context.Background()
	input := types.WorkflowInput{Name: "Flexible", Steps: []types.WorkflowStepInput{{Name: "Review", HandoffMode: "direct"}}}
	key := uuid.NewString()
	d, err := f.workflow.Publish(ctx, f.tenant, f.owner, "", key, input)
	if err != nil {
		t.Fatal(err)
	}
	again, err := f.workflow.Publish(ctx, f.tenant, f.owner, "", key, input)
	if err != nil || again.ID != d.ID || len(again.Versions) != 1 {
		t.Fatal("workflow create replay failed")
	}
	input.Name = "Different"
	_, err = f.workflow.Publish(ctx, f.tenant, f.owner, "", key, input)
	requireSalesStatus(t, err, 409)
	input.ExpectedVersion = 1
	d, err = f.workflow.Publish(ctx, f.tenant, f.owner, d.ID, "", input)
	if err != nil || d.Version != 2 {
		t.Fatal("publication failed")
	}
	_, err = f.workflow.Publish(ctx, f.tenant, f.owner, d.ID, "", input)
	requireSalesStatus(t, err, 409)
	_, err = f.workflow.Publish(ctx, f.tenant, f.sale, "", uuid.NewString(), types.WorkflowInput{Name: "Denied", Steps: input.Steps})
	requireSalesStatus(t, err, 403)
	input.Name = " "
	_, err = f.workflow.Publish(ctx, f.tenant, f.owner, "", uuid.NewString(), input)
	requireSalesStatus(t, err, 400)
	rows, total, err := f.workflow.Definitions(ctx, f.tenant, f.sale, 1, 20)
	if err != nil || total != 2 {
		t.Fatal("definitions list failed")
	}
	for _, row := range rows {
		if row.ID == d.ID && (len(row.Versions) != 1 || row.Versions[0].Number != 2) {
			t.Fatal("list must return only latest version")
		}
	}
}
func TestWorkflowConcurrentAcceptanceAndReassignmentPostgres(t *testing.T) {
	if os.Getenv("SALES_TEST_DATABASE_URL") == "" && os.Getenv("API_TEST_DATABASE_URL") == "" {
		t.Skip("requires disposable PostgreSQL row locks")
	}
	f := workflowTestFixture(t, "acceptance")
	ctx := context.Background()
	row := f.create(t)
	row, err := f.workflow.Command(ctx, f.tenant, f.sale, row.ID, "handoff", types.WorkCommand{ExpectedVersion: 1, RecipientID: f.peer})
	if err != nil {
		t.Fatal(err)
	}
	start := make(chan struct{})
	results := make(chan *apperrors.Error, 2)
	go func() {
		<-start
		_, err := f.workflow.Command(ctx, f.tenant, f.peer, row.ID, "accept", types.WorkCommand{ExpectedVersion: 2})
		results <- err
	}()
	go func() {
		<-start
		_, err := f.workflow.Command(ctx, f.tenant, f.owner, row.ID, "assign", types.WorkCommand{ExpectedVersion: 2, ResponsibleID: &f.owner})
		results <- err
	}()
	close(start)
	wins, conflicts := 0, 0
	for i := 0; i < 2; i++ {
		err := <-results
		if err == nil {
			wins++
		} else if err.GetHttpStatus() == 409 {
			conflicts++
		} else {
			t.Fatal(err)
		}
	}
	if wins != 1 || conflicts != 1 {
		t.Fatalf("expected single winner: %d/%d", wins, conflicts)
	}
	stored, err := f.workflow.Item(ctx, f.tenant, f.owner, row.ID)
	if err != nil || stored.Version != 3 || len(stored.Events) != 3 {
		t.Fatal("concurrent transitions duplicated state/events")
	}
}

func TestWorkflowConcurrentCreationReplayPostgres(t *testing.T) {
	if os.Getenv("SALES_TEST_DATABASE_URL") == "" && os.Getenv("API_TEST_DATABASE_URL") == "" {
		t.Skip("requires disposable PostgreSQL row locks")
	}
	f := workflowTestFixture(t, "direct")
	ctx, key, input := context.Background(), uuid.NewString(), f.input()
	type result struct {
		row *repository.WorkItem
		err *apperrors.Error
	}
	start, results := make(chan struct{}), make(chan result, 2)
	for i := 0; i < 2; i++ {
		go func() {
			<-start
			row, err := f.workflow.Create(ctx, f.tenant, f.owner, key, input)
			results <- result{row, err}
		}()
	}
	close(start)
	a, b := <-results, <-results
	if a.err != nil || b.err != nil || a.row.ID != b.row.ID {
		t.Fatalf("concurrent creation did not replay: %v/%v", a.err, b.err)
	}
	var count int64
	f.db.Model(&repository.WorkEvent{}).Count(&count)
	if count != 1 {
		t.Fatalf("expected one creation event, got %d", count)
	}
}

// This check uses the actual Atlas schema, including composite foreign keys,
// and additionally probes constraint rejection on an explicit pre-migrated database.
func TestWorkflowMigratedSchemaPostgres(t *testing.T) {
	db := testutil.Database(t)
	var err error
	tx := db.Begin()
	if tx.Error != nil {
		t.Fatal(tx.Error)
	}
	defer tx.Rollback()
	tenant, other, order := uuid.NewString(), uuid.NewString(), uuid.NewString()
	for _, row := range []any{
		&model.Tenant{ID: tenant, Name: "Migrated", Slug: tenant, OwnerID: "owner", Status: model.TenantStatusActive},
		&model.Tenant{ID: other, Name: "Other", Slug: other, OwnerID: "owner", Status: model.TenantStatusActive},
		&model.TenantMember{TenantID: tenant, UserID: "owner", Role: model.RoleOwner},
		&model.TenantMember{TenantID: tenant, UserID: "employee", Role: model.RoleMember},
		&model.Order{ID: order, TenantID: tenant, OrderNumber: order, Status: model.OrderStatusConfirmed},
	} {
		if err := tx.Create(row).Error; err != nil {
			t.Fatal(err)
		}
	}
	s := NewWorkflowService(repository.NewWorkflowRepository(tx))
	ctx := context.Background()
	d, failure := s.Publish(ctx, tenant, "owner", "", uuid.NewString(), dtoWorkflowServiceInput())
	if failure != nil {
		t.Fatal(failure)
	}
	work, failure := s.Create(ctx, tenant, "owner", uuid.NewString(), types.WorkInput{OrderID: order, WorkflowVersionID: d.Versions[0].ID, ResponsibleID: "owner"})
	if failure != nil {
		t.Fatal(failure)
	}
	work, failure = s.Command(ctx, tenant, "owner", work.ID, "handoff", types.WorkCommand{ExpectedVersion: 1, RecipientID: "employee"})
	if failure != nil {
		t.Fatal(failure)
	}
	work, failure = s.Command(ctx, tenant, "employee", work.ID, "accept", types.WorkCommand{ExpectedVersion: 2})
	if failure != nil {
		t.Fatal(failure)
	}
	work, failure = s.Command(ctx, tenant, "employee", work.ID, "complete", types.WorkCommand{ExpectedVersion: 3})
	if failure != nil || work.State != "completed" {
		t.Fatal("migrated-schema transition path failed", failure)
	}
	bad := *work
	bad.ID = uuid.NewString()
	bad.TenantID = other
	bad.RequestKey = uuid.NewString()
	err = tx.Transaction(func(nested *gorm.DB) error { return nested.Omit("Workflow", "Events").Create(&bad).Error })
	if err == nil {
		t.Fatal("database accepted cross-tenant order/workflow references")
	}
	err = tx.Transaction(func(nested *gorm.DB) error {
		return nested.Model(&repository.WorkItem{}).Where("id = ?", work.ID).Update("current_step", 29).Error
	})
	if err == nil {
		t.Fatal("database accepted a step outside the pinned version")
	}
	err = tx.Transaction(func(nested *gorm.DB) error {
		duplicate := work.Events[0]
		duplicate.ID = uuid.NewString()
		return nested.Create(&duplicate).Error
	})
	if err == nil {
		t.Fatal("database accepted duplicate transition history version")
	}
}

func dtoWorkflowServiceInput() types.WorkflowInput {
	return types.WorkflowInput{Name: "Migrated workflow", Steps: []types.WorkflowStepInput{{Name: "Prepare", HandoffMode: "acceptance"}, {Name: "Review", HandoffMode: "direct"}}}
}
