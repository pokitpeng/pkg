package logger

import (
	"os"
	"path"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

type log struct {
	logger *zap.SugaredLogger
}

const (
	JsonEncoder = iota
	NormalEncoder
	CapitalColor
	Capital
	LowercaseColor
	Lowercase
)

const (
	DebugLevel  = "debug"
	InfoLevel   = "info"
	WarnLevel   = "warn"
	ErrorLevel  = "error"
	DPanicLevel = "dpanic"
	PanicLevel  = "panic"
	FatalLevel  = "fatal"
)

// Config 必填参数，终端输出基本配置
type Config struct {
	IsStdOut bool   // 是否输出到控制台
	Encoder  int    // json输出还是普通输出
	LEncoder int    // 输出大小写和颜色
	Level    string // 输出日志级别
}

// ConfigFile 选填参数，文件输出配置
type ConfigFile struct {
	IsFileOut  bool   // 是否输出到文件
	FilePath   string // 日志路径
	FileName   string // 日志名字
	MaxSize    int    // 每个日志文件保存的最大尺寸 单位：MB
	MaxBackups int    // 日志文件最多保存多少个备份
	MaxAge     int    // 文件最多保存多少天
	Compress   bool   // 是否压缩
}

type Option func(config *ConfigFile)

// IsStdOut 是否输出到文件
func IsFileOut(b bool) Option {
	return func(config *ConfigFile) {
		config.IsFileOut = b
	}
}

// FilePath 日志路径
func FilePath(s string) Option {
	return func(config *ConfigFile) {
		// 如果设置了路径就默认开启文件输出
		config.IsFileOut = true
		config.FilePath = s
	}
}

// FileName 日志名字
func FileName(s string) Option {
	return func(config *ConfigFile) {
		// 	如果设置了日志名字，就默认开启文件输出
		// 	如果没有给日志路径，默认输出到当前路径
		config.IsFileOut = true
		if strings.TrimSpace(config.FilePath) == "" {
			config.FilePath = "./"
		}
		config.FileName = s
	}
}

// MaxSize 每个日志文件保存的最大尺寸 单位：MB
func MaxSize(i int) Option {
	return func(config *ConfigFile) {
		config.MaxSize = i
	}
}

// MaxBackups 日志文件最多保存多少个备份
func MaxBackups(i int) Option {
	return func(config *ConfigFile) {
		config.MaxBackups = i
	}
}

// MaxAge 文件最多保存多少天
func MaxAge(i int) Option {
	return func(config *ConfigFile) {
		config.MaxAge = i
	}
}

// Compress 是否压缩日志
func Compress(b bool) Option {
	return func(config *ConfigFile) {
		config.Compress = b
	}
}

func NewLogger(config Config, options ...Option) *log {
	var configFile ConfigFile

	// 应用option
	for _, option := range options {
		option(&configFile)
	}

	var lenc func(l zapcore.Level, enc zapcore.PrimitiveArrayEncoder)
	switch config.LEncoder {
	case CapitalColor:
		lenc = zapcore.CapitalColorLevelEncoder
	case Capital:
		lenc = zapcore.CapitalLevelEncoder
	case LowercaseColor:
		lenc = zapcore.LowercaseColorLevelEncoder
	case Lowercase:
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
		Filename:   path.Join(configFile.FilePath, configFile.FileName), // log path
		MaxSize:    configFile.MaxSize,                                  // 每个日志文件保存的最大尺寸 单位：M
		MaxBackups: configFile.MaxBackups,                               // 日志文件最多保存多少个备份
		MaxAge:     configFile.MaxAge,                                   // 文件最多保存多少天
		Compress:   configFile.Compress,                                 // 是否压缩
	}

	// 设置日志级别
	atomicLevel := zap.NewAtomicLevel()
	atomicLevel.SetLevel(level)

	var ws []zapcore.WriteSyncer
	if config.IsStdOut {
		ws = append(ws, zapcore.AddSync(os.Stdout))
	}
	if configFile.IsFileOut {
		ws = append(ws, zapcore.AddSync(&hook))
	}

	var enc zapcore.Encoder

	switch config.Encoder {
	case JsonEncoder:
		enc = zapcore.NewJSONEncoder(encoderConfig)
	case NormalEncoder:
		enc = zapcore.NewConsoleEncoder(encoderConfig)
	default:
		enc = zapcore.NewJSONEncoder(encoderConfig)
	}

	core := zapcore.NewCore(
		enc,                                // 编码器配置
		zapcore.NewMultiWriteSyncer(ws...), // 输出方式
		atomicLevel,                        // 日志级别
	)

	// 设置初始化字段
	// filed := zap.Fields(zap.String("serviceName", "serviceName"))
	// 构造日志
	logger := zap.New(
		core,
		zap.AddCaller(),      // 堆栈跟踪
		zap.AddCallerSkip(1), // 行号
	).Sugar()
	return &log{logger}
}

func (log *log) Debug(args ...interface{}) {
	log.logger.Debug(args...)
}

func (log *log) Debugf(template string, args ...interface{}) {
	log.logger.Debugf(template, args...)
}

func (log *log) Debugw(msg string, keysAndValues ...interface{}) {
	log.logger.Debugw(msg, keysAndValues...)
}

func (log *log) Print(args ...interface{}) {
	log.logger.Debug(args...)
}

func (log *log) Printf(template string, args ...interface{}) {
	log.logger.Debugf(template, args...)
}

func (log *log) Printw(msg string, keysAndValues ...interface{}) {
	log.logger.Debugw(msg, keysAndValues...)
}

func (log *log) Info(args ...interface{}) {
	log.logger.Info(args...)
}

func (log *log) Infof(template string, args ...interface{}) {
	log.logger.Infof(template, args...)
}

func (log *log) Infow(msg string, keysAndValues ...interface{}) {
	log.logger.Infow(msg, keysAndValues...)
}

func (log *log) Warn(args ...interface{}) {
	log.logger.Warn(args...)
}

func (log *log) Warnf(template string, args ...interface{}) {
	log.logger.Warnf(template, args...)
}

func (log *log) Warnw(msg string, keysAndValues ...interface{}) {
	log.logger.Warnw(msg, keysAndValues...)
}

func (log *log) Error(args ...interface{}) {
	log.logger.Error(args...)
}

func (log *log) Errorf(template string, args ...interface{}) {
	log.logger.Errorf(template, args...)
}

func (log *log) Errorw(msg string, keysAndValues ...interface{}) {
	log.logger.Errorw(msg, keysAndValues...)
}

func (log *log) Fatal(args ...interface{}) {
	log.logger.Fatal(args...)
}

func (log *log) Fatalf(template string, args ...interface{}) {
	log.logger.Fatalf(template, args...)
}

func (log *log) Fatalw(msg string, keysAndValues ...interface{}) {
	log.logger.Fatalw(msg, keysAndValues...)
}

func (log *log) Panic(args ...interface{}) {
	log.logger.Panic(args...)
}

func (log *log) Panicf(template string, args ...interface{}) {
	log.logger.Panicf(template, args...)
}

func (log *log) Panicw(msg string, keysAndValues ...interface{}) {
	log.logger.Panicw(msg, keysAndValues...)
}
