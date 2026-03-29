package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"gitlab.com/ecommercehub1/api/internal/core/model"
)

const (
	sessionKeyPrefix      = "session:%d"
	sessionTokenKeyPrefix = "session_token:%s"
	sessionTTL            = 7 * 24 * time.Hour // Assuming long lived sessions
)

type SessionCache struct {
	redis *RedisClient
}

func NewSessionCache(redis *RedisClient) *SessionCache {
	return &SessionCache{redis: redis}
}

func (c *SessionCache) SetSession(ctx context.Context, session *model.Sessions) error {
	key := fmt.Sprintf(sessionKeyPrefix, session.ID)
	tokenKey := fmt.Sprintf(sessionTokenKeyPrefix, session.Token)
	data, err := json.Marshal(session)
	if err != nil {
		return err
	}

	// Cache both by ID and Token for fast retrieval.
	if err := c.redis.Set(ctx, key, data, sessionTTL); err != nil {
		return err
	}
	return c.redis.Set(ctx, tokenKey, data, sessionTTL)
}

func (c *SessionCache) GetSession(ctx context.Context, id int64) (*model.Sessions, error) {
	key := fmt.Sprintf(sessionKeyPrefix, id)
	data, err := c.redis.Get(ctx, key)
	if err != nil {
		return nil, err
	}

	var session model.Sessions
	if err := json.Unmarshal(data, &session); err != nil {
		return nil, err
	}

	return &session, nil
}

func (c *SessionCache) GetSessionByToken(ctx context.Context, token string) (*model.Sessions, error) {
	key := fmt.Sprintf(sessionTokenKeyPrefix, token)
	data, err := c.redis.Get(ctx, key)
	if err != nil {
		return nil, err
	}

	var session model.Sessions
	if err := json.Unmarshal(data, &session); err != nil {
		return nil, err
	}

	return &session, nil
}

func (c *SessionCache) DeleteSession(ctx context.Context, id int64, token string) error {
	key := fmt.Sprintf(sessionKeyPrefix, id)
	tokenKey := fmt.Sprintf(sessionTokenKeyPrefix, token)

	_ = c.redis.Delete(ctx, key)
	return c.redis.Delete(ctx, tokenKey)
}
