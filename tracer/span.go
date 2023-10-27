package tracer

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	otel_trace "go.opentelemetry.io/otel/trace"
)

func TraceInfoFromContext(ctx context.Context) (traceID, spanID string) {
	spanCtx := trace.SpanContextFromContext(ctx)
	if spanCtx.HasTraceID() {
		traceID = spanCtx.TraceID().String()
	}
	if spanCtx.HasSpanID() {
		spanID = spanCtx.SpanID().String()
	}
	return traceID, spanID
}

func StartSpan(ctx context.Context, spanName string, opts ...trace.SpanStartOption) (context.Context, otel_trace.Span) {
	ctx, span := otel.GetTracerProvider().Tracer(TraceName).Start(ctx,
		spanName,
		opts...,
	// otel_trace.WithSpanKind(otel_trace.SpanKindClient),
	)
	// span.SetAttributes(attribute.Key(key).String(value))

	return ctx, span
}

func EndSpan(span otel_trace.Span, err error) {
	defer span.End()

	if err == nil {
		span.SetStatus(codes.Ok, "")
		return
	}

	span.SetStatus(codes.Error, err.Error())
	span.RecordError(err)
}
