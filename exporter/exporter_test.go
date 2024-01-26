package exporter_test

import (
	"context"
	"errors"
	"testing"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	. "github.com/onsi/gomega"

	"github.com/nginxinc/telemetry-exporter/exporter"
	"github.com/nginxinc/telemetry-exporter/exporter/exporterfakes"
)

type data struct {
}

func (d *data) Attributes() []attribute.KeyValue {
	return []attribute.KeyValue{
		{
			Key:   "ProductID",
			Value: attribute.StringValue("NGF"),
		},
	}
}

func TestExporter(t *testing.T) {
	oldErrorHandler := otel.GetErrorHandler()
	defer otel.SetErrorHandler(oldErrorHandler)

	errorHandler := exporter.NewErrorHandler()
	logger := zap.New()

	exporter.InitOTel(logger, errorHandler)

	g := NewWithT(t)

	fakeSpanExporter := &exporterfakes.FakeSpanExporter{}

	provideSpanExporter := func(ctx context.Context) (sdktrace.SpanExporter, error) {
		return fakeSpanExporter, nil
	}

	exp, err := exporter.NewExporter(exporter.ExporterConfig{
		Endpoint:             "",
		Secure:               false,
		ProductName:          "a",
		ProductVersion:       "1",
		ErrorHandler:         errorHandler,
		SpanExporterProvider: provideSpanExporter,
	})

	g.Expect(err).ToNot(HaveOccurred())

	data := &data{}

	expectedAttributes := data.Attributes()

	err = exp.Export(context.Background(), data)

	g.Expect(err).ToNot(HaveOccurred())

	g.Expect(fakeSpanExporter.ExportSpansCallCount()).To(Equal(1))

	_, res := fakeSpanExporter.ExportSpansArgsForCall(0)

	g.Expect(res).To(HaveLen(1))
	g.Expect(res[0].Attributes()).To(Equal(expectedAttributes))

	g.Expect(fakeSpanExporter.ShutdownCallCount()).To(Equal(1))

	testError := errors.New("test error")

	fakeSpanExporter.ExportSpansReturns(testError)

	err = exp.Export(context.Background(), data)

	g.Expect(err).To(HaveOccurred())
	g.Expect(err).To(MatchError(testError))
}
