
package telemetry
/*
This is a generated file. DO NOT EDIT.
*/

import (
	"go.opentelemetry.io/otel/attribute"

	
)

func (d *Data) Attributes() []attribute.KeyValue {
	var attrs []attribute.KeyValue

	attrs = append(attrs, attribute.Int64("Nodes", d.Nodes))
	

	return attrs
}

var _ Exportable = (*Data)(nil)
