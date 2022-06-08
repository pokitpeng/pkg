package logger

/*
log模块创建的目的是封装zaplog，这里拆分出zapcore，可以更好得适应不同项目的需求。
并且额外增加了两种日志输出的io.writer，一种是按照文件大小分割，一种是按照文件时间分割。
可以根据自己的需求，选择性地进行配置。
*/

import (
	"io"
	"os"
	"path"
	"strings"
	"time"

	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

const (
	FormatJson            = "json"
	FormatConsole         = "console"
	EncoderCapitalColor   = "cc"
	EncoderCapital        = "c"
	EncoderLowercaseColor = "lc"
	EncoderLowercase      = "l"
	CallerPathFull        = "full"
	CallerPathShort       = "short"

	LevelDebug  = "debug"
	LevelInfo   = "info"
	LevelWarn   = "warn"
	LevelError  = "error"
	LevelDPanic = "dpanic"
	LevelPanic  = "panic"
	LevelFatal  = "fatal"
)

// Config 必填参数，终端输出基本配置
// type Config struct {
//	IsStdOut bool   // 是否输出到控制台
//	Format   string // json输出还是普通输出
//	Encoder  string // 输出大小写和颜色
//	Level    string // 输出日志级别
// }

// ConfigOption 选填参数，文件输出配置和appName
type ConfigOption struct {
	CallerSkip   int    // 调用栈第几层的栈帧信息
	IsStdOut     bool   // 是否输出到控制台
	Format       string // json输出还是普通输出
	Encoder      string // 输出大小写和颜色
	CallerPath   string // 日志调用路径
	Level        string // 输出日志级别
	IsFileOut    bool   // 是否输出到文件
	FilePath     string // 日志路径
	FileName     string // 日志名字
	MaxAge       string // 文件最大保存时间
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
func WithCallerSkipOption(c int) Option {
	return func(config *ConfigOption) {
		config.CallerSkip = c
	}
}

// WithIsStdOutOption 是否输出到控制台
func WithIsStdOutOption(s bool) Option {
	return func(config *ConfigOption) {
		config.IsStdOut = s
	}
}

// WithFormatOption 格式化方式
func WithFormatOption(s string) Option {
	return func(config *ConfigOption) {
		config.Format = s
	}
}

// WithEncoderOption 输出格式
func WithEncoderOption(s string) Option {
	return func(config *ConfigOption) {
		config.Encoder = s
	}
}

// WithCallerPathOption 日志调用路径
func WithCallerPathOption(s string) Option {
	return func(config *ConfigOption) {
		config.CallerPath = s
	}
}

// WithLevelOption 输出等级
func WithLevelOption(s string) Option {
	return func(config *ConfigOption) {
		config.Level = s
	}
}

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

// withLogAge 按照文件时间切割日志
func withLogAge(option *ConfigOption) io.Writer {
	// 应用option
	logName := path.Join(option.FilePath, option.FileName)
	rotationTime, err := time.ParseDuration(option.LogAgeSplit.RotationTime)
	if err != nil {
		panic("parse time err: " + err.Error())
	}
	maxAge, err := time.ParseDuration(option.MaxAge)
	if err != nil {
		panic(err)
	}
	writer, err := rotatelogs.New(
		logName+option.LogAgeSplit.Suffix,
		rotatelogs.WithLinkName(logName),
		rotatelogs.WithMaxAge(maxAge),             // 最多保存多久的日志
		rotatelogs.WithRotationTime(rotationTime), // 每隔多久分割
	)
	if err != nil {
		panic("new writer err: " + err.Error())
	}
	return writer
}

// withLogSize 按照文件大小切割日志
func withLogSize(option *ConfigOption) io.Writer {
	logName := path.Join(option.FilePath, option.FileName)
	maxAge, err := time.ParseDuration(option.MaxAge)
	if err != nil {
		panic("parse time err: " + err.Error())
	}
	writer := lumberjack.Logger{
		Filename:   logName,
		MaxSize:    option.LogSizeSplit.MaxSize,
		MaxBackups: option.LogSizeSplit.MaxBackups,
		MaxAge:     int(maxAge.Hours() / 24),
		Compress:   option.LogSizeSplit.Compress,
	}
	return &writer
}

// NewZapLogger ...
func NewZapLogger(settings ...Option) *zap.Logger {
	var configOption ConfigOption

	// default config
	configOption.IsStdOut = true
	configOption.Format = FormatConsole
	configOption.Encoder = EncoderCapitalColor
	configOption.Level = LevelDebug
	configOption.CallerSkip = 0
	configOption.CallerPath = CallerPathShort

	// enable option
	for _, setting := range settings {
		setting(&configOption)
	}

	var ws []zapcore.WriteSyncer
	var enc zapcore.Encoder

	var lenc func(l zapcore.Level, enc zapcore.PrimitiveArrayEncoder)
	switch strings.ToLower(configOption.Encoder) {
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

	var encodeCaller func(caller zapcore.EntryCaller, enc zapcore.PrimitiveArrayEncoder)
	switch strings.ToLower(configOption.CallerPath) {
	case CallerPathFull:
		encodeCaller = zapcore.FullCallerEncoder
	case CallerPathShort:
		encodeCaller = zapcore.ShortCallerEncoder
	default:
		encodeCaller = zapcore.ShortCallerEncoder
	}

	encoderConfig := zapcore.EncoderConfig{
		TimeKey:          "time",
		LevelKey:         "level",
		NameKey:          "logger",
		CallerKey:        "caller",
		MessageKey:       "msg",
		StacktraceKey:    "stacktrace",
		LineEnding:       zapcore.DefaultLineEnding,
		EncodeLevel:      lenc,                       // 大小写
		EncodeTime:       zapcore.RFC3339TimeEncoder, // 时间格式
		EncodeDuration:   zapcore.SecondsDurationEncoder,
		EncodeCaller:     encodeCaller, // 追踪路径
		EncodeName:       zapcore.FullNameEncoder,
		ConsoleSeparator: "  ", // console格式，字段间隔符
	}

	var level zapcore.Level
	switch strings.ToLower(configOption.Level) {
	case LevelDebug:
		level = zapcore.DebugLevel
	case LevelInfo:
		level = zapcore.InfoLevel
	case LevelWarn:
		level = zapcore.WarnLevel
	case LevelError:
		level = zapcore.ErrorLevel
	case LevelFatal:
		level = zapcore.FatalLevel
	case LevelPanic:
		level = zapcore.PanicLevel
	default:
		level = zapcore.InfoLevel
	}
	atomicLevel := zap.NewAtomicLevel()
	atomicLevel.SetLevel(level)

	switch strings.ToLower(configOption.Format) {
	case FormatJson:
		enc = zapcore.NewJSONEncoder(encoderConfig)
	case FormatConsole:
		enc = zapcore.NewConsoleEncoder(encoderConfig)
	default:
		enc = zapcore.NewJSONEncoder(encoderConfig)
	}

	if configOption.IsStdOut {
		ws = append(ws, os.Stdout)
	}

	if configOption.LogAgeSplit != nil && configOption.LogSizeSplit != nil {
		panic("LogAgeSplit and LogSizeSplit can't both set.")
	}

	if configOption.LogAgeSplit != nil {
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
	opts := []zap.Option{
		zap.AddCaller(),
		zap.Development(), // 开启开发模式，堆栈跟踪
		zap.AddStacktrace(zapcore.FatalLevel),
		zap.AddCallerSkip(configOption.CallerSkip),
	}
	return zap.New(core, opts...)
}
