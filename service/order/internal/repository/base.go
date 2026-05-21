package repository

import (
	"context"

	"gitlab.com/ecommercehub1/lib/pkg/errors"
	"gorm.io/gorm"
)

type baseRepository struct {
	db *gorm.DB
}

func NewBaseRepository(db *gorm.DB) *baseRepository {
	return &baseRepository{db: db}
}

func (b *baseRepository) returnError(ctx context.Context, err error) *errors.Error {
	return errors.ErrSystemError(ctx, err.Error())
}
