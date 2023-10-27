package log

import (
	"context"

	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

var helper *zap.SugaredLogger

func init() {
	if helper == nil {
		helper = NewLogger()
	}
}

func Init(opts ...Option) {
	helper = NewLogger(opts...)
}

// WithKV must have key and value
func WithKV(args ...interface{}) *zap.SugaredLogger {
	l := helper.WithOptions(zap.AddCallerSkip(-1))
	return l.With(args...)
}

// WithName only work in FormatConsole setting
func WithName(name string) *zap.SugaredLogger {
	l := helper.WithOptions(zap.AddCallerSkip(-1))
	return l.Named(name)
}

func WithContext(ctx context.Context) *zap.SugaredLogger {
	l := helper.WithOptions(zap.AddCallerSkip(-1))

	var fields []interface{}

	traceID, spanID := TraceInfoFromContext(ctx)
	if len(traceID) > 0 {
		fields = append(fields, "trace_id", traceID)
	}
	if len(spanID) > 0 {
		fields = append(fields, "span_id", spanID)
	}

	return l.With(fields...)
}

func TraceInfoFromContext(ctx context.Context) (traceID, spanID string) {
	spanCtx := trace.SpanContextFromContext(ctx)
	if spanCtx.HasTraceID() {
		traceID = spanCtx.TraceID().String()
	}
	if spanCtx.HasSpanID() {
		spanID = spanCtx.SpanID().String()
	}
	return traceID, spanID
}

func Sync() error {
	return helper.Sync()
}

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

func DPanic(args ...interface{}) {
	helper.DPanic(args...)
}

func Panic(args ...interface{}) {
	helper.Panic(args...)
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

func DPanicf(format string, a ...interface{}) {
	helper.DPanicf(format, a...)
}

func Panicf(format string, a ...interface{}) {
	helper.Panicf(format, a...)
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

func DPanicw(msg string, kvs ...interface{}) {
	helper.DPanicw(msg, kvs...)
}

func Panicw(msg string, kvs ...interface{}) {
	helper.Panicw(msg, kvs...)
}

// =========================================================================

func Debugln(args ...interface{}) {
	helper.Debugln(args...)
}

func Infoln(args ...interface{}) {
	helper.Infoln(args...)
}

func Warnln(args ...interface{}) {
	helper.Warnln(args...)
}

func Errorln(args ...interface{}) {
	helper.Errorln(args...)
}

func Fatalln(args ...interface{}) {
	helper.Fatalln(args...)
}

func DPanicln(args ...interface{}) {
	helper.DPanicln(args...)
}

func Panicln(args ...interface{}) {
	helper.Panicln(args...)
}
