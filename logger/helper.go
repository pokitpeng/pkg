package logger

import "go.uber.org/zap"

var helper *zap.SugaredLogger

func init() {
	if helper == nil {
		helper = NewZapLogger(WithCallerSkipOption(1)).Sugar()
	}
}

func GetLogger() *zap.SugaredLogger {
	return helper
}

func Init(options ...Option) {
	helper = NewZapLogger(options...).Sugar()
}

// WithKV must have key and value
func WithKV(args ...interface{}) *zap.SugaredLogger {
	return helper.With(args...)
}

// WithName only work in FormatConsole setting
func WithName(name string) *zap.SugaredLogger {
	return helper.Named(name)
}

func Sync() error {
	return helper.Sync()
}

// =========================================================================

func Debug(args ...interface{}) {
	helper.Debug(args...)
}

func Info(args ...interface{}) {
	helper.Info(args...)
}

func Println(args ...interface{}) {
	helper.Info(args...)
}

func Warn(args ...interface{}) {
	helper.Warn(args...)
}

func Error(args ...interface{}) {
	helper.Error(args...)
}

func Fatal(args ...interface{}) {
	helper.Fatal(args...)
}

// =========================================================================

func Debugf(format string, a ...interface{}) {
	helper.Debugf(format, a...)
}

func Infof(format string, a ...interface{}) {
	helper.Infof(format, a...)
}

func Printf(format string, a ...interface{}) {
	helper.Infof(format, a...)
}

func Warnf(format string, a ...interface{}) {
	helper.Warnf(format, a...)
}

func Errorf(format string, a ...interface{}) {
	helper.Errorf(format, a...)
}

func Fatalf(format string, a ...interface{}) {
	helper.Fatalf(format, a...)
}

// =========================================================================

func Debugw(msg string, kvs ...interface{}) {
	helper.Debugw(msg, kvs...)
}

func Infow(msg string, kvs ...interface{}) {
	helper.Infow(msg, kvs...)
}

func Warnw(msg string, kvs ...interface{}) {
	helper.Warnw(msg, kvs...)
}

func Errorw(msg string, kvs ...interface{}) {
	helper.Errorw(msg, kvs...)
}

func Fatalw(msg string, kvs ...interface{}) {
	helper.Fatalw(msg, kvs...)
}
