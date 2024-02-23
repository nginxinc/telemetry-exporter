//go:build generator

package main

import (
	"go/types"
	"testing"

	. "github.com/onsi/gomega"

	"github.com/nginxinc/telemetry-exporter/cmd/generator/tests"
)

type DataUnexportedBasicTypeField struct {
	clusterID string //nolint:unused
}

type DataUnexportedSliceField struct {
	someStrings []string //nolint:unused
}

type DataUnexportedEmbeddedStructField struct {
	someStruct //nolint:unused
}

//nolint:unused
type someStruct struct{}

type SomeStruct struct{}

type DataNotEmbeddedStructField struct {
	SomeField SomeStruct //nolint:unused
}

type SomeInterface interface{}

type DataEmbeddedInterface struct {
	SomeInterface
}

type IntType int64

type UnsupportedEmbeddedType struct {
	IntType
}

type EmbeddedBasicType struct {
	int64 //nolint:unused
}

type UnsupportedBasicType struct {
	Counter int
}

type MissingBasicFieldDocString struct {
	Counter int64 // doc string above is missing
}

type MissingSliceFieldDocString struct {
	Counters []int64 // doc string above is missing
}

type EmptyFieldDocString struct {
	/*
	 */
	Counter int64 // empty doc string
}

type UnsupportedSliceType struct {
	Structs []SomeStruct
}

type UnsupportedBasicTypeSlice struct {
	Counters []int
}

type DuplicateFields struct {
	// Counter is a counter.
	Counter int64
	EmbeddedDuplicateFields
}

type EmbeddedDuplicateFields struct {
	// Counter is a counter.
	Counter int64
}

func TestParseErrors(t *testing.T) {
	tests := []struct {
		name           string
		expectedErrMsg string
		typeName       string
	}{
		{
			name:           "not a struct",
			expectedErrMsg: "expected struct, got interface{}",
			typeName:       "SomeInterface",
		},
		{
			name:           "unexported field",
			expectedErrMsg: "field clusterID: must be exported",
			typeName:       "DataUnexportedBasicTypeField",
		},
		{
			name:           "unexported slice field",
			expectedErrMsg: "field someStrings: must be exported",
			typeName:       "DataUnexportedSliceField",
		},
		{
			name:           "unexported embedded struct",
			expectedErrMsg: "field someStruct: must be exported",
			typeName:       "DataUnexportedEmbeddedStructField",
		},
		{
			name:           "not embedded struct",
			expectedErrMsg: "field SomeField: structs must be embedded",
			typeName:       "DataNotEmbeddedStructField",
		},
		{
			name:           "embedded interface",
			expectedErrMsg: "field SomeInterface: must be struct, got interface{}",
			typeName:       "DataEmbeddedInterface",
		},
		{
			name:           "unsupported embedded type",
			expectedErrMsg: "field IntType: must be struct, got int",
			typeName:       "UnsupportedEmbeddedType",
		},
		{
			name:           "embedded basic type",
			expectedErrMsg: "field int64: embedded basic types are not allowed",
			typeName:       "EmbeddedBasicType",
		},
		{
			name:           "unsupported basic type",
			expectedErrMsg: "field Counter: type of field must be one of bool, float64, int64, string, got int",
			typeName:       "UnsupportedBasicType",
		},
		{
			name:           "missing field doc string",
			expectedErrMsg: "field Counter: doc string not found",
			typeName:       "MissingBasicFieldDocString",
		},
		{
			name:           "missing slice field doc string",
			expectedErrMsg: "field Counters: doc string not found",
			typeName:       "MissingSliceFieldDocString",
		},
		{
			name:           "empty field doc string",
			expectedErrMsg: "field Counter: doc string not found",
			typeName:       "EmptyFieldDocString",
		},
		{
			name: "unsupported slice type",
			expectedErrMsg: "field Structs: type of field must be one of bool, float64, int64, string, " +
				"got []github.com/nginxinc/telemetry-exporter/cmd/generator.SomeStruct",
			typeName: "UnsupportedSliceType",
		},
		{
			name:           "unsupported basic type slice",
			expectedErrMsg: "field Counters: type of field must be one of bool, float64, int64, string, got []int",
			typeName:       "UnsupportedBasicTypeSlice",
		},
		{
			name: "duplicate fields",
			expectedErrMsg: "field Counter: already exists in " +
				"github.com/nginxinc/telemetry-exporter/cmd/generator.DuplicateFields",
			typeName: "DuplicateFields",
		},
		{
			name:           "type not found",
			expectedErrMsg: "type NotFoundType not found",
			typeName:       "NotFoundType",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			g := NewGomegaWithT(t)

			cfg := parsingConfig{
				pkgName:    "main",
				typeName:   test.typeName,
				loadTests:  true,
				buildFlags: []string{"-tags=generator"},
			}

			_, err := parse(cfg)

			g.Expect(err).To(MatchError(ContainSubstring(test.expectedErrMsg)))
		})
	}
}

