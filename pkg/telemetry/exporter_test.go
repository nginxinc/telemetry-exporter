package telemetry_test

import (
	"context"
	"errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.opentelemetry.io/otel/attribute"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"

	"github.com/nginxinc/telemetry-exporter/pkg/telemetry"
	"github.com/nginxinc/telemetry-exporter/pkg/telemetry/telemetryfakes"
)

type exportableData struct {
	attributes []attribute.KeyValue
}

func (d exportableData) Attributes() []attribute.KeyValue {
	return d.attributes
}

var _ = Describe("Exporter", func() {
	When("SpanProvider works correctly", func() {
		var (
			fakeSpanExporter *telemetryfakes.FakeSpanExporter
			exporter         *telemetry.Exporter
			data             exportableData
		)

		BeforeEach(func() {
			fakeSpanExporter = &telemetryfakes.FakeSpanExporter{}
			provideSpanExporter := func(_ context.Context) (sdktrace.SpanExporter, error) {
				return fakeSpanExporter, nil
			}

			errorHandler := telemetry.NewErrorHandler()

			var err error
			exporter, err = telemetry.NewExporter(
				telemetry.ExporterConfig{
					SpanProvider: provideSpanExporter,
				},
				telemetry.WithGlobalOTelErrorHandler(errorHandler),
			)

			Expect(err).ToNot(HaveOccurred())

			data = exportableData{
				attributes: []attribute.KeyValue{
					attribute.String("key", "value"),
				},
			}
		})

		When("no errors occur", func() {
			It("exports data successfully", func() {
				Expect(exporter.Export(context.Background(), data)).To(Succeed())

				Expect(fakeSpanExporter.ExportSpansCallCount()).To(Equal(1))

				_, res := fakeSpanExporter.ExportSpansArgsForCall(0)

				Expect(res).To(HaveLen(1))
				Expect(res[0].Attributes()).To(Equal(data.attributes))

				Expect(fakeSpanExporter.ShutdownCallCount()).To(Equal(1))
			})
		})

		When("SpanExporter returns an error", func() {
			It("fails to export data", func() {
				testError := errors.New("test error")

				fakeSpanExporter.ExportSpansReturns(testError)

				err := exporter.Export(context.Background(), data)

				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError(testError))
				Expect(fakeSpanExporter.ShutdownCallCount()).To(Equal(1))
			})
		})

		AfterEach(func() {
			err := exporter.Shutdown(context.Background())
			Expect(err).ToNot(HaveOccurred())
		})
	})

	When("SpanProvider returns an error", func() {
		It("fails to export data", func() {
			testError := errors.New("test error")
			provideSpanExporter := func(_ context.Context) (sdktrace.SpanExporter, error) {
				return nil, testError
			}

			exporter, err := telemetry.NewExporter(
				telemetry.ExporterConfig{
					SpanProvider: provideSpanExporter,
				},
			)
			Expect(err).ToNot(HaveOccurred())

			data := exportableData{
				attributes: []attribute.KeyValue{
					attribute.String("key", "value"),
				},
			}

			err = exporter.Export(context.Background(), data)

			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError(testError))

			err = exporter.Shutdown(context.Background())
			Expect(err).ToNot(HaveOccurred())
		})
	})
})
