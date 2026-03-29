package common

import (
	"context"

	"gitlab.com/ecommercehub1/api/internal/core/types"
)

type contextKey struct{}

var userKey = contextKey{}

func SetAuthInfo(ctx context.Context, info *types.AuthInfo) context.Context {
	return context.WithValue(ctx, userKey, info)
}

func GetAuthInfo(ctx context.Context) (*types.AuthInfo, bool) {
	info, ok := ctx.Value(userKey).(*types.AuthInfo)
	return info, ok
}
