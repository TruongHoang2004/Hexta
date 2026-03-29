package repository

import (
	"context"

	"gitlab.com/ecommercehub1/lib/pkg/errors"
	"gitlab.com/ecommercehub1/user/internal/core/model"
	"gitlab.com/ecommercehub1/user/internal/infrastructure/cache"
	"gitlab.com/ecommercehub1/user/internal/present/request"
)

type userCacheWrapper struct {
	dbRepo *UserRepository
	cache  *cache.UserCache
}

func NewUserCacheWrapper(dbRepo *UserRepository, cache *cache.UserCache) IUserRepository {
	return &userCacheWrapper{
		dbRepo: dbRepo,
		cache:  cache,
	}
}

func (w *userCacheWrapper) Create(ctx context.Context, user *model.User) (*model.User, *errors.Error) {
	return w.dbRepo.Create(ctx, user)
}

func (w *userCacheWrapper) Update(ctx context.Context, user *model.User) (*model.User, *errors.Error) {
	updatedUser, err := w.dbRepo.Update(ctx, user)
	if err == nil {
		_ = w.cache.DeleteUser(ctx, user.ID) // Invalidate cache
	}
	return updatedUser, err
}

func (w *userCacheWrapper) Delete(ctx context.Context, user *model.User) (*model.User, *errors.Error) {
	deletedUser, err := w.dbRepo.Delete(ctx, user)
	if err == nil {
		_ = w.cache.DeleteUser(ctx, user.ID) // Invalidate cache
	}
	return deletedUser, err
}

func (w *userCacheWrapper) GetByID(ctx context.Context, id string) (*model.User, *errors.Error) {
	// Try cache first
	if user, err := w.cache.GetUser(ctx, id); err == nil {
		return user, nil
	}

	// Fetch from DB
	user, err := w.dbRepo.GetByID(ctx, id)
	if err == nil {
		_ = w.cache.SetUser(ctx, user) // Populate cache
	}
	return user, err
}

func (w *userCacheWrapper) List(ctx context.Context, req *request.PaginationRequest) ([]*model.User, int64, *errors.Error) {
	return w.dbRepo.List(ctx, req)
}

func (w *userCacheWrapper) GetByIDs(ctx context.Context, ids []string) ([]*model.User, *errors.Error) {
	return w.dbRepo.GetByIDs(ctx, ids)
}

func (w *userCacheWrapper) GetByEmail(ctx context.Context, email string) (*model.User, *errors.Error) {
	return w.dbRepo.GetByEmail(ctx, email)
}

func (w *userCacheWrapper) IsFieldExist(ctx context.Context, fields []string, values []string) (string, bool, *errors.Error) {
	return w.dbRepo.IsFieldExist(ctx, fields, values)
}
