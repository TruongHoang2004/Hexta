package service

import (
	"context"
	"fmt"
	apperrors "github.com/TruongHoang2004/Hexta/packages/shared/pkg/errors"
	"github.com/TruongHoang2004/Hexta/services/api/internal/core/model"
	"github.com/TruongHoang2004/Hexta/services/api/internal/core/types"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
	"testing"
)

type consultCatalogStub struct {
	calls    int
	deniedAt int
}

func (c *consultCatalogStub) Products(ctx context.Context, tenant, actor string, page, size int) ([]model.Product, int64, *apperrors.Error) {
	c.calls++
	if actor != "member" || tenant != "tenant" || c.calls == c.deniedAt {
		return nil, 0, apperrors.ErrForbidden(ctx)
	}
	return []model.Product{{ID: "product", Name: "Real product", Variants: []model.ProductVariant{{ID: "variant", PriceProvided: true, Price: decimal.NewFromInt(25)}}}}, 1, nil
}

type consultClientStub struct {
	calls  int
	fail   bool
	result *types.ConsultResponse
}

func (c *consultClientStub) Consult(ctx context.Context, r *types.ConsultRequest) (*types.ConsultResponse, error) {
	c.calls++
	if c.fail {
		return nil, fmt.Errorf("private provider failure")
	}
	return c.result, nil
}
func TestConsultantAuthorityAndCandidateValidation(t *testing.T) {
	for _, tc := range []struct {
		name, tenant, actor string
		deniedAt            int
	}{{"wrong tenant", "other", "member", 0}, {"wrong actor", "tenant", "stranger", 0}, {"revoked during call", "tenant", "member", 2}} {
		t.Run(tc.name, func(t *testing.T) {
			client := &consultClientStub{result: &types.ConsultResponse{}}
			catalog := &consultCatalogStub{deniedAt: tc.deniedAt}
			s := &ConsultantService{client: client, catalog: catalog}
			emitted := 0
			err := s.Converse(context.Background(), tc.tenant, tc.actor, &types.AIConversationInput{Prompt: "hello"}, func(*types.AIEvent) error { emitted++; return nil })
			require.NotNil(t, err)
			require.Zero(t, emitted)
			if tc.deniedAt == 0 {
				require.Zero(t, client.calls)
			}
		})
	}
	client := &consultClientStub{result: &types.ConsultResponse{Message: "Hello", RecommendedProducts: []types.ConsultProduct{{ProductID: "foreign"}, {ProductID: "product", Price: decimal.NewFromInt(1)}}, OrderDraft: &types.AIOrderDraft{Items: []types.AIDraftItem{{ProductID: "variant", Quantity: 2, UnitPrice: decimal.NewFromInt(1)}}}}}
	s := &ConsultantService{client: client, catalog: &consultCatalogStub{}}
	var events []*types.AIEvent
	require.Nil(t, s.Converse(context.Background(), "tenant", "member", &types.AIConversationInput{Prompt: "hello"}, func(e *types.AIEvent) error { events = append(events, e); return nil }))
	require.Len(t, events[1].Products, 1)
	require.Equal(t, "25", events[1].Products[0].Price.String())
	require.Equal(t, "50", events[2].Draft.TotalAmount.String())
	require.Equal(t, types.AIEventDone, events[len(events)-1].Type)
	require.NotNil(t, s.Converse(context.Background(), "tenant", "member", &types.AIConversationInput{Prompt: "hello"}, func(*types.AIEvent) error { return fmt.Errorf("disconnected") }))
	client.result.OrderDraft.Items[0].ProductID = "foreign"
	events = nil
	require.Nil(t, s.Converse(context.Background(), "tenant", "member", &types.AIConversationInput{Prompt: "hello"}, func(e *types.AIEvent) error { events = append(events, e); return nil }))
	for _, e := range events {
		require.NotEqual(t, types.AIEventType("draft_order"), e.Type)
	}
	client.fail = true
	err := s.Converse(context.Background(), "tenant", "member", &types.AIConversationInput{Prompt: "hello"}, nil)
	require.NotNil(t, err)
	require.NotContains(t, err.GetDetail(), "private")
}
