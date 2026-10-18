package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/TruongHoang2004/Hexta/packages/shared/pkg/errors"
	"github.com/TruongHoang2004/Hexta/services/api/common/log"
	"github.com/TruongHoang2004/Hexta/services/api/internal/core/model"
	"github.com/TruongHoang2004/Hexta/services/api/internal/infrastructure/gemini"
	"github.com/TruongHoang2004/Hexta/services/api/internal/present/http/dto"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

const (
	DraftTTL            = 30 * time.Minute
	DraftKeyPrefixFmt   = "hexta:draft:%s:%s" // hexta:draft:<tenant_id>:<draft_id>
	DraftRedisKeyPrefix = "hexta:draft:%s:%s"
)

type IDraftCache interface {
	Set(ctx context.Context, key string, value []byte, ttl time.Duration) error
	Get(ctx context.Context, key string) ([]byte, error)
	Delete(ctx context.Context, key string) error
}

type IAIAgentService interface {
	Converse(ctx context.Context, tenantID, userID string, req *dto.AIConverseRequest, emitEvent func(*dto.SSEEvent) error) error
	GetDraft(ctx context.Context, tenantID, draftID string) (*dto.OrderDraftDTO, *errors.Error)
}

type AIAgentService struct {
	*baseService
	db           *gorm.DB
	draftCache   IDraftCache
	geminiClient gemini.IGeminiClient
}

func NewAIAgentService(
	db *gorm.DB,
	draftCache IDraftCache,
	geminiClient gemini.IGeminiClient,
) *AIAgentService {
	return &AIAgentService{
		baseService:  NewBaseService(),
		db:           db,
		draftCache:   draftCache,
		geminiClient: geminiClient,
	}
}

func (s *AIAgentService) getToolDeclarations() []gemini.Tool {
	return []gemini.Tool{
		{
			FunctionDeclarations: []gemini.FunctionDeclaration{
				{
					Name:        "search_products",
					Description: "Search active product catalog items by name, barcode, or SKU with live prices and stock.",
					Parameters: map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"query": map[string]interface{}{
								"type":        "string",
								"description": "Keywords, product title, barcode, or SKU to search in catalog",
							},
						},
						"required": []string{"query"},
					},
				},
				{
					Name:        "check_stock_availability",
					Description: "Check available and on-hand inventory quantity for a specific SKU or variant before drafting order.",
					Parameters: map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"sku": map[string]interface{}{
								"type":        "string",
								"description": "The SKU code of the product variant",
							},
						},
						"required": []string{"sku"},
					},
				},
				{
					Name:        "lookup_customer",
					Description: "Search customer profile by phone number or full name.",
					Parameters: map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"phone": map[string]interface{}{
								"type":        "string",
								"description": "Customer phone number (e.g. 0912345678)",
							},
							"name": map[string]interface{}{
								"type":        "string",
								"description": "Customer full or partial name",
							},
						},
					},
				},
				{
					Name:        "propose_order_draft",
					Description: "Generates a structured order draft for customer purchase to be confirmed by user. Does NOT mutate order database directly.",
					Parameters: map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"customer_name": map[string]interface{}{
								"type":        "string",
								"description": "Customer full name",
							},
							"customer_phone": map[string]interface{}{
								"type":        "string",
								"description": "Customer phone number",
							},
							"shipping_address": map[string]interface{}{
								"type":        "string",
								"description": "Delivery shipping address",
							},
							"items": map[string]interface{}{
								"type": "array",
								"items": map[string]interface{}{
									"type": "object",
									"properties": map[string]interface{}{
										"product_id": map[string]interface{}{
											"type":        "string",
											"description": "UUID of product or variant if known",
										},
										"product_name": map[string]interface{}{
											"type":        "string",
											"description": "Product display name",
										},
										"variant": map[string]interface{}{
											"type":        "string",
											"description": "Variant or size specification",
										},
										"quantity": map[string]interface{}{
											"type":        "integer",
											"description": "Quantity purchased (must be >= 1)",
											"minimum":     1,
										},
										"unit_price": map[string]interface{}{
											"type":        "number",
											"description": "Unit selling price in VND",
											"minimum":     0,
										},
									},
									"required": []string{"product_name", "quantity", "unit_price"},
								},
							},
							"payment_method": map[string]interface{}{
								"type":        "string",
								"enum":        []string{"cod", "bank_transfer", "cash"},
								"description": "Selected payment method (default: cod)",
							},
							"notes": map[string]interface{}{
								"type":        "string",
								"description": "Optional order notes or delivery instructions",
							},
						},
						"required": []string{"items"},
					},
				},
			},
		},
	}
}

