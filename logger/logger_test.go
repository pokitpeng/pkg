package log

import (
	"context"
	"os"
	"testing"
	"time"

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
	log := NewDevelopLog()
	log.Infow("name", "kratos", "from", "opensource")
	log.Infow("name", "kratos", "demo")
	log.Error("err msg")
}

func TestZapLogger3(t *testing.T) {
	log := NewLog(Config{
		IsStdOut: true,
		Format:   FormatJson,
		Encoder:  EncoderLowercase,
		Level:    LevelDebug,
	},
		// WithCallerSkipOption(4),
		WithFilePathOption("./"),
		WithFileNameOption("aaa"),
		WithMaxAgeOption("3s"), // 保留3s内日志
		WithLogAgeOption(&LogAgeSplitConfig{
			Suffix:       ".%Y-%m-%d-%H:%M:%S",
			RotationTime: "1s", // 每秒切割日志
		}),
	)

	for i := 0; i < 12; i++ {
		log.Infow("name", "kratos", "from", "opensource")
		log.Infow("name", "kratos")
		log.Infow("name", "kratos", "form")
		log.Info("hello log")
		time.Sleep(time.Second)
	}
}

func TestZapLogger4(t *testing.T) {
	logger := NewZapLogger(Config{
		IsStdOut: true,
		Format:   FormatJson,
		Encoder:  EncoderLowercase,
		Level:    LevelDebug,
	})
	ctx := context.WithValue(context.Background(), "trace_id", "2233")
	logger = log.WithContext(ctx, logger)
	zlog := log.NewHelper(log.With(logger, "k", "v"))
	zlog.Infow("name", "kratos", "from", "opensource")
	zlog.Infow("name", "kratos")
	zlog.Infow("name", "kratos", "form")
	zlog.Info("hello log")
}

func BenchmarkZapLogger(b *testing.B) {
	log := NewDevelopLog()
	for i := 0; i < b.N; i++ {
		log.Debug("hello zap log")
		log.Info("hello zap log")
		log.Warn("hello zap log")
		log.Error("hello zap log")
	}

	// 26300 ns/op
}
