//go:build generator

package main

import (
	"fmt"
	"go/types"
	"io"
	"strings"
	"text/template"
)

const schemeTemplate = `@namespace("{{ .Namespace }}") protocol {{ .Protocol }} {
	@df_datatype("{{ .DataFabricDataType }}") record {{ .Record }} {
		/** The field that identifies what type of data this is. */
		string dataType;
		/** The time the event occurred */
		long eventTime;
		/** The time our edge ingested the event */
		long ingestTime;

		{{ range .Fields }}
		/** {{ .Comment }} */
		{{ .Type }} {{ .Name }} = null;
		{{ end }}
	}
}
`

func getAvroPrimitiveType(kind types.BasicKind) string {
	switch kind {
	case types.Int64:
		return "long"
	case types.Float64:
		return "double"
	case types.String:
		return "string"
	case types.Bool:
		return "boolean"
	default:
		panic(fmt.Sprintf("unexpected kind %v", kind))
	}
}

func getAvroFieldName(name string) string {
	return strings.ToLower(name[:1]) + name[1:]
}

type schemeGen struct {
	Namespace          string
	Protocol           string
	DataFabricDataType string
	Record             string
	Fields             []schemeField
}

type schemeField struct {
	Comment string
	Type    string
	Name    string
}

type schemeGenConfig struct {
	namespace          string
	protocol           string
	dataFabricDataType string
	record             string
	fields             []field
}

func generateScheme(writer io.Writer, cfg schemeGenConfig) error {
	var schemeFields []schemeField

	var createSchemeFields func([]field)
	createSchemeFields = func(fields []field) {
		for _, f := range fields {
			if f.slice {
				schemeFields = append(schemeFields, schemeField{
					Comment: f.docString,
					Type:    fmt.Sprintf("union {null, array<%s>}", getAvroPrimitiveType(f.fieldType)),
					Name:    getAvroFieldName(f.name),
				})
			} else if f.embeddedStruct {
				createSchemeFields(f.embeddedStructFields)
			} else {
				schemeFields = append(schemeFields, schemeField{
					Comment: f.docString,
					Type:    getAvroPrimitiveType(f.fieldType) + "?",
					Name:    getAvroFieldName(f.name),
				})
			}
		}
	}

	createSchemeFields(cfg.fields)

	sg := schemeGen{
		Namespace:          cfg.namespace,
		Protocol:           cfg.protocol,
		DataFabricDataType: cfg.dataFabricDataType,
		Record:             cfg.record,
		Fields:             schemeFields,
	}

	funcMap := template.FuncMap{
		"getAttributeType": getAttributeType,
	}

	tmpl := template.Must(template.New("scheme").Funcs(funcMap).Parse(schemeTemplate))

	if err := tmpl.Execute(writer, sg); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	return nil
}
