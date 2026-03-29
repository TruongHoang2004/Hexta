package logger

import (
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const callerSkip = 2

type LoggerOption struct {
	IsProd bool
}

type Logger struct {
	zap *zap.SugaredLogger
	opt LoggerOption
}

func SyslogTimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format("2006-01-02 15:04:05"))
}

func CustomLevelEncoder(level zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString("[" + level.CapitalString() + "]")
}

func NewLogger(opt LoggerOption) *Logger {
	encoderConfig := zapcore.EncoderConfig{
		MessageKey:   "message",
		LevelKey:     "level",
		TimeKey:      "time",
		CallerKey:    "caller",
		EncodeCaller: zapcore.FullCallerEncoder,
		EncodeTime:   SyslogTimeEncoder,
		EncodeLevel:  CustomLevelEncoder,
	}

	var encoder zapcore.Encoder
	var level zapcore.Level
	if opt.IsProd {
		encoder = zapcore.NewJSONEncoder(encoderConfig)
		level = zap.InfoLevel
	} else {
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
		level = zap.DebugLevel
	}
	core := zapcore.NewCore(encoder, zapcore.AddSync(os.Stderr), level)
	//set log
	return &Logger{
		zap: zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel), zap.AddCallerSkip(callerSkip)).Sugar(),
		opt: opt,
	}
}

func (l *Logger) Info(msg string, args ...interface{}) {
	l.zap.Infof(msg, args...)
}

func (l *Logger) Debug(msg string, args ...interface{}) {
	l.zap.Debugf(msg, args...)
}

func (l *Logger) Warn(msg string, args ...interface{}) {
	l.zap.Warnf(msg, args...)
}

func (l *Logger) Error(msg string, args ...interface{}) {
	l.zap.Errorf(msg, args...)
}

func (l *Logger) Fatal(msg string, args ...interface{}) {
	l.zap.Fatalf(msg, args...)
}

func (l *Logger) GetZap() *zap.SugaredLogger {
	return l.zap
}
