// Package httpmap converts order exception case HTTP requests and use-case values.
package httpmap

import (
	"github.com/TruongHoang2004/Hexta/services/api/internal/core/types"
	"github.com/TruongHoang2004/Hexta/services/api/internal/present/http/dto"
)

func OpenInput(req dto.OpenOrderCaseRequest) types.OpenOrderCaseInput {
	return types.OpenOrderCaseInput{Type: req.Type, HandlerID: req.HandlerID, Note: req.Note}
}

func UpdateInput(req dto.UpdateOrderCaseRequest) types.UpdateOrderCaseInput {
	lines := make([]types.OrderCaseLineInput, len(req.Lines))
	for i, line := range req.Lines {
		lines[i] = types.OrderCaseLineInput{OrderItemID: line.OrderItemID, Quantity: line.Quantity}
	}
	return types.UpdateOrderCaseInput{Version: req.Version, Note: req.Note, HandlerID: req.HandlerID, Lines: lines}
}

func DecisionInput(req dto.OrderCaseDecisionRequest) types.OrderCaseDecisionInput {
	return types.OrderCaseDecisionInput{Version: req.Version, Note: req.Note}
}

func lines(values []types.OrderCaseLine) []dto.OrderCaseLineDTO {
	result := make([]dto.OrderCaseLineDTO, len(values))
	for i, line := range values {
		result[i] = dto.OrderCaseLineDTO{OrderItemID: line.OrderItemID, OriginalQuantity: line.OriginalQuantity, Quantity: line.Quantity}
	}
	return result
}

// Case maps a case with lines and history; empty collections serialize as [] rather than null.
func Case(value types.OrderCase) dto.OrderCaseDTO {
	result := dto.OrderCaseDTO{ID: value.ID, OrderID: value.OrderID, Type: value.Type, Status: value.Status, HandlerID: value.HandlerID, OpenedBy: value.OpenedBy, ProgressNote: value.ProgressNote, FinalizedBy: value.FinalizedBy, FinalizedAt: value.FinalizedAt, Version: value.Version, CreatedAt: value.CreatedAt, UpdatedAt: value.UpdatedAt, Lines: lines(value.Lines), Events: make([]dto.OrderCaseEventDTO, len(value.Events))}
	for i, event := range value.Events {
		result.Events[i] = dto.OrderCaseEventDTO{ID: event.ID, Version: event.Version, ActorID: event.ActorID, Action: event.Action, Status: event.Status, Note: event.Note, Lines: lines(event.Lines), CreatedAt: event.CreatedAt}
	}
	return result
}

// Lookup wraps an optional case.
func Lookup(value *types.OrderCase) dto.OrderCaseLookupResponse {
	if value == nil {
		return dto.OrderCaseLookupResponse{}
	}
	mapped := Case(*value)
	return dto.OrderCaseLookupResponse{Case: &mapped}
}
