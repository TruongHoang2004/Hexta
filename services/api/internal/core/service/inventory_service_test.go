package service

import (
	"context"
	"testing"

	"github.com/TruongHoang2004/Hexta/packages/shared/pkg/errors"
	"github.com/TruongHoang2004/Hexta/services/api/internal/core/model"
	"github.com/google/uuid"
)

type mockInventoryRepo struct {
	items     map[string]*model.InventoryItem
	movements []*model.StockMovement
}

func newMockInventoryRepo() *mockInventoryRepo {
	return &mockInventoryRepo{
		items:     make(map[string]*model.InventoryItem),
		movements: make([]*model.StockMovement, 0),
	}
}

func (m *mockInventoryRepo) key(tenantID, variantID string) string {
	return tenantID + ":" + variantID
}

func (m *mockInventoryRepo) GetByVariantID(ctx context.Context, tenantID, variantID string) (*model.InventoryItem, *errors.Error) {
	if item, ok := m.items[m.key(tenantID, variantID)]; ok {
		return item, nil
	}
	return nil, errors.ErrNotFound(ctx, "inventory item", "not found")
}

func (m *mockInventoryRepo) GetByID(ctx context.Context, tenantID, id string) (*model.InventoryItem, *errors.Error) {
	for _, item := range m.items {
		if item.TenantID == tenantID && item.ID == id {
			return item, nil
		}
	}
	return nil, errors.ErrNotFound(ctx, "inventory item", "not found")
}

func (m *mockInventoryRepo) CreateInventoryItem(ctx context.Context, item *model.InventoryItem) (*model.InventoryItem, *errors.Error) {
	if item.ID == "" {
		item.ID = uuid.New().String()
	}
	m.items[m.key(item.TenantID, item.VariantID)] = item
	return item, nil
}

func (m *mockInventoryRepo) ReserveStock(ctx context.Context, tenantID, variantID string, quantity int) (*model.InventoryItem, *errors.Error) {
	item, ok := m.items[m.key(tenantID, variantID)]
	if !ok {
		return nil, errors.ErrNotFound(ctx, "inventory item", "not found")
	}
	if item.AvailableQty < quantity {
		return nil, errors.ErrConflict(ctx, "inventory stock", "insufficient available quantity to reserve")
	}
	item.ReservedQty += quantity
	item.AvailableQty -= quantity
	return item, nil
}

func (m *mockInventoryRepo) ReleaseStock(ctx context.Context, tenantID, variantID string, quantity int) (*model.InventoryItem, *errors.Error) {
	item, ok := m.items[m.key(tenantID, variantID)]
	if !ok {
		return nil, errors.ErrNotFound(ctx, "inventory item", "not found")
	}
	if item.ReservedQty < quantity {
		return nil, errors.ErrBadRequest(ctx).SetDetail("release quantity exceeds currently reserved quantity")
	}
	item.ReservedQty -= quantity
	item.AvailableQty += quantity
	return item, nil
}

func (m *mockInventoryRepo) DeductStock(ctx context.Context, tenantID, variantID string, quantity int, refType string, refID *string) (*model.InventoryItem, *errors.Error) {
	item, ok := m.items[m.key(tenantID, variantID)]
	if !ok {
		return nil, errors.ErrNotFound(ctx, "inventory item", "not found")
	}
	if item.ReservedQty < quantity || item.OnHandQty < quantity {
		return nil, errors.ErrBadRequest(ctx).SetDetail("deduction quantity exceeds reserved or on-hand stock")
	}
	item.OnHandQty -= quantity
	item.ReservedQty -= quantity
	m.movements = append(m.movements, &model.StockMovement{
		ID:            uuid.New().String(),
		TenantID:      tenantID,
		InventoryID:   item.ID,
		MovementType:  model.MovementOutbound,
		Quantity:      quantity,
		BalanceAfter:  item.OnHandQty,
		ReferenceType: refType,
		ReferenceID:   refID,
	})
	return item, nil
}

func (m *mockInventoryRepo) AdjustStock(ctx context.Context, tenantID, variantID string, newOnHand int, notes string) (*model.InventoryItem, *errors.Error) {
	item, ok := m.items[m.key(tenantID, variantID)]
	if !ok {
		return nil, errors.ErrNotFound(ctx, "inventory item", "not found")
	}
	if newOnHand < item.ReservedQty {
		return nil, errors.ErrConflict(ctx, "inventory stock", "new on-hand stock cannot be less than current reserved stock")
	}
	diff := newOnHand - item.OnHandQty
	item.OnHandQty = newOnHand
	item.AvailableQty = newOnHand - item.ReservedQty
	m.movements = append(m.movements, &model.StockMovement{
		ID:           uuid.New().String(),
		TenantID:     tenantID,
		InventoryID:  item.ID,
		MovementType: model.MovementAdjustment,
		Quantity:     diff,
		BalanceAfter: newOnHand,
		Notes:        notes,
	})
	return item, nil
}

func (m *mockInventoryRepo) RecordMovement(ctx context.Context, movement *model.StockMovement) (*model.StockMovement, *errors.Error) {
	m.movements = append(m.movements, movement)
	return movement, nil
}

func (m *mockInventoryRepo) ListMovements(ctx context.Context, tenantID, inventoryID string, limit, offset int) ([]*model.StockMovement, int64, *errors.Error) {
	var result []*model.StockMovement
	for _, mv := range m.movements {
		if mv.TenantID == tenantID && mv.InventoryID == inventoryID {
			result = append(result, mv)
		}
	}
	return result, int64(len(result)), nil
}

