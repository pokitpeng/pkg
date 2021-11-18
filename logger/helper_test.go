package logger

import "testing"

func TestDebug(t *testing.T) {
	Debug("this is a debug msg")
	Info("this is a info msg")

	Warnf("this is a warnf msg")
	Errorf("this is a errorf msg")

	WithoutColor()

	Debug("this is a debug msg")
}
