package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/TruongHoang2004/Hexta/packages/shared/pkg/errors"
	"github.com/TruongHoang2004/Hexta/services/api/internal/core/model"
	"github.com/TruongHoang2004/Hexta/services/api/internal/repository"
)

type IInventoryService interface {
	GetInventory(ctx context.Context, tenantID, variantID string) (*model.InventoryItem, *errors.Error)
	GetInventoryByID(ctx context.Context, tenantID, id string) (*model.InventoryItem, *errors.Error)
	InitInventory(ctx context.Context, tenantID, variantID string, onHand, safetyThreshold int) (*model.InventoryItem, *errors.Error)
	CheckStock(ctx context.Context, tenantID, variantID string, quantity int) (bool, *model.InventoryItem, *errors.Error)
	ReserveStock(ctx context.Context, tenantID, variantID string, quantity int) (*model.InventoryItem, *errors.Error)
	ReleaseStock(ctx context.Context, tenantID, variantID string, quantity int) (*model.InventoryItem, *errors.Error)
	DeductStock(ctx context.Context, tenantID, variantID string, quantity int, refType string, refID *string) (*model.InventoryItem, *errors.Error)
	AdjustStock(ctx context.Context, tenantID, variantID string, newOnHand int, notes string) (*model.InventoryItem, *errors.Error)
	ListMovements(ctx context.Context, tenantID, inventoryID string, limit, offset int) ([]*model.StockMovement, int64, *errors.Error)
}

type InventoryService struct {
	*baseService
	inventoryRepo repository.IInventoryRepository
}

func NewInventoryService(inventoryRepo repository.IInventoryRepository) *InventoryService {
	return &InventoryService{
		baseService:   NewBaseService(),
		inventoryRepo: inventoryRepo,
	}
}

func (s *InventoryService) GetInventory(ctx context.Context, tenantID, variantID string) (*model.InventoryItem, *errors.Error) {
	if tenantID == "" || variantID == "" {
		return nil, errors.ErrBadRequest(ctx).SetDetail("tenant_id and variant_id are required")
	}
	return s.inventoryRepo.GetByVariantID(ctx, tenantID, variantID)
}

func (s *InventoryService) GetInventoryByID(ctx context.Context, tenantID, id string) (*model.InventoryItem, *errors.Error) {
	if tenantID == "" || id == "" {
		return nil, errors.ErrBadRequest(ctx).SetDetail("tenant_id and id are required")
	}
	return s.inventoryRepo.GetByID(ctx, tenantID, id)
}

func (s *InventoryService) InitInventory(ctx context.Context, tenantID, variantID string, onHand, safetyThreshold int) (*model.InventoryItem, *errors.Error) {
	if tenantID == "" || variantID == "" {
		return nil, errors.ErrBadRequest(ctx).SetDetail("tenant_id and variant_id are required")
	}
	if onHand < 0 {
		return nil, errors.ErrBadRequest(ctx).SetDetail("initial on-hand stock cannot be negative")
	}
	if safetyThreshold < 0 {
		safetyThreshold = 5
	}

	existing, _ := s.inventoryRepo.GetByVariantID(ctx, tenantID, variantID)
	if existing != nil {
		return nil, errors.ErrConflict(ctx, "inventory item", "already exists for this variant")
	}

	item := &model.InventoryItem{
		ID:              uuid.New().String(),
		TenantID:        tenantID,
		VariantID:       variantID,
		OnHandQty:       onHand,
		ReservedQty:     0,
		AvailableQty:    onHand,
		SafetyThreshold: safetyThreshold,
	}

	return s.inventoryRepo.CreateInventoryItem(ctx, item)
}

func (s *InventoryService) CheckStock(ctx context.Context, tenantID, variantID string, quantity int) (bool, *model.InventoryItem, *errors.Error) {
	if tenantID == "" || variantID == "" {
		return false, nil, errors.ErrBadRequest(ctx).SetDetail("tenant_id and variant_id are required")
	}
	if quantity <= 0 {
		return false, nil, errors.ErrBadRequest(ctx).SetDetail("quantity must be greater than zero")
	}

	item, err := s.inventoryRepo.GetByVariantID(ctx, tenantID, variantID)
	if err != nil {
		return false, nil, err
	}

	isAvailable := item.AvailableQty >= quantity
	return isAvailable, item, nil
}

func (s *InventoryService) ReserveStock(ctx context.Context, tenantID, variantID string, quantity int) (*model.InventoryItem, *errors.Error) {
	if tenantID == "" || variantID == "" {
		return nil, errors.ErrBadRequest(ctx).SetDetail("tenant_id and variant_id are required")
	}
	if quantity <= 0 {
		return nil, errors.ErrBadRequest(ctx).SetDetail("quantity must be greater than zero")
	}

	return s.inventoryRepo.ReserveStock(ctx, tenantID, variantID, quantity)
}

func (s *InventoryService) ReleaseStock(ctx context.Context, tenantID, variantID string, quantity int) (*model.InventoryItem, *errors.Error) {
	if tenantID == "" || variantID == "" {
		return nil, errors.ErrBadRequest(ctx).SetDetail("tenant_id and variant_id are required")
	}
	if quantity <= 0 {
		return nil, errors.ErrBadRequest(ctx).SetDetail("quantity must be greater than zero")
	}

	return s.inventoryRepo.ReleaseStock(ctx, tenantID, variantID, quantity)
}

func (s *InventoryService) DeductStock(ctx context.Context, tenantID, variantID string, quantity int, refType string, refID *string) (*model.InventoryItem, *errors.Error) {
	if tenantID == "" || variantID == "" {
		return nil, errors.ErrBadRequest(ctx).SetDetail("tenant_id and variant_id are required")
	}
	if quantity <= 0 {
		return nil, errors.ErrBadRequest(ctx).SetDetail("quantity must be greater than zero")
	}

	return s.inventoryRepo.DeductStock(ctx, tenantID, variantID, quantity, refType, refID)
}

func (s *InventoryService) AdjustStock(ctx context.Context, tenantID, variantID string, newOnHand int, notes string) (*model.InventoryItem, *errors.Error) {
	if tenantID == "" || variantID == "" {
		return nil, errors.ErrBadRequest(ctx).SetDetail("tenant_id and variant_id are required")
	}
	if newOnHand < 0 {
		return nil, errors.ErrBadRequest(ctx).SetDetail("new on-hand quantity cannot be negative")
	}

	return s.inventoryRepo.AdjustStock(ctx, tenantID, variantID, newOnHand, notes)
}

func (s *InventoryService) ListMovements(ctx context.Context, tenantID, inventoryID string, limit, offset int) ([]*model.StockMovement, int64, *errors.Error) {
	if tenantID == "" || inventoryID == "" {
		return nil, 0, errors.ErrBadRequest(ctx).SetDetail("tenant_id and inventory_id are required")
	}
	return s.inventoryRepo.ListMovements(ctx, tenantID, inventoryID, limit, offset)
}
