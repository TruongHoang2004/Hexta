package controller

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	apperrors "github.com/TruongHoang2004/Hexta/packages/shared/pkg/errors"
	"github.com/TruongHoang2004/Hexta/services/api/common/log"
	"github.com/TruongHoang2004/Hexta/services/api/config"
	"github.com/TruongHoang2004/Hexta/services/api/internal/core/types"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
)

type aiControllerServiceStub struct {
	events  []*types.AIEvent
	input   *types.AIConversationInput
	tenant  string
	failure *apperrors.Error
}

func (s *aiControllerServiceStub) Converse(ctx context.Context, tenant, user string, input *types.AIConversationInput, emit func(*types.AIEvent) error) *apperrors.Error {
	s.input, s.tenant = input, tenant
	if s.events != nil {
		for _, event := range s.events {
			if err := emit(event); err != nil {
				return apperrors.ErrSystemError(ctx, "interrupted")
			}
		}
		return nil
	}
	if s.failure != nil {
		return s.failure
	}
	_ = emit(&types.AIEvent{Type: types.AIEventToken, Text: "Hello"})
	_ = emit(&types.AIEvent{Type: types.AIEventDraftProposed, Draft: &types.AIOrderDraft{DraftID: "draft", TenantID: tenant, Items: []types.AIDraftItem{{ProductName: "Item", Quantity: 2, UnitPrice: decimal.NewFromFloat(12.5), Subtotal: decimal.NewFromInt(25)}}, TotalAmount: decimal.NewFromInt(25)}})
	return nil
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
	require.NoError(t, json.Unmarshal([]byte(strings.SplitN(chunks[1], "data: ", 2)[1]), &event))
	require.Equal(t, "draft_proposed", event.Type)
	require.Equal(t, "draft", event.Draft.DraftID)
	require.Equal(t, "25", event.Draft.Total)
	require.Equal(t, "12.5", event.Draft.Items[0].Price)
	require.Equal(t, 2, event.Draft.Items[0].Quantity)
}

func TestConsultantNamedEventsAndValidation(t *testing.T) {
	previous := config.AppConfig
	config.AppConfig = &config.Config{}
	log.GetLogger()
	t.Cleanup(func() { config.AppConfig = previous })
	for _, body := range []string{`{"prompt":"Hello","history":[{"role":"system","content":"forged"}]}`, `{"prompt":""}`} {
		stub := &aiControllerServiceStub{}
		ctrl := NewAIController(validator.New(), stub)
		engine := gin.New()
		engine.POST("/converse", ctrl.Converse)
		response := httptest.NewRecorder()
		request := httptest.NewRequest("POST", "/converse", strings.NewReader(body))
		request.Header.Set("X-Tenant-ID", "tenant")
		request.Header.Set("Content-Type", "application/json")
		engine.ServeHTTP(response, request)
		require.Equal(t, 400, response.Code)
		require.Nil(t, stub.input)
	}
	stub := &aiControllerServiceStub{events: []*types.AIEvent{{Type: "message", Text: "Hello\nworld"}, {Type: "products", Products: []types.ConsultProduct{}}, {Type: "draft_order", Draft: &types.AIOrderDraft{DraftID: "draft"}}, {Type: "quick_replies", QuickReplies: []string{"Buy"}}, {Type: "done"}}}
	ctrl := NewAIController(validator.New(), stub)
	engine := gin.New()
	engine.POST("/converse", ctrl.Converse)
	response := httptest.NewRecorder()
	request := httptest.NewRequest("POST", "/converse", strings.NewReader(`{"prompt":"Hello"}`))
	request.Header.Set("X-Tenant-ID", "tenant")
	engine.ServeHTTP(response, request)
	require.Equal(t, 200, response.Code)
	for _, name := range []string{"message", "products", "draft_order", "quick_replies", "done"} {
		require.Contains(t, response.Body.String(), "event: "+name+"\n")
	}
}
