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
	MaxAge       int    // 文件最多保存天数
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
func WithMaxAgeOption(s int) Option {
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

// withLogAge 按照文件时间切割日志
func withLogAge(option *ConfigOption) io.Writer {
	// 应用option
	logname := path.Join(option.FilePath, option.FileName)
	rotationTime, err := time.ParseDuration(option.LogAgeSplit.RotationTime)
	if err != nil {
		panic("parse log rotationTime err: " + err.Error())
	}
	hook, err := rotatelogs.New(
		logname+option.LogAgeSplit.Suffix,
		rotatelogs.WithLinkName(logname),
		rotatelogs.WithMaxAge(time.Hour*time.Duration(24*option.MaxAge)), // 保留多少天的日志
		rotatelogs.WithRotationTime(rotationTime),                        // 每隔多久分割
	)
	if err != nil {
		panic("new rotatelogs writer err: " + err.Error())
	}
	return hook
}

// withLogSize 按照文件大小切割日志
func withLogSize(option *ConfigOption) io.Writer {
	logname := path.Join(option.FilePath, option.FileName)
	hook := lumberjack.Logger{
		Filename:   logname,
		MaxSize:    option.LogSizeSplit.MaxSize,
		MaxBackups: option.LogSizeSplit.MaxBackups,
		MaxAge:     option.MaxAge,
		Compress:   option.LogSizeSplit.Compress,
	}
	return &hook
}

// NewZapLog return a zap logger.
func NewZapCore(config Config, settings ...Option) zapcore.Core {
	var configOption ConfigOption

	// enable option
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
	return core
}
