package service

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/TruongHoang2004/Hexta/services/api/internal/core/model"
	"github.com/TruongHoang2004/Hexta/services/api/internal/infrastructure/gemini"
	"github.com/TruongHoang2004/Hexta/services/api/internal/present/http/dto"
)

type mockMemoryRedis struct {
	store map[string][]byte
}

func newMockMemoryRedis() *mockMemoryRedis {
	return &mockMemoryRedis{store: make(map[string][]byte)}
}

func (m *mockMemoryRedis) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	m.store[key] = value
	return nil
}

func (m *mockMemoryRedis) Get(ctx context.Context, key string) ([]byte, error) {
	val, ok := m.store[key]
	if !ok {
		return nil, fmt.Errorf("key not found: %s", key)
	}
	return val, nil
}

func (m *mockMemoryRedis) Delete(ctx context.Context, key string) error {
	delete(m.store, key)
	return nil
}

type mockAIGeminiClient struct {
	generateFunc func(req *gemini.GenerateContentRequest) (*gemini.GenerateContentResponse, error)
	streamFunc   func(req *gemini.GenerateContentRequest, cb func(*gemini.GenerateContentResponse) error) error
}

func (m *mockAIGeminiClient) GenerateContent(ctx context.Context, req *gemini.GenerateContentRequest) (*gemini.GenerateContentResponse, error) {
	if m.generateFunc != nil {
		return m.generateFunc(req)
	}
	return &gemini.GenerateContentResponse{}, nil
}

func (m *mockAIGeminiClient) StreamGenerateContent(ctx context.Context, req *gemini.GenerateContentRequest, cb func(*gemini.GenerateContentResponse) error) error {
	if m.streamFunc != nil {
		return m.streamFunc(req, cb)
	}
	return cb(&gemini.GenerateContentResponse{
		Candidates: []gemini.Candidate{
			{Content: gemini.Content{Parts: []gemini.Part{{Text: "Đã xong."}}}},
		},
	})
}

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	err = db.AutoMigrate(
		&model.Product{},
		&model.ProductVariant{},
		&model.InventoryItem{},
		&model.Customer{},
		&model.Order{},
		&model.OrderItem{},
	)
	require.NoError(t, err)
	return db
}

func TestAIAgentService_ZeroDBMutations(t *testing.T) {
	db := setupTestDB(t)
	redisMock := newMockMemoryRedis()

	tenantID := "tenant-alpha"
	userID := "user-staff"

	turn := 0
	geminiMock := &mockAIGeminiClient{
		generateFunc: func(req *gemini.GenerateContentRequest) (*gemini.GenerateContentResponse, error) {
			turn++
			if turn == 1 {
				return &gemini.GenerateContentResponse{
					Candidates: []gemini.Candidate{
						{
							Content: gemini.Content{
								Role: "model",
								Parts: []gemini.Part{
									{
										FunctionCall: &gemini.FunctionCall{
											Name: "propose_order_draft",
											Args: map[string]interface{}{
												"customer_name":  "Nguyen Van A",
												"customer_phone": "0912345678",
												"items": []interface{}{
													map[string]interface{}{
														"product_name": "Sữa bắp non",
														"quantity":     float64(2),
														"unit_price":   float64(25000),
													},
												},
												"payment_method": "cod",
											},
										},
									},
								},
							},
						},
					},
				}, nil
			}
			return &gemini.GenerateContentResponse{
				Candidates: []gemini.Candidate{
					{
						Content: gemini.Content{
							Parts: []gemini.Part{{Text: "Em đã tạo bản nháp đơn hàng rồi ạ."}},
						},
					},
				},
			}, nil
		},
	}

	aiService := NewAIAgentService(db, redisMock, geminiMock)

	// Check orders count BEFORE converse
	var countBefore int64
	db.Model(&model.Order{}).Count(&countBefore)
	assert.Equal(t, int64(0), countBefore)

	var emittedEvents []*dto.SSEEvent
	err := aiService.Converse(context.Background(), tenantID, userID, &dto.AIConverseRequest{
		Prompt: "Tạo đơn cho anh A 2 hộp sữa bắp",
	}, func(event *dto.SSEEvent) error {
		emittedEvents = append(emittedEvents, event)
		return nil
	})

	require.NoError(t, err)

	// Invariant: ZERO direct database mutations executed by AI Agent
	var countAfter int64
	db.Model(&model.Order{}).Count(&countAfter)
	assert.Equal(t, int64(0), countAfter, "AI Agent must NEVER insert into real orders table directly")

	var itemCountAfter int64
	db.Model(&model.OrderItem{}).Count(&itemCountAfter)
	assert.Equal(t, int64(0), itemCountAfter, "AI Agent must NEVER insert into real order_items table directly")

	// Verify events were emitted properly
	var foundDraftEvent bool
	var proposedDraftID string
	for _, ev := range emittedEvents {
		if ev.Type == dto.SSEEventDraftProposed {
			foundDraftEvent = true
			require.NotNil(t, ev.Draft)
			proposedDraftID = ev.Draft.DraftID
			assert.Equal(t, tenantID, ev.Draft.TenantID)
			assert.Equal(t, "Nguyen Van A", ev.Draft.CustomerName)
			assert.True(t, ev.Draft.TotalAmount.Equal(decimal.NewFromInt(50000)), "TotalAmount must be 50000")
			assert.Equal(t, 1, len(ev.Draft.Items))
			assert.Equal(t, 2, ev.Draft.Items[0].Quantity)
		}
	}
	assert.True(t, foundDraftEvent, "SSE stream must emit draft_proposed event")

	// Verify draft was cached in Redis with 30m TTL
	cachedDraft, err := aiService.GetDraft(context.Background(), tenantID, proposedDraftID)
	require.Nil(t, err)
	require.NotNil(t, cachedDraft)
	assert.Equal(t, proposedDraftID, cachedDraft.DraftID)
	assert.Equal(t, "proposed", cachedDraft.Status)
}

