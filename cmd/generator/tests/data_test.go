//go:build generator

package tests

import (
	"testing"

	. "github.com/onsi/gomega"
	"go.opentelemetry.io/otel/attribute"

	"github.com/nginxinc/telemetry-exporter/cmd/generator/tests/subtests"
)

func TestData_Attributes(t *testing.T) {
	g := NewGomegaWithT(t)

	data := Data{
		SomeString:  "some string",
		SomeInt:     42,
		SomeFloat:   3.14,
		SomeBool:    true,
		SomeStrings: []string{"a", "b", "c"},
		SomeInts:    []int64{1, 2, 3},
		SomeFloats:  []float64{1.1, 2.2, 3.3},
		SomeBools:   []bool{true, false, true},
		AnotherData: subtests.AnotherData{
			AnotherSomeString:  "another string",
			AnotherSomeInt:     24,
			AnotherSomeFloat:   1.41,
			AnotherSomeBool:    false,
			AnotherSomeStrings: []string{"d", "e", "f"},
			AnotherSomeInts:    []int64{4, 5, 6},
			AnotherSomeFloats:  []float64{4.4, 5.5, 6.6},
			AnotherSomeBools:   []bool{false, true, false},
		},
	}

	expectedAttributes := []attribute.KeyValue{
		attribute.String("dataType", "ngf-product-telemetry"),
		attribute.String("SomeString", "some string"),
		attribute.Int64("SomeInt", 42),
		attribute.Float64("SomeFloat", 3.14),
		attribute.Bool("SomeBool", true),
		attribute.StringSlice("SomeStrings", []string{"a", "b", "c"}),
		attribute.Int64Slice("SomeInts", []int64{1, 2, 3}),
		attribute.Float64Slice("SomeFloats", []float64{1.1, 2.2, 3.3}),
		attribute.BoolSlice("SomeBools", []bool{true, false, true}),
		attribute.String("AnotherSomeString", "another string"),
		attribute.Int64("AnotherSomeInt", 24),
		attribute.Float64("AnotherSomeFloat", 1.41),
		attribute.Bool("AnotherSomeBool", false),
		attribute.StringSlice("AnotherSomeStrings", []string{"d", "e", "f"}),
		attribute.Int64Slice("AnotherSomeInts", []int64{4, 5, 6}),
		attribute.Float64Slice("AnotherSomeFloats", []float64{4.4, 5.5, 6.6}),
		attribute.BoolSlice("AnotherSomeBools", []bool{false, true, false}),
	}

	attributes := data.Attributes()

	g.Expect(attributes).To(ConsistOf(expectedAttributes))
}

func TestData_AttributesEmpty(t *testing.T) {
	g := NewGomegaWithT(t)

	data := Data{}

	expectedAttributes := []attribute.KeyValue{
		attribute.String("dataType", "ngf-product-telemetry"),
		attribute.String("SomeString", ""),
		attribute.Int64("SomeInt", 0),
		attribute.Float64("SomeFloat", 0),
		attribute.Bool("SomeBool", false),
		attribute.StringSlice("SomeStrings", []string{}),
		attribute.Int64Slice("SomeInts", []int64{}),
		attribute.Float64Slice("SomeFloats", []float64{}),
		attribute.BoolSlice("SomeBools", []bool{}),
		attribute.String("AnotherSomeString", ""),
		attribute.Int64("AnotherSomeInt", 0),
		attribute.Float64("AnotherSomeFloat", 0),
		attribute.Bool("AnotherSomeBool", false),
		attribute.StringSlice("AnotherSomeStrings", []string{}),
		attribute.Int64Slice("AnotherSomeInts", []int64{}),
		attribute.Float64Slice("AnotherSomeFloats", []float64{}),
		attribute.BoolSlice("AnotherSomeBools", []bool{}),
	}

	attributes := data.Attributes()

	g.Expect(attributes).To(ConsistOf(expectedAttributes))
}
