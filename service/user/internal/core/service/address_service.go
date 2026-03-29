package service

import (
	"gitlab.com/ecommercehub1/user/internal/core/model"
	"gitlab.com/ecommercehub1/user/internal/repository"
	"gitlab.com/ecommercehub1/user/internal/utils"
)

type AddressService struct {
	*baseService
	addressRepo *repository.AddressRepository
}

func NewAddressService(addressRepo *repository.AddressRepository) *AddressService {
	return &AddressService{
		baseService: NewBaseService(),
		addressRepo: addressRepo,
	}
}

func (s *AddressService) CreateAddress(address *model.Address) error {
	address.ID = utils.NewULID()
	return s.addressRepo.Create(address)
}

func (s *AddressService) UpdateAddress(address *model.Address) error {
	return s.addressRepo.Update(address)
}

func (s *AddressService) DeleteAddress(id string) error {
	address, err := s.addressRepo.GetByID(id)
	if err != nil {
		return err
	}
	return s.addressRepo.Delete(address)
}

func (s *AddressService) GetAddressByID(id string) (*model.Address, error) {
	return s.addressRepo.GetByID(id)
}

func (s *AddressService) ListUserAddresses(userID string, page, limit int32) ([]*model.Address, int64, error) {
	return s.addressRepo.ListByUserID(userID, page, limit)
}
