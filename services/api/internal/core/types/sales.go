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

type DraftInput struct {
	Version          int64
	CustomerID       *string
	EmployeeID       *string
	RequiresDelivery bool
	DeliveryAddress  string
	Notes            string
	Items            []DraftItemInput
}
