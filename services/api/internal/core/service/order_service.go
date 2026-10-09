package service

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/TruongHoang2004/Hexta/packages/shared/pkg/errors"
	"github.com/TruongHoang2004/Hexta/services/api/internal/core/model"
	"github.com/TruongHoang2004/Hexta/services/api/internal/repository"
)

type CreateOrderItemParam struct {
	VariantID string          `json:"variant_id"`
	Quantity  int             `json:"quantity"`
	UnitPrice decimal.Decimal `json:"unit_price"`
}

type CreateOrderParam struct {
	TenantID    string                 `json:"tenant_id"`
	CustomerID  *string                `json:"customer_id,omitempty"`
	EmployeeID  *string                `json:"employee_id,omitempty"`
	Items       []CreateOrderItemParam `json:"items"`
	Discount    decimal.Decimal        `json:"discount"`
	Source      model.OrderSource      `json:"source"`
	DraftID     *string                `json:"draft_id,omitempty"`
	Notes       string                 `json:"notes,omitempty"`
	AutoConfirm bool                   `json:"auto_confirm"`
}

type IOrderService interface {
	CreateOrder(ctx context.Context, param CreateOrderParam) (*model.Order, *errors.Error)
	GetOrder(ctx context.Context, tenantID, orderID string) (*model.Order, *errors.Error)
	ListOrders(ctx context.Context, tenantID string, status *model.OrderStatus, limit, offset int) ([]*model.Order, int64, *errors.Error)
	TransitionStatus(ctx context.Context, tenantID, orderID string, targetStatus model.OrderStatus) (*model.Order, *errors.Error)
	UpdatePaymentStatus(ctx context.Context, tenantID, orderID string, paymentStatus model.PaymentStatus) (*model.Order, *errors.Error)
}

type OrderService struct {
	*baseService
	orderRepo     repository.IOrderRepository
	inventorySvc  IInventoryService
}

func NewOrderService(orderRepo repository.IOrderRepository, inventorySvc IInventoryService) *OrderService {
	return &OrderService{
		baseService:  NewBaseService(),
		orderRepo:    orderRepo,
		inventorySvc: inventorySvc,
	}
}

func generateOrderNumber() string {
	now := time.Now().Format("20060102")
	randomNum := rand.Intn(90000) + 10000
	return fmt.Sprintf("ORD-%s-%d", now, randomNum)
}

func (s *OrderService) CreateOrder(ctx context.Context, param CreateOrderParam) (*model.Order, *errors.Error) {
	if param.TenantID == "" {
		return nil, errors.ErrBadRequest(ctx).SetDetail("tenant_id is required")
	}
	if len(param.Items) == 0 {
		return nil, errors.ErrBadRequest(ctx).SetDetail("order must contain at least one line item")
	}

	subtotal := decimal.Zero
	orderItems := make([]model.OrderItem, len(param.Items))

	for i, item := range param.Items {
		if item.VariantID == "" {
			return nil, errors.ErrBadRequest(ctx).SetDetail("variant_id is required for every line item")
		}
		if item.Quantity <= 0 {
			return nil, errors.ErrBadRequest(ctx).SetDetail("quantity must be greater than zero")
		}
		if item.UnitPrice.LessThan(decimal.Zero) {
			return nil, errors.ErrBadRequest(ctx).SetDetail("unit_price cannot be negative")
		}

		lineTotal := item.UnitPrice.Mul(decimal.NewFromInt(int64(item.Quantity)))
		subtotal = subtotal.Add(lineTotal)

		orderItems[i] = model.OrderItem{
			VariantID: item.VariantID,
			Quantity:  item.Quantity,
			UnitPrice: item.UnitPrice,
			LineTotal: lineTotal,
		}
	}

	if param.Discount.LessThan(decimal.Zero) {
		param.Discount = decimal.Zero
	}
	totalAmount := subtotal.Sub(param.Discount)
	if totalAmount.LessThan(decimal.Zero) {
		totalAmount = decimal.Zero
	}

	source := param.Source
	if source == "" {
		source = model.OrderSourceWebForm
	}

	initialStatus := model.OrderStatusDraft
	if param.AutoConfirm {
		// Reserve stock for all items
		for i, item := range orderItems {
			_, err := s.inventorySvc.ReserveStock(ctx, param.TenantID, item.VariantID, item.Quantity)
			if err != nil {
				// Rollback already reserved items
				for j := 0; j < i; j++ {
					_, _ = s.inventorySvc.ReleaseStock(ctx, param.TenantID, orderItems[j].VariantID, orderItems[j].Quantity)
				}
				return nil, err
			}
		}
		initialStatus = model.OrderStatusConfirmed
	}

	order := &model.Order{
		ID:            uuid.New().String(),
		TenantID:      param.TenantID,
		OrderNumber:   generateOrderNumber(),
		CustomerID:    param.CustomerID,
		EmployeeID:    param.EmployeeID,
		Status:        initialStatus,
		PaymentStatus: model.PaymentStatusUnpaid,
		Subtotal:      subtotal,
		Discount:      param.Discount,
		TotalAmount:   totalAmount,
		Source:        source,
		DraftID:       param.DraftID,
		Notes:         param.Notes,
	}

	return s.orderRepo.CreateOrder(ctx, order, orderItems)
}

