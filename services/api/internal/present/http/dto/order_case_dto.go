package dto

import "time"

type OpenOrderCaseRequest struct {
	Type string `json:"type" validate:"required,oneof=cancellation return" enums:"cancellation,return"`
	// HandlerID is optional; omitted assigns the caller. Only the owner may assign someone else.
	HandlerID *string `json:"handler_id" validate:"omitempty,max=255"`
	Note      string  `json:"note" validate:"max=5000"`
}

type OrderCaseLineRequest struct {
	OrderItemID string `json:"order_item_id" validate:"required,max=36"`
	// Quantity 0 removes the line; it can never exceed the original line quantity.
	Quantity int64 `json:"quantity" validate:"min=0,max=1000000"`
}

// UpdateOrderCaseRequest replaces the progress note and the full set of return lines on the expected version.
type UpdateOrderCaseRequest struct {
	Version   int64                  `json:"version" validate:"min=1"`
	Note      string                 `json:"note" validate:"max=5000"`
	HandlerID *string                `json:"handler_id" validate:"omitempty,max=255"`
	Lines     []OrderCaseLineRequest `json:"lines" validate:"max=100,dive"`
}

// OrderCaseDecisionRequest is an owner send-back (note required) or finalize on the expected version.
type OrderCaseDecisionRequest struct {
	Version int64  `json:"version" validate:"min=1"`
	Note    string `json:"note" validate:"max=5000"`
}

type OrderCaseLineDTO struct {
	OrderItemID      string `json:"order_item_id"`
	OriginalQuantity int64  `json:"original_quantity"`
	Quantity         int64  `json:"quantity"`
}

type OrderCaseEventDTO struct {
	ID        string             `json:"id"`
	Version   int64              `json:"version"`
	ActorID   string             `json:"actor_id"`
	Action    string             `json:"action" enums:"opened,updated,sent_back,finalized"`
	Status    string             `json:"status" enums:"open,sent_back,finalized"`
	Note      string             `json:"note"`
	Lines     []OrderCaseLineDTO `json:"lines"`
	CreatedAt time.Time          `json:"created_at"`
}

type OrderCaseDTO struct {
	ID           string              `json:"id"`
	OrderID      string              `json:"order_id"`
	Type         string              `json:"type" enums:"cancellation,return"`
	Status       string              `json:"status" enums:"open,sent_back,finalized"`
	HandlerID    string              `json:"handler_id"`
	OpenedBy     string              `json:"opened_by"`
	ProgressNote string              `json:"progress_note"`
	FinalizedBy  *string             `json:"finalized_by"`
	FinalizedAt  *time.Time          `json:"finalized_at"`
	Version      int64               `json:"version"`
	CreatedAt    time.Time           `json:"created_at"`
	UpdatedAt    time.Time           `json:"updated_at"`
	Lines        []OrderCaseLineDTO  `json:"lines"`
	Events       []OrderCaseEventDTO `json:"events"`
}

// OrderCaseLookupResponse carries the order's case, or null when none has been opened.
type OrderCaseLookupResponse struct {
	Case *OrderCaseDTO `json:"case"`
}
