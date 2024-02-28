//go:build generator

package main

import (
	"fmt"
	"go/types"
	"io"
	"reflect"
	"strings"
	"text/template"

	"github.com/nginxinc/telemetry-exporter/pkg/telemetry"
)

var telemetryPackagePath = reflect.TypeOf((*telemetry.Exportable)(nil)).Elem().PkgPath()

const codeTemplate = `{{ if .BuildTags }}//go:build {{ .BuildTags }}{{ end }}
package {{ .PackageName }}
/*
This is a generated file. DO NOT EDIT.
*/

import (
	"go.opentelemetry.io/otel/attribute"

	{{ if .TelemetryPackagePath }}"{{ .TelemetryPackagePath }}"{{ end }}
)

func (d *{{ .StructName }}) Attributes() []attribute.KeyValue {
	var attrs []attribute.KeyValue

	{{ range .Fields -}}
	attrs = append(attrs, {{ .AttributesSource }})
	{{ end }}

	return attrs
}

var _ {{ .ExportablePackagePrefix }}Exportable = (*{{ .StructName }})(nil)
`

type codeGen struct {
	PackageName             string
	TelemetryPackagePath    string
	ExportablePackagePrefix string
	StructName              string
	BuildTags               string
	Fields                  []codeField
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
	packagePath string
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

	var telemetryPkg string
	var exportablePkgPrefix string

	// check if we generate code for the type in the telemetry package or any other package
	if cfg.packagePath != telemetryPackagePath {
		telemetryPkg = telemetryPackagePath
		exportablePkgPrefix = getPackageName(telemetryPackagePath) + "."
	}

	cg := codeGen{
		PackageName:             getPackageName(cfg.packagePath),
		ExportablePackagePrefix: exportablePkgPrefix,
		TelemetryPackagePath:    telemetryPkg,
		StructName:              cfg.typeName,
		Fields:                  codeFields,
		BuildTags:               cfg.buildTags,
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

func getPackageName(packagePath string) string {
	packageParts := strings.Split(packagePath, "/")
	return packageParts[len(packageParts)-1]
}
