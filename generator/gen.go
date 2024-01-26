package main

import (
	"fmt"
	"go/types"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"golang.org/x/tools/go/packages"
)

// Based on https://dev.to/hlubek/metaprogramming-with-go-or-how-to-build-code-generators-that-parse-go-code-2k3j
// and https://github.com/hlubek/metaprogramming-go
// Under the LICENSE
//Copyright 2020 Christopher Hlubek (networkteam GmbH)
//
//Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the "Software"), to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject to the following conditions:
//
//The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.
//
//THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

func main() {
	// Handle arguments to command
	if len(os.Args) != 2 {
		panic(fmt.Errorf("expected exactly one argument: <source type>"))
	}

	sourceType := os.Args[1]
	sourceTypePackage, sourceTypeName := splitSourceType(sourceType)

	cfg := &packages.Config{Mode: packages.NeedTypes | packages.NeedImports}
	pkgs, err := packages.Load(cfg, sourceTypePackage)
	if err != nil {
		panic(fmt.Errorf("loading packages for inspection: %v", err))
	}

	pkg := pkgs[0]

	// Lookup the given source type name in the package declarations
	obj := pkg.Types.Scope().Lookup(sourceTypeName)
	if obj == nil {
		panic(fmt.Errorf("%s not found in declared types of %s",
			sourceTypeName, pkg))
	}

	// We check if it is a declared type
	if _, ok := obj.(*types.TypeName); !ok {
		panic(fmt.Errorf("%v is not a named type", obj))
	}
	// We expect the underlying type to be a struct
	structType, ok := obj.Type().Underlying().(*types.Struct)
	if !ok {
		panic(fmt.Errorf("type %v is not a struct", obj))
	}

	err = generateCode(sourceTypeName, structType)
	if err != nil {
		panic(fmt.Errorf("generating code: %v", err))
	}

	err = generateScheme(sourceTypeName, structType)
	if err != nil {
		panic(fmt.Errorf("generating scheme: %v", err))
	}
}

func splitSourceType(sourceType string) (string, string) {
	idx := strings.LastIndexByte(sourceType, '.')
	if idx == -1 {
		panic(fmt.Errorf(`expected qualified type as "pkg/path.MyType"`))
	}
	sourceTypePackage := sourceType[0:idx]
	sourceTypeName := sourceType[idx+1:]
	return sourceTypePackage, sourceTypeName
}

func generateCode(sourceTypeName string, structType *types.Struct) error {
	// Get the package of the file with go:generate comment
	goPackage := os.Getenv("GOPACKAGE")

	// Build the target file name
	goFile := os.Getenv("GOFILE")
	ext := filepath.Ext(goFile)
	baseFilename := goFile[0 : len(goFile)-len(ext)]
	targetFilename := baseFilename + "_" + strings.ToLower(sourceTypeName) + "_gen.go"

	tmpl := template.Must(template.New("impl").Parse(implTemplate))

	var fields []Field

	for i := 0; i < structType.NumFields(); i++ {
		field := structType.Field(i)

		// check if field is struct
		switch v := field.Type().(type) {
		case *types.Named:
			fields = append(fields, Field{
				Name:   field.Name(),
				Struct: true,
			})
		case *types.Basic:
			// check if field is string
			if v.Kind() == types.String {
				fields = append(fields, Field{
					Name:   field.Name(),
					String: true,
				})
			} else if v.Kind() == types.Int {
				fields = append(fields, Field{
					Name: field.Name(),
					Int:  true,
				})
			} else {
				return fmt.Errorf("field %s is not expected, got %s", field.Name(), field.Type().String())
			}
		default:
			return fmt.Errorf("field %s is not expected, got %s", field.Name(), field.Type().String())
		}
	}

	code := Code{
		Package:    goPackage,
		StructName: sourceTypeName,
		Fields:     fields,
	}

	file, err := os.Create(targetFilename)
	if err != nil {
		return err
	}
	defer file.Close()

	err = tmpl.Execute(file, code)
	if err != nil {
		return err
	}

	return nil
}

