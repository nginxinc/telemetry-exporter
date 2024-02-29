//go:build generator

package subtests

// AnotherData is a struct that can be exported by a struct in another package to test cross-package referencing
// when generating code and scheme.
// AnotherData includes a field of each supported data type except an embedded struct.
//
//go:generate go run -tags generator github.com/nginxinc/telemetry-exporter/cmd/generator -type=AnotherData -build-tags=generator
//nolint:govet // Disable fieldalignment linter (part of govet), to control the order of fields for better readability.
type AnotherData struct {
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
