package logger

import (
	"os"
	"testing"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
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
	logger := zap.New(core, opts...)
	logger.Sugar().Infow("name", "kratos", "from")
	logger.Sugar().Infow("name", "kratos", "demo")
}

func TestZapLog2(t *testing.T) {
	log := NewZapLog(Config{
		IsStdOut: true,
		Format:   "json",
		Encoder:  "c",
		Level:    "debug",
	})
	log.Infow("infow msg", "kratos", "from")
	log.Infow("infow msg", "kratos", "demo")
	log.Error("err msg")
	log.Errorf("errf msg")
	log = log.With("model", "data")
	log.Warn("warn msg")
}

func TestZapLog3(t *testing.T) {
	log := NewZapLog(Config{
		IsStdOut: true,
		Format:   "json",
		Encoder:  "c",
		Level:    "debug",
	},
		WithFilePathOption("/tmp"),
		WithFileNameOption("aaa"),
		WithMaxAgeOption(1), // 保留1天内日志
		WithLogAgeOption(&LogAgeSplitConfig{
			Suffix:       ".%Y-%m-%d-%H:%M:%S",
			RotationTime: "3s", // 每小时切割日志
		}),
	)
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
