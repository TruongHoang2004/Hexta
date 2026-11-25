package repository

import (
	"context"
	"testing"

	apperrors "github.com/TruongHoang2004/Hexta/packages/shared/pkg/errors"
	"github.com/TruongHoang2004/Hexta/services/api/internal/core/model"
	"github.com/TruongHoang2004/Hexta/services/api/internal/testutil"
	"github.com/google/uuid"
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
