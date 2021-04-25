package logger

import (
	"testing"
	"time"
)

// go test -timeout 30s -v -count=1 -run ^TestNewLoggerBase$ .
func TestNewLoggerBase(t *testing.T) {
	log := NewLogger(Config{
		IsStdOut: true,
		Encoder:  JsonEncoder,
		LEncoder: Lowercase,
		Level:    "debug",
	})
	log.Print("debug message")
	log.Printf("debug %s", "message")
	log.Printw("debug message", "model", "user")

	log.Debug("debug message")
	log.Debugf("debug %s", "message")
	log.Debugw("debug message", "model", "biz")

	log.Info("info message")
	log.Infof("info %s", "message")
	log.Infow("info message", "model", "data")

	log.Warn("warn message")
	log.Warnf("warn %s", "message")
	log.Warnw("warn message", "model", "middleware")

	log.Error("warn message")
	log.Errorf("warn %s", "message")
	log.Errorw("warn message", "model", "core")

	log = NewLogger(Config{
		IsStdOut: true,
		Encoder:  NormalEncoder,
		LEncoder: LowercaseColor,
		Level:    "debug",
	})
	log.Print("debug message")
	log.Printf("debug %s", "message")
	log.Printw("debug message", "model", "user")

	log.Debug("debug message")
	log.Debugf("debug %s", "message")
	log.Debugw("debug message", "model", "biz")

	log.Info("info message")
	log.Infof("info %s", "message")
	log.Infow("info message", "model", "data")

	log.Warn("warn message")
	log.Warnf("warn %s", "message")
	log.Warnw("warn message", "model", "middleware")

	log.Error("warn message")
	log.Errorf("warn %s", "message")
	log.Errorw("warn message", "model", "core")
}

// go test -timeout 30s -v -count=1 -run ^TestNewLoggerFile$ .
func TestNewLoggerFile(t *testing.T) {
	log := NewLogger(Config{
		IsStdOut: true,
		Encoder:  JsonEncoder,
		LEncoder: Lowercase,
		Level:    "debug",
	},
		WithFileOutOption(true),
		WithFilePathOption("./"),
		WithFileNameOption("test.log"),
		WithMaxAgeOption(10),
		WithMaxSizeOption(1),
		WithMaxBackupsOption(3),
		WithCompressOption(true),
	)
	log.Print("debug message")
	log.Printf("debug %s", "message")
	log.Printw("debug message", "model", "user")

	log.Debug("debug message")
	log.Debugf("debug %s", "message")
	log.Debugw("debug message", "model", "biz")

	log.Info("info message")
	log.Infof("info %s", "message")
	log.Infow("info message", "model", "data")

	log.Warn("warn message")
	log.Warnf("warn %s", "message")
	log.Warnw("warn message", "model", "middleware")

	log.Error("warn message")
	log.Errorf("warn %s", "message")
	log.Errorw("warn message", "model", "core")

	log.Print("debug message")
	log.Printf("debug %s", "message")
	log.Printw("debug message", "model", "user")

	log.Debug("debug message")
	log.Debugf("debug %s", "message")
	log.Debugw("debug message", "model", "biz")

	log.Info("info message")
	log.Infof("info %s", "message")
	log.Infow("info message", "model", "data")

	log.Warn("warn message")
	log.Warnf("warn %s", "message")
	log.Warnw("warn message", "model", "middleware")

	log.Error("warn message")
	log.Errorf("warn %s", "message")
	log.Errorw("warn message", "model", "core")

	for i := 0; i < 10000; i++ {
		log.Info("info message")
		log.Infof("info %s", "message")
		log.Infow("info message", "model", "data")

		log.Warn("warn message")
		log.Warnf("warn %s", "message")
		log.Warnw("warn message", "model", "middleware")
		time.Sleep(time.Millisecond * 10)
	}
}
