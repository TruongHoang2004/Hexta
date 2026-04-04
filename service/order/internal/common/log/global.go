package log

import (
	"context"
	"fmt"

	"gitlab.com/ecommercehub1/lib/pkg/errors"
	"gitlab.com/ecommercehub1/lib/pkg/logger"
)

var globalLogger *logger.Logger

func Info(ctx context.Context, msg string, args ...interface{}) {
	globalLogger.Info(addCtxValue(msg), args...)
}

func Debug(ctx context.Context, msg string, args ...interface{}) {
	globalLogger.Debug(addCtxValue(msg), args...)
}

func Warn(ctx context.Context, msg string, args ...interface{}) {
	globalLogger.Warn(addCtxValue(msg), args...)
}

func Error(ctx context.Context, msg string, args ...interface{}) {
	globalLogger.Error(addCtxValue(msg), args...)
}

func Fatal(msg string, args ...interface{}) {
	globalLogger.Fatal(msg, args...)
}

func IErr(ctx context.Context, err *errors.Error) {
	if errors.IsInternalError(err) {
		globalLogger.Error(addCtxValue(err.GetDetail()))
	} else if errors.IsClientError(err) {
		globalLogger.Warn(addCtxValue(err.ToJSon()))
	}

}

func GetLogger() *logger.Logger {
	if globalLogger == nil {
		globalLogger = logger.NewLogger(logger.LoggerOption{
			IsProd: false,
		})
	}
	return globalLogger
}

func addCtxValue(msg string) string {
	return fmt.Sprintf("%s", msg)
}
