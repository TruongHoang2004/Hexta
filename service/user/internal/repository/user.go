package repository

import (
	"context"
	"fmt"

	"gitlab.com/ecommercehub1/lib/pkg/errors"
	"gitlab.com/ecommercehub1/user/internal/core/model"
	"gitlab.com/ecommercehub1/user/internal/present/request"
	"gorm.io/gorm"
)

type IUserRepository interface {
	Create(ctx context.Context, user *model.User) (*model.User, *errors.Error)
	Update(ctx context.Context, user *model.User) (*model.User, *errors.Error)
	Delete(ctx context.Context, user *model.User) (*model.User, *errors.Error)
	GetByID(ctx context.Context, id string) (*model.User, *errors.Error)
	List(ctx context.Context, req *request.PaginationRequest) ([]*model.User, int64, *errors.Error)
	GetByIDs(ctx context.Context, ids []string) ([]*model.User, *errors.Error)
	GetByEmail(ctx context.Context, email string) (*model.User, *errors.Error)
	IsFieldExist(ctx context.Context, fields []string, values []string) (string, bool, *errors.Error)
}

type UserRepository struct {
	*baseRepository
}

func NewUserDBRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{
		baseRepository: NewBaseRepository(db),
	}
}

func (r *UserRepository) Create(ctx context.Context, user *model.User) (*model.User, *errors.Error) {
	if err := r.db.Create(user).Error; err != nil {
		return nil, r.returnError(ctx, err)
	}
	return user, nil
}

func (r *UserRepository) Update(ctx context.Context, user *model.User) (*model.User, *errors.Error) {
	if err := r.db.Save(user).Error; err != nil {
		return nil, r.returnError(ctx, err)
	}
	return user, nil
}

func (r *UserRepository) Delete(ctx context.Context, user *model.User) (*model.User, *errors.Error) {
	if err := r.db.Delete(user).Error; err != nil {
		return nil, r.returnError(ctx, err)
	}
	return user, nil
}

func (r *UserRepository) GetByID(ctx context.Context, id string) (*model.User, *errors.Error) {
	var user model.User
	if err := r.db.First(&user, "id = ?", id).Error; err != nil {
		return nil, r.returnError(ctx, err)
	}
	return &user, nil
}

func (r *UserRepository) List(ctx context.Context, req *request.PaginationRequest) ([]*model.User, int64, *errors.Error) {
	var users []*model.User
	var total int64
	offset := (req.Page - 1) * req.Limit
	query := r.db.Model(&model.User{})

	for field, value := range req.Filter {
		query = query.Where(field+" = ?", value)
	}

	if req.Search != "" {
		query = query.Where("name LIKE ?", "%"+req.Search+"%")
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, r.returnError(ctx, err)
	}

	if req.Sort != "" && req.Order != "" {
		query = query.Order(req.Sort + " " + req.Order)
	} else {
		query = query.Order("created_at DESC")
	}

	if err := query.Offset(int(offset)).Limit(int(req.Limit)).Find(&users).Error; err != nil {
		return nil, 0, r.returnError(ctx, err)
	}

	return users, total, nil
}

func (r *UserRepository) GetByIDs(ctx context.Context, ids []string) ([]*model.User, *errors.Error) {
	var users []*model.User
	if err := r.db.Where("id IN ?", ids).Find(&users).Error; err != nil {
		return nil, r.returnError(ctx, err)
	}
	return users, nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*model.User, *errors.Error) {
	var user model.User
	if err := r.db.Where("email = ?", email).First(&user).Error; err != nil {
		return nil, r.returnError(ctx, err)
	}
	return &user, nil
}

func (r *UserRepository) IsFieldExist(ctx context.Context, fields []string, values []string) (string, bool, *errors.Error) {
	var result map[string]interface{}

	query := r.db.Model(&model.User{})
	for i := range fields {
		if values[i] != "" {
			query = query.Or(fmt.Sprintf("%s = ?", fields[i]), values[i])
		}
	}

	if err := query.First(&result).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return "", false, nil
		}
		return "", false, r.returnError(ctx, err)
	}

	for i, field := range fields {
		if values[i] != "" {
			if val, ok := result[field]; ok {
				if fmt.Sprintf("%v", val) == values[i] {
					return field, true, nil
				}
			}
		}
	}

	return "", true, nil
}