func (s *AIAgentService) getSystemInstruction() *gemini.Content {
	return &gemini.Content{
		Role: "system",
		Parts: []gemini.Part{
			{
				Text: `You are Hexta AI, an intelligent business operations assistant for SMEs.
You assist store owners and staff in managing orders, looking up products, checking stock, and serving customers.
Rules:
1. Always be professional, helpful, and concise in Vietnamese.
2. When a user requests to create or place an order, you MUST use the search_products and lookup_customer tools to resolve real items, prices, and customer info.
3. Then call propose_order_draft to generate a structured draft for the user to review and confirm.
4. Never assume stock or invent prices; always check the catalog.
5. You cannot directly create confirmed orders in the database; you only propose drafts.`,
			},
		},
	}
}

func (s *AIAgentService) toolSearchProducts(ctx context.Context, tenantID string, args map[string]interface{}) (map[string]interface{}, error) {
	return s.executeSearchProducts(ctx, tenantID, args)
}

func (s *AIAgentService) toolCheckStockAvailability(ctx context.Context, tenantID string, args map[string]interface{}) (map[string]interface{}, error) {
	return s.executeCheckStock(ctx, tenantID, args)
}

func (s *AIAgentService) toolLookupCustomer(ctx context.Context, tenantID string, args map[string]interface{}) (map[string]interface{}, error) {
	return s.executeLookupCustomer(ctx, tenantID, args)
}

func (s *AIAgentService) toolProposeOrderDraft(ctx context.Context, tenantID string, args map[string]interface{}, emitEvent func(*dto.SSEEvent) error) (map[string]interface{}, error) {
	return s.executeProposeOrderDraft(ctx, tenantID, args, emitEvent)
}

func (s *AIAgentService) executeSearchProducts(ctx context.Context, tenantID string, args map[string]interface{}) (map[string]interface{}, error) {
	query, _ := args["query"].(string)
	query = strings.TrimSpace(query)

	var products []model.Product
	dbQuery := s.db.WithContext(ctx).
		Preload("Variants").
		Where("tenant_id = ? AND status = ?", tenantID, model.ProductStatusActive)

	if query != "" {
		dbQuery = dbQuery.Where("name LIKE ? OR category LIKE ?", "%"+query+"%", "%"+query+"%")
	}

	if err := dbQuery.Limit(10).Find(&products).Error; err != nil {
		return nil, err
	}

	resultList := make([]map[string]interface{}, 0, len(products))
	for _, p := range products {
		var variants []map[string]interface{}
		for _, v := range p.Variants {
			variants = append(variants, map[string]interface{}{
				"variant_id":   v.ID,
				"sku":          v.SKU,
				"variant_name": v.VariantName,
				"price":        v.Price.String(),
			})
		}
		resultList = append(resultList, map[string]interface{}{
			"product_id":   p.ID,
			"product_name": p.Name,
			"category":     p.Category,
			"variants":     variants,
		})
	}

	return map[string]interface{}{
		"query":    query,
		"count":    len(resultList),
		"matches":  len(resultList),
		"products": resultList,
	}, nil
}

func (s *AIAgentService) executeCheckStock(ctx context.Context, tenantID string, args map[string]interface{}) (map[string]interface{}, error) {
	sku, _ := args["sku"].(string)
	sku = strings.TrimSpace(sku)
	if sku == "" {
		return map[string]interface{}{"found": false, "error": "sku is required"}, nil
	}

	var variant model.ProductVariant
	err := s.db.WithContext(ctx).
		Where("tenant_id = ? AND sku = ?", tenantID, sku).
		First(&variant).Error
	if err != nil {
		return map[string]interface{}{
			"found":         false,
			"sku":           sku,
			"available_qty": 0,
			"on_hand_qty":   0,
		}, nil
	}

	var inv model.InventoryItem
	invErr := s.db.WithContext(ctx).
		Where("tenant_id = ? AND variant_id = ?", tenantID, variant.ID).
		First(&inv).Error
	if invErr != nil {
		return map[string]interface{}{
			"found":         true,
			"sku":           sku,
			"variant_id":    variant.ID,
			"variant_name":  variant.VariantName,
			"available_qty": 0,
			"on_hand_qty":   0,
		}, nil
	}

	return map[string]interface{}{
		"found":         true,
		"sku":           sku,
		"variant_id":    variant.ID,
		"variant_name":  variant.VariantName,
		"available_qty": inv.AvailableQty,
		"on_hand_qty":   inv.OnHandQty,
		"reserved_qty":  inv.ReservedQty,
	}, nil
}

