package log

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/lestrrat-go/file-rotatelogs"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var _ log.Logger = (*ZapLog)(nil)

// ZapLog is a logger impl.
type ZapLog struct {
	log  *zap.Logger
	ctx  context.Context
	pool *sync.Pool
}

const ()

const (
	FormatJson            = "json"
	FormatConsole         = "console"
	EncoderCapitalColor   = "cc"
	EncoderCapital        = "c"
	EncoderLowercaseColor = "lc"
	EncoderLowercase      = "l"

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
	Format   string // json输出还是普通输出
	Encoder  string // 输出大小写和颜色
	Level    string // 输出日志级别
}

// ConfigOption 选填参数，文件输出配置和appName
type ConfigOption struct {
	// CalllerSkip  int
	IsFileOut    bool   // 是否输出到文件
	FilePath     string // 日志路径
	FileName     string // 日志名字
	MaxAge       string // 文件最多保存时间
	LogSizeSplit *LogSizeSplitConfig
	LogAgeSplit  *LogAgeSplitConfig
}

type LogSizeSplitConfig struct {
	MaxSize    int  // 每个日志文件保存的最大尺寸 单位：MB
	MaxBackups int  // 日志文件最多保存多少个备份
	Compress   bool // 是否压缩
}

type LogAgeSplitConfig struct {
	Suffix       string // 分割后的文件后缀
	RotationTime string // 每多久分割一次
}

type Option func(config *ConfigOption)

// WithCallerSkipOption 控制当前调用栈第几层的栈帧信息
// func WithCallerSkipOption(c int) Option {
// 	return func(config *ConfigOption) {
// 		config.CalllerSkip = c
// 	}
// }

// WithFilePathOption 日志路径
func WithFilePathOption(s string) Option {
	return func(config *ConfigOption) {
		// 	如果没有给日志路径为空，默认输出到当前路径
		if strings.TrimSpace(config.FilePath) == "" {
			config.FilePath = "./"
		}
		config.FilePath = s
	}
}

// WithFileNameOption 日志名字
func WithFileNameOption(s string) Option {
	return func(config *ConfigOption) {
		// 	如果没有给日志名字为空，报错
		if strings.TrimSpace(s) == "" {
			panic("log file name is nil.")
		}
		config.FileName = s
	}
}

// WithMaxAgeOption 文件最多保存多少天
func WithMaxAgeOption(s string) Option {
	return func(config *ConfigOption) {
		config.MaxAge = s
	}
}

// WithLogSizeOption 文件按照大小分割配置
func WithLogSizeOption(logSizeSplitConfig *LogSizeSplitConfig) Option {
	return func(config *ConfigOption) {
		config.LogSizeSplit = logSizeSplitConfig
	}
}

// WithLogAgeOption 文件按照时间分割配置
func WithLogAgeOption(logAgeSplitConfig *LogAgeSplitConfig) Option {
	return func(config *ConfigOption) {
		config.LogAgeSplit = logAgeSplitConfig
	}
}

func NewCoreOpts(config Config) (zapcore.EncoderConfig, zap.AtomicLevel) {
	var lenc func(l zapcore.Level, enc zapcore.PrimitiveArrayEncoder)
	switch strings.ToLower(config.Encoder) {
	case EncoderCapitalColor:
		lenc = zapcore.CapitalColorLevelEncoder
	case EncoderCapital:
		lenc = zapcore.CapitalLevelEncoder
	case EncoderLowercaseColor:
		lenc = zapcore.LowercaseColorLevelEncoder
	case EncoderLowercase:
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
		StacktraceKey:    "stacktrace",
		LineEnding:       zapcore.DefaultLineEnding,
		EncodeLevel:      lenc,                       // 大小写
		EncodeTime:       zapcore.RFC3339TimeEncoder, // 时间格式
		EncodeDuration:   zapcore.SecondsDurationEncoder,
		EncodeCaller:     zapcore.ShortCallerEncoder, // 追踪路径
		EncodeName:       zapcore.FullNameEncoder,
		ConsoleSeparator: "  ", // console格式，字段间隔符
	}

	var level zapcore.Level
	switch strings.ToLower(config.Level) {
	case "debug":
		level = zapcore.DebugLevel
	case "info":
		level = zapcore.InfoLevel
	case "warn":
		level = zapcore.WarnLevel
	case "error":
		level = zapcore.ErrorLevel
	case "fatal":
		level = zapcore.FatalLevel
	case "panic":
		level = zapcore.PanicLevel
	default:
		level = zapcore.InfoLevel
	}
	atomicLevel := zap.NewAtomicLevel()
	atomicLevel.SetLevel(level)

	return encoderConfig, atomicLevel
}

