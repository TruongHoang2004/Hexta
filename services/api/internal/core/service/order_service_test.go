package service

import (
	"context"
	"testing"
	"time"

	"github.com/TruongHoang2004/Hexta/packages/shared/pkg/errors"
	"github.com/TruongHoang2004/Hexta/services/api/internal/core/model"
	"github.com/TruongHoang2004/Hexta/services/api/internal/repository"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type mockOrderRepo struct {
	orders map[string]*model.Order
}

func newMockOrderRepo() *mockOrderRepo {
	return &mockOrderRepo{
		orders: make(map[string]*model.Order),
	}
}

func (m *mockOrderRepo) CreateOrder(ctx context.Context, order *model.Order, items []model.OrderItem) (*model.Order, *errors.Error) {
	if order.ID == "" {
		order.ID = uuid.New().String()
	}
	for i := range items {
		if items[i].ID == "" {
			items[i].ID = uuid.New().String()
		}
		items[i].OrderID = order.ID
		items[i].TenantID = order.TenantID
	}
	order.Items = items
	order.CreatedAt = time.Now()
	order.UpdatedAt = time.Now()
	m.orders[order.TenantID+":"+order.ID] = order
	return order, nil
}

func (m *mockOrderRepo) GetByID(ctx context.Context, tenantID, orderID string) (*model.Order, *errors.Error) {
	if order, ok := m.orders[tenantID+":"+orderID]; ok {
		return order, nil
	}
	return nil, errors.ErrNotFound(ctx, "order", "not found")
}

func (m *mockOrderRepo) GetByOrderNumber(ctx context.Context, tenantID, orderNumber string) (*model.Order, *errors.Error) {
	for _, ord := range m.orders {
		if ord.TenantID == tenantID && ord.OrderNumber == orderNumber {
			return ord, nil
		}
	}
	return nil, errors.ErrNotFound(ctx, "order", "not found")
}

func (m *mockOrderRepo) ListOrders(ctx context.Context, tenantID string, status *model.OrderStatus, limit, offset int) ([]*model.Order, int64, *errors.Error) {
	var list []*model.Order
	for _, ord := range m.orders {
		if ord.TenantID == tenantID {
			if status != nil && *status != "" && ord.Status != *status {
				continue
			}
			list = append(list, ord)
		}
	}
	return list, int64(len(list)), nil
}

func (m *mockOrderRepo) UpdateOrderStatus(ctx context.Context, tenantID, orderID string, status model.OrderStatus) (*model.Order, *errors.Error) {
	ord, ok := m.orders[tenantID+":"+orderID]
	if !ok {
		return nil, errors.ErrNotFound(ctx, "order", "not found")
	}
	ord.Status = status
	ord.UpdatedAt = time.Now()
	return ord, nil
}

func (m *mockOrderRepo) UpdatePaymentStatus(ctx context.Context, tenantID, orderID string, paymentStatus model.PaymentStatus) (*model.Order, *errors.Error) {
	ord, ok := m.orders[tenantID+":"+orderID]
	if !ok {
		return nil, errors.ErrNotFound(ctx, "order", "not found")
	}
	ord.PaymentStatus = paymentStatus
	ord.UpdatedAt = time.Now()
	return ord, nil
}

var _ repository.IOrderRepository = (*mockOrderRepo)(nil)

func TestOrderService_CreateOrder_Draft(t *testing.T) {
	orderRepo := newMockOrderRepo()
	invRepo := newMockInventoryRepo()
	invSvc := NewInventoryService(invRepo)
	svc := NewOrderService(orderRepo, invSvc)
	ctx := context.Background()

	tenantID := "tenant-alpha"
	param := CreateOrderParam{
		TenantID: tenantID,
		Items: []CreateOrderItemParam{
			{
				VariantID: "var-1",
				Quantity:  2,
				UnitPrice: decimal.NewFromFloat(50.00),
			},
			{
				VariantID: "var-2",
				Quantity:  1,
				UnitPrice: decimal.NewFromFloat(30.00),
			},
		},
		Discount:    decimal.NewFromFloat(10.00),
		Source:      model.OrderSourceWebForm,
		AutoConfirm: false,
	}

	order, err := svc.CreateOrder(ctx, param)
	if err != nil {
		t.Fatalf("CreateOrder failed: %v", err)
	}

	if order.Status != model.OrderStatusDraft {
		t.Errorf("expected status %s, got %s", model.OrderStatusDraft, order.Status)
	}
	expectedSubtotal := decimal.NewFromFloat(130.00)
	if !order.Subtotal.Equal(expectedSubtotal) {
		t.Errorf("expected subtotal %s, got %s", expectedSubtotal, order.Subtotal)
	}
	expectedTotal := decimal.NewFromFloat(120.00)
	if !order.TotalAmount.Equal(expectedTotal) {
		t.Errorf("expected total %s, got %s", expectedTotal, order.TotalAmount)
	}
	if len(order.Items) != 2 {
		t.Errorf("expected 2 items, got %d", len(order.Items))
	}
}

func TestOrderService_CreateOrder_AutoConfirm_WithInventory(t *testing.T) {
	orderRepo := newMockOrderRepo()
	invRepo := newMockInventoryRepo()
	invSvc := NewInventoryService(invRepo)
	svc := NewOrderService(orderRepo, invSvc)
	ctx := context.Background()

	tenantID := "tenant-beta"
	variantID := "var-shoes"

	// 1. Initialize stock with 10 units
	_, err := invSvc.InitInventory(ctx, tenantID, variantID, 10, 2)
	if err != nil {
		t.Fatalf("InitInventory failed: %v", err)
	}

	// 2. Create order with AutoConfirm
	param := CreateOrderParam{
		TenantID: tenantID,
		Items: []CreateOrderItemParam{
			{
				VariantID: variantID,
				Quantity:  4,
				UnitPrice: decimal.NewFromFloat(100.00),
			},
		},
		AutoConfirm: true,
	}

	order, err := svc.CreateOrder(ctx, param)
	if err != nil {
		t.Fatalf("CreateOrder with AutoConfirm failed: %v", err)
	}

	if order.Status != model.OrderStatusConfirmed {
		t.Errorf("expected status confirmed, got %s", order.Status)
	}

	// Verify inventory stock was reserved
	inv, err := invSvc.GetInventory(ctx, tenantID, variantID)
	if err != nil {
		t.Fatalf("GetInventory failed: %v", err)
	}
	if inv.ReservedQty != 4 || inv.AvailableQty != 6 {
		t.Errorf("expected Reserved=4, Available=6; got Reserved=%d, Available=%d", inv.ReservedQty, inv.AvailableQty)
	}
}

func TestOrderService_CreateOrder_AutoConfirm_RollbackOnInsufficientStock(t *testing.T) {
	orderRepo := newMockOrderRepo()
	invRepo := newMockInventoryRepo()
	invSvc := NewInventoryService(invRepo)
	svc := NewOrderService(orderRepo, invSvc)
	ctx := context.Background()

	tenantID := "tenant-gamma"
	var1 := "var-hat"
	var2 := "var-gloves"

	// Init var1 with 10 units, but var2 with only 1 unit
	_, _ = invSvc.InitInventory(ctx, tenantID, var1, 10, 0)
	_, _ = invSvc.InitInventory(ctx, tenantID, var2, 1, 0)

	// Order wants 5 of var1 and 5 of var2 (which fails)
	param := CreateOrderParam{
		TenantID: tenantID,
		Items: []CreateOrderItemParam{
			{VariantID: var1, Quantity: 5, UnitPrice: decimal.NewFromFloat(20.00)},
			{VariantID: var2, Quantity: 5, UnitPrice: decimal.NewFromFloat(15.00)},
		},
		AutoConfirm: true,
	}

	_, err := svc.CreateOrder(ctx, param)
	if err == nil {
		t.Fatalf("expected error due to insufficient stock on var2")
	}

	// Verify var1 was rolled back and ReservedQty is 0
	inv1, _ := invSvc.GetInventory(ctx, tenantID, var1)
	if inv1.ReservedQty != 0 || inv1.AvailableQty != 10 {
		t.Errorf("expected var1 reserved to rollback to 0, got Reserved=%d, Available=%d", inv1.ReservedQty, inv1.AvailableQty)
	}
}

func TestOrderService_FSM_Transitions(t *testing.T) {
	orderRepo := newMockOrderRepo()
	invRepo := newMockInventoryRepo()
	invSvc := NewInventoryService(invRepo)
	svc := NewOrderService(orderRepo, invSvc)
	ctx := context.Background()

	tenantID := "tenant-fsm"
	variantID := "var-watch"

	// 1. Initialize stock
	_, _ = invSvc.InitInventory(ctx, tenantID, variantID, 10, 1)

	// 2. Create Draft order
	order, err := svc.CreateOrder(ctx, CreateOrderParam{
		TenantID: tenantID,
		Items: []CreateOrderItemParam{
			{VariantID: variantID, Quantity: 3, UnitPrice: decimal.NewFromFloat(200.00)},
		},
	})
	if err != nil {
		t.Fatalf("CreateOrder failed: %v", err)
	}

	// Invalid transition: Draft -> Fulfilled directly
	_, err = svc.TransitionStatus(ctx, tenantID, order.ID, model.OrderStatusFulfilled)
	if err == nil {
		t.Fatalf("expected error transitioning Draft -> Fulfilled")
	}

	// Valid transition: Draft -> Confirmed (reserves stock)
	order, err = svc.TransitionStatus(ctx, tenantID, order.ID, model.OrderStatusConfirmed)
	if err != nil {
		t.Fatalf("transition to Confirmed failed: %v", err)
	}
	if order.Status != model.OrderStatusConfirmed {
		t.Errorf("expected status confirmed, got %s", order.Status)
	}
	inv, _ := invSvc.GetInventory(ctx, tenantID, variantID)
	if inv.ReservedQty != 3 || inv.AvailableQty != 7 {
		t.Errorf("expected Reserved=3, Available=7; got Reserved=%d, Available=%d", inv.ReservedQty, inv.AvailableQty)
	}

	// Valid transition: Confirmed -> Processing
	order, err = svc.TransitionStatus(ctx, tenantID, order.ID, model.OrderStatusProcessing)
	if err != nil {
		t.Fatalf("transition to Processing failed: %v", err)
	}
	if order.Status != model.OrderStatusProcessing {
		t.Errorf("expected status processing, got %s", order.Status)
	}

	// Valid transition: Processing -> Fulfilled (deducts stock)
	order, err = svc.TransitionStatus(ctx, tenantID, order.ID, model.OrderStatusFulfilled)
	if err != nil {
		t.Fatalf("transition to Fulfilled failed: %v", err)
	}
	if order.Status != model.OrderStatusFulfilled {
		t.Errorf("expected status fulfilled, got %s", order.Status)
	}
	inv, _ = invSvc.GetInventory(ctx, tenantID, variantID)
	if inv.OnHandQty != 7 || inv.ReservedQty != 0 || inv.AvailableQty != 7 {
		t.Errorf("expected OnHand=7, Reserved=0, Available=7; got OnHand=%d, Reserved=%d, Available=%d",
			inv.OnHandQty, inv.ReservedQty, inv.AvailableQty)
	}

	// Terminal transition: Fulfilled -> Draft must fail
	_, err = svc.TransitionStatus(ctx, tenantID, order.ID, model.OrderStatusDraft)
	if err == nil {
		t.Fatalf("expected error transitioning from terminal Fulfilled to Draft")
	}
}

func TestOrderService_FSM_Cancellation_ReleasesStock(t *testing.T) {
	orderRepo := newMockOrderRepo()
	invRepo := newMockInventoryRepo()
	invSvc := NewInventoryService(invRepo)
	svc := NewOrderService(orderRepo, invSvc)
	ctx := context.Background()

	tenantID := "tenant-fsm-cancel"
	variantID := "var-camera"

	// 1. Initialize stock
	_, _ = invSvc.InitInventory(ctx, tenantID, variantID, 5, 0)

	// 2. Create Confirmed order (holds 3 items)
	order, err := svc.CreateOrder(ctx, CreateOrderParam{
		TenantID: tenantID,
		Items: []CreateOrderItemParam{
			{VariantID: variantID, Quantity: 3, UnitPrice: decimal.NewFromFloat(500.00)},
		},
		AutoConfirm: true,
	})
	if err != nil {
		t.Fatalf("CreateOrder failed: %v", err)
	}

	inv, _ := invSvc.GetInventory(ctx, tenantID, variantID)
	if inv.ReservedQty != 3 {
		t.Fatalf("expected ReservedQty=3, got %d", inv.ReservedQty)
	}

	// 3. Cancel Confirmed order -> should release held stock
	order, err = svc.TransitionStatus(ctx, tenantID, order.ID, model.OrderStatusCancelled)
	if err != nil {
		t.Fatalf("Transition to Cancelled failed: %v", err)
	}
	if order.Status != model.OrderStatusCancelled {
		t.Errorf("expected Cancelled, got %s", order.Status)
	}

	inv, _ = invSvc.GetInventory(ctx, tenantID, variantID)
	if inv.ReservedQty != 0 || inv.AvailableQty != 5 {
		t.Errorf("expected Reserved=0, Available=5 after cancellation; got Reserved=%d, Available=%d", inv.ReservedQty, inv.AvailableQty)
	}

	// 4. Cannot transition cancelled order back to confirmed
	_, err = svc.TransitionStatus(ctx, tenantID, order.ID, model.OrderStatusConfirmed)
	if err == nil {
		t.Fatalf("expected error transitioning from terminal Cancelled to Confirmed")
	}
}

func TestOrderService_ValidationErrors(t *testing.T) {
	orderRepo := newMockOrderRepo()
	invRepo := newMockInventoryRepo()
	invSvc := NewInventoryService(invRepo)
	svc := NewOrderService(orderRepo, invSvc)
	ctx := context.Background()

	// Missing tenant ID
	_, err := svc.CreateOrder(ctx, CreateOrderParam{
		Items: []CreateOrderItemParam{
			{VariantID: "v1", Quantity: 1, UnitPrice: decimal.NewFromFloat(10)},
		},
	})
	if err == nil {
		t.Errorf("expected error for missing tenant_id")
	}

	// Empty items
	_, err = svc.CreateOrder(ctx, CreateOrderParam{
		TenantID: "tenant-1",
		Items:    []CreateOrderItemParam{},
	})
	if err == nil {
		t.Errorf("expected error for empty items")
	}

	// Zero quantity
	_, err = svc.CreateOrder(ctx, CreateOrderParam{
		TenantID: "tenant-1",
		Items: []CreateOrderItemParam{
			{VariantID: "v1", Quantity: 0, UnitPrice: decimal.NewFromFloat(10)},
		},
	})
	if err == nil {
		t.Errorf("expected error for zero quantity")
	}

	// Negative unit price
	_, err = svc.CreateOrder(ctx, CreateOrderParam{
		TenantID: "tenant-1",
		Items: []CreateOrderItemParam{
			{VariantID: "v1", Quantity: 1, UnitPrice: decimal.NewFromFloat(-5)},
		},
	})
	if err == nil {
		t.Errorf("expected error for negative unit price")
	}
}
