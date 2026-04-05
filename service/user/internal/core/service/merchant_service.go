package service

import (
	"context"

	"gitlab.com/ecommercehub1/lib/pkg/errors"
	"gitlab.com/ecommercehub1/user/internal/core/model"
	"gitlab.com/ecommercehub1/user/internal/present/request"
	"gitlab.com/ecommercehub1/user/internal/repository"
	"gitlab.com/ecommercehub1/user/internal/utils"
)

type MerchantService struct {
	*baseService
	merchantRepo repository.IMerchantRepository
}

func NewMerchantService(merchantRepo repository.IMerchantRepository) *MerchantService {
	return &MerchantService{
		baseService:  NewBaseService(),
		merchantRepo: merchantRepo,
	}
}

func (s *MerchantService) CreateMerchant(ctx context.Context, merchant *model.Merchant) (*model.Merchant, *errors.Error) {
	merchant.ID = utils.NewULID()

	exsited, err := s.merchantRepo.IsExited(ctx, merchant.Name, merchant.Email, merchant.Phone)
	if err != nil {
		return nil, err
	}
	if exsited {
		return nil, errors.ErrConflict(ctx, "merchant", "name, email or phonenumber already exists")
	}

	return s.merchantRepo.Create(ctx, merchant)
}

func (s *MerchantService) UpdateMerchant(ctx context.Context, merchant *model.Merchant) (*model.Merchant, *errors.Error) {
	return s.merchantRepo.Update(ctx, merchant)
}

func (s *MerchantService) DeleteMerchant(ctx context.Context, merchant *model.Merchant) (*model.Merchant, *errors.Error) {
	return s.merchantRepo.Delete(ctx, merchant)
}

func (s *MerchantService) GetMerchantByID(ctx context.Context, id string) (*model.Merchant, *errors.Error) {
	return s.merchantRepo.GetByID(ctx, id)
}

func (s *MerchantService) GetMerchantByOwnerID(ctx context.Context, ownerID string) (*model.Merchant, *errors.Error) {
	return s.merchantRepo.GetByOwnerID(ctx, ownerID)
}

func (s *MerchantService) ListMerchants(ctx context.Context, req *request.PaginationRequest) ([]*model.Merchant, int64, *errors.Error) {
	return s.merchantRepo.List(ctx, req)
}