func (s *AIAgentService) executeLookupCustomer(ctx context.Context, tenantID string, args map[string]interface{}) (map[string]interface{}, error) {
	phone, _ := args["phone"].(string)
	name, _ := args["name"].(string)
	phone = strings.TrimSpace(phone)
	name = strings.TrimSpace(name)

	var customers []model.Customer
	dbQuery := s.db.WithContext(ctx).Where("tenant_id = ?", tenantID)

	if phone != "" {
		dbQuery = dbQuery.Where("phone LIKE ?", "%"+phone+"%")
	}
	if name != "" {
		dbQuery = dbQuery.Where("full_name LIKE ?", "%"+name+"%")
	}

	if err := dbQuery.Limit(5).Find(&customers).Error; err != nil {
		return nil, err
	}

	result := make([]map[string]interface{}, 0, len(customers))
	for _, c := range customers {
		result = append(result, map[string]interface{}{
			"customer_id": c.ID,
			"full_name":   c.FullName,
			"phone":       c.Phone,
			"address":     c.Address,
		})
	}

	return map[string]interface{}{
		"count":     len(result),
		"customers": result,
	}, nil
}

func (s *AIAgentService) executeProposeOrderDraft(
	ctx context.Context,
	tenantID string,
	args map[string]interface{},
	emitEvent func(*dto.SSEEvent) error,
) (map[string]interface{}, error) {
	customerName, _ := args["customer_name"].(string)
	customerPhone, _ := args["customer_phone"].(string)
	shippingAddress, _ := args["shipping_address"].(string)
	paymentMethod, _ := args["payment_method"].(string)
	if paymentMethod == "" {
		paymentMethod = "cod"
	}
	notes, _ := args["notes"].(string)

	rawItems, _ := args["items"].([]interface{})
	if len(rawItems) == 0 {
		return map[string]interface{}{"error": "items cannot be empty"}, nil
	}

	draftItems := make([]dto.DraftItemDTO, 0, len(rawItems))
	totalAmount := decimal.Zero

	for _, rawItem := range rawItems {
		itemMap, ok := rawItem.(map[string]interface{})
		if !ok {
			continue
		}

		productID, _ := itemMap["product_id"].(string)
		productName, _ := itemMap["product_name"].(string)
		variant, _ := itemMap["variant"].(string)

		qty := 1
		if qVal, ok := itemMap["quantity"].(float64); ok && qVal > 0 {
			qty = int(qVal)
		} else if qVal, ok := itemMap["quantity"].(int); ok && qVal > 0 {
			qty = qVal
		}

		unitPrice := decimal.Zero
		if pVal, ok := itemMap["unit_price"].(float64); ok {
			unitPrice = decimal.NewFromFloat(pVal)
		} else if pVal, ok := itemMap["unit_price"].(string); ok {
			unitPrice, _ = decimal.NewFromString(pVal)
		}

		subtotal := unitPrice.Mul(decimal.NewFromInt(int64(qty)))
		totalAmount = totalAmount.Add(subtotal)

		draftItems = append(draftItems, dto.DraftItemDTO{
			ProductID:   productID,
			ProductName: productName,
			Variant:     variant,
			Quantity:    qty,
			UnitPrice:   unitPrice,
			Subtotal:    subtotal,
		})
	}

	draftID := fmt.Sprintf("draft_%s", uuid.New().String())
	now := time.Now()
	expiresAt := now.Add(DraftTTL)

	draft := &dto.OrderDraftDTO{
		DraftID:         draftID,
		TenantID:        tenantID,
		CustomerName:    customerName,
		CustomerPhone:   customerPhone,
		ShippingAddress: shippingAddress,
		Items:           draftItems,
		TotalAmount:     totalAmount,
		PaymentMethod:   paymentMethod,
		Notes:           notes,
		Status:          "proposed",
		CreatedAt:       now,
		ExpiresAt:       expiresAt,
	}

	// Cache in Redis with 30-minute TTL
	cacheKey := fmt.Sprintf(DraftKeyPrefixFmt, tenantID, draftID)
	draftBytes, err := json.Marshal(draft)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal order draft: %w", err)
	}

	if s.draftCache != nil {
		if err := s.draftCache.Set(ctx, cacheKey, draftBytes, DraftTTL); err != nil {
			log.Warn(ctx, "failed to cache draft in Redis: %v", err)
		}
	}

	// Emit SSE draft_proposed event
	if emitEvent != nil {
		_ = emitEvent(&dto.SSEEvent{
			Type:  dto.SSEEventDraftProposed,
			Draft: draft,
		})
	}

	return map[string]interface{}{
		"status":   "draft_proposed",
		"draft_id": draftID,
		"total":    totalAmount.String(),
		"message":  "Order draft created and awaiting human confirmation in chat.",
	}, nil
}

