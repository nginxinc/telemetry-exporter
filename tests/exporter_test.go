package tests

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"

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
	sync               sync.Mutex
	expectedSubstrings map[string]struct{}
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
		ctx, _ = context.WithTimeout(context.Background(), 30*time.Second)

		//  Run the collector container

		relayCfgPath := "./relay.yaml"

		lc = &matchingLogConsumer{}

		req := testcontainers.ContainerRequest{
			Image: "otel/opentelemetry-collector-contrib:0.88.0",
			Files: []testcontainers.ContainerFile{
				{
					HostFilePath:      relayCfgPath,
					ContainerFilePath: "/relay.yaml",
					FileMode:          0o777,
				},
			},
			ExposedPorts: []string{"4317/tcp"},
			WaitingFor:   wait.ForLog("Everything is ready. Begin running and processing data."),
			LogConsumerCfg: &testcontainers.LogConsumerConfig{
				Consumers: []testcontainers.LogConsumer{lc},
			},
			Cmd: []string{"--config=/relay.yaml"},
		}

		collector, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
			ContainerRequest: req,
			Started:          true,
		})
		Expect(err).ToNot(HaveOccurred())

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
			err := collector.Terminate(ctx)
			Expect(err).ToNot(HaveOccurred())
		}
		if exporter != nil {
			err := exporter.Shutdown(ctx)
			Expect(err).ToNot(HaveOccurred())
		}
	})

	It("exports data successfully", func() {
		lc.setExpectedSubstrings([]string{
			"resourceCount: Int(1)",
		})

		data := &telemetryData{
			ResourceCount: 1,
		}

		err := exporter.Export(context.Background(), data)
		Expect(err).ToNot(HaveOccurred())

		Eventually(lc.unmatchedCount, "5s").Should(BeZero())
	})
})
