// Package persistencemap converts order exception case rows to core values and back.
package persistencemap

import (
	"encoding/json"
	"errors"

	"github.com/TruongHoang2004/Hexta/services/api/internal/core/model"
	"github.com/TruongHoang2004/Hexta/services/api/internal/core/types"
	"github.com/TruongHoang2004/Hexta/services/api/internal/repository/ordercasemodel"
)

// ErrUnknownValue reports a stored type, status or action this code does not understand.
var ErrUnknownValue = errors.New("unknown stored order case value")

type storedLine struct {
	OrderItemID      string `json:"order_item_id"`
	OriginalQuantity int64  `json:"original_quantity"`
	Quantity         int64  `json:"quantity"`
}

func knownType(value string) bool {
	return value == types.OrderCaseTypeCancellation || value == types.OrderCaseTypeReturn
}
func knownStatus(value string) bool {
	return value == types.OrderCaseStatusOpen || value == types.OrderCaseStatusSentBack || value == types.OrderCaseStatusFinalized
}
func knownAction(value string) bool {
	switch value {
	case types.OrderCaseActionOpened, types.OrderCaseActionUpdated, types.OrderCaseActionSentBack, types.OrderCaseActionFinalized:
		return true
	}
	return false
}

// Case converts a case row with its line and event rows; unknown stored enums fail instead of inventing a state.
func Case(row ordercasemodel.Case, lines []ordercasemodel.Line, events []ordercasemodel.Event) (types.OrderCase, error) {
	if !knownType(row.CaseType) || !knownStatus(row.Status) {
		return types.OrderCase{}, ErrUnknownValue
	}
	value := types.OrderCase{ID: row.ID, TenantID: row.TenantID, OrderID: row.OrderID, Type: row.CaseType, Status: row.Status, HandlerID: row.HandlerID, OpenedBy: row.OpenedBy, ProgressNote: row.ProgressNote, FinalizedBy: row.FinalizedBy, FinalizedAt: row.FinalizedAt, Version: row.Version, CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt, Lines: make([]types.OrderCaseLine, len(lines)), Events: make([]types.OrderCaseEvent, len(events))}
	for i, line := range lines {
		value.Lines[i] = types.OrderCaseLine{OrderItemID: line.OrderItemID, OriginalQuantity: line.OriginalQuantity, Quantity: line.Quantity}
	}
	for i, event := range events {
		converted, err := Event(event)
		if err != nil {
			return types.OrderCase{}, err
		}
		value.Events[i] = converted
	}
	return value, nil
}

// CaseRow converts the case header; lines and events are written through their own rows.
func CaseRow(value types.OrderCase) ordercasemodel.Case {
	return ordercasemodel.Case{ID: value.ID, TenantID: value.TenantID, OrderID: value.OrderID, CaseType: value.Type, Status: value.Status, HandlerID: value.HandlerID, OpenedBy: value.OpenedBy, ProgressNote: value.ProgressNote, FinalizedBy: value.FinalizedBy, FinalizedAt: value.FinalizedAt, Version: value.Version, CreatedAt: value.CreatedAt, UpdatedAt: value.UpdatedAt}
}

// LineRows converts the case lines; ids are supplied by the caller so the mapper stays deterministic.
func LineRows(value types.OrderCase, ids []string) []ordercasemodel.Line {
	rows := make([]ordercasemodel.Line, len(value.Lines))
	for i, line := range value.Lines {
		rows[i] = ordercasemodel.Line{ID: ids[i], TenantID: value.TenantID, CaseID: value.ID, OrderID: value.OrderID, OrderItemID: line.OrderItemID, OriginalQuantity: line.OriginalQuantity, Quantity: line.Quantity, CreatedAt: value.UpdatedAt}
	}
	return rows
}

func Event(row ordercasemodel.Event) (types.OrderCaseEvent, error) {
	if !knownAction(row.Action) || !knownStatus(row.Status) {
		return types.OrderCaseEvent{}, ErrUnknownValue
	}
	var stored []storedLine
	if row.Lines != "" {
		if err := json.Unmarshal([]byte(row.Lines), &stored); err != nil {
			return types.OrderCaseEvent{}, err
		}
	}
	value := types.OrderCaseEvent{ID: row.ID, ActorID: row.ActorID, Action: row.Action, Status: row.Status, Note: row.Note, Version: row.Version, CreatedAt: row.CreatedAt, Lines: make([]types.OrderCaseLine, len(stored))}
	for i, line := range stored {
		value.Lines[i] = types.OrderCaseLine{OrderItemID: line.OrderItemID, OriginalQuantity: line.OriginalQuantity, Quantity: line.Quantity}
	}
	return value, nil
}

func EventRow(tenantID, caseID string, value types.OrderCaseEvent) ordercasemodel.Event {
	stored := make([]storedLine, len(value.Lines))
	for i, line := range value.Lines {
		stored[i] = storedLine{OrderItemID: line.OrderItemID, OriginalQuantity: line.OriginalQuantity, Quantity: line.Quantity}
	}
	data, _ := json.Marshal(stored)
	return ordercasemodel.Event{ID: value.ID, TenantID: tenantID, CaseID: caseID, Version: value.Version, ActorID: value.ActorID, Action: value.Action, Status: value.Status, Note: value.Note, Lines: string(data), CreatedAt: value.CreatedAt}
}

// Order keeps only the original order fields a case decision needs. The order row stays a legacy core ORM type.
func Order(row model.Order) types.OrderCaseOrder {
	value := types.OrderCaseOrder{ID: row.ID, Status: string(row.Status), Items: make([]types.OrderCaseOrderItem, len(row.Items))}
	for i, item := range row.Items {
		value.Items[i] = types.OrderCaseOrderItem{ID: item.ID, Quantity: int64(item.Quantity)}
	}
	return value
}
