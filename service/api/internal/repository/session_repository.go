package repository

import (
	"context"
	stdErrors "errors"
	"time"

	"gitlab.com/ecommercehub1/api/internal/core/model"
	"gitlab.com/ecommercehub1/lib/pkg/errors"
	"gorm.io/gorm"
)

type ISessionRepository interface {
	CreateSession(ctx context.Context, session *model.Sessions) (*model.Sessions, *errors.Error)
	GetSessionByID(ctx context.Context, id int64) (*model.Sessions, *errors.Error)
	GetSessionByToken(ctx context.Context, token string) (*model.Sessions, *errors.Error)
	RevokeSession(ctx context.Context, id int64) *errors.Error
}

type SessionRepository struct {
	*baseRepository
}

func NewSessionDBRepository(db *gorm.DB) *SessionRepository {
	return &SessionRepository{baseRepository: NewBaseRepository(db)}
}

func (r *SessionRepository) CreateSession(ctx context.Context, session *model.Sessions) (*model.Sessions, *errors.Error) {
	if err := r.db.Create(session).Error; err != nil {
		return nil, r.returnError(ctx, err)
	}
	return session, nil
}

func (r *SessionRepository) GetSessionByID(ctx context.Context, id int64) (*model.Sessions, *errors.Error) {
	var session model.Sessions
	if err := r.db.Where("id = ?", id).First(&session).Error; err != nil {
		if stdErrors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, r.returnError(ctx, err)
	}
	return &session, nil
}

func (r *SessionRepository) GetSessionByToken(ctx context.Context, token string) (*model.Sessions, *errors.Error) {
	var session model.Sessions
	if err := r.db.Where("token = ?", token).First(&session).Error; err != nil {
		return nil, r.returnError(ctx, err)
	}
	return &session, nil
}

func (r *SessionRepository) RevokeSession(ctx context.Context, id int64) *errors.Error {
	var session model.Sessions
	if err := r.db.Where("id = ?", id).First(&session).Error; err != nil {
		return r.returnError(ctx, err)
	}
	session.IsActive = false
	session.RevokedAt = time.Now()
	if err := r.db.Save(&session).Error; err != nil {
		return r.returnError(ctx, err)
	}
	return nil
}
