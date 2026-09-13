package log

import (
	"context"
	"fmt"

	"github.com/TruongHoang2004/Hexta/services/api/common"
	"github.com/TruongHoang2004/Hexta/services/api/config"
	"github.com/TruongHoang2004/Hexta/packages/shared/pkg/errors"
	"github.com/TruongHoang2004/Hexta/packages/shared/pkg/logger"
)

var globalLogger *logger.Logger

func Info(ctx context.Context, msg string, args ...interface{}) {
	globalLogger.Info(addCtxValue(ctx, msg), args...)
}

func Debug(ctx context.Context, msg string, args ...interface{}) {
	globalLogger.Debug(addCtxValue(ctx, msg), args...)
}

func Warn(ctx context.Context, msg string, args ...interface{}) {
	globalLogger.Warn(addCtxValue(ctx, msg), args...)
}

func Error(ctx context.Context, msg string, args ...interface{}) {
	globalLogger.Error(addCtxValue(ctx, msg), args...)
}

func Fatal(msg string, args ...interface{}) {
	globalLogger.Fatal(msg, args...)
}

func IErr(ctx context.Context, err *errors.Error) {
	if errors.IsInternalError(err) {
		globalLogger.Error(addCtxValue(ctx, err.GetDetail()))
	} else if errors.IsClientError(err) {
		globalLogger.Warn(addCtxValue(ctx, err.ToJSon()))
	}

}

func GetLogger() *logger.Logger {
	if globalLogger == nil {
		globalLogger = logger.NewLogger(logger.LoggerOption{
			IsProd: config.AppConfig.Server.Production,
		})
	}
	return globalLogger
}

func addCtxValue(ctx context.Context, msg string) string {
	traceId := common.GetTraceId(ctx)

	return fmt.Sprintf("TraceId: %s | %s", traceId, msg)
}
