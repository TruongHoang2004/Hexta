package repository

import (
	"gitlab.com/ecommercehub1/user/internal/core/model"
	"gorm.io/gorm"
)

type AddressRepository struct {
	*baseRepository
}

func NewAddressRepository(db *gorm.DB) *AddressRepository {
	return &AddressRepository{
		baseRepository: NewBaseRepository(db),
	}
}

func (r *AddressRepository) Create(address *model.Address) error {
	return r.db.Create(address).Error
}

func (r *AddressRepository) Update(address *model.Address) error {
	return r.db.Save(address).Error
}

func (r *AddressRepository) Delete(address *model.Address) error {
	return r.db.Delete(address).Error
}

func (r *AddressRepository) GetByID(id string) (*model.Address, error) {
	var address model.Address
	return &address, r.db.First(&address, "id = ?", id).Error
}

func (r *AddressRepository) ListByUserID(userID string, page, limit int32) ([]*model.Address, int64, error) {
	var addresses []*model.Address
	var total int64
	offset := (page - 1) * limit

	if err := r.db.Model(&model.Address{}).Where("user_id = ?", userID).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := r.db.Where("user_id = ?", userID).Offset(int(offset)).Limit(int(limit)).Find(&addresses).Error; err != nil {
		return nil, 0, err
	}
	return addresses, total, nil
}