func (s *AIAgentService) Converse(
	ctx context.Context,
	tenantID, userID string,
	req *dto.AIConverseRequest,
	emitEvent func(*dto.SSEEvent) error,
) error {
	if tenantID == "" {
		return fmt.Errorf("tenant_id is required")
	}

	tools := s.getToolDeclarations()
	sysInstruction := s.getSystemInstruction()

	var contents []gemini.Content

	// 1. Append past conversation history
	if req != nil {
		for _, msg := range req.History {
			role := msg.Role
			if role == "assistant" {
				role = "model"
			}
			contents = append(contents, gemini.Content{
				Role:  role,
				Parts: []gemini.Part{{Text: msg.Content}},
			})
		}

		// 2. Append current user message
		contents = append(contents, gemini.Content{
			Role:  "user",
			Parts: []gemini.Part{{Text: req.Prompt}},
		})
	}

	geminiReq := &gemini.GenerateContentRequest{
		Contents:          contents,
		SystemInstruction: sysInstruction,
		Tools:             tools,
	}

	// First pass: generate initial response or function calls
	res, err := s.geminiClient.GenerateContent(ctx, geminiReq)
	if err != nil {
		if emitEvent != nil {
			_ = emitEvent(&dto.SSEEvent{Type: dto.SSEEventError, Error: err.Error()})
		}
		return err
	}

	fc := res.FirstFunctionCall()
	if fc != nil {
		// Emit tool call event
		if emitEvent != nil {
			_ = emitEvent(&dto.SSEEvent{
				Type: dto.SSEEventToolCall,
				Tool: fc.Name,
			})
		}

		var toolResult map[string]interface{}
		var toolErr error

		switch fc.Name {
		case "search_products":
			toolResult, toolErr = s.executeSearchProducts(ctx, tenantID, fc.Args)
		case "check_stock_availability":
			toolResult, toolErr = s.executeCheckStock(ctx, tenantID, fc.Args)
		case "lookup_customer":
			toolResult, toolErr = s.executeLookupCustomer(ctx, tenantID, fc.Args)
		case "propose_order_draft":
			toolResult, toolErr = s.executeProposeOrderDraft(ctx, tenantID, fc.Args, emitEvent)
		default:
			toolResult = map[string]interface{}{"error": fmt.Sprintf("unsupported tool %s", fc.Name)}
		}

		if toolErr != nil {
			toolResult = map[string]interface{}{"error": toolErr.Error()}
		}

		// Append model turn with functionCall
		geminiReq.Contents = append(geminiReq.Contents, gemini.Content{
			Role: "model",
			Parts: []gemini.Part{
				{FunctionCall: fc},
			},
		})

		// Append function response turn
		geminiReq.Contents = append(geminiReq.Contents, gemini.Content{
			Role: "function",
			Parts: []gemini.Part{
				{
					FunctionResponse: &gemini.FunctionResponse{
						Name:     fc.Name,
						Response: toolResult,
					},
				},
			},
		})

		// Second pass: stream natural response explaining the result
		return s.geminiClient.StreamGenerateContent(ctx, geminiReq, func(chunk *gemini.GenerateContentResponse) error {
			text := chunk.FirstText()
			if text != "" && emitEvent != nil {
				return emitEvent(&dto.SSEEvent{
					Type: dto.SSEEventToken,
					Text: text,
				})
			}
			return nil
		})
	}

	// No function call: stream direct textual output
	text := res.FirstText()
	if text != "" && emitEvent != nil {
		_ = emitEvent(&dto.SSEEvent{
			Type: dto.SSEEventToken,
			Text: text,
		})
	}

	if emitEvent != nil {
		_ = emitEvent(&dto.SSEEvent{Type: dto.SSEEventDone})
	}

	return nil
}

func (s *AIAgentService) GetDraft(ctx context.Context, tenantID, draftID string) (*dto.OrderDraftDTO, *errors.Error) {
	if tenantID == "" || draftID == "" {
		return nil, errors.ErrBadRequest(ctx).SetDetail("tenant_id and draft_id are required")
	}

	if s.draftCache == nil {
		return nil, errors.ErrNotFound(ctx, "draft", "draft cache unavailable")
	}

	cacheKey := fmt.Sprintf(DraftRedisKeyPrefix, tenantID, draftID)
	data, err := s.draftCache.Get(ctx, cacheKey)
	if err != nil {
		return nil, errors.ErrNotFound(ctx, "draft", "draft not found or expired")
	}

	var draft dto.OrderDraftDTO
	if err := json.Unmarshal(data, &draft); err != nil {
		return nil, errors.ErrSystemError(ctx, "failed to parse draft payload")
	}

	if draft.TenantID != tenantID {
		return nil, errors.ErrForbidden(ctx).SetDetail("access to this draft is forbidden")
	}

	return &draft, nil
}
