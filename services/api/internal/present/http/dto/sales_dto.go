package dto

import (
	"github.com/TruongHoang2004/Hexta/services/api/internal/core/model"
	"github.com/shopspring/decimal"
)

type SalesCustomerRequest struct {
	FullName string `json:"full_name" validate:"required,max=255"`
	Phone    string `json:"phone" validate:"max=32"`
	Email    string `json:"email" validate:"omitempty,email,max=255"`
	Address  string `json:"address" validate:"max=5000"`
}
type SalesProductRequest struct {
	Name        string           `json:"name" validate:"required,max=255"`
	Description string           `json:"description" validate:"max=10000"`
	Category    string           `json:"category" validate:"max=100"`
	SKU         string           `json:"sku" validate:"required,max=64"`
	VariantName string           `json:"variant_name" validate:"required,max=128"`
	Price       *decimal.Decimal `json:"price" swaggertype:"string"`
}
type SalesDraftItemRequest struct {
	VariantID string           `json:"variant_id" validate:"required,max=36"`
	Quantity  int              `json:"quantity" validate:"min=1,max=1000000"`
	UnitPrice *decimal.Decimal `json:"unit_price" swaggertype:"string"`
}
type SalesDraftRequest struct {
	Version          int64                   `json:"version" validate:"min=0"`
	CustomerID       *string                 `json:"customer_id" validate:"omitempty,max=36"`
	EmployeeID       *string                 `json:"employee_id" validate:"omitempty,max=255"`
	RequiresDelivery bool                    `json:"requires_delivery"`
	DeliveryAddress  string                  `json:"delivery_address" validate:"max=5000"`
	Notes            string                  `json:"notes" validate:"max=10000"`
	Items            []SalesDraftItemRequest `json:"items" validate:"max=100,dive"`
	// Legacy callers must not silently bypass explicit confirmation.
	AutoConfirm bool `json:"auto_confirm"`
}
type SalesConfirmRequest struct {
	Version int64 `json:"version" validate:"min=1"`
}
type SalesCustomerResponse = model.Customer
type SalesProductResponse = model.Product
type SalesOrderResponse = model.Order
type SalesCustomersResponse struct {
	Items []model.Customer `json:"items"`
	Total int64            `json:"total"`
	Page  int              `json:"page"`
	Size  int              `json:"size"`
}
type SalesProductsResponse struct {
	Items []model.Product `json:"items"`
	Total int64           `json:"total"`
	Page  int             `json:"page"`
	Size  int             `json:"size"`
}
type SalesOrdersResponse struct {
	Items []model.Order `json:"items"`
	Total int64         `json:"total"`
	Page  int           `json:"page"`
	Size  int           `json:"size"`
}
type SalesArchiveResponse struct {
	ID       string `json:"id"`
	Archived bool   `json:"archived"`
}