func (s *OrderService) GetOrder(ctx context.Context, tenantID, orderID string) (*model.Order, *errors.Error) {
	if tenantID == "" || orderID == "" {
		return nil, errors.ErrBadRequest(ctx).SetDetail("tenant_id and order_id are required")
	}
	return s.orderRepo.GetByID(ctx, tenantID, orderID)
}

func (s *OrderService) ListOrders(ctx context.Context, tenantID string, status *model.OrderStatus, limit, offset int) ([]*model.Order, int64, *errors.Error) {
	if tenantID == "" {
		return nil, 0, errors.ErrBadRequest(ctx).SetDetail("tenant_id is required")
	}
	return s.orderRepo.ListOrders(ctx, tenantID, status, limit, offset)
}

func (s *OrderService) TransitionStatus(ctx context.Context, tenantID, orderID string, targetStatus model.OrderStatus) (*model.Order, *errors.Error) {
	if tenantID == "" || orderID == "" {
		return nil, errors.ErrBadRequest(ctx).SetDetail("tenant_id and order_id are required")
	}

	order, err := s.orderRepo.GetByID(ctx, tenantID, orderID)
	if err != nil {
		return nil, err
	}

	if order.Status == targetStatus {
		return order, nil
	}

	// Validate Finite State Machine transitions
	switch order.Status {
	case model.OrderStatusDraft:
		switch targetStatus {
		case model.OrderStatusConfirmed:
			// Reserve stock for all items
			for i, item := range order.Items {
				_, err := s.inventorySvc.ReserveStock(ctx, tenantID, item.VariantID, item.Quantity)
				if err != nil {
					// Rollback already reserved items
					for j := 0; j < i; j++ {
						_, _ = s.inventorySvc.ReleaseStock(ctx, tenantID, order.Items[j].VariantID, order.Items[j].Quantity)
					}
					return nil, err
				}
			}
		case model.OrderStatusCancelled:
			// No stock was held, directly cancel
		default:
			return nil, errors.ErrBadRequest(ctx).SetDetail(fmt.Sprintf("invalid transition from %s to %s", order.Status, targetStatus))
		}

	case model.OrderStatusConfirmed:
		switch targetStatus {
		case model.OrderStatusProcessing:
			// Valid progression
		case model.OrderStatusCancelled:
			// Release reserved stock
			for _, item := range order.Items {
				_, _ = s.inventorySvc.ReleaseStock(ctx, tenantID, item.VariantID, item.Quantity)
			}
		default:
			return nil, errors.ErrBadRequest(ctx).SetDetail(fmt.Sprintf("invalid transition from %s to %s", order.Status, targetStatus))
		}

	case model.OrderStatusProcessing:
		switch targetStatus {
		case model.OrderStatusFulfilled:
			// Deduct held stock permanently and record outbound movement
			for _, item := range order.Items {
				refType := "ORDER"
				refID := order.ID
				_, err := s.inventorySvc.DeductStock(ctx, tenantID, item.VariantID, item.Quantity, refType, &refID)
				if err != nil {
					return nil, err
				}
			}
		case model.OrderStatusCancelled:
			// Release reserved stock
			for _, item := range order.Items {
				_, _ = s.inventorySvc.ReleaseStock(ctx, tenantID, item.VariantID, item.Quantity)
			}
		default:
			return nil, errors.ErrBadRequest(ctx).SetDetail(fmt.Sprintf("invalid transition from %s to %s", order.Status, targetStatus))
		}

	case model.OrderStatusFulfilled:
		switch targetStatus {
		case model.OrderStatusReturned:
			// Return progression
		default:
			return nil, errors.ErrBadRequest(ctx).SetDetail(fmt.Sprintf("cannot transition from fulfilled order to %s", targetStatus))
		}

	default:
		return nil, errors.ErrBadRequest(ctx).SetDetail(fmt.Sprintf("cannot transition order from terminal state %s", order.Status))
	}

	return s.orderRepo.UpdateOrderStatus(ctx, tenantID, orderID, targetStatus)
}

func (s *OrderService) UpdatePaymentStatus(ctx context.Context, tenantID, orderID string, paymentStatus model.PaymentStatus) (*model.Order, *errors.Error) {
	if tenantID == "" || orderID == "" {
		return nil, errors.ErrBadRequest(ctx).SetDetail("tenant_id and order_id are required")
	}
	return s.orderRepo.UpdatePaymentStatus(ctx, tenantID, orderID, paymentStatus)
}
