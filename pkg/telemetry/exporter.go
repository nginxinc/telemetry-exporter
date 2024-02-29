package telemetry

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

// Data defines common telemetry data points for NGINX Kubernetes-related projects.
//
//go:generate go run -tags=generator github.com/nginxinc/telemetry-exporter/cmd/generator -type Data
type Data struct {
	// ProjectName is the name of the project.
	ProjectName string
	// ProjectVersion is the version of the project.
	ProjectVersion string
	// ProjectArchitecture is the architecture the project. For example, "amd64".
	ProjectArchitecture string
	// ClusterID is the unique id of the Kubernetes cluster where the project is installed.
	// It is the UID of the `kube-system` Namespace.
	ClusterID string
	// ClusterVersion is the Kubernetes version of the cluster.
	ClusterVersion string
	// ClusterPlatform is the Kubernetes platform of the cluster.
	ClusterPlatform string
	// DeploymentID is the unique id of the project installation in the cluster.
	DeploymentID string
	// ClusterNodeCount is the number of nodes in the cluster.
	ClusterNodeCount int64
}

// Exportable allows exporting telemetry data using the Exporter.
type Exportable interface {
	// Attributes returns a list of key-value pairs that represent the telemetry data.
	Attributes() []attribute.KeyValue
}

// ExporterConfig contains the configuration for the Exporter.
type ExporterConfig struct {
	// SpanProvider contains SpanProvider for exporting spans.
	SpanProvider SpanProvider
}

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . SpanExporter

// SpanExporter is used to generate a fake for the unit test.
type SpanExporter interface {
	sdktrace.SpanExporter
}

// Exporter exports telemetry data.
type Exporter struct {
	spanProvider  SpanProvider
	traceProvider *sdktrace.TracerProvider
	handler       *ErrorHandler
}

type optionsCfg struct {
	errorHandler *ErrorHandler
	logger       logr.Logger
}

// Option is a configuration option for the Exporter.
type Option func(*optionsCfg)

// WithGlobalOTelErrorHandler sets the global OpenTelemetry error handler.
//
// Note that the error handler captures all errors generated by the OpenTelemetry SDK.
// The Exporter uses it to catch errors that occur during exporting.
// If this option is not used, the Exporter will not be able to catch errors that occur during the export process.
//
// Warning: This option changes the global OpenTelemetry state. If OpenTelemetry is used in other parts of
// your application, the error handler will catch errors from those parts as well. As a result, the Exporter might
// return errors when exporting telemetry data, even if no error occurred.
func WithGlobalOTelErrorHandler(errorHandler *ErrorHandler) Option {
	return func(o *optionsCfg) {
		o.errorHandler = errorHandler
	}
}

// WithGlobalOTelLogger sets the global OpenTelemetry logger.
// The logger is used by the OpenTelemetry SDK to log messages.
//
// Warning: This option changes the global OpenTelemetry state. If OpenTelemetry is used in other parts of your
// application, the logger will be used for those parts as well.
func WithGlobalOTelLogger(logger logr.Logger) Option {
	return func(o *optionsCfg) {
		o.logger = logger
	}
}

// NewExporter creates a new Exporter.
func NewExporter(cfg ExporterConfig, options ...Option) (*Exporter, error) {
	var optCfg optionsCfg
	for _, opt := range options {
		opt(&optCfg)
	}

	if optCfg.errorHandler != nil {
		otel.SetErrorHandler(optCfg.errorHandler)
	}
	if (optCfg.logger != logr.Logger{}) {
		otel.SetLogger(optCfg.logger)
	}

	res, err := resource.Merge(
		resource.Default(),
		resource.NewSchemaless(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create an OTel resource: %w", err)
	}

	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithResource(res),
	)

	return &Exporter{
		spanProvider:  cfg.SpanProvider,
		traceProvider: tracerProvider,
		handler:       optCfg.errorHandler,
	}, nil
}

// Export exports telemetry data.
func (e *Exporter) Export(ctx context.Context, exportable Exportable) error {
	spanExporter, err := e.spanProvider(ctx)
	if err != nil {
		return fmt.Errorf("failed to create span exporter: %w", err)
	}

	// We create a new span processor for each export to ensure the Exporter doesn't keep a GRPC connection to
	// the OTLP endpoint in between exports.

	// We create a SpanProcessor that synchronously exports spans to the OTLP endpoint.
	// As mentioned in the NewSimpleSpanProcessor doc, it is not recommended to use this in production,
	// because it is synchronous. However, in our case, we only send one span and we want to catch errors during
	// sending, so synchronous is good for us.
	spanProcessor := sdktrace.NewSimpleSpanProcessor(spanExporter)
	defer func() {
		// This error is ignored because it happens after the span has been exported, so it is not useful.
		_ = spanProcessor.Shutdown(ctx)
	}()

	e.traceProvider.RegisterSpanProcessor(spanProcessor)
	defer e.traceProvider.UnregisterSpanProcessor(spanProcessor)

	if e.handler != nil {
		e.handler.Clear()
	}

	tracer := e.traceProvider.Tracer("product-telemetry")

	_, span := tracer.Start(ctx, "report")

	span.SetAttributes(exportable.Attributes()...)

	// Because we use a synchronous span processor, the span is exported immediately and synchronously.
	// Any error will be caught by the error handler.
	span.End()

	if e.handler != nil {
		if handlerErr := e.handler.Error(); handlerErr != nil {
			return fmt.Errorf("failed to export telemetry: %w", handlerErr)
		}
	}

	return nil
}

// Shutdown shuts down the Exporter.
func (e *Exporter) Shutdown(ctx context.Context) error {
	if err := e.traceProvider.Shutdown(ctx); err != nil {
		return fmt.Errorf("failed to shutdown OTel trace provider: %w", err)
	}
	return nil
}
