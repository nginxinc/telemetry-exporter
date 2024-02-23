//go:build generator

package subtests

// Data2 is a struct that can be exported by a struct in another package to test cross-package referencing
// when generating code and scheme.
// Data2 includes a field of each supported data type except an embedded struct.
//
//go:generate go run -tags generator github.com/nginxinc/telemetry-exporter/cmd/generator -type=Data2 -build-tags=generator
//nolint:govet
type Data2 struct {
	// AnotherSomeString is a string field.
	AnotherSomeString string
	// AnotherSomeInt is an int64 field.
	AnotherSomeInt int64
	// AnotherSomeFloat is a float64 field.
	AnotherSomeFloat float64
	// AnotherSomeBool is a bool field.
	AnotherSomeBool bool
	// AnotherSomeStrings is a slice of strings.
	AnotherSomeStrings []string
	// AnotherSomeInts is a slice of int64.
	AnotherSomeInts []int64
	// AnotherSomeFloats is a slice of float64.
	AnotherSomeFloats []float64
	// AnotherSomeBools is a slice of bool.
	AnotherSomeBools []bool
}
