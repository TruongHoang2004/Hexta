package common

import (
	"context"

	"gitlab.com/ecommercehub1/api/internal/constant"
)

func GetTraceId(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	traceId := ""
	if ctx.Value(constant.TraceIdName) != nil {
		traceId = ctx.Value(constant.TraceIdName).(string)
	}
	return traceId
}
