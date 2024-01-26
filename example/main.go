package main

import (
	"context"

	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"github.com/nginxinc/telemetry-exporter/common"
	"github.com/nginxinc/telemetry-exporter/exporter"
)

//go:generate go run  github.com/nginxinc/telemetry-exporter/generator github.com/nginxinc/telemetry-exporter/example.TelemetryData
type TelemetryData struct {
	common.Data
	ResourceCount int
}

func main() {
	logger := zap.New()

	logger.Info("starting the exporter")

	errorHandler := exporter.NewErrorHandler()

	exporter.InitOTel(logger, errorHandler)

	exp, err := exporter.NewExporter(exporter.ExporterConfig{
		ProductName:          "NGF",
		ProductVersion:       "1.0",
		ErrorHandler:         errorHandler,
		SpanExporterProvider: exporter.CreateOTLPSpanExporterProvider("localhost:4317", false /*not secure*/),
	})
	if err != nil {
		panic(err)
	}

	data := TelemetryData{
		Data: common.Data{
			ProductID: "NGF",
		},
		ResourceCount: 1,
	}

	logger.Info("exporting the data", "data", data)

	err = exp.Export(context.Background(), &data)
	if err != nil {
		panic(err)
	}
}
