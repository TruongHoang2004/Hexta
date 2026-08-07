package repository

import (
	"context"

	"gitlab.com/ecommercehub1/shared/pkg/errors"
	"gorm.io/gorm"
)

type baseRepository struct {
	db *gorm.DB
}

func NewBaseRepository(db *gorm.DB) *baseRepository {
	return &baseRepository{db: db}
}

func (r *baseRepository) returnError(ctx context.Context, err error) *errors.Error {
	return errors.ErrSystemError(ctx, err.Error())
}
