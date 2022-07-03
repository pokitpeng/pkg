package logger

// 调用栈深度有问题

import (
	"context"
	"fmt"
	"io"
	"testing"

	"github.com/opentracing/opentracing-go"
	"github.com/uber/jaeger-client-go"
	jcf "github.com/uber/jaeger-client-go/config"
)

func TestDebug(t *testing.T) {
	Debug("this is a debug msg")
	Info("this is a info msg")

	Warnf("this is a warnf msg")
	Errorf("this is a errorf msg")

	Debug("this is a debug msg")

	Init(
		WithEncoderOption(EncoderLowercase),
		WithFormatOption(FormatJson),
		WithCallerSkipOption(1),
	)
	Debug("this is a debug msg")

	l1 := WithKV("model", "test")
	l1.Info("this is a info msg")

	l2 := WithName("hello")
	l2.Info("this is a info msg")
}

func TestStdLogger(t *testing.T) {
	InitStandardLogger()
	Debug("this is debug msg")
	Info("this is info msg")
}

func TestDevLogger(t *testing.T) {
	InitDevelopmentLogger()
	Debug("this is debug msg")
	Info("this is info msg")
}

// Set global trace provider
func setTracerProvider(serviceName, collectorEndpoint string) io.Closer {
	cfg := jcf.Configuration{
		ServiceName: serviceName,
		// 将采样频率设置为1，每一个span都记录，方便查看测试结果
		Sampler: &jcf.SamplerConfig{
			Type:  jaeger.SamplerTypeConst,
			Param: 1,
		},
		Reporter: &jcf.ReporterConfig{
			LogSpans: false,
			// 将span发往jaeger-collector的服务地址
			CollectorEndpoint: collectorEndpoint,
		},
	}
	closer, err := cfg.InitGlobalTracer(serviceName, jcf.Logger(jaeger.StdLogger))
	if err != nil {
		panic(fmt.Sprintf("ERROR: cannot init Jaeger: %v\n", err))
	}
	return closer
}

func TestTraceContext(t *testing.T) {
	url := "http://106.75.217.186:14268/api/traces"
	closer := setTracerProvider("LogTest", url)
	defer closer.Close()

	ctx := context.Background()
	span := opentracing.GlobalTracer().StartSpan("GetFeed")
	defer span.Finish()
	newctx := opentracing.ContextWithSpan(ctx, span)
	WithContext(newctx).Info("with trace info")
}
