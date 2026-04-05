package repository

import (
	"context"

	"gitlab.com/ecommercehub1/lib/pkg/errors"
	"gitlab.com/ecommercehub1/user/internal/core/model"
	"gitlab.com/ecommercehub1/user/internal/present/request"
	"gorm.io/gorm"
)

type IMerchantRepository interface {
	IsExited(ctx context.Context, name, email, phone string) (bool, *errors.Error)
	Create(ctx context.Context, merchant *model.Merchant) (*model.Merchant, *errors.Error)
	Update(ctx context.Context, merchant *model.Merchant) (*model.Merchant, *errors.Error)
	Delete(ctx context.Context, merchant *model.Merchant) (*model.Merchant, *errors.Error)
	GetByID(ctx context.Context, id string) (*model.Merchant, *errors.Error)
	GetByOwnerID(ctx context.Context, ownerID string) (*model.Merchant, *errors.Error)
	List(ctx context.Context, req *request.PaginationRequest) ([]*model.Merchant, int64, *errors.Error)
}

type MerchantRepository struct {
	*baseRepository
}

func NewMerchantDBRepository(db *gorm.DB) *MerchantRepository {
	return &MerchantRepository{
		baseRepository: NewBaseRepository(db),
	}
}

func (r *MerchantRepository) IsExited(ctx context.Context, name, email, phone string) (bool, *errors.Error) {
	var id string
	err := r.db.WithContext(ctx).Model(&model.Merchant{}).
		Select("id").
		Where("name = ? OR email = ? OR phone = ?", name, email, phone).
		Limit(1).
		Find(&id).Error

	if err != nil {
		return false, r.returnError(ctx, err)
	}
	return id != "", nil
}

func (r *MerchantRepository) Create(ctx context.Context, merchant *model.Merchant) (*model.Merchant, *errors.Error) {
	if err := r.db.WithContext(ctx).Create(merchant).Error; err != nil {

		return nil, r.returnError(ctx, err)
	}
	return merchant, nil
}

func (r *MerchantRepository) Update(ctx context.Context, merchant *model.Merchant) (*model.Merchant, *errors.Error) {
	if err := r.db.WithContext(ctx).Save(merchant).Error; err != nil {
		return nil, r.returnError(ctx, err)
	}
	return merchant, nil
}

func (r *MerchantRepository) Delete(ctx context.Context, merchant *model.Merchant) (*model.Merchant, *errors.Error) {
	if err := r.db.WithContext(ctx).Delete(merchant).Error; err != nil {
		return nil, r.returnError(ctx, err)
	}
	return merchant, nil
}

func (r *MerchantRepository) GetByID(ctx context.Context, id string) (*model.Merchant, *errors.Error) {
	var merchant model.Merchant
	if err := r.db.WithContext(ctx).First(&merchant, "id = ?", id).Error; err != nil {
		return nil, r.returnError(ctx, err)
	}
	return &merchant, nil
}

func (r *MerchantRepository) GetByOwnerID(ctx context.Context, ownerID string) (*model.Merchant, *errors.Error) {
	var merchant model.Merchant
	if err := r.db.WithContext(ctx).First(&merchant, "owner_id = ?", ownerID).Error; err != nil {
		return nil, r.returnError(ctx, err)
	}
	return &merchant, nil
}

func (r *MerchantRepository) List(ctx context.Context, req *request.PaginationRequest) ([]*model.Merchant, int64, *errors.Error) {
	var merchants []*model.Merchant
	var total int64

	query := r.db.WithContext(ctx).Model(&model.Merchant{})

	if req.Search != "" {
		query = query.Where("name ILIKE ? OR email ILIKE ?", "%"+req.Search+"%", "%"+req.Search+"%")
	}
	for k, v := range req.Filter {
		query = query.Where(k+" = ?", v)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, r.returnError(ctx, err)
	}

	offset := (req.Page - 1) * req.Limit
	if err := query.Offset(int(offset)).Limit(int(req.Limit)).Find(&merchants).Error; err != nil {
		return nil, 0, r.returnError(ctx, err)
	}

	return merchants, total, nil
}
