package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/TruongHoang2004/Hexta/services/api/common/log"
)

func Log() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		log.Info(c.Request.Context(), "path: [%v], status: [%v], method: [%v], user_agent: [%v]",
			c.Request.URL.Path, c.Writer.Status(), c.Request.Method, c.Request.UserAgent())
	}
}
