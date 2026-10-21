package controller

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	apperrors "github.com/TruongHoang2004/Hexta/packages/shared/pkg/errors"
	"github.com/TruongHoang2004/Hexta/services/api/internal/core/types"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
)

type aiControllerServiceStub struct {
	input   *types.AIConversationInput
	tenant  string
	failure *apperrors.Error
}

func (s *aiControllerServiceStub) Converse(ctx context.Context, tenant, user string, input *types.AIConversationInput, emit func(*types.AIEvent) error) *apperrors.Error {
	s.input, s.tenant = input, tenant
	if s.failure != nil {
		return s.failure
	}
	_ = emit(&types.AIEvent{Type: types.AIEventToken, Text: "Hello"})
	_ = emit(&types.AIEvent{Type: types.AIEventDraftProposed, Draft: &types.AIOrderDraft{DraftID: "draft", TenantID: tenant, Items: []types.AIDraftItem{{ProductName: "Item", Quantity: 2, UnitPrice: decimal.NewFromFloat(12.5), Subtotal: decimal.NewFromInt(25)}}, TotalAmount: decimal.NewFromInt(25)}})
	return nil
}

func (s *aiControllerServiceStub) GetDraft(context.Context, string, string) (*types.AIOrderDraft, *apperrors.Error) {
	return &types.AIOrderDraft{DraftID: "draft", TenantID: "tenant", TotalAmount: decimal.NewFromInt(25)}, nil
}

func TestAIControllerMapsDomainConversationToExistingSSEContract(t *testing.T) {
	gin.SetMode(gin.TestMode)
	stub := &aiControllerServiceStub{}
	ctrl := NewAIController(validator.New(), stub)
	router := gin.New()
	// Test-only route registration; dormant application routes remain unregistered.
	router.POST("/converse", ctrl.Converse)
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/converse", strings.NewReader(`{"prompt":"Hello","conversation_id":"conversation","history":[{"role":"assistant","content":"Previous"}]}`))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-Tenant-ID", "tenant")
	router.ServeHTTP(recorder, request)
	require.Equal(t, http.StatusOK, recorder.Code)
	require.Equal(t, "tenant", stub.tenant)
	require.Equal(t, "conversation", stub.input.ConversationID)
	require.Equal(t, "Previous", stub.input.History[0].Content)
	require.Contains(t, recorder.Header().Get("Content-Type"), "text/event-stream")
	chunks := strings.Split(strings.TrimSpace(recorder.Body.String()), "\n\n")
	require.Len(t, chunks, 2)
	var event struct {
		Type  string `json:"type"`
		Draft struct {
			DraftID string `json:"draft_id"`
			Total   string `json:"total_amount"`
			Items   []struct {
				Quantity int    `json:"quantity"`
				Price    string `json:"unit_price"`
			} `json:"items"`
		} `json:"draft"`
	}
	require.NoError(t, json.Unmarshal([]byte(strings.TrimPrefix(chunks[1], "data: ")), &event))
	require.Equal(t, "draft_proposed", event.Type)
	require.Equal(t, "draft", event.Draft.DraftID)
	require.Equal(t, "25", event.Draft.Total)
	require.Equal(t, "12.5", event.Draft.Items[0].Price)
	require.Equal(t, 2, event.Draft.Items[0].Quantity)
}

func TestAIControllerMapsDraftToStandardEnvelope(t *testing.T) {
	ctrl := NewAIController(validator.New(), &aiControllerServiceStub{})
	router := gin.New()
	router.GET("/drafts/:id", ctrl.GetDraft)
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/drafts/draft?tenant_id=tenant", nil)
	router.ServeHTTP(recorder, request)
	require.Equal(t, http.StatusOK, recorder.Code)
	var body struct {
		Data struct {
			DraftID string `json:"draft_id"`
			Total   string `json:"total_amount"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &body))
	require.Equal(t, "draft", body.Data.DraftID)
	require.Equal(t, "25", body.Data.Total)
}
