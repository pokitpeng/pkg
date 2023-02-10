package tracer

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/pokitpeng/pkg/log"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	"google.golang.org/grpc/metadata"
)

func TestSetupProvider(t *testing.T) {
	tp, err := SetupProvider("http://127.0.0.1:14268/api/traces", "tracer", 1)
	if err != nil {
		t.Fatal(err)
	}
	defer tp.Shutdown(context.Background())

	ctx := context.Background()
	redis(ctx)
}

func redis(ctx context.Context) {
	var err error
	newCtx, span := StartSpan(ctx, "redis")
	// 设置tag
	span.SetAttributes(attribute.Bool("connected", false))
	span.SetAttributes(attribute.String("module", "cache"))

	// log output
	span.AddEvent("event-example")

	defer EndSpan(span, err)

	err = errors.New("connect to resid 127.0.0.1:6379 refused")

	time.Sleep(time.Millisecond * 10)
	log.WithContext(newCtx).Error(err)

	mysql(newCtx)
}

func mysql(ctx context.Context) {
	newCtx, span := StartSpan(ctx, "mysql")
	defer EndSpan(span, nil)
	time.Sleep(time.Millisecond * 200)
	log.WithContext(newCtx).Info("mysql connect success")
}

func TestExtract(t *testing.T) {
	ctx := context.Background()
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{}, propagation.Baggage{}))
	propagator := otel.GetTextMapPropagator()
	md := metadata.MD{}
	Inject(ctx, propagator, &md)

	ctx = context.Background()
	md = metadata.MD{}
	md.Set("traceparent", "1")
	md.Set("tracestate", "2")
	_, spanCtx := Extract(ctx, propagator, &md)
	_ = spanCtx
}
