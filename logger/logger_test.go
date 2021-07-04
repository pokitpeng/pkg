package log

import (
	"context"
	"os"
	"testing"

	"github.com/go-kratos/kratos/v2/log"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func TestZapLogger1(t *testing.T) {
	encoder := zapcore.EncoderConfig{
		TimeKey:        "t",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stack",
		EncodeTime:     zapcore.RFC3339TimeEncoder,
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.FullCallerEncoder,
	}
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoder),
		zapcore.NewMultiWriteSyncer(
			zapcore.AddSync(os.Stdout),
		),
		zap.NewAtomicLevelAt(zapcore.DebugLevel),
	)
	opts := []zap.Option{
		zap.AddStacktrace(
			zap.NewAtomicLevelAt(zapcore.ErrorLevel)),
		zap.AddCaller(),
		zap.AddCallerSkip(2),
		zap.Development(),
	}
	logger := zap.New(core, opts...)
	// zlog := log.NewHelper(logger)
	logger.Sugar().Infow("name", "kratos", "from")
	logger.Sugar().Infow("name", "kratos", "demo")
}

func TestZapLogger2(t *testing.T) {
	logger := NewDevLog()
	zlog := log.NewHelper(logger)
	zlog.Infow("name", "kratos", "from", "opensource")
	zlog.Infow("name", "kratos", "demo")
}

func TestZapLogger3(t *testing.T) {
	logger := NewZapLogger(Config{
		IsStdOut: true,
		Encoder:  EncoderJson,
		LEncoder: LEncoderLowercase,
		Level:    LevelDebug,
	},
		WithServiceNameOption("tests"),
	)
	logger = log.With(logger, "k", "v")
	zlog := log.NewHelper(logger)
	zlog.Infow("name", "kratos", "from", "opensource")
	zlog.Infow("name", "kratos")
	zlog.Infow("name", "kratos", "form")
	zlog.Info("hello log")
}

func TestZapLogger4(t *testing.T) {
	logger := NewZapLogger(Config{
		IsStdOut: true,
		Encoder:  EncoderJson,
		LEncoder: LEncoderLowercase,
		Level:    LevelDebug,
	},
		WithServiceNameOption("tests"),
	)
	ctx := context.WithValue(context.Background(), "trace_id", "2233")
	logger = log.With(logger, "k", "v")
	logger = log.WithContext(ctx, logger)
	zlog := log.NewHelper(logger)
	zlog.Infow("name", "kratos", "from", "opensource")
	zlog.Infow("name", "kratos")
	zlog.Infow("name", "kratos", "form")
	zlog.Info("hello log")
}

func BenchmarkZapLogger(b *testing.B) {
	devLog := NewDevLog()
	helper := log.NewHelper(devLog)
	for i := 0; i < b.N; i++ {
		helper.Debug("hello zap log")
		helper.Info("hello zap log")
		helper.Warn("hello zap log")
		helper.Error("hello zap log")
	}

	// 26300 ns/op
}