func TestAIAgentService_ToolSearchProducts_MultiTenant(t *testing.T) {
	db := setupTestDB(t)
	redisMock := newMockMemoryRedis()

	tenantAlpha := "tenant-alpha"
	tenantBeta := "tenant-beta"

	// Seed catalog in tenant-alpha
	prodAlpha := model.Product{
		ID:       "p-alpha-1",
		TenantID: tenantAlpha,
		Name:     "Sữa bắp non",
		Status:   model.ProductStatusActive,
	}
	db.Create(&prodAlpha)

	variantAlpha := model.ProductVariant{
		ID:          "pv-alpha-1",
		TenantID:    tenantAlpha,
		ProductID:   prodAlpha.ID,
		SKU:         "SKU-SUA-BAP",
		VariantName: "Chai 500ml",
		Price:       decimal.NewFromInt(25000),
		Barcode:     "8930001",
	}
	db.Create(&variantAlpha)

	invAlpha := model.InventoryItem{
		ID:           "inv-1",
		TenantID:     tenantAlpha,
		VariantID:    variantAlpha.ID,
		OnHandQty:    50,
		AvailableQty: 45,
	}
	db.Create(&invAlpha)

	// Seed catalog in tenant-beta (should be isolated!)
	prodBeta := model.Product{
		ID:       "p-beta-1",
		TenantID: tenantBeta,
		Name:     "Sữa bắp non Beta",
		Status:   model.ProductStatusActive,
	}
	db.Create(&prodBeta)

	geminiMock := &mockAIGeminiClient{}
	aiService := NewAIAgentService(db, redisMock, geminiMock)

	// Execute search_products in tenantAlpha
	resAlpha, err := aiService.toolSearchProducts(context.Background(), tenantAlpha, map[string]interface{}{
		"query": "sữa bắp",
	})
	require.NoError(t, err)
	assert.Equal(t, 1, resAlpha["matches"])

	// Check stock availability in tenantAlpha
	stockRes, err := aiService.toolCheckStockAvailability(context.Background(), tenantAlpha, map[string]interface{}{
		"sku": "SKU-SUA-BAP",
	})
	require.NoError(t, err)
	assert.Equal(t, true, stockRes["found"])
	assert.Equal(t, 45, stockRes["available_qty"])

	// Check stock availability for tenantAlpha SKU from tenantBeta context -> should NOT find it
	stockResBeta, err := aiService.toolCheckStockAvailability(context.Background(), tenantBeta, map[string]interface{}{
		"sku": "SKU-SUA-BAP",
	})
	require.NoError(t, err)
	assert.Equal(t, false, stockResBeta["found"], "Tenant beta must not see tenant alpha SKU")
}

func TestAIAgentService_TenantDraftIsolation(t *testing.T) {
	db := setupTestDB(t)
	redisMock := newMockMemoryRedis()
	aiService := NewAIAgentService(db, redisMock, &mockAIGeminiClient{})

	// Store draft for tenant-alpha
	draftID := "draft_isolated_123"
	draft := dto.OrderDraftDTO{
		DraftID:  draftID,
		TenantID: "tenant-alpha",
		Status:   "proposed",
	}
	b, _ := json.Marshal(draft)
	_ = redisMock.Set(context.Background(), fmt.Sprintf(DraftRedisKeyPrefix, "tenant-alpha", draftID), b, 30*time.Minute)

	// Tenant alpha can access
	fetched, err := aiService.GetDraft(context.Background(), "tenant-alpha", draftID)
	require.Nil(t, err)
	assert.Equal(t, draftID, fetched.DraftID)

	// Tenant beta cannot access
	_, errBeta := aiService.GetDraft(context.Background(), "tenant-beta", draftID)
	assert.NotNil(t, errBeta)
}
