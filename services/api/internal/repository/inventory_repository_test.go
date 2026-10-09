package repository

import (
	"context"
	"testing"

	"github.com/TruongHoang2004/Hexta/services/api/internal/core/model"
	"github.com/google/uuid"
)

func TestInventoryRepository_CRUD_And_Reservation(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}

	tx := db.Begin()
	defer tx.Rollback()

	repo := NewInventoryDBRepository(tx)
	ctx := context.Background()

	tenantID := uuid.New().String()
	variantID := uuid.New().String()

	// 1. Create inventory item
	item := &model.InventoryItem{
		ID:              uuid.New().String(),
		TenantID:        tenantID,
		VariantID:       variantID,
		OnHandQty:       20,
		ReservedQty:     0,
		AvailableQty:    20,
		SafetyThreshold: 5,
	}

	created, err := repo.CreateInventoryItem(ctx, item)
	if err != nil {
		t.Fatalf("failed to create inventory item: %v", err)
	}
	if created == nil || created.OnHandQty != 20 {
		t.Fatalf("expected on-hand 20, got %+v", created)
	}

	// 2. Query by variant ID
	queried, err := repo.GetByVariantID(ctx, tenantID, variantID)
	if err != nil {
		t.Fatalf("failed to get inventory by variant: %v", err)
	}
	if queried.AvailableQty != 20 {
		t.Errorf("expected available 20, got %d", queried.AvailableQty)
	}

	// 3. Atomic reservation
	reserved, err := repo.ReserveStock(ctx, tenantID, variantID, 5)
	if err != nil {
		t.Fatalf("failed to reserve stock: %v", err)
	}
	if reserved.ReservedQty != 5 || reserved.AvailableQty != 15 {
		t.Errorf("expected reserved=5, available=15, got reserved=%d, available=%d",
			reserved.ReservedQty, reserved.AvailableQty)
	}

	// 4. Over-reservation should fail
	_, err = repo.ReserveStock(ctx, tenantID, variantID, 16)
	if err == nil {
		t.Fatalf("expected error reserving 16 with only 15 available")
	}

	// 5. Release 2 units
	released, err := repo.ReleaseStock(ctx, tenantID, variantID, 2)
	if err != nil {
		t.Fatalf("failed to release stock: %v", err)
	}
	if released.ReservedQty != 3 || released.AvailableQty != 17 {
		t.Errorf("expected reserved=3, available=17, got reserved=%d, available=%d",
			released.ReservedQty, released.AvailableQty)
	}

	// 6. Deduct 3 units (fulfill order)
	refType := "ORDER"
	refID := uuid.New().String()
	deducted, err := repo.DeductStock(ctx, tenantID, variantID, 3, refType, &refID)
	if err != nil {
		t.Fatalf("failed to deduct stock: %v", err)
	}
	if deducted.OnHandQty != 17 || deducted.ReservedQty != 0 || deducted.AvailableQty != 17 {
		t.Errorf("expected on_hand=17, reserved=0, available=17, got on_hand=%d, reserved=%d, available=%d",
			deducted.OnHandQty, deducted.ReservedQty, deducted.AvailableQty)
	}

	// 7. Verify movement recorded
	movements, total, err := repo.ListMovements(ctx, tenantID, created.ID, 10, 0)
	if err != nil || total != 1 || len(movements) != 1 {
		t.Fatalf("expected 1 movement logged, got %d, err=%v", total, err)
	}
	if movements[0].MovementType != model.MovementOutbound || movements[0].Quantity != 3 {
		t.Errorf("expected outbound movement qty 3, got %s qty %d", movements[0].MovementType, movements[0].Quantity)
	}
}