// LogSize 按照文件时间切割日志
func withLogAge(option *ConfigOption) io.Writer {
	// 应用option
	logname := path.Join(option.FilePath, option.FileName)
	maxAge, err := time.ParseDuration(option.MaxAge)
	if err != nil {
		panic("parse log maxAge err: " + err.Error())
	}
	rotationTime, err := time.ParseDuration(option.LogAgeSplit.RotationTime)
	if err != nil {
		panic("parse log rotationTime err: " + err.Error())
	}
	hook, err := rotatelogs.New(
		logname+option.LogAgeSplit.Suffix,
		rotatelogs.WithLinkName(logname),
		rotatelogs.WithMaxAge(maxAge),             // 保留多久的日志
		rotatelogs.WithRotationTime(rotationTime), // 每隔多久分割
	)
	if err != nil {
		panic("new rotatelogs writer err: " + err.Error())
	}
	return hook
}

// LogSize 按照文件时间切割日志
func withLogSize(option *ConfigOption) io.Writer {
	logname := path.Join(option.FilePath, option.FileName)
	maxAge, err := time.ParseDuration(option.MaxAge)
	if err != nil {
		panic("parse log maxAge err: " + err.Error())
	}
	// 由于两个库输入类型并不相同，使用统一风格转化，可能会有一定精度损失的代价
	maxAgeDay := int(maxAge.Hours() / 24)
	hook := lumberjack.Logger{
		Filename:   logname,
		MaxSize:    option.LogSizeSplit.MaxSize,
		MaxBackups: option.LogSizeSplit.MaxBackups,
		MaxAge:     maxAgeDay,
		Compress:   option.LogSizeSplit.Compress,
	}
	return &hook
}

// NewZapLogger return a zap logger.
func NewZapLogger(config Config, settings ...Option) log.Logger {
	var configOption ConfigOption

	// 应用option
	for _, setting := range settings {
		setting(&configOption)
	}

	if configOption.LogAgeSplit != nil && configOption.LogSizeSplit != nil {
		panic("LogAgeSplit and LogSizeSplit can't both set.")
	}

	var ws []zapcore.WriteSyncer
	var enc zapcore.Encoder

	coreOpts, atomicLevel := NewCoreOpts(config)

	switch strings.ToLower(config.Format) {
	case FormatJson:
		enc = zapcore.NewJSONEncoder(coreOpts)
	case FormatConsole:
		enc = zapcore.NewConsoleEncoder(coreOpts)
	default:
		enc = zapcore.NewJSONEncoder(coreOpts)
	}

	if config.IsStdOut {
		ws = append(ws, os.Stdout)
	}

	if configOption.LogAgeSplit != nil {
		fmt.Println(configOption)
		logAgeSplit := withLogAge(&configOption)
		ws = append(ws, zapcore.AddSync(logAgeSplit))
	}
	if configOption.LogSizeSplit != nil {
		logSizeSplit := withLogSize(&configOption)
		ws = append(ws, zapcore.AddSync(logSizeSplit))
	}

	core := zapcore.NewCore(
		enc,                                // 编码器配置
		zapcore.NewMultiWriteSyncer(ws...), // 输出方式
		atomicLevel,                        // 日志级别
	)

	// var cs int
	// if configOption.CalllerSkip == 0 {
	// 	cs = 3
	// }

	opts := []zap.Option{
		zap.AddCaller(),
		zap.Development(), // 开启开发模式，堆栈跟踪
		zap.AddStacktrace(zapcore.FatalLevel),
		zap.AddCallerSkip(2),
	}

	return &ZapLog{
		log: zap.New(core, opts...),
		pool: &sync.Pool{
			New: func() interface{} {
				return new(bytes.Buffer)
			},
		}}
}

func NewLog(config Config, settings ...Option) *log.Helper {
	logger := NewZapLogger(config, settings...)
	return log.NewHelper(logger)
}

// NewDevLog 用于开发环境的log
func NewDevelopLog() *log.Helper {
	logger := NewLog(Config{
		IsStdOut: true,
		Format:   FormatConsole,
		Encoder:  EncoderCapitalColor,
		Level:    LevelDebug,
	})
	return logger
}

// Log Implementation of logger interface.
func (l *ZapLog) Log(level log.Level, keyvals ...interface{}) error {
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
