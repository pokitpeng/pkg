package log

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path"
	"strings"
	"sync"

	"github.com/go-kratos/kratos/v2/log"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var _ log.Logger = (*ZapLogger)(nil)

// ZapLogger is a logger impl.
type ZapLogger struct {
	log  *zap.Logger
	ctx  context.Context
	pool *sync.Pool
}

const (
	EncoderJson = iota
	EncoderConsole
	LEncoderCapitalColor
	LEncoderCapital
	LEncoderLowercaseColor
	LEncoderLowercase
)

const (
	LevelDebug  = "debug"
	LevelInfo   = "info"
	LevelWarn   = "warn"
	LevelError  = "error"
	LevelDPanic = "dpanic"
	LevelPanic  = "panic"
	LevelFatal  = "fatal"
)

// Config 必填参数，终端输出基本配置
type Config struct {
	IsStdOut bool   // 是否输出到控制台
	Encoder  int    // json输出还是普通输出
	LEncoder int    // 输出大小写和颜色
	Level    string // 输出日志级别
}

// ConfigOption 选填参数，文件输出配置和appName
type ConfigOption struct {
	ServiceName string
	IsFileOut   bool   // 是否输出到文件
	FilePath    string // 日志路径
	FileName    string // 日志名字
	MaxSize     int    // 每个日志文件保存的最大尺寸 单位：MB
	MaxBackups  int    // 日志文件最多保存多少个备份
	MaxAge      int    // 文件最多保存多少天
	Compress    bool   // 是否压缩
}

type Option func(config *ConfigOption)

// WithServiceNameOption app名字设置
func WithServiceNameOption(s string) Option {
	return func(config *ConfigOption) {
		config.ServiceName = s
	}
}

// WithFileOutOption 是否输出到文件
func WithFileOutOption(b bool) Option {
	return func(config *ConfigOption) {
		config.IsFileOut = b
	}
}

// WithFilePathOption 日志路径
func WithFilePathOption(s string) Option {
	return func(config *ConfigOption) {
		// 如果设置了路径就默认开启文件输出
		config.IsFileOut = true
		config.FilePath = s
	}
}

// WithFileNameOption 日志名字
func WithFileNameOption(s string) Option {
	return func(config *ConfigOption) {
		// 	如果设置了日志名字，就默认开启文件输出
		// 	如果没有给日志路径，默认输出到当前路径
		config.IsFileOut = true
		if strings.TrimSpace(config.FilePath) == "" {
			config.FilePath = "./"
		}
		config.FileName = s
	}
}

// WithMaxSizeOption 每个日志文件保存的最大尺寸 单位：MB
func WithMaxSizeOption(i int) Option {
	return func(config *ConfigOption) {
		config.MaxSize = i
	}
}

// WithMaxBackupsOption 日志文件最多保存多少个备份
func WithMaxBackupsOption(i int) Option {
	return func(config *ConfigOption) {
		config.MaxBackups = i
	}
}

// WithMaxAgeOption 文件最多保存多少天
func WithMaxAgeOption(i int) Option {
	return func(config *ConfigOption) {
		config.MaxAge = i
	}
}

// WithCompressOption 是否压缩日志
func WithCompressOption(b bool) Option {
	return func(config *ConfigOption) {
		config.Compress = b
	}
}

