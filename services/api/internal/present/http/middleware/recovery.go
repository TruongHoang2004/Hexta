package middleware

import (
	"fmt"
	"net/http/httputil"
	"runtime/debug"
	"time"

	"github.com/gin-gonic/gin"
	"gitlab.com/ecommercehub1/api/common/log"
	"gitlab.com/ecommercehub1/shared/pkg/errors"
)

func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				httpRequest, _ := httputil.DumpRequest(c.Request, false)
				log.Error(c, "[Recovery from panic]\ntime: [%v]\nerror: [%v] \nrequest: [%v]\nstack: [%v]\n",
					time.Now(), err, string(httpRequest), string(debug.Stack()))
				e := errors.ErrSystemError(c, fmt.Sprintf("recovery, err:[%s]", err))
				c.JSON(e.GetHttpStatus(), errors.ConvertErrorToResponse(e))
				c.Abort()
			}
		}()
		c.Next()
	}
}
