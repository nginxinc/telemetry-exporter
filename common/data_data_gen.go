/* DO NOT EDIT */
package common 

import (
	"go.opentelemetry.io/otel/attribute"
	"github.com/nginxinc/telemetry-exporter/exporter"
)

func (d *Data) Attributes() []attribute.KeyValue {
	var attrs []attribute.KeyValue

	
	attrs = append(
		attrs,
		
		attribute.KeyValue{
			Key:   attribute.Key("ProductID"),
			Value: attribute.StringValue(d.ProductID),
		},
		
	)
		

	return attrs 
}

var _ exporter.Exportable = (*Data)(nil)
