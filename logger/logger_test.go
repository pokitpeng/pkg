package logger

import (
	"testing"
)

// go test -v -count=1 -run ^TestNewLogger$ .
func TestNewLogger(t *testing.T) {
	log := NewLogger(Config{
		IsStdOut: true,
		Encoder:  EncoderJson,
		LEncoder: LEncoderLowercase,
		Level:    LevelDebug,
	},
		WithServiceNameOption("test"),
	)

	log.Print("debug message")
	log.Printf("debug %s", "message")
	log.Printw("debug message", "model", "user")

	log.Debug("debug message")
	log.Debugf("debug %s", "message")
	log.Debugw("debug message", "model", "dao")

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

// go test -v -count=1 -run ^TestNewLogger .
func TestNewDevLog(t *testing.T) {
	log := NewDevLog()

	log.Print("debug message")
	log.Printf("debug %s", "message")
	log.Printw("debug message", "model", "user")

	log.Debug("debug message")
	log.Debugf("debug %s", "message")
	log.Debugw("debug message", "model", "dao")

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
