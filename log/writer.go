package log

import (
	"fmt"
	"io"
	"strings"
	"time"

	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"gopkg.in/natefinch/lumberjack.v2"
)

// NewWriterWithAge 根据时间切割日志
// logName eg: "./tmp" or "./tmp.log"
func NewWriterWithAge(logName string, opts ...SplitByAgeOption) io.Writer {
	suffix := ".log"
	if strings.HasSuffix(logName, suffix) {
		logName, _, _ = strings.Cut(logName, suffix)
	}

	var config = &SplitByAgeConfig{
		Format:       "%Y-%m-%dT%H:%M:%S",
		MaxAge:       time.Hour * 24 * 60,
		RotationTime: time.Hour * 24,
	}
	for _, opt := range opts {
		opt(config)
	}

	writer, err := rotatelogs.New(
		logName+"."+config.Format+suffix,
		rotatelogs.WithLinkName(logName+suffix),
		rotatelogs.WithMaxAge(config.MaxAge),             // 最多保存多久的日志
		rotatelogs.WithRotationTime(config.RotationTime), // 每隔多久分割
	)
	if err != nil {
		panic(err)
	}
	return writer
}

type SplitBySizeConfig struct {
	MaxSize    int  // 日志文件大小，单位是 MB
	MaxBackups int  // 最大过期日志保留个数
	MaxAge     int  // 保留过期文件最大时间，单位 天
	Compress   bool // 是否压缩日志，默认不压缩
}

type SplitBySizeOption func(*SplitBySizeConfig)

func SplitBySizeWithMaxSize(ms int) SplitBySizeOption {
	return func(config *SplitBySizeConfig) {
		config.MaxSize = ms
	}
}

func SplitBySizeWithMaxAge(ma int) SplitBySizeOption {
	return func(config *SplitBySizeConfig) {
		config.MaxAge = ma
	}
}

func SplitBySizeWithMaxBackups(mb int) SplitBySizeOption {
	return func(config *SplitBySizeConfig) {
		config.MaxBackups = mb
	}
}

func SplitBySizeWithCompress(c bool) SplitBySizeOption {
	return func(config *SplitBySizeConfig) {
		config.Compress = c
	}
}

// NewWriterWithSize 根据日志大小切割文件
// logName eg: "./tmp" or "./tmp.log"
func NewWriterWithSize(logName string, opts ...SplitBySizeOption) io.Writer {
	suffix := ".log"
	if strings.HasSuffix(logName, suffix) {
		logName, _, _ = strings.Cut(logName, suffix)
	}

	var config = &SplitBySizeConfig{
		MaxSize:    200,
		MaxBackups: 5,
		MaxAge:     60,
		Compress:   true,
	}
	for _, opt := range opts {
		opt(config)
	}

	return &lumberjack.Logger{
		Filename:   logName + suffix,
		MaxSize:    config.MaxSize,    // 日志文件大小，单位是 MB
		MaxBackups: config.MaxBackups, // 最大过期日志保留个数
		MaxAge:     config.MaxAge,     // 保留过期文件最大时间，单位 天
		Compress:   config.Compress,   // 是否压缩日志，默认是不压缩。这里设置为true，压缩日志
	}
}

type SplitByAgeConfig struct {
	Format       string        // 切割日期格式
	MaxAge       time.Duration // 保留过期文件最大时间
	RotationTime time.Duration // 每隔多久切割一次
}

type SplitByAgeOption func(*SplitByAgeConfig)

func SplitByAgeWithFormat(f string) SplitByAgeOption {
	return func(config *SplitByAgeConfig) {
		config.Format = f
	}
}

func SplitByAgeWithMaxAge(ma time.Duration) SplitByAgeOption {
	return func(config *SplitByAgeConfig) {
		config.MaxAge = ma
	}
}

func SplitByAgeWithRotationTime(rt string) SplitByAgeOption {
	return func(config *SplitByAgeConfig) {
		duration, err := time.ParseDuration(rt)
		if err != nil {
			panic(err)
		}
		config.RotationTime = duration
	}
}
