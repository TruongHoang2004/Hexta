package middleware

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"

	"github.com/gin-gonic/gin"
	"gitlab.com/ecommercehub1/api/internal/constant"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc/metadata"
)

func Tracer() gin.HandlerFunc {
	return func(c *gin.Context) {
		spanContext := trace.SpanContextFromContext(c.Request.Context())
		span := trace.SpanFromContext(c.Request.Context())
		span.SetName(fmt.Sprintf("[%s] %s", c.Request.Method, c.FullPath()))
		var traceId string
		if spanContext.TraceID().IsValid() {
			traceId = spanContext.TraceID().String()
		} else {
			traceIdByte := make([]byte, 16)
			rand.Read(traceIdByte)
			traceId = hex.EncodeToString(traceIdByte[:])
		}
		traceContext := context.WithValue(c.Request.Context(), constant.TraceIdName, traceId)
		ctxMetaData := metadata.AppendToOutgoingContext(traceContext, []string{constant.TraceIdName, traceId}...)
		c.Request = c.Request.WithContext(ctxMetaData)
		c.Next()
	}
}
