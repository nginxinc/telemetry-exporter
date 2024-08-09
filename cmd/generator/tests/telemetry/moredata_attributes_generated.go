//go:build generator
package telemetry

/*
This is a generated file. DO NOT EDIT.
*/

import (
	"go.opentelemetry.io/otel/attribute"

	ngxTelemetry "github.com/nginxinc/telemetry-exporter/pkg/telemetry"
)

func (d *MoreData) Attributes() []attribute.KeyValue {
	var attrs []attribute.KeyValue
	attrs = append(attrs, attribute.String("StringField", d.StringField))

	return attrs
}

var _ ngxTelemetry.Exportable = (*MoreData)(nil)