func TestInventoryService_CheckAndReserveStock(t *testing.T) {
	repo := newMockInventoryRepo()
	svc := NewInventoryService(repo)
	ctx := context.Background()

	tenantID := "tenant-test-1"
	variantID := "variant-item-100"

	// 1. Initialize stock with 10 units
	item, err := svc.InitInventory(ctx, tenantID, variantID, 10, 5)
	if err != nil {
		t.Fatalf("InitInventory failed: %v", err)
	}
	if item.AvailableQty != 10 || item.OnHandQty != 10 {
		t.Errorf("expected AvailableQty=10, got %d", item.AvailableQty)
	}

	// 2. Check stock availability
	ok, checkedItem, err := svc.CheckStock(ctx, tenantID, variantID, 8)
	if err != nil || !ok || checkedItem.AvailableQty != 10 {
		t.Fatalf("CheckStock expected true, got %v, err=%v", ok, err)
	}

	ok, _, err = svc.CheckStock(ctx, tenantID, variantID, 15)
	if err != nil || ok {
		t.Fatalf("CheckStock expected false for 15, got %v", ok)
	}

	// 3. Reserve 4 units
	reservedItem, err := svc.ReserveStock(ctx, tenantID, variantID, 4)
	if err != nil {
		t.Fatalf("ReserveStock failed: %v", err)
	}
	if reservedItem.ReservedQty != 4 || reservedItem.AvailableQty != 6 {
		t.Errorf("expected Reserved=4, Available=6, got Reserved=%d, Available=%d", reservedItem.ReservedQty, reservedItem.AvailableQty)
	}

	// 4. Attempt to reserve more than available (7 units, only 6 available)
	_, err = svc.ReserveStock(ctx, tenantID, variantID, 7)
	if err == nil {
		t.Fatalf("expected error when reserving 7 units with only 6 available")
	}

	// 5. Release 2 units
	releasedItem, err := svc.ReleaseStock(ctx, tenantID, variantID, 2)
	if err != nil {
		t.Fatalf("ReleaseStock failed: %v", err)
	}
	if releasedItem.ReservedQty != 2 || releasedItem.AvailableQty != 8 {
		t.Errorf("expected Reserved=2, Available=8, got Reserved=%d, Available=%d", releasedItem.ReservedQty, releasedItem.AvailableQty)
	}

	// 6. Deduct 2 units (order fulfilled)
	refType := "ORDER"
	refID := "order-1234"
	deductedItem, err := svc.DeductStock(ctx, tenantID, variantID, 2, refType, &refID)
	if err != nil {
		t.Fatalf("DeductStock failed: %v", err)
	}
	if deductedItem.OnHandQty != 8 || deductedItem.ReservedQty != 0 || deductedItem.AvailableQty != 8 {
		t.Errorf("expected OnHand=8, Reserved=0, Available=8, got OnHand=%d, Reserved=%d, Available=%d",
			deductedItem.OnHandQty, deductedItem.ReservedQty, deductedItem.AvailableQty)
	}

	// 7. Verify stock movement was recorded
	movements, total, err := svc.ListMovements(ctx, tenantID, item.ID, 10, 0)
	if err != nil || total != 1 || len(movements) != 1 {
		t.Fatalf("expected 1 movement, got %d, err=%v", total, err)
	}
	if movements[0].MovementType != model.MovementOutbound || movements[0].Quantity != 2 {
		t.Errorf("expected OUTBOUND movement of qty 2, got %s qty %d", movements[0].MovementType, movements[0].Quantity)
	}
}

func TestInventoryService_AdjustStock(t *testing.T) {
	repo := newMockInventoryRepo()
	svc := NewInventoryService(repo)
	ctx := context.Background()

	tenantID := "tenant-test-2"
	variantID := "variant-item-200"

	// 1. Initialize stock with 15 units
	item, err := svc.InitInventory(ctx, tenantID, variantID, 15, 3)
	if err != nil {
		t.Fatalf("InitInventory failed: %v", err)
	}

	// 2. Reserve 5 units
	_, err = svc.ReserveStock(ctx, tenantID, variantID, 5)
	if err != nil {
		t.Fatalf("ReserveStock failed: %v", err)
	}

	// 3. Adjust on-hand to 20 units
	adjusted, err := svc.AdjustStock(ctx, tenantID, variantID, 20, "Physical inventory audit count")
	if err != nil {
		t.Fatalf("AdjustStock failed: %v", err)
	}
	if adjusted.OnHandQty != 20 || adjusted.ReservedQty != 5 || adjusted.AvailableQty != 15 {
		t.Errorf("expected OnHand=20, Reserved=5, Available=15, got OnHand=%d, Reserved=%d, Available=%d",
			adjusted.OnHandQty, adjusted.ReservedQty, adjusted.AvailableQty)
	}

	// 4. Adjust on-hand below reserved (3 units when 5 reserved) should fail
	_, err = svc.AdjustStock(ctx, tenantID, variantID, 3, "Invalid count")
	if err == nil {
		t.Fatalf("expected error when adjusting stock below current reserved quantity")
	}

	// 5. Verify audit movement logged
	movements, total, err := svc.ListMovements(ctx, tenantID, item.ID, 10, 0)
	if err != nil || total != 1 {
		t.Fatalf("expected 1 adjustment movement, got %d", total)
	}
	if movements[0].MovementType != model.MovementAdjustment {
		t.Errorf("expected ADJUSTMENT movement, got %s", movements[0].MovementType)
	}
}
