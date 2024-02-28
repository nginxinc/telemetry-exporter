//go:build generator

package tests

import "github.com/nginxinc/telemetry-exporter/cmd/generator/tests/subtests"

// Data includes a field of each supported data type.
// We use this struct to test the generation of code and scheme.
// We also use it to test that the generated code compiles and runs as expected.
//
//go:generate go run -tags generator github.com/nginxinc/telemetry-exporter/cmd/generator -type=Data -build-tags=generator -scheme -scheme-protocol=NGFProductTelemetry -scheme-df-datatype=ngf-product-telemetry
//nolint:govet // Disable fieldalignment linter (part of govet), to control the order of fields for better readability.
type Data struct {
	// SomeString is a string field.
	SomeString string
	/* SomeInt is an int64 field. */
	SomeInt int64
	// SomeFloat is a float64 field.
	// More comments.
	SomeFloat float64
	// SomeBool is a bool field.
	SomeBool bool
	/*
		SomeStrings is a slice of strings.
	*/
	SomeStrings []string
	// SomeInts is a slice of int64.
	SomeInts []int64
	// SomeFloats is a slice of float64.
	SomeFloats []float64
	// SomeBools is a slice of bool.
	SomeBools []bool

	subtests.Data2
}
