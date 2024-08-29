package tests

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"sync"

	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"

	"github.com/nginxinc/telemetry-exporter/pkg/telemetry"
)

type telemetryData struct {
	ResourceCount int
}

func (d *telemetryData) Attributes() []attribute.KeyValue {
	return []attribute.KeyValue{
		attribute.Int("resourceCount", d.ResourceCount),
	}
}

type matchingLogConsumer struct {
	expectedSubstrings map[string]struct{}
	sync               sync.Mutex
}

func (c *matchingLogConsumer) Accept(log testcontainers.Log) {
	c.sync.Lock()
	defer c.sync.Unlock()

	line := string(log.Content)

	for k := range c.expectedSubstrings {
		if strings.Contains(line, k) {
			delete(c.expectedSubstrings, k)
			break
		}
	}
}

func (c *matchingLogConsumer) setExpectedSubstrings(substrings []string) {
	c.sync.Lock()
	defer c.sync.Unlock()

	c.expectedSubstrings = make(map[string]struct{}, len(substrings))
	for _, s := range substrings {
		c.expectedSubstrings[s] = struct{}{}
	}
}

func (c *matchingLogConsumer) unmatchedCount() int {
	c.sync.Lock()
	defer c.sync.Unlock()

	return len(c.expectedSubstrings)
}

var _ = Describe("Exporter", func() {
	var (
		lc        *matchingLogConsumer
		exporter  *telemetry.Exporter
		collector testcontainers.Container
		ctx       context.Context
	)

	BeforeEach(func() {
		ctx = context.Background()

		//  Run the collector container

		collectorImage := "otel/opentelemetry-collector-contrib"
		// renovate: datasource=docker depName=otel/opentelemetry-collector-contrib
		collectorImageVersion := "0.108.0"

		const collectorCfgName = "collector.yaml"

		lc = &matchingLogConsumer{}

		req := testcontainers.ContainerRequest{
			Image: collectorImage + ":" + collectorImageVersion,
			Files: []testcontainers.ContainerFile{
				{
					HostFilePath:      "./" + collectorCfgName,
					ContainerFilePath: "/" + collectorCfgName,
					FileMode:          0o444,
				},
			},
			ExposedPorts: []string{"4317/tcp"},
			WaitingFor:   wait.ForLog("Everything is ready. Begin running and processing data."),
			LogConsumerCfg: &testcontainers.LogConsumerConfig{
				Consumers: []testcontainers.LogConsumer{lc},
			},
			Cmd: []string{"--config=/" + collectorCfgName},
		}

		var errCollector error
		collector, errCollector = testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
			ContainerRequest: req,
			Started:          true,
		})
		Expect(errCollector).ToNot(HaveOccurred())

		// Create the exporter

		ip, err := collector.Host(ctx)
		Expect(err).ToNot(HaveOccurred())

		port, err := collector.MappedPort(ctx, "4317")
		Expect(err).ToNot(HaveOccurred())

		endpoint := fmt.Sprintf("%s:%s", ip, port.Port())

		logger := logr.FromSlogHandler(slog.Default().Handler())

		errorHandler := telemetry.NewErrorHandler()

		exporter, err = telemetry.NewExporter(
			telemetry.ExporterConfig{
				SpanProvider: telemetry.CreateOTLPSpanProvider(
					otlptracegrpc.WithEndpoint(endpoint),
					otlptracegrpc.WithInsecure(),
				),
			},
			telemetry.WithGlobalOTelLogger(logger),
			telemetry.WithGlobalOTelErrorHandler(errorHandler),
		)
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		if collector != nil {
			Expect(collector.Terminate(ctx)).To(Succeed())
		}
		if exporter != nil {
			Expect(exporter.Shutdown(ctx)).To(Succeed())
		}
	})

	It("exports data successfully", func() {
		lc.setExpectedSubstrings([]string{
			"resourceCount: Int(1)",
		})

		data := &telemetryData{
			ResourceCount: 1,
		}

		Expect(exporter.Export(ctx, data)).To(Succeed())

		Eventually(lc.unmatchedCount, "10s").Should(BeZero())
	})
})
