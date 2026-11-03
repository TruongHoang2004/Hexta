package types

import "github.com/shopspring/decimal"

// ConsultRequest carries only server-authorized context to the internal AI service.
type ConsultRequest struct {
	TenantID        string                 `json:"tenant_id"`
	SessionID       string                 `json:"session_id"`
	Message         string                 `json:"message"`
	History         []AIChatMessage        `json:"history"`
	CustomerContext map[string]interface{} `json:"customer_context"`
}
type ConsultProduct struct {
	ProductID   string          `json:"product_id"`
	ProductName string          `json:"product_name"`
	Variant     string          `json:"variant,omitempty"`
	VariantID   string          `json:"variant_id,omitempty"`
	Price       decimal.Decimal `json:"price"`
	ImageURL    string          `json:"image_url,omitempty"`
}
type ConsultResponse struct {
	Message             string           `json:"message"`
	Intent              string           `json:"intent"`
	RecommendedProducts []ConsultProduct `json:"recommended_products"`
	OrderDraft          *AIOrderDraft    `json:"order_draft"`
	QuickReplies        []string         `json:"quick_replies"`
	NeedsHuman          bool             `json:"needs_human"`
}
