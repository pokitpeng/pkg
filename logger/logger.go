package logger

/*
为了继承kratos logger的接口而创建，详细可参考：https://github.com/go-kratos/kratos
*/

import (
	"bytes"
	"context"
	"fmt"
	"sync"

	"github.com/go-kratos/kratos/v2/log"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

var _ log.Logger = (*ZapLogger)(nil)

// ZapLogger is a logger impl.
type ZapLogger struct {
	log  *zap.Logger
	ctx  context.Context
	pool *sync.Pool
}

// NewLogger return a kratos logger.
func NewLogger(settings ...Option) log.Logger {
	zapLogger := NewZapLogger(settings...)
	return &ZapLogger{
		log: zapLogger,
		pool: &sync.Pool{
			New: func() interface{} {
				return new(bytes.Buffer)
			},
		}}
}

func NewLog(settings ...Option) *log.Helper {
	logger := NewLogger(settings...)
	return log.NewHelper(logger)
}

// NewDevelopLog 用于开发环境的log
func NewDevelopLog() *log.Helper {
	logger := NewLog()
	return logger
}

// Log Implementation of logger interface.
func (l *ZapLogger) Log(level log.Level, keyvals ...interface{}) error {
	if len(keyvals) == 0 {
		return nil
	}
	if len(keyvals)%2 != 0 {
		keyvals = append(keyvals, "")
	}
	buf := l.pool.Get().(*bytes.Buffer)
	var fields []zap.Field
	if traceId := getTraceId(l.ctx); traceId != "" {
		fields = append(fields, zap.String("trace_id", traceId))
	}
	for i := 0; i < len(keyvals); i += 2 {
		fields = append(fields, zap.Any(fmt.Sprint(keyvals[i]), fmt.Sprint(keyvals[i+1])))
	}
	switch level {
	case log.LevelDebug:
		l.log.Debug(buf.String(), fields...)
	case log.LevelInfo:
		l.log.Info(buf.String(), fields...)
	case log.LevelWarn:
		l.log.Warn(buf.String(), fields...)
	case log.LevelError:
		l.log.Error(buf.String(), fields...)
	}
	buf.Reset()
	l.pool.Put(buf)
	return nil
}

// get trace id
func getTraceId(ctx context.Context) string {
	var traceID string
	if tid := trace.SpanContextFromContext(ctx).TraceID(); tid.IsValid() {
		traceID = tid.String()
	}
	return traceID
}
