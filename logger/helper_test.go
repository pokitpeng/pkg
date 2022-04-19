package logger

// 调用栈深度有问题

import "testing"

func TestDebug(t *testing.T) {
	Debug("this is a debug msg")
	Info("this is a info msg")

	Warnf("this is a warnf msg")
	Errorf("this is a errorf msg")

	Debug("this is a debug msg")

	Init(
		WithEncoderOption(EncoderLowercase),
		WithFormatOption(FormatJson),
		WithCallerSkipOption(1),
	)
	Debug("this is a debug msg")

	l1 := WithKV("model", "test")
	l1.Info("this is a info msg")

	l2 := WithName("hello")
	l2.Info("this is a info msg")
}
