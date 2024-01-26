package exporter

import (
	"github.com/go-logr/logr"
	"go.opentelemetry.io/otel"
)

func InitOTel(logger logr.Logger, errorHandler otel.ErrorHandler) {
	otel.SetLogger(logger)
	otel.SetErrorHandler(errorHandler)
}