// NewZapLogger return a zap logger.
func NewZapLogger(config Config, options ...Option) log.Logger {
	var configOption ConfigOption

	// 应用option
	for _, option := range options {
		option(&configOption)
	}

	var lenc func(l zapcore.Level, enc zapcore.PrimitiveArrayEncoder)
	switch config.LEncoder {
	case LEncoderCapitalColor:
		lenc = zapcore.CapitalColorLevelEncoder
	case LEncoderCapital:
		lenc = zapcore.CapitalLevelEncoder
	case LEncoderLowercaseColor:
		lenc = zapcore.LowercaseColorLevelEncoder
	case LEncoderLowercase:
		lenc = zapcore.LowercaseLevelEncoder
	default:
		lenc = zapcore.CapitalLevelEncoder
	}

	encoderConfig := zapcore.EncoderConfig{
		TimeKey:   "time",
		LevelKey:  "level",
		NameKey:   "logger",
		CallerKey: "caller",
		// MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    lenc,                           // 大小写
		EncodeTime:     zapcore.RFC3339TimeEncoder,     // 时间格式
		EncodeDuration: zapcore.SecondsDurationEncoder, //
		EncodeCaller:   zapcore.ShortCallerEncoder,     // 追踪路径
		EncodeName:     zapcore.FullNameEncoder,
	}

	var level zapcore.Level
	switch strings.ToUpper(config.Level) {
	default:
		level = zapcore.InfoLevel
	case "DEBUG":
		level = zapcore.DebugLevel
	case "INFO":
		level = zapcore.InfoLevel
	case "WARN":
		level = zapcore.WarnLevel
	case "ERROR":
		level = zapcore.ErrorLevel
	case "FATAL":
		level = zapcore.FatalLevel
	case "PANIC":
		level = zapcore.PanicLevel
	}

	hook := lumberjack.Logger{
		Filename:   path.Join(configOption.FilePath, configOption.FileName), // log path
		MaxSize:    configOption.MaxSize,                                    // 每个日志文件保存的最大尺寸 单位：M
		MaxBackups: configOption.MaxBackups,                                 // 日志文件最多保存多少个备份
		MaxAge:     configOption.MaxAge,                                     // 文件最多保存多少天
		Compress:   configOption.Compress,                                   // 是否压缩
	}

	// 设置日志级别
	atomicLevel := zap.NewAtomicLevel()
	atomicLevel.SetLevel(level)

	var ws []zapcore.WriteSyncer
	if config.IsStdOut {
		ws = append(ws, zapcore.AddSync(os.Stdout))
	}
	if configOption.IsFileOut {
		ws = append(ws, zapcore.AddSync(&hook))
	}

	var enc zapcore.Encoder

	switch config.Encoder {
	case EncoderJson:
		enc = zapcore.NewJSONEncoder(encoderConfig)
	case EncoderConsole:
		enc = zapcore.NewConsoleEncoder(encoderConfig)
	default:
		enc = zapcore.NewJSONEncoder(encoderConfig)
	}

	core := zapcore.NewCore(
		enc,                                // 编码器配置
		zapcore.NewMultiWriteSyncer(ws...), // 输出方式
		atomicLevel,                        // 日志级别
	)

	opts := []zap.Option{
		zap.AddCaller(),
		zap.Development(), // 开启开发模式，堆栈跟踪
		zap.AddStacktrace(zapcore.FatalLevel),
		zap.AddCallerSkip(3),
	}
	if configOption.ServiceName != "" {
		opts = append(opts, zap.Fields(zap.String("serviceName", configOption.ServiceName)))
	}

	return &ZapLogger{
		log: zap.New(core, opts...),
		pool: &sync.Pool{
			New: func() interface{} {
				return new(bytes.Buffer)
			},
		}}
}

// NewDevLog 用于测试环境的log
func NewDevLog() log.Logger {
	return NewZapLogger(Config{
		IsStdOut: true,
		Encoder:  EncoderConsole,
		LEncoder: LEncoderLowercaseColor,
		Level:    LevelDebug,
	})
}

// Log Implementation of logger interface.
func (l *ZapLogger) Log(level log.Level, keyvals ...interface{}) error {
	if len(keyvals) == 0 {
		return nil
	}
	if len(keyvals)%2 != 0 {
		keyvals = append(keyvals, "")
	}
	buf := l.pool.Get().(*bytes.Buffer)
	var fields []zap.Field
	if traceId := getTraceId(l.ctx); traceId != "" {
		fields = append(fields, zap.String("trace_id", traceId))
	}
	for i := 0; i < len(keyvals); i += 2 {
		fields = append(fields, zap.Any(fmt.Sprint(keyvals[i]), fmt.Sprint(keyvals[i+1])))
	}
	switch level {
	case log.LevelDebug:
		l.log.Debug(buf.String(), fields...)
	case log.LevelInfo:
		l.log.Info(buf.String(), fields...)
	case log.LevelWarn:
		l.log.Warn(buf.String(), fields...)
	case log.LevelError:
		l.log.Error(buf.String(), fields...)
	}
	buf.Reset()
	l.pool.Put(buf)
	return nil
}

// get trace id
func getTraceId(ctx context.Context) string {
	var traceID string
	if tid := trace.SpanContextFromContext(ctx).TraceID(); tid.IsValid() {
		traceID = tid.String()
	}
	return traceID
}
