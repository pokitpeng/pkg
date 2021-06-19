package logger

import (
	"os"
	"path"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

type Log struct {
	logger *zap.Logger
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
	AppName    string
	IsFileOut  bool   // 是否输出到文件
	FilePath   string // 日志路径
	FileName   string // 日志名字
	MaxSize    int    // 每个日志文件保存的最大尺寸 单位：MB
	MaxBackups int    // 日志文件最多保存多少个备份
	MaxAge     int    // 文件最多保存多少天
	Compress   bool   // 是否压缩
}

type Option func(config *ConfigOption)

// WithServiceNameOption app名字设置
func WithServiceNameOption(s string) Option {
	return func(config *ConfigOption) {
		config.AppName = s
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

func NewLogger(config Config, options ...Option) *Log {
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
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
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
		zap.AddCallerSkip(1),
	}
	if configOption.AppName != "" {
		opts = append(opts, zap.Fields(zap.String("serviceName", configOption.AppName)))
	}

	return &Log{logger: zap.New(core, opts...)}
}

func (log *Log) Debug(args ...interface{}) {
	log.logger.Sugar().Debug(args...)
}

func (log *Log) Debugf(template string, args ...interface{}) {
	log.logger.Sugar().Debugf(template, args...)
}

func (log *Log) Debugw(msg string, keysAndValues ...interface{}) {
	log.logger.Sugar().Debugw(msg, keysAndValues...)
}

func (log *Log) Print(args ...interface{}) {
	log.logger.Sugar().Debug(args...)
}

func (log *Log) Printf(template string, args ...interface{}) {
	log.logger.Sugar().Debugf(template, args...)
}

func (log *Log) Printw(msg string, keysAndValues ...interface{}) {
	log.logger.Sugar().Debugw(msg, keysAndValues...)
}

func (log *Log) Info(args ...interface{}) {
	log.logger.Sugar().Info(args...)
}

func (log *Log) Infof(template string, args ...interface{}) {
	log.logger.Sugar().Infof(template, args...)
}

func (log *Log) Infow(msg string, keysAndValues ...interface{}) {
	log.logger.Sugar().Infow(msg, keysAndValues...)
}

func (log *Log) Warn(args ...interface{}) {
	log.logger.Sugar().Warn(args...)
}

func (log *Log) Warnf(template string, args ...interface{}) {
	log.logger.Sugar().Warnf(template, args...)
}

func (log *Log) Warnw(msg string, keysAndValues ...interface{}) {
	log.logger.Sugar().Warnw(msg, keysAndValues...)
}

func (log *Log) Error(args ...interface{}) {
	log.logger.Sugar().Error(args...)
}

func (log *Log) Errorf(template string, args ...interface{}) {
	log.logger.Sugar().Errorf(template, args...)
}

func (log *Log) Errorw(msg string, keysAndValues ...interface{}) {
	log.logger.Sugar().Errorw(msg, keysAndValues...)
}

func (log *Log) Fatal(args ...interface{}) {
	log.logger.Sugar().Fatal(args...)
}

func (log *Log) Fatalf(template string, args ...interface{}) {
	log.logger.Sugar().Fatalf(template, args...)
}

func (log *Log) Fatalw(msg string, keysAndValues ...interface{}) {
	log.logger.Sugar().Fatalw(msg, keysAndValues...)
}

func (log *Log) Panic(args ...interface{}) {
	log.logger.Sugar().Panic(args...)
}

func (log *Log) Panicf(template string, args ...interface{}) {
	log.logger.Sugar().Panicf(template, args...)
}

func (log *Log) Panicw(msg string, keysAndValues ...interface{}) {
	log.logger.Sugar().Panicw(msg, keysAndValues...)
}

// ======================================

// NewDevLog 用于测试环境的log
func NewDevLog() *Log {
	return NewLogger(Config{
		IsStdOut: true,
		Encoder:  EncoderConsole,
		LEncoder: LEncoderLowercaseColor,
		Level:    LevelDebug,
	})
}
