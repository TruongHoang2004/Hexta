package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"gitlab.com/ecommercehub1/api/common"
	"gitlab.com/ecommercehub1/api/internal/core/service"
	"gitlab.com/ecommercehub1/shared/pkg/errors"
)

type AuthMiddleware struct {
	authService *service.AuthService
}

func NewAuthMiddleware(authService *service.AuthService) *AuthMiddleware {
	return &AuthMiddleware{
		authService: authService,
	}
}

func (m *AuthMiddleware) Authenticate() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenStr := ""

		// 1. Extract from Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" {
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
				tokenStr = strings.TrimSpace(parts[1])
			}
		}

		// 2. Fallback to auth_token cookie
		if tokenStr == "" {
			if cookieToken, err := c.Cookie("auth_token"); err == nil && cookieToken != "" {
				tokenStr = cookieToken
			}
		}

		if tokenStr == "" {
			err := errors.ErrUnauthorized(c.Request.Context()).SetDetail("Authorization token is required")
			c.JSON(err.GetHttpStatus(), errors.ConvertErrorToResponse(err))
			c.Abort()
			return
		}

		// 3. Validate access token
		authInfo, err := m.authService.ValidateAccessToken(c.Request.Context(), tokenStr)
		if err != nil {
			c.JSON(err.GetHttpStatus(), errors.ConvertErrorToResponse(err))
			c.Abort()
			return
		}

		// 4. Inject auth info into Gin context and Request context
		c.Set("auth_info", authInfo)
		c.Set("user_id", authInfo.UserID)
		c.Set("session_id", authInfo.SessionID)
		c.Request = c.Request.WithContext(common.SetAuthInfo(c.Request.Context(), authInfo))

		c.Next()
	}
}
