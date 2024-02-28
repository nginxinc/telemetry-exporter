//go:build generator
package subtests
/*
This is a generated file. DO NOT EDIT.
*/

import (
	"go.opentelemetry.io/otel/attribute"
	"github.com/nginxinc/telemetry-exporter/pkg/telemetry"
)

func (d *AnotherData) Attributes() []attribute.KeyValue {
	var attrs []attribute.KeyValue

	attrs = append(attrs, attribute.String("AnotherSomeString", d.AnotherSomeString))
	attrs = append(attrs, attribute.Int64("AnotherSomeInt", d.AnotherSomeInt))
	attrs = append(attrs, attribute.Float64("AnotherSomeFloat", d.AnotherSomeFloat))
	attrs = append(attrs, attribute.Bool("AnotherSomeBool", d.AnotherSomeBool))
	attrs = append(attrs, attribute.StringSlice("AnotherSomeStrings", d.AnotherSomeStrings))
	attrs = append(attrs, attribute.Int64Slice("AnotherSomeInts", d.AnotherSomeInts))
	attrs = append(attrs, attribute.Float64Slice("AnotherSomeFloats", d.AnotherSomeFloats))
	attrs = append(attrs, attribute.BoolSlice("AnotherSomeBools", d.AnotherSomeBools))
	

	return attrs
}

var _ telemetry.Exportable = (*AnotherData)(nil)
