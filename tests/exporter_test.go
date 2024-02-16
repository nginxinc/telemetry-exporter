package tests

import (
	"go.opentelemetry.io/otel/attribute"
)

type telemetryData struct {
	ResourceCount int
}

func (d *telemetryData) Attributes() []attribute.KeyValue {
	return []attribute.KeyValue{
		attribute.Int("resourceCount", d.ResourceCount),
	}
}
