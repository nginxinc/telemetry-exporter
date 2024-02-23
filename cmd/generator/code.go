//go:build generator

package main

import (
	"fmt"
	"go/types"
	"io"
	"text/template"
)

const codeTemplate = `{{ if .BuildTags }}//go:build {{ .BuildTags }}{{ end }}
package {{ .PackageName }}
/*
This is a generated file. DO NOT EDIT.
*/

import (
	"go.opentelemetry.io/otel/attribute"
	"github.com/nginxinc/telemetry-exporter/pkg/telemetry"
)

func (d *{{ .StructName }}) Attributes() []attribute.KeyValue {
	var attrs []attribute.KeyValue

	{{ range .Fields -}}
	attrs = append(attrs, {{ .AttributesSource }})
	{{ end }}

	return attrs
}

var _ telemetry.Exportable = (*{{ .StructName }})(nil)
`

type codeGen struct {
	PackageName string
	StructName  string
	BuildTags   string
	Fields      []codeField
}

type codeField struct {
	AttributesSource string
}

func getAttributeType(kind types.BasicKind) string {
	switch kind {
	case types.Int64:
		return "Int64"
	case types.Float64:
		return "Float64"
	case types.String:
		return "String"
	case types.Bool:
		return "Bool"
	default:
		panic(fmt.Sprintf("unexpected kind %v", kind))
	}
}

type codeGenConfig struct {
	packageName string
	typeName    string
	buildTags   string
	fields      []field
}

func generateCode(writer io.Writer, cfg codeGenConfig) error {
	codeFields := make([]codeField, 0, len(cfg.fields))

	for _, f := range cfg.fields {
		var cf codeField

		if f.embeddedStruct {
			cf = codeField{
				AttributesSource: fmt.Sprintf(`d.%s.Attributes()...`, f.name),
			}
		} else if f.slice {
			cf = codeField{
				AttributesSource: fmt.Sprintf(`attribute.%sSlice("%s", d.%s)`, getAttributeType(f.fieldType), f.name, f.name),
			}
		} else {
			cf = codeField{
				AttributesSource: fmt.Sprintf(`attribute.%s("%s", d.%s)`, getAttributeType(f.fieldType), f.name, f.name),
			}
		}

		codeFields = append(codeFields, cf)
	}

	cg := codeGen{
		PackageName: cfg.packageName,
		StructName:  cfg.typeName,
		Fields:      codeFields,
		BuildTags:   cfg.buildTags,
	}

	funcMap := template.FuncMap{
		"getAttributeType": getAttributeType,
	}

	tmpl := template.Must(template.New("scheme").Funcs(funcMap).Parse(codeTemplate))

	if err := tmpl.Execute(writer, cg); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	return nil
}
