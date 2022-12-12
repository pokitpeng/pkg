package logger

import (
	"context"

	"github.com/opentracing/opentracing-go"
	"github.com/uber/jaeger-client-go"
	"go.uber.org/zap"
)

var helper *zap.SugaredLogger

func init() {
	if helper == nil {
		helper = NewZapLogger(ConfigWithCallerSkipOption(1)).Sugar()
	}
}

func GetLogger() *zap.SugaredLogger {
	return helper
}

func Init(options ...Option) {
	helper = NewZapLogger(options...).Sugar()
}

func InitStandardLogger() {
	helper = NewZapLogger(
		ConfigWithEncoderOption(EncoderLowercase),
		ConfigWithFormatOption(FormatJson),
		ConfigWithCallerSkipOption(1),
	).Sugar()
}

func InitDevelopmentLogger() {
	helper = NewZapLogger(
		ConfigWithEncoderOption(EncoderCapitalColor),
		ConfigWithFormatOption(FormatConsole),
		ConfigWithCallerSkipOption(1),
	).Sugar()
}

// InitProductionLogger ...
func InitProductionLogger(logPath, logName string) {
	helper = NewZapLogger(
		ConfigWithIsStdOutOption(false),
		ConfigWithEncoderOption(EncoderLowercase),
		ConfigWithFormatOption(FormatJson),
		ConfigWithCallerSkipOption(1),
		ConfigWithLevelOption(LevelDebug),
		ConfigWithFilePathOption(logPath),
		ConfigWithFileNameOption(logName),
		ConfigWithLogSizeOption(&LogSizeSplitConfig{
			MaxAge:     "720h",
			MaxSize:    1,
			MaxBackups: 60,
			Compress:   true,
		}),
	).Sugar()
}

// WithKV must have key and value
func WithKV(args ...interface{}) *zap.SugaredLogger {
	return helper.With(args...)
}

// WithName only work in FormatConsole setting
func WithName(name string) *zap.SugaredLogger {
	return helper.Named(name)
}

func Sync() error {
	return helper.Sync()
}

// =========================================================================

// WithContext 解析context中的trace信息
func WithContext(ctx context.Context) *zap.SugaredLogger {
	var traceLog = helper
	if span := opentracing.SpanFromContext(ctx); span != nil {
		if jaegerCtx, ok := span.Context().(jaeger.SpanContext); ok {
			traceLog = helper.With(
				"trace_id", jaegerCtx.TraceID().String(),
				"span_id", jaegerCtx.SpanID().String(),
			)
			return traceLog
		}
	}
	return helper
}

// =========================================================================

func Debug(args ...interface{}) {
	helper.Debug(args...)
}

func Info(args ...interface{}) {
	helper.Info(args...)
}

func Println(args ...interface{}) {
	helper.Info(args...)
}

func Warn(args ...interface{}) {
	helper.Warn(args...)
}

func Error(args ...interface{}) {
	helper.Error(args...)
}

func Fatal(args ...interface{}) {
	helper.Fatal(args...)
}

// =========================================================================

func Debugf(format string, a ...interface{}) {
	helper.Debugf(format, a...)
}

func Infof(format string, a ...interface{}) {
	helper.Infof(format, a...)
}

func Printf(format string, a ...interface{}) {
	helper.Infof(format, a...)
}

func Warnf(format string, a ...interface{}) {
	helper.Warnf(format, a...)
}

func Errorf(format string, a ...interface{}) {
	helper.Errorf(format, a...)
}

func Fatalf(format string, a ...interface{}) {
	helper.Fatalf(format, a...)
}

// =========================================================================

func Debugw(msg string, kvs ...interface{}) {
	helper.Debugw(msg, kvs...)
}

func Infow(msg string, kvs ...interface{}) {
	helper.Infow(msg, kvs...)
}

func Warnw(msg string, kvs ...interface{}) {
	helper.Warnw(msg, kvs...)
}

func Errorw(msg string, kvs ...interface{}) {
	helper.Errorw(msg, kvs...)
}

func Fatalw(msg string, kvs ...interface{}) {
	helper.Fatalw(msg, kvs...)
}
