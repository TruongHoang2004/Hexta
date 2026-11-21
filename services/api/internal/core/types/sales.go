package types

import "github.com/shopspring/decimal"

type CustomerInput struct {
	FullName string
	Phone    string
	Email    string
	Address  string
}

type ProductInput struct {
	Name        string
	Description string
	Category    string
	SKU         string
	VariantName string
	Price       *decimal.Decimal
}

type DraftItemInput struct {
	VariantID string
	Quantity  int
	UnitPrice *decimal.Decimal
}

// OrderListFilter narrows the order list. Empty fields mean no filter.
// Status is "draft" or "confirmed"; Assignee is "me" (the caller) or "unassigned".
type OrderListFilter struct {
	Status   string
	Assignee string
}

type DraftInput struct {
	Version          int64
	CustomerID       *string
	EmployeeID       *string
	RequiresDelivery bool
	DeliveryAddress  string
	Notes            string
	Items            []DraftItemInput
}
