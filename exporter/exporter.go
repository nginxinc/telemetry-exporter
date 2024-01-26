package exporter

import (
	"context"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
)

type Exportable interface {
	Attributes() []attribute.KeyValue
}

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . SpanExporter

type SpanExporter interface {
	sdktrace.SpanExporter
}

type SpanExporterProvider func(ctx context.Context) (sdktrace.SpanExporter, error)

func CreateOTLPSpanExporterProvider(endpoint string, secure bool) SpanExporterProvider {
	return func(ctx context.Context) (sdktrace.SpanExporter, error) {
		return newOTLPExporter(ctx, endpoint, secure)
	}
}

type ExporterConfig struct {
	ProductName          string
	ProductVersion       string
	ErrorHandler         *ErrorHandler
	SpanExporterProvider SpanExporterProvider
}

type Exporter struct {
	spanExporterProvider SpanExporterProvider
	provider             *sdktrace.TracerProvider
	handler              *ErrorHandler
}

func NewExporter(cfg ExporterConfig) (*Exporter, error) {
	res, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			"",                                   // TO-DO: do we need to provide it?
			semconv.ServiceName(cfg.ProductName), // TO-DO: do we need to provide it?
			semconv.ServiceVersion(cfg.ProductVersion), // TO-DO: do we need to provide it?
		),
	)
	if err != nil {
		return nil, err
	}

	provider := sdktrace.NewTracerProvider(
		sdktrace.WithResource(res),
	)

	return &Exporter{
		spanExporterProvider: cfg.SpanExporterProvider,
		provider:             provider,
		handler:              cfg.ErrorHandler,
	}, nil
}

func (e *Exporter) Export(ctx context.Context, data Exportable) error {
	// We register and unregister the exporter on each call to Export,
	// so that we don't keep the connection open.

	spanExporter, err := e.spanExporterProvider(ctx)
	if err != nil {
		return err
	}

	spanProcessor := sdktrace.NewSimpleSpanProcessor(spanExporter)
	defer spanProcessor.Shutdown(ctx)

	e.provider.RegisterSpanProcessor(spanProcessor)
	defer e.provider.UnregisterSpanProcessor(spanProcessor)

	// clear errors
	e.handler.Clear()

	tracer := e.provider.Tracer("product-telemetry")

	_, span := tracer.Start(ctx, "report")

	span.SetAttributes(
		data.Attributes()...,
	)

	span.End()

	if err := e.handler.Error(); err != nil {
		return err
	}

	return nil
}

func (e *Exporter) Shutdown(ctx context.Context) error {
	return e.provider.Shutdown(ctx)
}

func newOTLPExporter(
	ctx context.Context,
	endpoint string,
	secure bool,
) (*otlptrace.Exporter, error) {
	options := []otlptracegrpc.Option{
		otlptracegrpc.WithEndpoint(endpoint),
		// Uncomment the block bellow to make sure a connection to the endpoint is established before otlptrace.New() returns.
		// Not recommended. See https://github.com/grpc/grpc-go/blob/master/Documentation/anti-patterns.md#dialing-in-grpc
		//otlptracegrpc.WithDialOption(
		//	grpc.WithBlock(),
		//),
	}

	if !secure {
		options = append(options, otlptracegrpc.WithInsecure())
	}

	traceClient := otlptracegrpc.NewClient(options...)
	return otlptrace.New(ctx, traceClient)
}
