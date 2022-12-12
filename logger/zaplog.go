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

// Config 选填参数，文件输出配置和appName
type Config struct {
	CallerSkip   int    // 调用栈第几层的栈帧信息
	IsStdOut     bool   // 是否输出到控制台
	Format       string // json输出还是普通输出
	Encoder      string // 输出大小写和颜色
	Level        string // 输出日志级别
	IsFileOut    bool   // 是否输出到文件
	FilePath     string // 日志路径
	FileName     string // 日志名字
	LogSizeSplit *LogSizeSplitConfig
	LogAgeSplit  *LogAgeSplitConfig
}

type LogSizeSplitConfig struct {
	MaxAge     string // 文件最大保存时间
	MaxSize    int    // 每个日志文件保存的最大尺寸 单位：MB
	MaxBackups int    // 日志文件最多保存多少个备份
	Compress   bool   // 是否压缩
}

type LogAgeSplitConfig struct {
	MaxAge       string // 文件最大保存时间
	Suffix       string // 分割后的文件后缀
	RotationTime string // 每多久分割一次
}

type Option func(config *Config)

// ConfigWithCallerSkipOption 控制当前调用栈第几层的栈帧信息
func ConfigWithCallerSkipOption(c int) Option {
	return func(config *Config) {
		config.CallerSkip = c
	}
}

// ConfigWithIsStdOutOption 是否输出到控制台
func ConfigWithIsStdOutOption(s bool) Option {
	return func(config *Config) {
		config.IsStdOut = s
	}
}

// ConfigWithFormatOption 格式化方式
func ConfigWithFormatOption(s string) Option {
	return func(config *Config) {
		config.Format = s
	}
}

// ConfigWithEncoderOption 输出格式
func ConfigWithEncoderOption(s string) Option {
	return func(config *Config) {
		config.Encoder = s
	}
}

// ConfigWithLevelOption 输出等级
func ConfigWithLevelOption(s string) Option {
	return func(config *Config) {
		config.Level = s
	}
}

// ConfigWithFilePathOption 日志路径
func ConfigWithFilePathOption(s string) Option {
	return func(config *Config) {
		// 	如果没有给日志路径为空，默认输出到当前路径
		if strings.TrimSpace(config.FilePath) == "" {
			config.FilePath = "./"
		}
		config.FilePath = s
	}
}

// ConfigWithFileNameOption 日志名字
func ConfigWithFileNameOption(s string) Option {
	return func(config *Config) {
		// 	如果没有给日志名字为空，报错
		if strings.TrimSpace(s) == "" {
			panic("log file name is nil.")
		}
		config.FileName = s
	}
}

// ConfigWithLogSizeOption 文件按照大小分割配置
func ConfigWithLogSizeOption(logSizeSplitConfig *LogSizeSplitConfig) Option {
	return func(config *Config) {
		config.LogSizeSplit = logSizeSplitConfig
	}
}

// ConfigWithLogAgeOption 文件按照时间分割配置
func ConfigWithLogAgeOption(logAgeSplitConfig *LogAgeSplitConfig) Option {
	return func(config *Config) {
		config.LogAgeSplit = logAgeSplitConfig
	}
}

// ConfigWithLogAge 按照文件时间切割日志
func ConfigWithLogAge(option *Config) io.Writer {
	// 应用option
	logName := path.Join(option.FilePath, option.FileName)
	rotationTime, err := time.ParseDuration(option.LogAgeSplit.RotationTime)
	if err != nil {
		panic("parse time err: " + err.Error())
	}
	maxAge, err := time.ParseDuration(option.LogAgeSplit.MaxAge)
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

// ConfigWithLogSize 按照文件大小切割日志
func ConfigWithLogSize(option *Config) io.Writer {
	logName := path.Join(option.FilePath, option.FileName)
	maxAge, err := time.ParseDuration(option.LogSizeSplit.MaxAge)
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
	var Config Config

	// default config
	Config.IsStdOut = true
	Config.Format = FormatConsole
	Config.Encoder = EncoderCapital
	Config.Level = LevelDebug
	Config.CallerSkip = 0

	// enable option
	for _, setting := range settings {
		setting(&Config)
	}

	var ws []zapcore.WriteSyncer
	var enc zapcore.Encoder

	var lenc func(l zapcore.Level, enc zapcore.PrimitiveArrayEncoder)
	switch strings.ToLower(Config.Encoder) {
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
	switch strings.ToLower(Config.Level) {
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

	switch strings.ToLower(Config.Format) {
	case FormatJson:
		enc = zapcore.NewJSONEncoder(encoderConfig)
	case FormatConsole:
		enc = zapcore.NewConsoleEncoder(encoderConfig)
	default:
		enc = zapcore.NewJSONEncoder(encoderConfig)
	}

	if Config.IsStdOut {
		ws = append(ws, os.Stdout)
	}

	if Config.LogAgeSplit != nil && Config.LogSizeSplit != nil {
		panic("LogAgeSplit and LogSizeSplit can't both set.")
	}

	if Config.LogAgeSplit != nil {
		logAgeSplit := ConfigWithLogAge(&Config)
		ws = append(ws, zapcore.AddSync(logAgeSplit))
	}
	if Config.LogSizeSplit != nil {
		logSizeSplit := ConfigWithLogSize(&Config)
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
		zap.AddCallerSkip(Config.CallerSkip),
	}
	return zap.New(core, opts...)
}
