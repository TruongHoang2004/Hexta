package dto

import (
	"time"

	"github.com/TruongHoang2004/Hexta/services/api/internal/core/model"
	"github.com/shopspring/decimal"
)

type CreateOrderItemRequest struct {
	VariantID string          `json:"variant_id" validate:"required"`
	Quantity  int             `json:"quantity" validate:"required,min=1"`
	UnitPrice decimal.Decimal `json:"unit_price" validate:"required"`
}

type CreateOrderRequest struct {
	TenantID    string                   `json:"tenant_id"`
	CustomerID  *string                  `json:"customer_id,omitempty"`
	EmployeeID  *string                  `json:"employee_id,omitempty"`
	Items       []CreateOrderItemRequest `json:"items" validate:"required,min=1,dive"`
	Discount    decimal.Decimal          `json:"discount"`
	Source      model.OrderSource        `json:"source"`
	DraftID     *string                  `json:"draft_id,omitempty"`
	Notes       string                   `json:"notes,omitempty"`
	AutoConfirm bool                     `json:"auto_confirm"`
}

type UpdateOrderStatusRequest struct {
	Status model.OrderStatus `json:"status" validate:"required"`
}

type UpdatePaymentStatusRequest struct {
	PaymentStatus model.PaymentStatus `json:"payment_status" validate:"required"`
}

type OrderItemResponse struct {
	ID        string          `json:"id"`
	VariantID string          `json:"variant_id"`
	Quantity  int             `json:"quantity"`
	UnitPrice decimal.Decimal `json:"unit_price"`
	LineTotal decimal.Decimal `json:"line_total"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
}

type OrderResponse struct {
	ID            string              `json:"id"`
	TenantID      string              `json:"tenant_id"`
	OrderNumber   string              `json:"order_number"`
	CustomerID    *string             `json:"customer_id,omitempty"`
	EmployeeID    *string             `json:"employee_id,omitempty"`
	Status        model.OrderStatus   `json:"status"`
	PaymentStatus model.PaymentStatus `json:"payment_status"`
	Subtotal      decimal.Decimal     `json:"subtotal"`
	Discount      decimal.Decimal     `json:"discount"`
	TotalAmount   decimal.Decimal     `json:"total_amount"`
	Source        model.OrderSource   `json:"source"`
	DraftID       *string             `json:"draft_id,omitempty"`
	Notes         string              `json:"notes,omitempty"`
	Items         []OrderItemResponse `json:"items,omitempty"`
	CreatedAt     time.Time           `json:"created_at"`
	UpdatedAt     time.Time           `json:"updated_at"`
}

type OrderListResponse struct {
	Items      []OrderResponse `json:"items"`
	Total      int64           `json:"total"`
	Page       int             `json:"page"`
	Size       int             `json:"size"`
	TotalPages int             `json:"total_pages"`
}
