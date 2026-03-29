package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"gitlab.com/ecommercehub1/api/common"
	"gitlab.com/ecommercehub1/api/internal/core/service"
	"gitlab.com/ecommercehub1/api/internal/core/types"
	"gitlab.com/ecommercehub1/lib/pkg/errors"
)

type AuthMiddleware struct {
	authService *service.AuthService
}

func NewAuthMiddleware(authService *service.AuthService) *AuthMiddleware {
	return &AuthMiddleware{authService: authService}
}

func extractTokenFromHeader(c *gin.Context) (string, *errors.Error) {
	token := c.GetHeader("Authorization")
	if token == "" {
		return "", errors.ErrUnauthorized(c.Request.Context()).SetDetail("Token is missing")
	}
	if !strings.HasPrefix(token, "Bearer ") {
		return "", errors.ErrUnauthorized(c.Request.Context()).SetDetail("Token format is invalid")
	}
	return token[7:], nil
}

func (m *AuthMiddleware) Authentication() gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := extractTokenFromHeader(c)
		if err != nil {
			c.AbortWithStatusJSON(err.GetHttpStatus(), errors.ConvertErrorToResponse(err))
			return
		}

		claims, err := m.authService.ValidateToken(c.Request.Context(), token)
		if err != nil {
			c.AbortWithStatusJSON(err.GetHttpStatus(), errors.ConvertErrorToResponse(err))
			return
		}

		authInfo := &types.AuthInfo{
			UserID:    claims.UserID,
			SessionID: claims.SessionID,
		}

		ctx := common.SetAuthInfo(c.Request.Context(), authInfo)
		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}