type Code struct {
	Package    string
	Imports    []string
	StructName string
	Fields     []Field
}

type Field struct {
	Name   string
	Int    bool
	String bool
	Struct bool
}

const implTemplate = `/* DO NOT EDIT */
package {{ .Package }} 

import (
	"go.opentelemetry.io/otel/attribute"
	"github.com/nginxinc/telemetry-exporter/exporter"
)

func (d *{{ .StructName }}) Attributes() []attribute.KeyValue {
	var attrs []attribute.KeyValue

	{{ range .Fields }}
	attrs = append(
		attrs,
		{{ if .String }}
		attribute.KeyValue{
			Key:   attribute.Key("{{ .Name }}"),
			Value: attribute.StringValue(d.{{ .Name }}),
		},
		{{ else if .Int }}
		attribute.KeyValue{
			Key:   "{{ .Name }}",
			Value: attribute.IntValue(d.{{ .Name }}),
		},
		{{ else if .Struct }}
        d.{{ .Name }}.Attributes()...,
		{{ end }}
	)
	{{ end }}	

	return attrs 
}

var _ exporter.Exportable = (*{{ .StructName }})(nil)
`

const schemeTemplate = `@namespace("nginx.org") protocol ProductData {
	@df_datatype("ngf") record {{ .Record }} {
		/** The field that identifies what type of data this is. */
		string dataType;
		/** The time the event occurred */
		long eventTime;
		/** The time our edge ingested the event */
		long ingestTime;

	
		{{ range .Fields }}
		/** {{ .Comment }} */
		{{ .Type }}? {{ .Name }} = null;
		{{ end }}
	}
}
`

type Scheme struct {
	Record string
	Fields []SchemeField
}

type SchemeField struct {
	Comment string
	Name    string
	Type    string
}

func generateScheme(sourceTypeName string, structType *types.Struct) error {
	// Build the target file name
	goFile := os.Getenv("GOFILE")
	ext := filepath.Ext(goFile)
	baseFilename := goFile[0 : len(goFile)-len(ext)]
	targetFilename := baseFilename + "_" + strings.ToLower(sourceTypeName) + "_gen.txt"

	tmpl := template.Must(template.New("scheme").Parse(schemeTemplate))

	var fields []SchemeField

	var build func(*types.Struct)
	build = func(structType *types.Struct) {
		for i := 0; i < structType.NumFields(); i++ {
			field := structType.Field(i)

			// check if field is struct
			switch v := field.Type().(type) {
			case *types.Named:
				nextStructType, ok := v.Underlying().(*types.Struct)
				if !ok {
					panic(fmt.Errorf("type %v is not a struct", v))
				}
				build(nextStructType)
			case *types.Basic:
				// check if field is string
				if v.Kind() == types.String {
					fields = append(fields, SchemeField{
						Name:    field.Name(),
						Type:    "string",
						Comment: "The " + field.Name() + " of the product.", // TO-DO: Extract comment from the doc string of the field
					})
				} else if v.Kind() == types.Int {
					fields = append(fields, SchemeField{
						Name:    field.Name(),
						Type:    "int",
						Comment: "The " + field.Name() + " of the product.",
					})
				} else {
					panic(fmt.Errorf("field %s is not expected, got %s", field.Name(), field.Type().String()))
				}
			default:
				panic(fmt.Errorf("field %s is not expected, got %s", field.Name(), field.Type().String()))
			}
		}
	}

	build(structType)

	scheme := Scheme{
		Record: sourceTypeName,
		Fields: fields,
	}

	file, err := os.Create(targetFilename)
	if err != nil {
		return err
	}
	defer file.Close()

	err = tmpl.Execute(file, scheme)
	if err != nil {
		return err
	}

	return nil
}

func genFieldName(s string) string {
	return strings.ToLower(s[0:1]) + s[1:]
}
