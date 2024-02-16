package tests

import (
	"archive/tar"
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"

	"github.com/nginxinc/telemetry-exporter/pkg/telemetry"
)

func getDockerUnixSocket() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	// TO-DO: Ensure it works for Linux too
	return filepath.Join(homeDir, ".docker/run/docker.sock"), nil
}

func tarFileIntoWriter(filePath string, writer io.Writer) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	fileInfo, err := os.Lstat(filePath)
	if err != nil {
		return err
	}

	tarWriter := tar.NewWriter(writer)

	header, err := tar.FileInfoHeader(fileInfo, filePath)
	if err != nil {
		return err
	}
	err = tarWriter.WriteHeader(header)
	if err != nil {
		return err
	}

	_, err = io.Copy(tarWriter, file)
	if err != nil {
		return err
	}

	tarWriter.Close()
	return nil
}

var _ = Describe("Exporter 2", func() {
	var (
		ctx                  context.Context
		dockerClient         *client.Client
		collectorContainerID string
		exporter             *telemetry.Exporter
	)

	readLogs := func() []byte {
		out, err := dockerClient.ContainerLogs(ctx, collectorContainerID, container.LogsOptions{
			ShowStdout: true,
			ShowStderr: true,
		})
		Expect(err).ToNot(HaveOccurred())

		defer out.Close()

		body, err := io.ReadAll(out)
		Expect(err).ToNot(HaveOccurred())

		return body
	}

	logContainsSubstring := func(subs string) bool {
		return bytes.Contains(readLogs(), []byte(subs))
	}

	BeforeEach(func() {
		ctx, _ = context.WithTimeout(context.Background(), 30*time.Second)

		dockerSocket, err := getDockerUnixSocket()
		if err != nil {
			panic(err)
		}

		// Create Docker client
		dockerClient, err = client.NewClientWithOpts(
			client.FromEnv,
			client.WithAPIVersionNegotiation(),
			client.WithHost(fmt.Sprintf("unix://%s", dockerSocket)),
		)
		Expect(err).ToNot(HaveOccurred())
		DeferCleanup(dockerClient.Close)

		// Pull the collector Docker image
		reader, err := dockerClient.ImagePull(ctx, "otel/opentelemetry-collector-contrib:0.88.0", types.ImagePullOptions{})
		Expect(err).ToNot(HaveOccurred())
		_, err = io.Copy(io.Discard, reader)
		Expect(err).ToNot(HaveOccurred())

		Expect(reader.Close()).To(Succeed())
		// Run the collector with the relay configuration
		resp, err := dockerClient.ContainerCreate(ctx, &container.Config{
			Image: "otel/opentelemetry-collector-contrib:0.88.0",
			Tty:   false,
			Cmd:   []string{"--config=/relay.yaml"},
		},
			&container.HostConfig{
				PublishAllPorts: true,
			}, nil, nil, "")

		Expect(err).ToNot(HaveOccurred())

		collectorContainerID = resp.ID

		var buffer bytes.Buffer

		err = tarFileIntoWriter("relay.yaml", &buffer)
		Expect(err).ToNot(HaveOccurred())

		err = dockerClient.CopyToContainer(ctx, resp.ID, "/", &buffer, types.CopyToContainerOptions{})
		Expect(err).ToNot(HaveOccurred())

		err = dockerClient.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{})
		Expect(err).ToNot(HaveOccurred())

		// Wait for container to become ready

		checkLog := func() bool {
			return logContainsSubstring("Everything is ready. Begin running and processing data.")
		}

		Eventually(checkLog).WithTimeout(5 * time.Second).WithPolling(1 * time.Second).Should(BeTrue())

		// Get mapped port
		c, err := dockerClient.ContainerInspect(ctx, resp.ID)
		Expect(err).ToNot(HaveOccurred())

		ports := c.NetworkSettings.Ports["4317/tcp"]
		Expect(ports).To(HaveLen(1))

		// Create exporter
		endpoint := fmt.Sprintf("%s:%s", "127.0.0.1", ports[0].HostPort)

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
		statusCh, errCh := dockerClient.ContainerWait(ctx, collectorContainerID, container.WaitConditionNotRunning)

		err := dockerClient.ContainerStop(ctx, collectorContainerID, container.StopOptions{})
		Expect(err).ToNot(HaveOccurred())

		var statusCode int64

		select {
		case err := <-errCh:
			Expect(err).ToNot(HaveOccurred())
		case status := <-statusCh:
			statusCode = status.StatusCode
		}

		err = dockerClient.ContainerRemove(ctx, collectorContainerID, container.RemoveOptions{})
		Expect(err).ToNot(HaveOccurred())

		if statusCode != 0 {
			Fail(fmt.Sprintf("container exited with status %d, logs: %v", statusCode, readLogs()))
		}

		if exporter != nil {
			err := exporter.Shutdown(ctx)
			Expect(err).ToNot(HaveOccurred())
		}
	})

	It("exports data successfully", func() {
		data := &telemetryData{
			ResourceCount: 1,
		}

		Expect(exporter.Export(context.Background(), data)).To(Succeed())

		checkLog := func() bool {
			return logContainsSubstring("resourceCount: Int(1)")
		}

		Eventually(checkLog).WithTimeout(5 * time.Second).WithPolling(1 * time.Second).Should(BeTrue())
	})
})
