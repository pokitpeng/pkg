package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
	"testing"
	"time"
)

func TestZapLog1(t *testing.T) {
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
		zap.AddCallerSkip(0),
		zap.Development(),
	}
	logger := zap.New(core, opts...).Sugar()
	logger.Infow("name", "kratos", "from")
	logger.Infow("name", "kratos", "demo")
}

func TestZapLog2(t *testing.T) {
	log := NewZapLogger().Sugar()
	log.Infow("infow msg", "kratos", "from")
	log.Infow("infow msg", "kratos", "demo")
	log.Error("err msg")
	log.Errorf("errf msg")
	log = log.With("model", "data")
	log.Warn("warn msg")
}

func TestZapLog3(t *testing.T) {
	log := NewZapLogger(
		WithFilePathOption("/tmp"),
		WithFileNameOption("aaa"),
		WithMaxAgeOption("1d"), // 保留1天内日志
		WithLogAgeOption(&LogAgeSplitConfig{
			Suffix:       ".%Y-%m-%d-%H:%M:%S",
			RotationTime: "3s", // 每小时切割日志
		}),
	).Sugar()
	log.Infow("infow msg", "kratos", "from")
	log = log.With("model", "data")

	for i := 0; i < 10; i++ {
		log.Infow("infow msg", "kratos", "demo")
		log.Error("err msg")
		log.Errorf("errf msg")
		log.Warn("warn msg")
		time.Sleep(time.Second)
	}
}

func TestZapLog4(t *testing.T) {
	log := NewZapLogger(
		WithFormatOption(FormatJson),
		WithEncoderOption(EncoderLowercase),
	).Sugar()
	log.Infow("infow msg", "kratos", "from")
	log = log.With("model", "data")
	log.Infof("info msg")
}
