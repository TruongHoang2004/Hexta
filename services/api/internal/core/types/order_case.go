package types

import "time"

// Order exception case values (MVP 6). A case records a cancellation or return on a confirmed order. It never
// changes the original order, collects or refunds money, or moves stock.
const (
	OrderCaseTypeCancellation = "cancellation"
	OrderCaseTypeReturn       = "return"

	OrderCaseStatusOpen      = "open"
	OrderCaseStatusSentBack  = "sent_back"
	OrderCaseStatusFinalized = "finalized"

	OrderCaseActionOpened    = "opened"
	OrderCaseActionUpdated   = "updated"
	OrderCaseActionSentBack  = "sent_back"
	OrderCaseActionFinalized = "finalized"
)

// OrderCase is one exception case with its current return lines and its append-only history (oldest first).
type OrderCase struct {
	ID, TenantID, OrderID, Type, Status, HandlerID, OpenedBy, ProgressNote string
	FinalizedBy                                                            *string
	FinalizedAt                                                            *time.Time
	Version                                                                int64
	CreatedAt, UpdatedAt                                                   time.Time
	Lines                                                                  []OrderCaseLine
	Events                                                                 []OrderCaseEvent
}

// OrderCaseLine is a returned quantity of one original order line.
type OrderCaseLine struct {
	OrderItemID      string
	OriginalQuantity int64
	Quantity         int64
}

// OrderCaseEvent is one history entry; Version is the case version the entry produced.
type OrderCaseEvent struct {
	ID, ActorID, Action, Status, Note string
	Version                           int64
	Lines                             []OrderCaseLine
	CreatedAt                         time.Time
}

// OrderCaseOrder is the original order state a case decision needs.
type OrderCaseOrder struct {
	ID, Status string
	Items      []OrderCaseOrderItem
}

type OrderCaseOrderItem struct {
	ID       string
	Quantity int64
}

type OpenOrderCaseInput struct {
	Type string
	// HandlerID nil assigns the caller as the handler.
	HandlerID *string
	Note      string
}

type OrderCaseLineInput struct {
	OrderItemID string
	Quantity    int64
}

type UpdateOrderCaseInput struct {
	Version int64
	Note    string
	// HandlerID nil keeps the current handler; only the owner may change it.
	HandlerID *string
	Lines     []OrderCaseLineInput
}

// OrderCaseDecisionInput is an owner decision (send back or finalize) on the expected case version.
type OrderCaseDecisionInput struct {
	Version int64
	Note    string
}
