package repository

import (
	"context"
	"strconv"
	"testing"
	"time"

	apperrors "github.com/TruongHoang2004/Hexta/packages/shared/pkg/errors"
	"github.com/TruongHoang2004/Hexta/services/api/internal/core/model"
	"github.com/TruongHoang2004/Hexta/services/api/internal/core/types"
	"github.com/TruongHoang2004/Hexta/services/api/internal/testutil"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestOverviewSnapshotIsReadOnlyRepeatableRead(t *testing.T) {
	db := testutil.Database(t)
	ctx := context.Background()
	tenant := uuid.NewString()
	if err := db.Create(&model.Tenant{ID: tenant, Name: tenant, Slug: tenant, OwnerID: "owner", Status: model.TenantStatusActive}).Error; err != nil {
		t.Fatal(err)
	}
	var isolation, readOnly string
	var writeErr error
	err := NewOverviewRepository(db).WithinSnapshot(ctx, func(store OverviewStore) *apperrors.Error {
		r := store.(*OverviewRepository)
		r.db.Raw("SHOW transaction_isolation").Scan(&isolation)
		r.db.Raw("SHOW transaction_read_only").Scan(&readOnly)
		if _, err := r.Tenant(ctx, tenant); err != nil {
			return err
		}
		// A write must be refused by PostgreSQL inside the overview snapshot.
		writeErr = r.db.Exec("UPDATE tenants SET name = 'changed' WHERE id = ?", tenant).Error
		return nil
	})
	if isolation != "repeatable read" || readOnly != "on" {
		t.Fatalf("overview snapshot isolation %q read_only %q", isolation, readOnly)
	}
	if writeErr == nil {
		t.Fatal("overview snapshot must reject writes")
	}
	// The aborted transaction may surface at commit; if it does, only a generic system error is public.
	if err != nil && err.GetHttpStatus() != 500 {
		t.Fatalf("expected a safe system error after an aborted snapshot, got %v", err)
	}
	var row model.Tenant
	if db.Where("id = ?", tenant).First(&row).Error != nil || row.Name != tenant {
		t.Fatal("overview snapshot must not change data")
	}
	if _, err := NewOverviewRepository(db).Tenant(ctx, uuid.NewString()); err == nil || err.GetHttpStatus() != 404 {
		t.Fatal("missing tenant must be not found")
	}
}

// Cases are counted through the order cohort: tenant, period, sale and soft deletion follow the order.
func TestOverviewCaseTotalsAndPendingCasesFollowTheOrderCohort(t *testing.T) {
	db := testutil.Database(t)
	ctx := context.Background()
	base := time.Date(2026, 11, 10, 0, 0, 0, 0, time.UTC)
	for _, statement := range []string{
		`INSERT INTO orders(id,tenant_id,order_number,status,employee_id,total_amount,created_at) VALUES
 ('o1','alpha','A-1','confirmed','sale',100,'2026-11-11T00:00:00Z'),
 ('o2','alpha','A-2','confirmed','peer',200,'2026-11-12T00:00:00Z'),
 ('o3','alpha','A-3','fulfilled','sale',300,'2026-11-13T00:00:00Z'),
 ('o4','alpha','A-4','confirmed','sale',400,'2026-11-14T00:00:00Z'),
 ('o5','alpha','A-5','confirmed','sale',500,'2026-11-01T00:00:00Z'),
 ('o6','alpha','A-6','confirmed','sale',600,'2026-11-15T00:00:00Z'),
 ('b1','beta','B-1','confirmed','sale',700,'2026-11-11T00:00:00Z')`,
		`UPDATE orders SET deleted_at = '2026-11-16T00:00:00Z' WHERE id = 'o6'`,
		`INSERT INTO order_cases(id,tenant_id,order_id,case_type,status,handler_id,opened_by,finalized_by,finalized_at,version,updated_at) VALUES
 ('c1','alpha','o1','return','open','sale','sale',NULL,NULL,1,'2026-11-20T03:00:00Z'),
 ('c2','alpha','o2','cancellation','open','peer','peer',NULL,NULL,1,'2026-11-20T01:00:00Z'),
 ('c3','alpha','o3','cancellation','finalized','sale','sale','owner','2026-11-20T00:00:00Z',3,'2026-11-20T00:00:00Z'),
 ('c4','alpha','o4','return','sent_back','sale','sale',NULL,NULL,2,'2026-11-19T00:00:00Z'),
 ('c5','alpha','o5','return','open','sale','sale',NULL,NULL,1,'2026-11-18T00:00:00Z'),
 ('c6','alpha','o6','return','open','sale','sale',NULL,NULL,1,'2026-11-17T00:00:00Z'),
 ('d1','beta','b1','return','open','sale','sale',NULL,NULL,1,'2026-11-17T00:00:00Z')`,
	} {
		require.NoError(t, db.Exec(statement).Error)
	}
	repo := NewOverviewRepository(db)
	from, to := base, base.Add(7*24*time.Hour)
	period := types.OverviewFilter{From: &from, To: &to}

	totals, err := repo.CaseTotals(ctx, "alpha", period)
	require.Nil(t, err)
	got := map[string]string{}
	for _, row := range totals {
		got[row.OrderStatus+"/"+row.CaseType+"/"+row.CaseStatus] = strconv.FormatInt(row.Count, 10) + ":" + row.Value.String()
	}
	// o5 is outside the period, o6 is soft-deleted and b1 belongs to another tenant.
	require.Equal(t, map[string]string{
		"confirmed/return/open":            "1:100",
		"confirmed/cancellation/open":      "1:200",
		"fulfilled/cancellation/finalized": "1:300",
		"confirmed/return/sent_back":       "1:400",
	}, got)

	sale := "sale"
	totals, err = repo.CaseTotals(ctx, "alpha", types.OverviewFilter{From: &from, To: &to, EmployeeID: &sale})
	require.Nil(t, err)
	require.Len(t, totals, 3, "the employee filter follows the order's sale")

	pending, total, err := repo.PendingCases(ctx, "alpha", period, 1)
	require.Nil(t, err)
	require.Equal(t, int64(2), total, "only open cases wait for the owner; sent back and finalized do not")
	require.Len(t, pending, 1, "rows are capped while the total stays complete")
	require.Equal(t, types.OverviewPendingCase{CaseID: "c2", OrderID: "o2", OrderNumber: "A-2", CaseType: "cancellation", HandlerID: "peer", WaitingSince: time.Date(2026, 11, 20, 1, 0, 0, 0, time.UTC)}, pending[0])

	pending, total, err = repo.PendingCases(ctx, "alpha", types.OverviewFilter{}, 10)
	require.Nil(t, err)
	require.Equal(t, int64(3), total)
	require.Equal(t, []string{"c5", "c2", "c1"}, []string{pending[0].CaseID, pending[1].CaseID, pending[2].CaseID}, "longest waiting first")

	pending, total, err = repo.PendingCases(ctx, "gamma", types.OverviewFilter{}, 10)
	require.Nil(t, err)
	require.Zero(t, total)
	require.NotNil(t, pending)
	require.Empty(t, pending)
}
