package telemetry

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

// SpanProvider provides a span exporter.
type SpanProvider func(ctx context.Context) (sdktrace.SpanExporter, error)

// CreateOTLPSpanProvider creates a new gRPC OTLP span provider.
// The options allow you to configure the remote endpoint and tune the behavior of the exporter.
// See https://pkg.go.dev/go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc#Option for details.
func CreateOTLPSpanProvider(options ...otlptracegrpc.Option) SpanProvider {
	return func(ctx context.Context) (sdktrace.SpanExporter, error) {
		return newOTLPExporter(ctx, options...)
	}
}

// newOTLPExporter creates a new gRPC OTLP exporter.
func newOTLPExporter(ctx context.Context, options ...otlptracegrpc.Option) (*otlptrace.Exporter, error) {
	defaultOptions := []otlptracegrpc.Option{
		otlptracegrpc.WithHeaders(map[string]string{
			"X-F5-OTEL": "GRPC",
		}),
	}

	traceClient := otlptracegrpc.NewClient(append(defaultOptions, options...)...)
	exp, err := otlptrace.New(ctx, traceClient)
	if err != nil {
		return nil, fmt.Errorf("failed to create OTLP exporter: %w", err)
	}
	return exp, nil
}
