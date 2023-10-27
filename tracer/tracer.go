package tracer

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/baggage"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	trace_sdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	otel_trace "go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc/metadata"
)

var TraceName = "tracer-example"

func SetupTraceName(s string) {
	TraceName = s
}

// SetupProvider
//
//	eg: provider url http://127.0.0.1:14268/api/traces
//	eg: web url http://127.0.0.1:16686
func SetupProvider(providerUrl, serviceName string, sample float64, attrs ...attribute.KeyValue) (*trace_sdk.TracerProvider, error) {
	// Create the Jaeger exporter
	exp, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(providerUrl)))
	if err != nil {
		return nil, err
	}

	attrs = append(attrs, semconv.ServiceNameKey.String(serviceName))

	tp := trace_sdk.NewTracerProvider(
		// Always be sure to batch in production.
		trace_sdk.WithBatcher(exp),
		trace_sdk.WithSampler(trace_sdk.TraceIDRatioBased(sample)),
		// Record information about this application in a Resource.
		trace_sdk.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			attrs...,
		)),
	)
	otel.SetTracerProvider(tp)
	return tp, nil
}

// assert that metadataSupplier implements the TextMapCarrier interface
var _ propagation.TextMapCarrier = (*metadataSupplier)(nil)

type metadataSupplier struct {
	metadata *metadata.MD
}

func (s *metadataSupplier) Get(key string) string {
	values := s.metadata.Get(key)
	if len(values) == 0 {
		return ""
	}

	return values[0]
}

func (s *metadataSupplier) Set(key, value string) {
	s.metadata.Set(key, value)
}

func (s *metadataSupplier) Keys() []string {
	out := make([]string, 0, len(*s.metadata))
	for key := range *s.metadata {
		out = append(out, key)
	}

	return out
}

// Inject injects cross-cutting concerns from the ctx into the metadata.
func Inject(ctx context.Context, p propagation.TextMapPropagator, metadata *metadata.MD) {
	p.Inject(ctx, &metadataSupplier{
		metadata: metadata,
	})
}

// Extract extracts the metadata from ctx.
func Extract(ctx context.Context, p propagation.TextMapPropagator, metadata *metadata.MD) (
	baggage.Baggage, otel_trace.SpanContext,
) {
	ctx = p.Extract(ctx, &metadataSupplier{
		metadata: metadata,
	})

	return baggage.FromContext(ctx), otel_trace.SpanContextFromContext(ctx)
}
