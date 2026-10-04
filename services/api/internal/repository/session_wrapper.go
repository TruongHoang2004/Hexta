package repository

import (
	"context"

	"github.com/TruongHoang2004/Hexta/services/api/internal/core/model"
	"github.com/TruongHoang2004/Hexta/services/api/internal/infrastructure/cache"
	"github.com/TruongHoang2004/Hexta/packages/shared/pkg/errors"
)

type SessionCacheWrapper struct {
	dbRepo *SessionRepository
	cache  *cache.SessionCache
}

func NewSessionCacheWrapper(dbRepo *SessionRepository, ch *cache.SessionCache) ISessionRepository {
	return &SessionCacheWrapper{
		dbRepo: dbRepo,
		cache:  ch,
	}
}

func (w *SessionCacheWrapper) CreateSession(ctx context.Context, session *model.Sessions) (*model.Sessions, *errors.Error) {
	createdSession, err := w.dbRepo.CreateSession(ctx, session)
	if err == nil {
		_ = w.cache.SetSession(ctx, createdSession)
	}
	return createdSession, err
}

func (w *SessionCacheWrapper) GetSessionByID(ctx context.Context, id int64) (*model.Sessions, *errors.Error) {
	if cachedSession, err := w.cache.GetSession(ctx, id); err == nil && cachedSession != nil {
		return cachedSession, nil
	}
	session, err := w.dbRepo.GetSessionByID(ctx, id)
	if err == nil && session != nil {
		_ = w.cache.SetSession(ctx, session)
	}
	return session, err
}

func (w *SessionCacheWrapper) GetSessionByToken(ctx context.Context, token string) (*model.Sessions, *errors.Error) {
	if cachedSession, err := w.cache.GetSessionByToken(ctx, token); err == nil && cachedSession != nil {
		return cachedSession, nil
	}

	session, err := w.dbRepo.GetSessionByToken(ctx, token)
	if err == nil && session != nil {
		_ = w.cache.SetSession(ctx, session)
	}
	return session, err
}

func (w *SessionCacheWrapper) RevokeSession(ctx context.Context, id int64) *errors.Error {
	err := w.dbRepo.RevokeSession(ctx, id)
	if err == nil {
		if session, findErr := w.dbRepo.GetSessionByID(ctx, id); findErr == nil && session != nil {
			_ = w.cache.DeleteSession(ctx, id, session.Token)
		}
	}
	return err
}
