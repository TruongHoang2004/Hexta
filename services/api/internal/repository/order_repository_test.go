package repository

import (
	"context"
	"testing"

	"github.com/TruongHoang2004/Hexta/services/api/internal/core/model"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

func TestOrderRepository_CRUD(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}

	tx := db.Begin()
	defer tx.Rollback()

	repo := NewOrderDBRepository(tx)
	ctx := context.Background()

	tenantID := uuid.New().String()
	orderNumber := "ORD-" + uuid.New().String()[:8]

	// 1. Create Order with Items
	order := &model.Order{
		ID:            uuid.New().String(),
		TenantID:      tenantID,
		OrderNumber:   orderNumber,
		Status:        model.OrderStatusDraft,
		PaymentStatus: model.PaymentStatusUnpaid,
		Subtotal:      decimal.NewFromInt(100),
		Discount:      decimal.NewFromInt(10),
		TotalAmount:   decimal.NewFromInt(90),
		Source:        model.OrderSourceWebForm,
		Notes:         "Test order notes",
	}

	items := []model.OrderItem{
		{
			ID:        uuid.New().String(),
			TenantID:  tenantID,
			VariantID: uuid.New().String(),
			Quantity:  2,
			UnitPrice: decimal.NewFromInt(30),
			LineTotal: decimal.NewFromInt(60),
		},
		{
			ID:        uuid.New().String(),
			TenantID:  tenantID,
			VariantID: uuid.New().String(),
			Quantity:  1,
			UnitPrice: decimal.NewFromInt(40),
			LineTotal: decimal.NewFromInt(40),
		},
	}

	created, err := repo.CreateOrder(ctx, order, items)
	if err != nil {
		t.Fatalf("failed to create order: %v", err)
	}
	if created == nil || created.ID == "" {
		t.Fatalf("expected created order with ID, got %+v", created)
	}
	if len(created.Items) != 2 {
		t.Fatalf("expected 2 order items, got %d", len(created.Items))
	}

	// 2. Get Order By ID
	byID, err := repo.GetByID(ctx, tenantID, created.ID)
	if err != nil {
		t.Fatalf("failed to get order by ID: %v", err)
	}
	if byID == nil || byID.OrderNumber != orderNumber {
		t.Fatalf("expected order number %s, got %+v", orderNumber, byID)
	}
	if len(byID.Items) != 2 {
		t.Fatalf("expected 2 eager-loaded items, got %d", len(byID.Items))
	}

	// 2b. Get Order By non-existent ID
	_, err = repo.GetByID(ctx, tenantID, uuid.New().String())
	if err == nil {
		t.Fatalf("expected error getting non-existent order by ID")
	}

	// 3. Get Order By Order Number
	byNum, err := repo.GetByOrderNumber(ctx, tenantID, orderNumber)
	if err != nil {
		t.Fatalf("failed to get order by number: %v", err)
	}
	if byNum == nil || byNum.ID != created.ID {
		t.Fatalf("expected order ID %s, got %+v", created.ID, byNum)
	}
	if len(byNum.Items) != 2 {
		t.Fatalf("expected 2 eager-loaded items, got %d", len(byNum.Items))
	}

	// 3b. Get Order By non-existent Order Number
	_, err = repo.GetByOrderNumber(ctx, tenantID, "ORD-NONEXISTENT")
	if err == nil {
		t.Fatalf("expected error getting non-existent order number")
	}

	// 4. List Orders with pagination and filters
	orders, total, err := repo.ListOrders(ctx, tenantID, nil, 10, 0)
	if err != nil {
		t.Fatalf("failed to list orders: %v", err)
	}
	if total != 1 || len(orders) != 1 {
		t.Fatalf("expected 1 order, got total=%d len=%d", total, len(orders))
	}

	statusFilter := model.OrderStatusDraft
	ordersFiltered, totalFiltered, err := repo.ListOrders(ctx, tenantID, &statusFilter, 10, 0)
	if err != nil || totalFiltered != 1 || len(ordersFiltered) != 1 {
		t.Fatalf("expected 1 filtered order, got total=%d len=%d", totalFiltered, len(ordersFiltered))
	}

	diffStatus := model.OrderStatusCancelled
	ordersCancelled, totalCancelled, err := repo.ListOrders(ctx, tenantID, &diffStatus, 10, 0)
	if err != nil || totalCancelled != 0 || len(ordersCancelled) != 0 {
		t.Fatalf("expected 0 cancelled orders, got total=%d len=%d", totalCancelled, len(ordersCancelled))
	}

	// 5. Update Order Status
	updatedStatus, err := repo.UpdateOrderStatus(ctx, tenantID, created.ID, model.OrderStatusConfirmed)
	if err != nil {
		t.Fatalf("failed to update order status: %v", err)
	}
	if updatedStatus.Status != model.OrderStatusConfirmed {
		t.Fatalf("expected status %s, got %s", model.OrderStatusConfirmed, updatedStatus.Status)
	}

	// 5b. Update Status of non-existent order
	_, err = repo.UpdateOrderStatus(ctx, tenantID, uuid.New().String(), model.OrderStatusProcessing)
	if err == nil {
		t.Fatalf("expected error updating status of non-existent order")
	}

	// 6. Update Payment Status
	updatedPayment, err := repo.UpdatePaymentStatus(ctx, tenantID, created.ID, model.PaymentStatusPaid)
	if err != nil {
		t.Fatalf("failed to update payment status: %v", err)
	}
	if updatedPayment.PaymentStatus != model.PaymentStatusPaid {
		t.Fatalf("expected payment status %s, got %s", model.PaymentStatusPaid, updatedPayment.PaymentStatus)
	}

	// 6b. Update Payment Status of non-existent order
	_, err = repo.UpdatePaymentStatus(ctx, tenantID, uuid.New().String(), model.PaymentStatusRefunded)
	if err == nil {
		t.Fatalf("expected error updating payment status of non-existent order")
	}
}
