package logger

import (
	"testing"
)

func TestLogger(t *testing.T) {
	InitLogger(&Config{
		IsFileOut:  false,
		IsStdOut:   true,
		Level:      DebugLevel,
		Encoder:    NormalEncoder,
		LEncoder:   CapitalColor,
		FilePath:   `E:\tmp`,
		FileName:   `test.log`,
		MaxSize:    1,
		MaxBackups: 3,
		MaxAge:     30,
		Compress:   true,
	})

	Debug("debug log")
	Info("info log")
	Warn("warn log")
	Error("error log")
	// Fatal("fatal log")
	// Panic("info log")

	Debugf("debug log %v", 1)
	Infof("info log %v", 1)
	Warnf("warn log %v", 1)
	Errorf("error log %v", 1)
	// Fatalf("fatal log")
	// Panic("info log")

	Debugw("debug log", "user", "login success")
	Infow("info log", "user", "logout")
	Warnw("warn log", "trade", "repeat payment")
	Errorw("error log", "trade", "incorrect transaction password")

}

func TestLoggerDefault(t *testing.T) {
	Debug("debug log")
	Info("info log")
	Warn("warn log")
	Error("error log")
	// Fatal("fatal log")
	// Panic("info log")

	Debugf("debug log %v", 1)
	Infof("info log %v", 1)
	Warnf("warn log %v", 1)
	Errorf("error log %v", 1)
	// Fatalf("fatal log")
	// Panic("info log")

	Debugw("debug log", "user", "login success")
	Infow("info log", "user", "logout")
	Warnw("warn log", "trade", "repeat payment")
	Errorw("error log", "trade", "incorrect transaction password")
}
