package types

import (
	"time"

	"github.com/shopspring/decimal"
)

// AIChatMessage represents a single conversational turn.
type AIChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// AIConversationInput represents the request payload for the conversational AI agent.
type AIConversationInput struct {
	Prompt         string          `json:"prompt"`
	ConversationID string          `json:"conversation_id,omitempty"`
	History        []AIChatMessage `json:"history,omitempty"`
}

// AIDraftItem represents an item in an interactive order draft.
type AIDraftItem struct {
	ProductID   string          `json:"product_id,omitempty"`
	ProductName string          `json:"product_name"`
	Quantity    int             `json:"quantity"`
	UnitPrice   decimal.Decimal `json:"unit_price"`
	Variant     string          `json:"variant,omitempty"`
	Subtotal    decimal.Decimal `json:"subtotal"`
}

// AIOrderDraft represents a structured, Redis-cached order proposal awaiting human confirmation.
type AIOrderDraft struct {
	DraftID         string          `json:"draft_id"`
	TenantID        string          `json:"tenant_id"`
	CustomerName    string          `json:"customer_name,omitempty"`
	CustomerPhone   string          `json:"customer_phone,omitempty"`
	ShippingAddress string          `json:"shipping_address,omitempty"`
	Items           []AIDraftItem   `json:"items"`
	TotalAmount     decimal.Decimal `json:"total_amount"`
	PaymentMethod   string          `json:"payment_method"`
	Notes           string          `json:"notes,omitempty"`
	Status          string          `json:"status"`
	CreatedAt       time.Time       `json:"created_at"`
	ExpiresAt       time.Time       `json:"expires_at"`
}

// AIEventType represents the type of Server-Sent Event emitted.
type AIEventType string

const (
	AIEventToken         AIEventType = "token"
	AIEventDraftProposed AIEventType = "draft_proposed"
	AIEventToolCall      AIEventType = "tool_call"
	AIEventError         AIEventType = "error"
	AIEventDone          AIEventType = "done"
)

// AIEvent represents a single Server-Sent Event message emitted during streaming conversation.
type AIEvent struct {
	Type  AIEventType   `json:"type"`
	Text  string        `json:"text,omitempty"`
	Draft *AIOrderDraft `json:"draft,omitempty"`
	Tool  string        `json:"tool,omitempty"`
	Error string        `json:"error,omitempty"`
}
