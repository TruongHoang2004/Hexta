package repository

import (
	"context"
	apperrors "github.com/TruongHoang2004/Hexta/packages/shared/pkg/errors"
	"github.com/TruongHoang2004/Hexta/services/api/internal/testutil"
	"testing"
)

func TestOverviewSnapshotOptionsAndReadOnly(t *testing.T) {
	db := testutil.Database(t)
	r := NewOverviewRepository(db)
	ctx := context.Background()
	failure := r.WithinSnapshot(ctx, func(snapshot *OverviewRepository) *apperrors.Error {
		var isolation, readOnly string
		if err := snapshot.db.Raw("SHOW transaction_isolation").Scan(&isolation).Error; err != nil {
			t.Fatal(err)
		}
		if err := snapshot.db.Raw("SHOW transaction_read_only").Scan(&readOnly).Error; err != nil {
			t.Fatal(err)
		}
		if isolation != "repeatable read" || readOnly != "on" {
			t.Fatalf("incorrect transaction options %q/%q", isolation, readOnly)
		}
		return nil
	})
	if failure != nil {
		t.Fatal(failure)
	}
	failure = r.WithinSnapshot(ctx, func(snapshot *OverviewRepository) *apperrors.Error {
		if err := snapshot.db.Exec("UPDATE work_items SET version = version + 1 WHERE false").Error; err == nil {
			t.Fatal("snapshot allowed a mutation")
		}
		return overviewDBError(ctx)
	})
	if failure == nil || failure.GetHttpStatus() != 500 {
		t.Fatal("failed snapshot returned success")
	}
}
