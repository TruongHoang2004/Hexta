package repository

import (
	"context"
	stdErrors "errors"

	"gitlab.com/ecommercehub1/api/internal/core/model"
	"gitlab.com/ecommercehub1/lib/pkg/errors"
	"gorm.io/gorm"
)

type IdentityRepository struct {
	*baseRepository
}

func NewIdentityRepository(db *gorm.DB) *IdentityRepository {
	return &IdentityRepository{baseRepository: NewBaseRepository(db)}
}

func (r *IdentityRepository) GetCredentialByIdentifier(ctx context.Context, identifier string, provider model.Provider) (*model.AuthIdentities, *errors.Error) {
	var authIdentity model.AuthIdentities
	if err := r.db.Where("identifier = ? AND provider = ?", identifier, provider).First(&authIdentity).Error; err != nil {
		if stdErrors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, r.returnError(ctx, err)
	}

	return &authIdentity, nil
}

func (r *IdentityRepository) CreateIdentity(ctx context.Context, authIdentity *model.AuthIdentities) (*model.AuthIdentities, *errors.Error) {
	if err := r.db.Create(authIdentity).Error; err != nil {
		return nil, r.returnError(ctx, err)
	}
	return authIdentity, nil
}
