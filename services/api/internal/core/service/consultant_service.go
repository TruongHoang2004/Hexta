package service

import (
	"context"
	"net/url"
	"strings"

	apperrors "github.com/TruongHoang2004/Hexta/packages/shared/pkg/errors"
	"github.com/TruongHoang2004/Hexta/services/api/internal/core/model"
	"github.com/TruongHoang2004/Hexta/services/api/internal/core/types"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type IConsultClient interface {
	Consult(context.Context, *types.ConsultRequest) (*types.ConsultResponse, error)
}
type consultantCatalog interface {
	Products(context.Context, string, string, int, int) ([]model.Product, int64, *apperrors.Error)
}
type ConsultantService struct {
	client  IConsultClient
	catalog consultantCatalog
}

func NewConsultantService(client IConsultClient, sales *SalesService) *ConsultantService {
	return &ConsultantService{client: client, catalog: sales}
}

func (s *ConsultantService) Converse(ctx context.Context, tenant, actor string, input *types.AIConversationInput, emit func(*types.AIEvent) error) *apperrors.Error {
	if input == nil || strings.TrimSpace(input.Prompt) == "" || len(input.Prompt) > 4000 || len(input.History) > 40 || len(input.ConversationID) > 128 {
		return apperrors.ErrBadRequest(ctx).SetDetail("invalid conversation input")
	}
	for _, h := range input.History {
		if (h.Role != "user" && h.Role != "assistant") || len(h.Content) > 4000 {
			return apperrors.ErrBadRequest(ctx).SetDetail("invalid history")
		}
	}
	// This read checks current tenant membership before any external call; no DB transaction spans the call.
	products, _, failure := s.catalog.Products(ctx, tenant, actor, 1, 100)
	if failure != nil {
		return failure
	}
	session := input.ConversationID
	if session == "" {
		session = uuid.NewString()
	}
	candidates := make([]types.ConsultProduct, 0)
	for _, p := range products {
		for _, v := range p.Variants {
			if v.PriceProvided {
				candidates = append(candidates, consultantProduct(p, v))
			}
		}
	}
	result, err := s.client.Consult(ctx, &types.ConsultRequest{TenantID: tenant, SessionID: tenant + ":" + actor + ":" + session, Message: input.Prompt, History: input.History, CustomerContext: map[string]interface{}{"products": candidates}})
	if err != nil || result == nil {
		return apperrors.ErrSystemError(ctx, "AI consultant unavailable. Please try again.")
	}
	// Recheck current access/catalog after the external call. All display prices and identifiers come from Go.
	products, _, failure = s.catalog.Products(ctx, tenant, actor, 1, 100)
	if failure != nil {
		return failure
	}
	verified := make(map[string]types.ConsultProduct)
	for _, p := range products {
		for _, v := range p.Variants {
			if v.PriceProvided {
				item := consultantProduct(p, v)
				verified[v.ID] = item
				if len(p.Variants) == 1 {
					verified[p.ID] = item
				}
			}
		}
	}
	recommendations := make([]types.ConsultProduct, 0)
	seenVariants := make(map[string]bool)
	for _, p := range result.RecommendedProducts {
		key := p.VariantID
		if key == "" {
			key = p.ProductID
		}
		if item, ok := verified[key]; ok && !seenVariants[item.VariantID] {
			seenVariants[item.VariantID] = true
			if image, err := url.Parse(p.ImageURL); err == nil && image.Scheme == "https" && image.Host != "" && image.User == nil {
				item.ImageURL = image.String()
			}
			recommendations = append(recommendations, item)
		}
	}
	var draft *types.AIOrderDraft
	if proposed := result.OrderDraft; proposed != nil && len(proposed.Items) > 0 && len(proposed.Items) <= 50 {
		items := make([]types.AIDraftItem, 0)
		total := decimal.Zero
		valid := true
		for _, item := range proposed.Items {
			key := item.VariantID
			if key == "" {
				key = item.ProductID
			}
			product, ok := verified[key]
			if !ok || item.Quantity < 1 || item.Quantity > 100 {
				valid = false
				break
			}
			subtotal := product.Price.Mul(decimal.NewFromInt(int64(item.Quantity)))
			total = total.Add(subtotal)
			items = append(items, types.AIDraftItem{ProductID: product.ProductID, VariantID: product.VariantID, Variant: product.Variant, ProductName: product.ProductName, Quantity: item.Quantity, UnitPrice: product.Price, Subtotal: subtotal})
		}
		if valid {
			draft = &types.AIOrderDraft{DraftID: uuid.NewString(), TenantID: tenant, Items: items, TotalAmount: total, Status: "proposed"}
		}
	}
	events := []*types.AIEvent{{Type: "message", Text: result.Message, NeedsHuman: result.NeedsHuman}, {Type: "products", Products: recommendations}}
	if draft != nil {
		events = append(events, &types.AIEvent{Type: "draft_order", Draft: draft})
	}
	events = append(events, &types.AIEvent{Type: "quick_replies", QuickReplies: result.QuickReplies}, &types.AIEvent{Type: types.AIEventDone})
	for _, event := range events {
		if err := ctx.Err(); err != nil {
			return apperrors.ErrSystemError(ctx, "consultation cancelled")
		}
		if emit != nil {
			if err := emit(event); err != nil {
				return apperrors.ErrSystemError(ctx, "consultation interrupted")
			}
		}
	}
	return nil
}

// The legacy cache API is intentionally not exposed for consultant candidates.
func (s *ConsultantService) GetDraft(ctx context.Context, tenant, id string) (*types.AIOrderDraft, *apperrors.Error) {
	return nil, apperrors.ErrNotFound(ctx, "draft", "draft unavailable")
}

// Keep catalog variant identity and its human-readable label together at both reads.
func consultantProduct(product model.Product, variant model.ProductVariant) types.ConsultProduct {
	name := product.Name
	label := strings.TrimSpace(variant.VariantName)
	if label != "" {
		name += " (" + label + ")"
	}
	return types.ConsultProduct{ProductID: product.ID, ProductName: name, VariantID: variant.ID, Variant: label, Price: variant.Price}
}
