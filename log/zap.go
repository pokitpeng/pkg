package log

import (
	"io"
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Encoder string

const (
	EncoderJson    Encoder = "json"
	EncoderConsole Encoder = "console"
)

type Option func(*config)

type config struct {
	level         zapcore.Level
	encoder       Encoder
	addCallerSkip int
	writers       []io.Writer
}

func ConfigWithLevel(l zapcore.Level) Option {
	return func(config *config) {
		config.level = l
	}
}

func ConfigWithEncoder(e Encoder) Option {
	return func(config *config) {
		config.encoder = e
	}
}

func ConfigWithAddCallerSkip(acs int) Option {
	return func(config *config) {
		config.addCallerSkip = acs
	}
}

func ConfigWithWriters(ws []io.Writer) Option {
	return func(config *config) {
		config.writers = ws
	}
}

func NewLogger(opts ...Option) *zap.SugaredLogger {
	config := &config{
		level:         zapcore.DebugLevel,
		encoder:       EncoderConsole,
		addCallerSkip: 1,
		writers:       []io.Writer{os.Stdout},
	}

	for _, opt := range opts {
		opt(config)
	}

	encoderConfig := zapcore.EncoderConfig{
		MessageKey:     "M",
		LevelKey:       "L",
		TimeKey:        "T",
		NameKey:        "N",
		CallerKey:      "C",
		StacktraceKey:  "Stack",
		SkipLineEnding: false,
		LineEnding:     "\n",
		EncodeLevel:    zapcore.CapitalLevelEncoder,
		// EncodeTime:     zapcore.RFC3339TimeEncoder,
		EncodeTime: func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
			enc.AppendString(t.Format("2006-01-02 15:04:05.000"))
		},
		EncodeDuration:   zapcore.MillisDurationEncoder,
		EncodeCaller:     zapcore.ShortCallerEncoder,
		EncodeName:       zapcore.FullNameEncoder,
		ConsoleSeparator: "  ",
	}

	var encoding zapcore.Encoder
	if config.encoder == EncoderConsole {
		encoding = zapcore.NewConsoleEncoder(encoderConfig)
	} else {
		encoding = zapcore.NewJSONEncoder(encoderConfig)
	}

	zapOpts := []zap.Option{
		zap.AddCaller(),
		zap.Development(),
		zap.AddStacktrace(zapcore.FatalLevel),
		zap.AddCallerSkip(config.addCallerSkip),
	}

	var ws []zapcore.WriteSyncer
	for _, writer := range config.writers {
		ws = append(ws, zapcore.AddSync(writer))
	}

	return zap.New(zapcore.NewTee(
		// zapcore.NewCore(encoding, zapcore.AddSync(utils.GetWriterWithAge("./tmp/debug.log")), zap.LevelEnablerFunc(func(l zapcore.Level) bool {
		// 	return l == zapcore.DebugLevel
		// })),
		// zapcore.NewCore(encoding, zapcore.AddSync(utils.GetWriterWithAge("./tmp/info.log")), zap.LevelEnablerFunc(func(l zapcore.Level) bool {
		// 	return l >= zapcore.InfoLevel
		// })),
		// zapcore.NewCore(encoding, zapcore.AddSync(utils.GetWriterWithAge("./tmp/error.log")), zap.LevelEnablerFunc(func(l zapcore.Level) bool {
		// 	return l >= zapcore.ErrorLevel
		// })),
		zapcore.NewCore(encoding, zapcore.NewMultiWriteSyncer(ws...), zap.LevelEnablerFunc(func(l zapcore.Level) bool {
			return l >= config.level
		})),
	), zapOpts...).Sugar()
}
