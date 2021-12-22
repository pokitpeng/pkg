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
	"go.uber.org/zap/zapcore"
)

var _ log.Logger = (*ZapLogger)(nil)

// ZapLogger is a logger impl.
type ZapLogger struct {
	log  *zap.Logger
	ctx  context.Context
	pool *sync.Pool
}

// NewZapLogger return a zap logger.
func NewZapLogger(config Config, settings ...Option) log.Logger {
	core := NewZapCore(config, settings...)
	opts := []zap.Option{
		zap.AddCaller(),
		zap.Development(), // 开启开发模式，堆栈跟踪
		zap.AddStacktrace(zapcore.FatalLevel),
		zap.AddCallerSkip(2),
	}
	return &ZapLogger{
		log: zap.New(core, opts...),
		pool: &sync.Pool{
			New: func() interface{} {
				return new(bytes.Buffer)
			},
		}}
}

func NewLog(config Config, settings ...Option) *log.Helper {
	logger := NewZapLogger(config, settings...)
	return log.NewHelper(logger)
}

// NewDevLog 用于开发环境的log
func NewDevelopLog() *log.Helper {
	logger := NewLog(Config{
		IsStdOut: true,
		Format:   FormatConsole,
		Encoder:  EncoderCapitalColor,
		Level:    LevelDebug,
	})
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
