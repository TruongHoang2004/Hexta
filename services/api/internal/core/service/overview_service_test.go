package service

import (
	"context"
	"testing"
	"time"

	apperrors "github.com/TruongHoang2004/Hexta/packages/shared/pkg/errors"
	"github.com/TruongHoang2004/Hexta/services/api/internal/core/model"
	"github.com/TruongHoang2004/Hexta/services/api/internal/core/types"
	"github.com/TruongHoang2004/Hexta/services/api/internal/repository"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// fakeOverviewStore records whether any business data was read so denial paths can prove no leak.
type fakeOverviewStore struct {
	tenant   *model.Tenant
	members  map[string]bool
	dataRead bool
	filter   types.OverviewFilter
	totals   []types.OverviewOrderStatusTotal
}

func (f *fakeOverviewStore) WithinSnapshot(_ context.Context, fn func(repository.OverviewStore) *apperrors.Error) *apperrors.Error {
	return fn(f)
}
func (f *fakeOverviewStore) Tenant(ctx context.Context, id string) (*model.Tenant, *apperrors.Error) {
	if f.tenant == nil || f.tenant.ID != id {
		return nil, apperrors.ErrNotFound(ctx, "tenant", "not found")
	}
	return f.tenant, nil
}
func (f *fakeOverviewStore) IsMember(_ context.Context, _, user string) (bool, *apperrors.Error) {
	return f.members[user], nil
}
func (f *fakeOverviewStore) HasActivity(context.Context, string) (bool, *apperrors.Error) {
	f.dataRead = true
	return len(f.totals) > 0, nil
}
func (f *fakeOverviewStore) OrderStatusTotals(_ context.Context, _ string, filter types.OverviewFilter) ([]types.OverviewOrderStatusTotal, *apperrors.Error) {
	f.dataRead, f.filter = true, filter
	return f.totals, nil
}
func (f *fakeOverviewStore) WorkCounts(context.Context, string, types.OverviewFilter, int) (types.OverviewWork, *apperrors.Error) {
	f.dataRead = true
	return types.OverviewWork{BySteps: []types.OverviewStepCount{}}, nil
}
func (f *fakeOverviewStore) OpenWork(context.Context, string, types.OverviewFilter, time.Time, int) ([]types.OverviewWorkRow, int64, *apperrors.Error) {
	f.dataRead = true
	return []types.OverviewWorkRow{}, 0, nil
}

func newFakeOverview() (*fakeOverviewStore, *OverviewService) {
	store := &fakeOverviewStore{tenant: &model.Tenant{ID: uuid.NewString(), OwnerID: "owner", Status: model.TenantStatusActive}, members: map[string]bool{"owner": true, "sale": true}}
	return store, NewOverviewService(store)
}

func TestOverviewAuthorizationDeniesBeforeReadingData(t *testing.T) {
	ctx := context.Background()
	for _, tc := range []struct {
		name, actor string
		suspend     bool
		status      int
	}{
		{"missing actor", "", false, 401},
		{"member", "sale", false, 403},
		{"non-member", "stranger", false, 403},
		{"suspended tenant owner", "owner", true, 403},
	} {
		t.Run(tc.name, func(t *testing.T) {
			store, s := newFakeOverview()
			if tc.suspend {
				store.tenant.Status = model.TenantStatusSuspended
			}
			_, err := s.Overview(ctx, store.tenant.ID, tc.actor, types.OverviewFilter{})
			requireSalesStatus(t, err, tc.status)
			if store.dataRead {
				t.Fatal("denied caller must not trigger business data reads")
			}
		})
	}
	store, s := newFakeOverview()
	// A member holding the "owner" role label is still not the tenant owner identity.
	store.members["co-owner"] = true
	_, err := s.Overview(ctx, store.tenant.ID, "co-owner", types.OverviewFilter{})
	requireSalesStatus(t, err, 403)
	_, err = s.Overview(ctx, uuid.NewString(), "owner", types.OverviewFilter{})
	requireSalesStatus(t, err, 404)
	_, err = s.Overview(ctx, "not-a-uuid", "owner", types.OverviewFilter{})
	requireSalesStatus(t, err, 400)
}

func TestOverviewFilterValidation(t *testing.T) {
	ctx := context.Background()
	store, s := newFakeOverview()
	from := time.Date(2026, 11, 1, 0, 0, 0, 0, time.FixedZone("ICT", 7*3600))
	to := from.Add(24 * time.Hour)
	for name, filter := range map[string]types.OverviewFilter{
		"only from":    {From: &from},
		"only to":      {To: &to},
		"equal bounds": {From: &from, To: &from},
		"inverted":     {From: &to, To: &from},
	} {
		if _, err := s.Overview(ctx, store.tenant.ID, "owner", filter); err == nil || err.GetHttpStatus() != 400 {
			t.Fatalf("%s: expected 400, got %v", name, err)
		}
	}
	nobody := "nobody"
	_, err := s.Overview(ctx, store.tenant.ID, "owner", types.OverviewFilter{EmployeeID: &nobody})
	requireSalesStatus(t, err, 400)
	if store.dataRead {
		t.Fatal("an invalid employee filter must be rejected before data reads")
	}
	result, err := s.Overview(ctx, store.tenant.ID, "owner", types.OverviewFilter{From: &from, To: &to})
	if err != nil {
		t.Fatal(err)
	}
	if store.filter.From.Location() != time.UTC || !store.filter.From.Equal(from) || !result.Filter.To.Equal(to) {
		t.Fatal("bounds must be normalized to UTC without changing the instant")
	}
}

func TestOverviewOrderBucketsAreStableAndExtensible(t *testing.T) {
	result := overviewOrders([]types.OverviewOrderStatusTotal{
		{Status: "confirmed", Count: 2, Value: decimal.RequireFromString("200.50")},
		{Status: "draft", Count: 1, Value: decimal.RequireFromString("99")},
		{Status: "processing", Count: 1, Value: decimal.RequireFromString("10")},
		{Status: "cancelled", Count: 1, Value: decimal.RequireFromString("70")},
		{Status: "exception_open", Count: 3, Value: decimal.RequireFromString("5")},
	})
	if result.Total != 8 || !result.OrderValue.Equal(decimal.RequireFromString("210.50")) {
		t.Fatalf("unexpected totals: %d %s", result.Total, result.OrderValue)
	}
	want := append(append([]string{}, overviewOrderStatuses...), OverviewOrderStatusOther)
	if len(result.ByStatus) != len(want) {
		t.Fatalf("expected %d buckets, got %d", len(want), len(result.ByStatus))
	}
	for i, status := range want {
		if result.ByStatus[i].Status != status {
			t.Fatalf("bucket %d is %s, want %s", i, result.ByStatus[i].Status, status)
		}
	}
	if other := result.ByStatus[len(want)-1]; other.Count != 3 || !other.Value.Equal(decimal.NewFromInt(5)) {
		t.Fatal("unknown statuses must be counted in other, never dropped")
	}
	empty := overviewOrders(nil)
	if empty.Total != 0 || !empty.OrderValue.IsZero() || len(empty.ByStatus) != len(want) {
		t.Fatal("empty cohorts must still report every bucket with zero")
	}
}

func TestOverviewStuckReasons(t *testing.T) {
	sale := "sale"
	result := overviewStuck([]types.OverviewWorkRow{{WorkItemID: "a"}, {WorkItemID: "b", Responsible: &sale}}, 7)
	if result.Total != 7 || result.ThresholdHours != 48 || result.Items[0].Reason != OverviewReasonUnassigned || result.Items[1].Reason != OverviewReasonWaitingTooLong {
		t.Fatalf("unexpected stuck result: %+v", result)
	}
}

// overviewDBFixture seeds two tenants with fixed timestamps relative to base; see TestOverviewDatabase.
type overviewDBFixture struct {
	salesFixture
	overview *OverviewService
	empty    string
	base     time.Time
	now      time.Time
}

func overviewDBTestFixture(t *testing.T) overviewDBFixture {
	t.Helper()
	f := salesTestFixture(t)
	ctx := context.Background()
	base := time.Date(2026, 11, 10, 0, 0, 0, 0, time.UTC)
	o := overviewDBFixture{salesFixture: f, overview: NewOverviewService(repository.NewOverviewRepository(f.db)), empty: uuid.NewString(), base: base, now: base.Add(10 * 24 * time.Hour)}
	o.overview.now = func() time.Time { return o.now }
	if err := f.db.Create(&model.Tenant{ID: o.empty, Name: o.empty, Slug: o.empty, OwnerID: f.owner, Status: model.TenantStatusActive}).Error; err != nil {
		t.Fatal(err)
	}
	if err := f.db.Create(&model.TenantMember{TenantID: o.empty, UserID: f.owner, Role: model.RoleOwner}).Error; err != nil {
		t.Fatal(err)
	}
	if err := f.db.Create(&model.TenantMember{TenantID: f.other, UserID: "outsider", Role: model.RoleMember}).Error; err != nil {
		t.Fatal(err)
	}
	day := 24 * time.Hour
	orders := map[string]string{}
	order := func(tenant, key, status string, employee *string, created time.Time, total int64, deleted bool) {
		id := uuid.NewString()
		orders[key] = id
		if err := f.db.Create(&model.Order{ID: id, TenantID: tenant, OrderNumber: key, EmployeeID: employee, Status: model.OrderStatus(status), TotalAmount: decimal.NewFromInt(total)}).Error; err != nil {
			t.Fatal(err)
		}
		var deletedAt *time.Time
		if deleted {
			deletedAt = &created
		}
		if err := f.db.Exec("UPDATE orders SET created_at = ?, deleted_at = ? WHERE id = ?", created, deletedAt, id).Error; err != nil {
			t.Fatal(err)
		}
	}
	sale, peer, owner := f.sale, f.peer, f.owner
	order(f.tenant, "O1", "draft", &sale, base.Add(1*day), 100, false)
	order(f.tenant, "O2", "confirmed", &sale, base.Add(2*day), 200, false)
	order(f.tenant, "O3", "confirmed", &peer, base.Add(3*day), 300, false)
	order(f.tenant, "O4", "cancelled", &sale, base.Add(4*day), 50, false)
	order(f.tenant, "O5", "legacy_unknown", nil, base.Add(5*day), 10, false)
	order(f.tenant, "O6", "confirmed", &sale, base.Add(-5*day), 1000, false)
	order(f.tenant, "O7", "confirmed", &sale, base.Add(2*day), 7000, true)
	order(f.other, "B1", "confirmed", &owner, base.Add(2*day), 999, false)

	workflows := NewWorkflowService(repository.NewWorkflowRepository(f.db))
	definition, err := workflows.Publish(ctx, f.tenant, f.owner, "", uuid.NewString(), types.WorkflowInput{Name: "Delivery", Steps: []types.WorkflowStepInput{{Name: "Prepare", HandoffMode: "acceptance"}, {Name: "Pack", HandoffMode: "direct"}, {Name: "Ship", HandoffMode: "direct"}}})
	if err != nil {
		t.Fatal(err)
	}
	version := definition.Versions[0].ID
	work := func(key, orderKey string, step int, state string, responsible, recipient *string, created, updated time.Time) {
		row := repository.WorkItem{ID: uuid.NewString(), TenantID: f.tenant, OrderID: orders[orderKey], WorkflowVersionID: version, CurrentStep: step, ResponsibleID: responsible, RecipientID: recipient, State: state, Version: 1, CreatedBy: f.owner, RequestKey: uuid.NewString(), CreateHash: key}
		if err := f.db.Omit("Workflow", "Events").Create(&row).Error; err != nil {
			t.Fatal(err)
		}
		if err := f.db.Exec("UPDATE work_items SET created_at = ?, updated_at = ? WHERE id = ?", created, updated, row.ID).Error; err != nil {
			t.Fatal(err)
		}
	}
	hour := time.Hour
	work("W1", "O2", 0, "active", &sale, nil, base.Add(2*day), o.now.Add(-1*hour))
	work("W2", "O2", 0, "waiting", &sale, &peer, base.Add(3*day), o.now.Add(-72*hour))
	work("W3", "O3", 1, "active", nil, nil, base.Add(3*day), o.now.Add(-1*hour))
	work("W4", "O3", 2, "completed", &peer, nil, base.Add(4*day), o.now.Add(-100*hour))
	work("W5", "O6", 0, "active", &peer, nil, base.Add(-4*day), o.now.Add(-200*hour))
	return o
}

func overviewBucket(t *testing.T, result *types.OverviewResult, status string) types.OverviewOrderBucket {
	t.Helper()
	for _, bucket := range result.Orders.ByStatus {
		if bucket.Status == status {
			return bucket
		}
	}
	t.Fatalf("missing bucket %s", status)
	return types.OverviewOrderBucket{}
}

func overviewStuckIDs(result *types.OverviewResult) []string {
	keys := []string{}
	for _, item := range result.Stuck.Items {
		keys = append(keys, item.OrderNumber+":"+item.StepName+":"+item.Reason)
	}
	return keys
}

func TestOverviewDatabase(t *testing.T) {
	o := overviewDBTestFixture(t)
	ctx := context.Background()
	from, to := o.base, o.base.Add(7*24*time.Hour)
	read := func(tenant, actor string, filter types.OverviewFilter) *types.OverviewResult {
		t.Helper()
		result, err := o.overview.Overview(ctx, tenant, actor, filter)
		if err != nil {
			t.Fatal(err)
		}
		return result
	}

	t.Run("period", func(t *testing.T) {
		r := read(o.tenant, o.owner, types.OverviewFilter{From: &from, To: &to})
		if !r.HasActivity || r.Orders.Total != 5 || !r.Orders.OrderValue.Equal(decimal.NewFromInt(500)) {
			t.Fatalf("orders: total %d value %s", r.Orders.Total, r.Orders.OrderValue)
		}
		if overviewBucket(t, r, "draft").Count != 1 || overviewBucket(t, r, "confirmed").Count != 2 || overviewBucket(t, r, "cancelled").Count != 1 || overviewBucket(t, r, OverviewOrderStatusOther).Count != 1 {
			t.Fatal("unexpected order buckets; soft-deleted and out-of-period orders must be excluded")
		}
		w := r.Work
		if w.Total != 4 || w.Active != 2 || w.Waiting != 1 || w.Completed != 1 || w.Unassigned != 1 || len(w.BySteps) != 3 {
			t.Fatalf("unexpected work counts: %+v", w)
		}
		if s := w.BySteps[0]; s.StepName != "Prepare" || s.StepOrdinal != 0 || s.Active != 1 || s.Waiting != 1 || s.WorkflowName != "Delivery" {
			t.Fatalf("unexpected first step: %+v", s)
		}
		if s := w.BySteps[1]; s.StepName != "Pack" || s.Unassigned != 1 || s.Active != 1 {
			t.Fatalf("unexpected second step: %+v", s)
		}
		if s := w.BySteps[2]; s.StepName != "Ship" || s.Completed != 1 {
			t.Fatalf("unexpected third step: %+v", s)
		}
		got := overviewStuckIDs(r)
		if r.Stuck.Total != 2 || len(got) != 2 || got[0] != "O2:Prepare:waiting_too_long" || got[1] != "O3:Pack:unassigned" {
			t.Fatalf("unexpected stuck rows: %v (total %d)", got, r.Stuck.Total)
		}
		if r.Stuck.Items[0].WaitingSince.Location() != time.UTC || !r.Stuck.Items[0].WaitingSince.Equal(o.now.Add(-72*time.Hour)) {
			t.Fatal("waiting_since must be the last change instant in UTC")
		}
	})

	t.Run("employee", func(t *testing.T) {
		sale := o.sale
		r := read(o.tenant, o.owner, types.OverviewFilter{From: &from, To: &to, EmployeeID: &sale})
		if r.Orders.Total != 3 || !r.Orders.OrderValue.Equal(decimal.NewFromInt(200)) || r.Work.Total != 2 || r.Work.Unassigned != 0 {
			t.Fatalf("employee filter: orders %d value %s work %d", r.Orders.Total, r.Orders.OrderValue, r.Work.Total)
		}
		if got := overviewStuckIDs(r); len(got) != 1 || got[0] != "O2:Prepare:waiting_too_long" {
			t.Fatalf("employee stuck rows follow current responsibility, not recipient: %v", got)
		}
		peer := o.peer
		r = read(o.tenant, o.owner, types.OverviewFilter{From: &from, To: &to, EmployeeID: &peer})
		if r.Work.Total != 1 || r.Work.Completed != 1 || r.Stuck.Total != 0 {
			t.Fatal("a nominated recipient must not be credited with waiting work")
		}
	})

	t.Run("all time", func(t *testing.T) {
		r := read(o.tenant, o.owner, types.OverviewFilter{})
		if r.Orders.Total != 6 || !r.Orders.OrderValue.Equal(decimal.NewFromInt(1500)) || r.Work.Total != 5 {
			t.Fatalf("all time: orders %d value %s work %d", r.Orders.Total, r.Orders.OrderValue, r.Work.Total)
		}
		if got := overviewStuckIDs(r); len(got) != 3 || got[0] != "O6:Prepare:waiting_too_long" {
			t.Fatalf("all time stuck rows, longest waiting first: %v", got)
		}
	})

	t.Run("no data in period", func(t *testing.T) {
		later, end := o.now, o.now.Add(24*time.Hour)
		r := read(o.tenant, o.owner, types.OverviewFilter{From: &later, To: &end})
		if !r.HasActivity || r.Orders.Total != 0 || r.Work.Total != 0 || len(r.Work.BySteps) != 0 || len(r.Stuck.Items) != 0 || len(r.Orders.ByStatus) != len(overviewOrderStatuses)+1 {
			t.Fatalf("empty period must keep zero buckets and empty lists: %+v", r)
		}
	})

	t.Run("tenant isolation and empty tenant", func(t *testing.T) {
		r := read(o.other, o.owner, types.OverviewFilter{})
		if !r.HasActivity || r.Orders.Total != 1 || !r.Orders.OrderValue.Equal(decimal.NewFromInt(999)) || r.Work.Total != 0 || r.Stuck.Total != 0 {
			t.Fatalf("other tenant must see only its own rows: %+v", r.Orders)
		}
		r = read(o.empty, o.owner, types.OverviewFilter{})
		if r.HasActivity || r.Orders.Total != 0 || r.Work.Total != 0 || r.Stuck.Items == nil || r.Work.BySteps == nil {
			t.Fatal("an empty tenant reports no activity with empty, non-nil lists")
		}
		outsider := "outsider"
		_, err := o.overview.Overview(ctx, o.tenant, o.owner, types.OverviewFilter{EmployeeID: &outsider})
		requireSalesStatus(t, err, 400)
	})

	t.Run("authorization", func(t *testing.T) {
		for _, actor := range []string{o.sale, o.peer, "stranger"} {
			_, err := o.overview.Overview(ctx, o.tenant, actor, types.OverviewFilter{})
			requireSalesStatus(t, err, 403)
		}
		if err := o.db.Model(&model.Tenant{}).Where("id = ?", o.other).Update("status", model.TenantStatusSuspended).Error; err != nil {
			t.Fatal(err)
		}
		_, err := o.overview.Overview(ctx, o.other, o.owner, types.OverviewFilter{})
		requireSalesStatus(t, err, 403)
	})
}