func TestParseLoadingFailures(t *testing.T) {
	g := NewGomegaWithT(t)

	cfg := parsingConfig{
		pkgName:    "notfound",
		typeName:   "Data",
		loadTests:  true,
		buildFlags: []string{"-tags=generator"},
	}

	_, err := parse(cfg)

	g.Expect(err).To(MatchError(ContainSubstring("package notfound not found")))
}

func TestParseSuccess(t *testing.T) {
	g := NewGomegaWithT(t)

	cfg := parsingConfig{
		pkgName:     "tests",
		typeName:    "Data",
		loadPattern: "github.com/nginxinc/telemetry-exporter/cmd/generator/tests",
		buildFlags:  []string{"-tags=generator"},
	}

	_ = tests.Data{} // depends on the type being defined

	expectedEmbeddedStructFields := []field{
		{
			docString:            "AnotherSomeString is a string field.",
			name:                 "AnotherSomeString",
			fieldType:            types.String,
			slice:                false,
			embeddedStruct:       false,
			embeddedStructFields: nil,
		},
		{
			docString:            "AnotherSomeInt is an int64 field.",
			name:                 "AnotherSomeInt",
			fieldType:            types.Int64,
			slice:                false,
			embeddedStruct:       false,
			embeddedStructFields: nil,
		},
		{
			docString:            "AnotherSomeFloat is a float64 field.",
			name:                 "AnotherSomeFloat",
			fieldType:            types.Float64,
			slice:                false,
			embeddedStruct:       false,
			embeddedStructFields: nil,
		},
		{
			docString:            "AnotherSomeBool is a bool field.",
			name:                 "AnotherSomeBool",
			fieldType:            types.Bool,
			slice:                false,
			embeddedStruct:       false,
			embeddedStructFields: nil,
		},
		{
			docString:            "AnotherSomeStrings is a slice of strings.",
			name:                 "AnotherSomeStrings",
			fieldType:            types.String,
			slice:                true,
			embeddedStruct:       false,
			embeddedStructFields: nil,
		},
		{
			docString:            "AnotherSomeInts is a slice of int64.",
			name:                 "AnotherSomeInts",
			fieldType:            types.Int64,
			slice:                true,
			embeddedStruct:       false,
			embeddedStructFields: nil,
		},
		{
			docString:            "AnotherSomeFloats is a slice of float64.",
			name:                 "AnotherSomeFloats",
			fieldType:            types.Float64,
			slice:                true,
			embeddedStruct:       false,
			embeddedStructFields: nil,
		},
		{
			docString:            "AnotherSomeBools is a slice of bool.",
			name:                 "AnotherSomeBools",
			fieldType:            types.Bool,
			slice:                true,
			embeddedStruct:       false,
			embeddedStructFields: nil,
		},
	}

	expectedFields := []field{
		{
			docString:            "SomeString is a string field.",
			name:                 "SomeString",
			fieldType:            types.String,
			slice:                false,
			embeddedStruct:       false,
			embeddedStructFields: nil,
		},
		{
			docString:            "SomeInt is an int64 field.",
			name:                 "SomeInt",
			fieldType:            types.Int64,
			slice:                false,
			embeddedStruct:       false,
			embeddedStructFields: nil,
		},
		{
			docString:            "SomeFloat is a float64 field.\nMore comments.",
			name:                 "SomeFloat",
			fieldType:            types.Float64,
			slice:                false,
			embeddedStruct:       false,
			embeddedStructFields: nil,
		},
		{
			docString:            "SomeBool is a bool field.",
			name:                 "SomeBool",
			fieldType:            types.Bool,
			slice:                false,
			embeddedStruct:       false,
			embeddedStructFields: nil,
		},
		{
			docString:            "SomeStrings is a slice of strings.",
			name:                 "SomeStrings",
			fieldType:            types.String,
			slice:                true,
			embeddedStruct:       false,
			embeddedStructFields: nil,
		},
		{
			docString:            "SomeInts is a slice of int64.",
			name:                 "SomeInts",
			fieldType:            types.Int64,
			slice:                true,
			embeddedStruct:       false,
			embeddedStructFields: nil,
		},
		{
			docString:            "SomeFloats is a slice of float64.",
			name:                 "SomeFloats",
			fieldType:            types.Float64,
			slice:                true,
			embeddedStruct:       false,
			embeddedStructFields: nil,
		},
		{
			docString:            "SomeBools is a slice of bool.",
			name:                 "SomeBools",
			fieldType:            types.Bool,
			slice:                true,
			embeddedStruct:       false,
			embeddedStructFields: nil,
		},
		{
			docString:            "",
			name:                 "Data2",
			fieldType:            0,
			slice:                false,
			embeddedStruct:       true,
			embeddedStructFields: expectedEmbeddedStructFields,
		},
	}

	fields, err := parse(cfg)

	g.Expect(err).To(BeNil())
	g.Expect(fields).To(Equal(expectedFields))
}
