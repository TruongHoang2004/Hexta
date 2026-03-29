package service

import (
	"context"

	"gitlab.com/ecommercehub1/lib/pkg/errors"
	"gitlab.com/ecommercehub1/user/internal/core/model"
	"gitlab.com/ecommercehub1/user/internal/present/request"
	"gitlab.com/ecommercehub1/user/internal/repository"
	"gitlab.com/ecommercehub1/user/internal/utils"
)

type UserService struct {
	*baseService
	userRepo repository.IUserRepository
}

func NewUserService(userRepo repository.IUserRepository) *UserService {
	return &UserService{
		baseService: NewBaseService(),
		userRepo:    userRepo,
	}
}

func (s *UserService) CreateUser(ctx context.Context, user *model.User) (*model.User, *errors.Error) {
	user.ID = utils.NewULID()
	conflictField, exists, err := s.userRepo.IsFieldExist(ctx, []string{"email", "phone", "user_name"}, []string{user.Email, user.Phone, user.UserName})
	if err != nil {
		return nil, err
	}
	if exists {
		// Provide specific conflict error message
		return nil, errors.ErrConflict(ctx, conflictField, "already exists")
	}
	return s.userRepo.Create(ctx, user)
}

func (s *UserService) UpdateUser(ctx context.Context, user *model.User) (*model.User, *errors.Error) {
	return s.userRepo.Update(ctx, user)
}

func (s *UserService) DeleteUser(ctx context.Context, user *model.User) (*model.User, *errors.Error) {
	return s.userRepo.Delete(ctx, user)
}

func (s *UserService) GetUserByID(ctx context.Context, id string) (*model.User, *errors.Error) {
	return s.userRepo.GetByID(ctx, id)
}

func (s *UserService) GetUserByEmail(ctx context.Context, email string) (*model.User, *errors.Error) {
	return s.userRepo.GetByEmail(ctx, email)
}

func (s *UserService) ListUsers(ctx context.Context, req *request.PaginationRequest) ([]*model.User, int64, *errors.Error) {
	return s.userRepo.List(ctx, req)
}

func (s *UserService) GetUsersByIDs(ctx context.Context, ids []string) ([]*model.User, *errors.Error) {
	return s.userRepo.GetByIDs(ctx, ids)
}
