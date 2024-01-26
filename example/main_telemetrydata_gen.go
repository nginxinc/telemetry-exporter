/* DO NOT EDIT */
package main

import (
	"go.opentelemetry.io/otel/attribute"

	"github.com/nginxinc/telemetry-exporter/exporter"
)

func (d *TelemetryData) Attributes() []attribute.KeyValue {
	var attrs []attribute.KeyValue

	attrs = append(
		attrs,

		d.Data.Attributes()...,
	)

	attrs = append(
		attrs,

		attribute.KeyValue{
			Key:   "ResourceCount",
			Value: attribute.IntValue(d.ResourceCount),
		},
	)

	return attrs
}

var _ exporter.Exportable = (*TelemetryData)(nil)
