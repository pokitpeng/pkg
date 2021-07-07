package log

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// NewZapLog return a zap logger.
func NewZapLog(config Config, settings ...Option) *zap.SugaredLogger {
	core := NewZapCore(config, settings...)
	opts := []zap.Option{
		zap.AddCaller(),
		zap.Development(), // 开启开发模式，堆栈跟踪
		zap.AddStacktrace(zapcore.FatalLevel),
		zap.AddCallerSkip(0),
	}
	return zap.New(core, opts...).Sugar()
}
