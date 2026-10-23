package dto

import (
	"github.com/TruongHoang2004/Hexta/services/api/internal/core/types"
	"time"

	"github.com/shopspring/decimal"
)

// ChatMessage represents a single conversational turn.
type ChatMessage struct {
	Role    string `json:"role" example:"user"`
	Content string `json:"content" example:"Tôi muốn đặt 2 hộp sữa bắp"`
}

// AIConverseRequest represents the request payload for the conversational AI agent.
type AIConverseRequest struct {
	Prompt         string        `json:"prompt" binding:"required" validate:"required" example:"Tạo đơn cho anh Nam 0912345678 lấy 2 hộp sữa bắp"`
	ConversationID string        `json:"conversation_id,omitempty" example:"conv_12345"`
	History        []ChatMessage `json:"history,omitempty"`
}

// DraftItemDTO represents an item in an interactive order draft.
type DraftItemDTO struct {
	ProductID   string          `json:"product_id,omitempty" example:"prod_123"`
	ProductName string          `json:"product_name" validate:"required" example:"Sữa bắp non 500ml"`
	Quantity    int             `json:"quantity" validate:"required,min=1" example:"2"`
	UnitPrice   decimal.Decimal `json:"unit_price" example:"25000.00"`
	Variant     string          `json:"variant,omitempty" example:"Chai 500ml"`
	Subtotal    decimal.Decimal `json:"subtotal" example:"50000.00"`
}

// OrderDraftDTO represents a structured, Redis-cached order proposal awaiting human confirmation.
type OrderDraftDTO struct {
	DraftID         string          `json:"draft_id" example:"draft_8f3a9b21-4c5d"`
	TenantID        string          `json:"tenant_id" example:"tenant_123"`
	CustomerName    string          `json:"customer_name,omitempty" example:"Anh Nam"`
	CustomerPhone   string          `json:"customer_phone,omitempty" example:"0912345678"`
	ShippingAddress string          `json:"shipping_address,omitempty" example:"123 Lê Lợi, Q1, TP.HCM"`
	Items           []DraftItemDTO  `json:"items"`
	TotalAmount     decimal.Decimal `json:"total_amount" example:"50000.00"`
	PaymentMethod   string          `json:"payment_method" example:"cod"`
	Notes           string          `json:"notes,omitempty" example:"Giao hàng giờ hành chính"`
	Status          string          `json:"status" example:"proposed"`
	CreatedAt       time.Time       `json:"created_at"`
	ExpiresAt       time.Time       `json:"expires_at"`
}

// SSEEventType represents the type of Server-Sent Event emitted.
type SSEEventType string

const (
	SSEEventToken         SSEEventType = "token"
	SSEEventDraftProposed SSEEventType = "draft_proposed"
	SSEEventToolCall      SSEEventType = "tool_call"
	SSEEventError         SSEEventType = "error"
	SSEEventDone          SSEEventType = "done"
)

// SSEEvent represents a single Server-Sent Event message emitted during streaming conversation.
type SSEEvent struct {
	Type  SSEEventType   `json:"type"`
	Text  string         `json:"text,omitempty"`
	Draft *OrderDraftDTO `json:"draft,omitempty"`
	Tool  string         `json:"tool,omitempty"`
	Error string         `json:"error,omitempty"`
}

// DomainInput keeps the HTTP request contract outside the business service.
func (r *AIConverseRequest) DomainInput() *types.AIConversationInput {
	if r == nil {
		return nil
	}
	history := make([]types.AIChatMessage, len(r.History))
	for i, message := range r.History {
		history[i] = types.AIChatMessage{Role: message.Role, Content: message.Content}
	}
	return &types.AIConversationInput{Prompt: r.Prompt, ConversationID: r.ConversationID, History: history}
}

func AIOrderDraftResponse(draft *types.AIOrderDraft) *OrderDraftDTO {
	if draft == nil {
		return nil
	}
	var items []DraftItemDTO
	if draft.Items != nil {
		items = make([]DraftItemDTO, len(draft.Items))
	}
	for i, item := range draft.Items {
		items[i] = DraftItemDTO{ProductID: item.ProductID, ProductName: item.ProductName, Quantity: item.Quantity, UnitPrice: item.UnitPrice, Variant: item.Variant, Subtotal: item.Subtotal}
	}
	return &OrderDraftDTO{DraftID: draft.DraftID, TenantID: draft.TenantID, CustomerName: draft.CustomerName, CustomerPhone: draft.CustomerPhone, ShippingAddress: draft.ShippingAddress, Items: items, TotalAmount: draft.TotalAmount, PaymentMethod: draft.PaymentMethod, Notes: draft.Notes, Status: draft.Status, CreatedAt: draft.CreatedAt, ExpiresAt: draft.ExpiresAt}
}

func AIEventResponse(event *types.AIEvent) *SSEEvent {
	if event == nil {
		return nil
	}
	return &SSEEvent{Type: SSEEventType(event.Type), Text: event.Text, Draft: AIOrderDraftResponse(event.Draft), Tool: event.Tool, Error: event.Error}
}
