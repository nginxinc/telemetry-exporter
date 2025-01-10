//go:build generator

package telemetry

// MoreData is used to ensure that the generator produces the correct code for a struct in a package with the name
// 'telemetry'.
// Correctness is confirmed by the fact the generated code compiles.
//
//go:generate go run -tags generator github.com/nginx/telemetry-exporter/cmd/generator -type=MoreData -build-tags=generator
type MoreData struct {
	// StringField is a string field.
	StringField string
}
