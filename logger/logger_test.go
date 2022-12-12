package logger

import (
	"context"
	"testing"
	"time"

	"github.com/go-kratos/kratos/v2/log"
)

func TestZapLogger3(t *testing.T) {
	log := log.NewHelper(NewLogger(ConfigWithCallerSkipOption(3),
		ConfigWithFilePathOption("./"),
		ConfigWithFileNameOption("aaa"),
		ConfigWithLogAgeOption(&LogAgeSplitConfig{
			MaxAge:       "720h",
			Suffix:       ".%Y-%m-%d-%H:%M:%S",
			RotationTime: "1s", // 每秒切割日志
		})))

	for i := 0; i < 12; i++ {
		log.Infow("name", "kratos", "from", "opensource")
		log.Infow("name", "kratos")
		log.Infow("name", "kratos", "form")
		log.Info("hello log")
		time.Sleep(time.Second)
	}
}

func TestZapLogger4(t *testing.T) {
	logger := NewLogger(ConfigWithCallerSkipOption(4))
	var traceKey = "trace_id"
	ctx := context.WithValue(context.Background(), traceKey, "2233")
	//logger = log.WithContext(ctx, logger)
	zlog := log.NewHelper(log.With(logger, traceKey, ctx.Value(traceKey)))
	zlog.Infow("name", "kratos", "from", "opensource")
	zlog.Infow("name", "kratos")
	zlog.Infow("name", "kratos", "form")
	zlog.Info("hello log")
}

func BenchmarkZapLogger(b *testing.B) {
	log := log.NewHelper(NewLogger(ConfigWithCallerSkipOption(3)))
	for i := 0; i < b.N; i++ {
		log.Debug("hello zap log")
		log.Info("hello zap log")
		log.Warn("hello zap log")
		log.Error("hello zap log")
	}

	// 26300 ns/op
}
