package logger

import (
	"os"
	"path"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// type Logger struct {
// 	l *zap.SugaredLogger
// }

var logger *zap.SugaredLogger

const (
	JsonEncoder    = "json"
	NormalEncoder  = "normal"
	CapitalColor   = "cc"
	Capital        = "c"
	LowercaseColor = "lc"
	Lowercase      = "l"
)

const (
	DebugLevel  = zapcore.DebugLevel
	InfoLevel   = zapcore.InfoLevel
	WarnLevel   = zapcore.WarnLevel
	ErrorLevel  = zapcore.ErrorLevel
	DPanicLevel = zapcore.DPanicLevel
	PanicLevel  = zapcore.PanicLevel
	FatalLevel  = zapcore.FatalLevel
)

type Config struct {
	IsStdOut   bool          // 是否输出到控制台
	IsFileOut  bool          // 是否输出到文件
	Encoder    string        // json输出还是普通输出
	LEncoder   string        // 输出大小写和颜色
	Level      zapcore.Level // 输出日志级别
	FilePath   string        // 日志路径
	FileName   string        // 日志名字
	MaxSize    int           // 每个日志文件保存的最大尺寸 单位：MB
	MaxBackups int           // 日志文件最多保存多少个备份
	MaxAge     int           // 文件最多保存多少天
	Compress   bool          // 是否压缩
}

// InitLogger init logger
func InitLogger(config *Config) {
	hook := lumberjack.Logger{
		Filename:   path.Join(config.FilePath, config.FileName), // log path
		MaxSize:    config.MaxSize,                              // 每个日志文件保存的最大尺寸 单位：M
		MaxBackups: config.MaxBackups,                           // 日志文件最多保存多少个备份
		MaxAge:     config.MaxAge,                               // 文件最多保存多少天
		Compress:   config.Compress,                             // 是否压缩
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

	// 设置日志级别
	atomicLevel := zap.NewAtomicLevel()
	atomicLevel.SetLevel(config.Level)

	var ws []zapcore.WriteSyncer
	if config.IsStdOut {
		ws = append(ws, zapcore.AddSync(os.Stdout))
	}
	if config.IsFileOut {
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
	logger = zap.New(
		core,
		zap.AddCaller(),      // 堆栈跟踪
		zap.AddCallerSkip(1), // 行号
	).Sugar()
}

func Debug(args ...interface{}) {
	logger.Debug(args...)
}

func Debugf(template string, args ...interface{}) {
	logger.Debugf(template, args...)
}

func Debugw(msg string, keysAndValues ...interface{}) {
	logger.Debugw(msg, keysAndValues...)
}

func Info(args ...interface{}) {
	logger.Info(args...)
}

func Infof(template string, args ...interface{}) {
	logger.Infof(template, args...)
}

func Infow(msg string, keysAndValues ...interface{}) {
	logger.Infow(msg, keysAndValues...)
}

func Warn(args ...interface{}) {
	logger.Warn(args...)
}

func Warnf(template string, args ...interface{}) {
	logger.Warnf(template, args...)
}

func Warnw(msg string, keysAndValues ...interface{}) {
	logger.Warnw(msg, keysAndValues...)
}

func Error(args ...interface{}) {
	logger.Error(args...)
}

func Errorf(template string, args ...interface{}) {
	logger.Errorf(template, args...)
}

func Errorw(msg string, keysAndValues ...interface{}) {
	logger.Errorw(msg, keysAndValues...)
}

func Panic(args ...interface{}) {
	logger.Panic(args...)
}

func Panicf(template string, args ...interface{}) {
	logger.Panicf(template, args...)
}

func Panicw(msg string, keysAndValues ...interface{}) {
	logger.Panicw(msg, keysAndValues...)
}

func NewDevEnv() {
	InitLogger(&Config{
		IsFileOut: false,
		IsStdOut:  true,
		Level:     DebugLevel,
		Encoder:   NormalEncoder,
	})
}

func NewDevEnvWithColor() {
	InitLogger(&Config{
		IsFileOut: false,
		IsStdOut:  true,
		Level:     DebugLevel,
		Encoder:   NormalEncoder,
		LEncoder:  CapitalColor,
	})
}
