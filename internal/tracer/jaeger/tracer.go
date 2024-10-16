package jaeger

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

// type Tracer interface {
// 	Start(ctx context.Context, spanName string, opts ...trace.SpanStartOption) (context.Context, trace.Span)
// }

func newExporter(ctx context.Context, endpoint string) (*otlptrace.Exporter, error) {
	client := otlptracegrpc.NewClient(
		// otlptracegrpc.WithHeaders(headers),
		otlptracegrpc.WithEndpoint(endpoint),
		otlptracegrpc.WithInsecure(),
	)

	// headers := map[string]string{
	// 	"content-type": "application/json",
	// }

	// client := otlptracehttp.NewClient(
	// 	otlptracehttp.WithHeaders(headers),
	// 	otlptracehttp.WithEndpoint(endpoint),
	// 	otlptracehttp.WithURLPath("/v1/traces"),
	// 	otlptracehttp.WithInsecure(),
	// 	// otlptracehttp.WithCompression(otlptracehttp.NoCompression),
	// )

	return otlptrace.New(ctx, client)

	// return otlptracehttp.New(ctx, otlptracehttp.WithEndpoint(url))
}

func newTraceProvider(exp *otlptrace.Exporter, serviceName string, traceIdRatioBased float64) (*sdktrace.TracerProvider, error) {
	// Ensure default SDK resources and the required service name are set.
	r, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(serviceName),
			// semconv.ServiceVersion("v0.1"),
			// attribute.String("environment", "dev"),
		),
	)

	if err != nil {
		return nil, err
	}

	return sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(
			exp,
			sdktrace.WithMaxExportBatchSize(sdktrace.DefaultMaxExportBatchSize),
			sdktrace.WithBatchTimeout(sdktrace.DefaultScheduleDelay*time.Millisecond),
			sdktrace.WithMaxExportBatchSize(sdktrace.DefaultMaxExportBatchSize),
		),
		sdktrace.WithResource(r),
		sdktrace.WithSampler(
			sdktrace.TraceIDRatioBased(traceIdRatioBased),
		),
	), nil
}

func InitTraceProvider(ctx context.Context, config Config, serviceName string) (*sdktrace.TracerProvider, error) {
	// ctx := context.Background()

	exporter, err := newExporter(ctx, config.Endpoint)
	if err != nil {
		return nil, fmt.Errorf("initialize exporter: %w", err)
	}

	tp, err := newTraceProvider(exporter, serviceName, config.TraceIDRatioBased)
	if err != nil {
		return nil, fmt.Errorf("initialize provider: %w", err)
	}

	// // Handle shutdown properly so nothing leaks.
	// defer func() { _ = tp.Shutdown(ctx) }()

	otel.SetTracerProvider(tp) // !!!!!!!!!!!

	return tp, nil
}

// func InitTracer(jaegerURL string, serviceName string) (trace.Tracer, func(ctx context.Context) error, error) {
// 	ctx := context.Background()

// 	exporter, err := newExporter(ctx, jaegerURL)
// 	if err != nil {
// 		return nil, nil, fmt.Errorf("initialize exporter: %w", err)
// 	}

// 	tp, err := newTraceProvider(exporter, serviceName)
// 	if err != nil {
// 		return nil, nil, fmt.Errorf("initialize provider: %w", err)
// 	}

// 	// // Handle shutdown properly so nothing leaks.
// 	// defer func() { _ = tp.Shutdown(ctx) }()

// 	otel.SetTracerProvider(tp) // !!!!!!!!!!!

// 	// Finally, set the tracer that can be used for this package.
// 	return tp.Tracer("rd-gate-go tracer"), tp.Shutdown, nil
// }
