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

const codeTemplate = `{{- if .BuildTags }}//go:build {{ .BuildTags }}
{{ end -}}
package {{ .PackageName }}

/*
This is a generated file. DO NOT EDIT.
*/

import (
	"go.opentelemetry.io/otel/attribute"
	{{- if .TelemetryPackagePath }}

	{{ if .TelemetryPackageAlias }}{{ .TelemetryPackageAlias }} {{ end }}"{{ .TelemetryPackagePath }}"
	{{- end }}
)

func (d *{{ .StructName }}) Attributes() []attribute.KeyValue {
	var attrs []attribute.KeyValue

	{{- if .SchemeDataType }}
	attrs = append(attrs, attribute.String("dataType", "{{ .SchemeDataType }}"))
	{{- end }}

	{{- range .Fields }}
	{{- if .AttributesSource }}
	attrs = append(attrs, {{ .AttributesSource }})
	{{- end }}
	{{- end }}

	return attrs
}

var _ {{ .ExportablePackagePrefix }}Exportable = (*{{ .StructName }})(nil)
`

type codeGen struct {
	PackageName             string
	TelemetryPackagePath    string
	TelemetryPackageAlias   string
	ExportablePackagePrefix string
	StructName              string
	SchemeDataType          string
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
	packagePath    string
	typeName       string
	schemeDataType string
	buildTags      string
	fields         []field
}

func generateCode(writer io.Writer, cfg codeGenConfig) error {
	codeFields := make([]codeField, 0, len(cfg.fields))

	for _, f := range cfg.fields {
		var cf codeField

		switch {
		case f.embeddedStruct:
			cf = codeField{
				AttributesSource: fmt.Sprintf(`d.%s.Attributes()...`, f.name),
			}
		case f.slice:
			cf = codeField{
				AttributesSource: fmt.Sprintf(`attribute.%sSlice("%s", d.%s)`, getAttributeType(f.fieldType), f.name, f.name),
			}
		default:
			cf = codeField{
				AttributesSource: fmt.Sprintf(`attribute.%s("%s", d.%s)`, getAttributeType(f.fieldType), f.name, f.name),
			}
		}

		codeFields = append(codeFields, cf)
	}

	const alias = "ngxTelemetry"

	var (
		telemetryPkg        string
		exportablePkgPrefix string
		telemetryPkgAlias   string
	)

	// check if we generate code for the type in the telemetry package or any other package
	if cfg.packagePath != telemetryPackagePath {
		telemetryPkg = telemetryPackagePath

		// if the name of the package is the same as the telemetry package, we need to use an alias
		if getPackageName(cfg.packagePath) == getPackageName(telemetryPackagePath) {
			exportablePkgPrefix = alias + "."
			telemetryPkgAlias = alias
		} else {
			exportablePkgPrefix = getPackageName(telemetryPackagePath) + "."
		}
	}

	cg := codeGen{
		PackageName:             getPackageName(cfg.packagePath),
		ExportablePackagePrefix: exportablePkgPrefix,
		TelemetryPackageAlias:   telemetryPkgAlias,
		TelemetryPackagePath:    telemetryPkg,
		StructName:              cfg.typeName,
		SchemeDataType:          cfg.schemeDataType,
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
