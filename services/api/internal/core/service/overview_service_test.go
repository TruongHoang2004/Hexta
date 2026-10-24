package service

import (
	"context"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/TruongHoang2004/Hexta/services/api/internal/core/model"
	"github.com/TruongHoang2004/Hexta/services/api/internal/core/types"
	"github.com/TruongHoang2004/Hexta/services/api/internal/repository"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type overviewFixture struct {
	salesFixture
	overview *OverviewService
	orders   []model.Order
	work     []repository.WorkItem
}

func overviewTime(value string) time.Time {
	stamp, err := time.Parse(time.RFC3339Nano, value)
	if err != nil {
		panic(err)
	}
	return stamp
}
func overviewTestFixture(t *testing.T) overviewFixture {
	t.Helper()
	f := salesTestFixture(t)
	ctx := context.Background()
	wf := NewWorkflowService(repository.NewWorkflowRepository(f.db))
	d, err := wf.Publish(ctx, f.tenant, f.owner, "", uuid.NewString(), types.WorkflowInput{Name: "Demo flow", Steps: []types.WorkflowStepInput{{Name: "Review", HandoffMode: "acceptance"}, {Name: "Finish", HandoffMode: "direct"}}})
	if err != nil {
		t.Fatal(err)
	}
	result := overviewFixture{salesFixture: f, overview: NewOverviewService(repository.NewOverviewRepository(f.db))}
	dates := []string{"2026-10-18T10:00:00+07:00", "2026-10-19T00:00:00+07:00", "2026-10-20T09:00:00+07:00", "2026-10-21T09:00:00+07:00", "2026-10-22T09:00:00+07:00", "2026-10-26T00:00:00+07:00"}
	for i, date := range dates {
		responsible := &f.sale
		if i == 3 {
			responsible = &f.peer
		}
		if i == 4 {
			responsible = nil
		}
		status := model.OrderStatusConfirmed
		if i == 1 || i == 5 {
			status = model.OrderStatusDraft
		}
		row := model.Order{ID: uuid.NewString(), TenantID: f.tenant, OrderNumber: uuid.NewString(), EmployeeID: responsible, Status: status, CreatedAt: overviewTime(date)}
		if e := f.db.Create(&row).Error; e != nil {
			t.Fatal(e)
		}
		result.orders = append(result.orders, row)
	}
	for i, orderIndex := range []int{2, 2, 3, 3, 0} {
		responsible := &f.sale
		if i >= 3 {
			responsible = &f.peer
		}
		if i == 2 {
			responsible = nil
		}
		state := "active"
		var recipient *string
		step := 0
		if i == 1 {
			state = "waiting"
			recipient = &f.peer
		}
		if i == 3 {
			state = "completed"
			step = 1
		}
		row := repository.WorkItem{ID: uuid.NewString(), TenantID: f.tenant, OrderID: result.orders[orderIndex].ID, WorkflowVersionID: d.Versions[0].ID, CurrentStep: step, ResponsibleID: responsible, RecipientID: recipient, State: state, Version: 2, CreatedBy: f.owner, RequestKey: uuid.NewString(), CreateHash: strings.Repeat("a", 64), CreatedAt: overviewTime([]string{"2026-10-20T10:00:00+07:00", "2026-10-21T10:00:00+07:00", "2026-10-22T10:00:00+07:00", "2026-10-23T10:00:00+07:00", "2026-10-24T10:00:00+07:00"}[i])}
		if e := f.db.Omit("Workflow", "Events").Create(&row).Error; e != nil {
			t.Fatal(e)
		}
		result.work = append(result.work, row)
	}
	_, err = wf.Publish(ctx, f.tenant, f.owner, d.ID, "", types.WorkflowInput{Name: "New flow", ExpectedVersion: 1, Steps: []types.WorkflowStepInput{{Name: "New review", HandoffMode: "direct"}}})
	if err != nil {
		t.Fatal(err)
	}
	return result
}
func overviewWeek() types.OverviewQuery {
	from, to := overviewTime("2026-10-19T00:00:00+07:00"), overviewTime("2026-10-26T00:00:00+07:00")
	return types.OverviewQuery{From: &from, To: &to, Assignment: "all", Page: 1, Size: 20}
}
func (f overviewFixture) read(t *testing.T, q types.OverviewQuery) *types.OverviewResult {
	t.Helper()
	r, e := f.overview.Read(context.Background(), f.tenant, f.owner, q)
	if e != nil {
		t.Fatal(e)
	}
	return r
}
func TestOverviewWorkedCohortsAndPagination(t *testing.T) {
	f := overviewTestFixture(t)
	q := overviewWeek()
	r := f.read(t, q)
	if r.Orders != (types.OverviewOrderCounts{Total: 4, Draft: 1, Confirmed: 3}) || r.Work != (types.OverviewWorkCounts{Total: 5, Active: 3, Waiting: 1, Completed: 1, Unassigned: 1}) {
		t.Fatalf("wrong full counts: %+v %+v", r.Orders, r.Work)
	}
	for i, row := range r.WorkItems.Items {
		if row.ID != f.work[4-i].ID || row.WorkflowName != "Demo flow" || row.WorkflowVersionNumber != 1 || (row.CurrentStep == 0 && row.CurrentStepName != "Review") || (row.CurrentStep == 1 && row.CurrentStepName != "Finish") {
			t.Fatalf("wrong pinned row: %+v", row)
		}
	}
	q.EmployeeID = &f.sale
	q.Size = 1
	r = f.read(t, q)
	if r.Orders.Total != 2 || r.Orders.Draft != 1 || r.Work.Total != 2 || r.Work.Waiting != 1 || r.WorkItems.Items[0].ID != f.work[1].ID {
		t.Fatal("employee A worked example mismatch")
	}
	q.Page = 2
	r = f.read(t, q)
	if r.Work.Total != 2 || r.WorkItems.Items[0].ID != f.work[0].ID {
		t.Fatal("page 2 mismatch")
	}
	q.Page = 3
	r = f.read(t, q)
	if r.WorkItems.Items == nil || len(r.WorkItems.Items) != 0 || r.Work.Total != 2 {
		t.Fatal("beyond last page lost counts/array")
	}
	q = overviewWeek()
	q.EmployeeID = &f.peer
	r = f.read(t, q)
	if r.Orders.Total != 1 || r.Work.Total != 2 || r.Work.Waiting != 0 || r.WorkItems.Items[0].ID != f.work[4].ID {
		t.Fatal("nominee incorrectly counted as responsible")
	}
	q.EmployeeID = nil
	q.Assignment = "unassigned"
	r = f.read(t, q)
	if r.Orders.Total != 1 || r.Work.Total != 1 || r.Work.Active != 1 || r.Work.Unassigned != 1 || r.WorkItems.Items[0].ID != f.work[2].ID {
		t.Fatal("unassigned subset mismatch")
	}
	q = overviewWeek()
	q.Assignment = "assigned"
	r = f.read(t, q)
	if r.Orders.Total != 3 || r.Work.Total != 4 {
		t.Fatal("assigned mismatch")
	}
	q = types.OverviewQuery{Page: 1, Size: 20}
	r = f.read(t, q)
	if r.Orders.Total != 6 || r.Orders.Draft != 2 || r.Work.Total != 5 {
		t.Fatal("all time mismatch")
	}
	// A deleted source order affects only O; work still projects its order number.
	if err := f.db.Delete(&f.orders[2]).Error; err != nil {
		t.Fatal(err)
	}
	r = f.read(t, overviewWeek())
	if r.Orders.Total != 3 || r.Work.Total != 5 || r.WorkItems.Items[3].OrderNumber != f.orders[2].OrderNumber {
		t.Fatal("deleted source changed work cohort")
	}
	// Tied times use UUID ascending after created_at descending.
	if err := f.db.Model(&repository.WorkItem{}).Where("id IN ?", []string{f.work[0].ID, f.work[1].ID}).Update("created_at", f.work[4].CreatedAt.Add(time.Hour)).Error; err != nil {
		t.Fatal(err)
	}
	r = f.read(t, overviewWeek())
	a, b := f.work[0].ID, f.work[1].ID
	if a > b {
		a, b = b, a
	}
	if r.WorkItems.Items[0].ID != a || r.WorkItems.Items[1].ID != b {
		t.Fatal("tie order unstable")
	}
}
func TestOverviewAuthorizationAndLegacyEmpty(t *testing.T) {
	f := overviewTestFixture(t)
	ctx := context.Background()
	q := overviewWeek()
	for _, test := range []struct {
		tenant, actor string
		status        int
	}{{f.tenant, "", 401}, {f.tenant, f.sale, 403}, {f.tenant, "outsider", 403}, {uuid.NewString(), f.owner, 404}, {"bad", f.owner, 400}} {
		_, err := f.overview.Read(ctx, test.tenant, test.actor, q)
		requireSalesStatus(t, err, test.status)
	}
	q.EmployeeID = &f.owner
	r := f.read(t, q)
	if r.Orders.Total != 0 || r.Work.Total != 0 || r.WorkItems.Items == nil {
		t.Fatal("empty current-owner filter")
	}
	for _, id := range []string{"foreign", "removed", "nonexistent"} {
		q.EmployeeID = &id
		_, err := f.overview.Read(ctx, f.tenant, f.owner, q)
		requireSalesStatus(t, err, 400)
	}
	q = overviewWeek()
	if err := f.db.Model(&model.Tenant{}).Where("id = ?", f.tenant).Update("status", "suspended").Error; err != nil {
		t.Fatal(err)
	}
	_, err := f.overview.Read(ctx, f.tenant, f.owner, q)
	requireSalesStatus(t, err, 403)
	if err := f.db.Model(&model.Tenant{}).Where("id = ?", f.tenant).Update("status", model.TenantStatusActive).Error; err != nil {
		t.Fatal(err)
	}
	if err := f.db.Model(&model.TenantMember{}).Where("tenant_id = ? AND user_id = ?", f.tenant, f.sale).Update("role", model.RoleOwner).Error; err != nil {
		t.Fatal(err)
	}
	_, err = f.overview.Read(ctx, f.tenant, f.sale, q)
	requireSalesStatus(t, err, 403)
	for _, status := range []model.OrderStatus{model.OrderStatusProcessing, model.OrderStatusFulfilled, model.OrderStatusCancelled, model.OrderStatusReturned, "legacy"} {
		row := model.Order{ID: uuid.NewString(), TenantID: f.tenant, OrderNumber: uuid.NewString(), Status: status}
		if e := f.db.Create(&row).Error; e != nil {
			t.Fatal(e)
		}
		if e := f.db.Model(&row).Update("created_at", nil).Error; e != nil {
			t.Fatal(e)
		}
	}
	r = f.read(t, types.OverviewQuery{Page: 1, Size: 20})
	if r.Orders.Total != 11 || r.Orders.Processing != 1 || r.Orders.Fulfilled != 1 || r.Orders.Cancelled != 1 || r.Orders.Returned != 1 || r.Orders.Other != 1 {
		t.Fatalf("legacy buckets: %+v", r.Orders)
	}
	if r = f.read(t, overviewWeek()); r.Orders.Total != 4 {
		t.Fatal("NULL creation dates in bounded cohort")
	}
	r, err = f.overview.Read(ctx, f.other, f.owner, overviewWeek())
	if err != nil || r.Orders.Total != 0 || r.WorkItems.Items == nil {
		t.Fatal("foreign tenant leaked records")
	}
	if e := f.db.Where("tenant_id = ? AND user_id = ?", f.tenant, f.owner).Delete(&model.TenantMember{}).Error; e != nil {
		t.Fatal(e)
	}
	_, err = f.overview.Read(ctx, f.tenant, f.owner, q)
	requireSalesStatus(t, err, 403)
}
func TestOverviewConsistentReadSnapshot(t *testing.T) {
	f := overviewTestFixture(t)
	paused, resume := make(chan struct{}), make(chan struct{})
	var once sync.Once
	name := "overview_snapshot_test"
	if err := f.db.Callback().Row().After("gorm:row").Register(name, func(db *gorm.DB) {
		if strings.Contains(db.Statement.SQL.String(), "count(*) AS total") && strings.Contains(db.Statement.SQL.String(), "work_items") {
			once.Do(func() { close(paused); <-resume })
		}
	}); err != nil {
		t.Fatal(err)
	}
	defer f.db.Callback().Row().Remove(name)
	type outcome struct {
		result *types.OverviewResult
		err    error
	}
	done := make(chan outcome, 1)
	go func() {
		r, e := f.overview.Read(context.Background(), f.tenant, f.owner, overviewWeek())
		if e != nil {
			done <- outcome{err: e}
			return
		}
		done <- outcome{result: r}
	}()
	select {
	case <-paused:
	case <-time.After(10 * time.Second):
		close(resume)
		t.Fatal("aggregate callback not reached")
	}
	err := f.db.Model(&repository.WorkItem{}).Where("id = ?", f.work[0].ID).Updates(map[string]any{"state": "completed", "current_step": 1, "responsible_id": f.peer}).Error
	close(resume)
	if err != nil {
		t.Fatal(err)
	}
	var got outcome
	select {
	case got = <-done:
	case <-time.After(10 * time.Second):
		t.Fatal("snapshot did not finish")
	}
	if got.err != nil {
		t.Fatal(got.err)
	}
	if got.result.Work.Active != 3 || got.result.Work.Completed != 1 || got.result.WorkItems.Items[4].State != "active" || *got.result.WorkItems.Items[4].ResponsibleID != f.sale {
		t.Fatal("aggregate/rows observed different snapshots")
	}
	fresh := f.read(t, overviewWeek())
	if fresh.Work.Active != 2 || fresh.Work.Completed != 2 || fresh.WorkItems.Items[4].State != "completed" || *fresh.WorkItems.Items[4].ResponsibleID != f.peer {
		t.Fatal("later request missed committed change")
	}
}
