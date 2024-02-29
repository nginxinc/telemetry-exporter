//go:build generator
package tests
/*
This is a generated file. DO NOT EDIT.
*/

import (
	"go.opentelemetry.io/otel/attribute"

	
	"github.com/nginxinc/telemetry-exporter/pkg/telemetry"
	
)

func (d *Data) Attributes() []attribute.KeyValue {
	var attrs []attribute.KeyValue

	attrs = append(attrs, attribute.String("SomeString", d.SomeString))
	attrs = append(attrs, attribute.Int64("SomeInt", d.SomeInt))
	attrs = append(attrs, attribute.Float64("SomeFloat", d.SomeFloat))
	attrs = append(attrs, attribute.Bool("SomeBool", d.SomeBool))
	attrs = append(attrs, attribute.StringSlice("SomeStrings", d.SomeStrings))
	attrs = append(attrs, attribute.Int64Slice("SomeInts", d.SomeInts))
	attrs = append(attrs, attribute.Float64Slice("SomeFloats", d.SomeFloats))
	attrs = append(attrs, attribute.BoolSlice("SomeBools", d.SomeBools))
	attrs = append(attrs, d.AnotherData.Attributes()...)
	

	return attrs
}

var _ telemetry.Exportable = (*Data)(nil)
