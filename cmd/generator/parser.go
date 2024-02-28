//go:build generator

package main

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"sort"
	"strings"

	"golang.org/x/tools/go/packages"
)

// docStringFieldsProvider parses the code and finds doc string comments for fields of the structs.
type docStringFieldsProvider struct {
	packages   map[string]struct{}
	docStrings map[string]string
	buildFlags []string
	loadTests  bool
}

// newDocStringFieldsProvider creates a new docStringFieldsProvider.
// loadTests specifies whether the parser will load test files (e.g. *_test.go).
// buildFlags are go build flags (e.g. -tags=foo).
func newDocStringFieldsProvider(loadTests bool, buildFlags []string) *docStringFieldsProvider {
	return &docStringFieldsProvider{
		loadTests:  loadTests,
		buildFlags: buildFlags,
		packages:   make(map[string]struct{}),
		docStrings: make(map[string]string),
	}
}

func parseFullTypeName(fullTypeName string) (pkgName, typeName string) {
	idx := strings.LastIndex(fullTypeName, ".")
	if idx == -1 {
		panic(fmt.Sprintf("invalid full type name: %s", fullTypeName))
	}

	return fullTypeName[:idx], fullTypeName[idx+1:]
}

func getDocStringKey(pkgName, typeName, fieldName string) string {
	return fmt.Sprintf("%s.%s.%s", pkgName, typeName, fieldName)
}

// getDocString returns the doc string comment for the field of the struct.
// fullTypeName is the full type name of the struct
// (e.g. "github.com/nginxinc/nginx-gateway-fabric/pkg/mypackage.MyStruct").
func (p *docStringFieldsProvider) getDocString(fullTypeName, fieldName string) (string, error) {
	pkgName, typeName := parseFullTypeName(fullTypeName)

	_, exists := p.packages[pkgName]
	if !exists {
		if err := p.parseDocStringsFromPackage(pkgName); err != nil {
			return "", fmt.Errorf("failed to load struct comments from package %s: %w", pkgName, err)
		}
	}

	doc, exists := p.docStrings[getDocStringKey(pkgName, typeName, fieldName)]
	if !exists {
		return "", fmt.Errorf("doc string not found")
	}

	trimmedComment := strings.TrimSpace(doc)
	if trimmedComment == "" {
		return "", fmt.Errorf("trimmed doc string is empty")
	}

	return trimmedComment, nil
}

func (p *docStringFieldsProvider) parseDocStringsFromPackage(pkgName string) error {
	mode := packages.NeedName | packages.NeedSyntax | packages.NeedTypes

	cfg := packages.Config{
		Mode:       mode,
		Fset:       token.NewFileSet(),
		Tests:      p.loadTests,
		BuildFlags: p.buildFlags,
	}

	pkgs, err := packages.Load(&cfg, pkgName)
	if err != nil {
		return fmt.Errorf("failed to load package: %w", err)
	}

	var loadedPkg *packages.Package

	for _, pkg := range pkgs {
		if p.loadTests && !strings.HasSuffix(pkg.ID, ".test]") {
			continue
		}

		if pkgName == pkg.PkgPath {
			loadedPkg = pkg
			break
		}
	}

	if loadedPkg == nil {
		return fmt.Errorf("package %s not found", pkgName)
	}

	p.packages[pkgName] = struct{}{}

	// for each struct in the package,
	// save the doc string comments for the fields of the struct
	for _, fileAst := range loadedPkg.Syntax {
		ast.Inspect(fileAst, func(n ast.Node) bool {
			structTypeSpec, ok := n.(*ast.TypeSpec)
			if !ok {
				return true
			}

			structType, ok := structTypeSpec.Type.(*ast.StructType)
			if !ok {
				return true
			}

			for _, f := range structType.Fields.List {
				for _, name := range f.Names {
					comment := f.Doc.Text()
					if comment == "" {
						continue
					}

					p.docStrings[getDocStringKey(loadedPkg.PkgPath, structTypeSpec.Name.String(), name.Name)] = comment
				}
			}
			return true
		})
	}

	return nil
}

type parsingError struct {
	typeName  string
	fieldName string
	msg       string
}

func (e parsingError) Error() string {
	return fmt.Sprintf("type %s: field %s: %s", e.typeName, e.fieldName, e.msg)
}

// parsingConfig is a configuration for the parser.
type parsingConfig struct {
	// pkgName is the name of the package where the struct is located.
	pkgName string
	// typeName is the name of the struct.
	typeName string
	// loadPattern is the pattern to load the package.
	// For example, "github.com/nginxinc/nginx-gateway-fabric/pkg/mypackage" or "."
	// That path in the pattern are relative to the current working directory.
	loadPattern string
	// buildFlags are go build flags (e.g. -tags=foo).
	buildFlags []string
	// loadTests specifies whether the parser will load test files (e.g. *_test.go).
	loadTests bool
}

// field represents a field of a struct.
// the field is either a basic type, a slice of basic type or an embedded struct.
type field struct {
	docString            string
	name                 string
	embeddedStructFields []field
	fieldType            types.BasicKind
	slice                bool
	embeddedStruct       bool
}

// parse parses the struct defined by the config.
// The fields of the struct must satisfy the following rules:
// - Must be exported.
// - Must be of basic type, slice of basic type or embedded struct, where the embedded struct must satisfy the same
// rules.
// - Must have unique names across all embedded structs.
// - Must have a doc string comment for each field.
func parse(parsingCfg parsingConfig) ([]field, error) {
	mode := packages.NeedName | packages.NeedTypes | packages.NeedTypesInfo

	cfg := packages.Config{
		Mode:       mode,
		Fset:       token.NewFileSet(),
		Tests:      parsingCfg.loadTests,
		BuildFlags: parsingCfg.buildFlags,
	}

	pattern := "."
	if parsingCfg.loadPattern != "" {
		pattern = parsingCfg.loadPattern
	}

	loadedPackages, err := packages.Load(&cfg, pattern)
	if err != nil {
		return nil, fmt.Errorf("failed to load package: %w", err)
	}

	var pkg *packages.Package

	for _, p := range loadedPackages {
		if cfg.Tests && !strings.HasSuffix(p.ID, ".test]") {
			continue
		}

		if p.Name == parsingCfg.pkgName {
			pkg = p
			break
		}
	}

	if pkg == nil {
		return nil, fmt.Errorf("package %s not found", parsingCfg.pkgName)
	}

	targetType := pkg.Types.Scope().Lookup(parsingCfg.typeName)
	if targetType == nil {
		return nil, fmt.Errorf("type %s not found", parsingCfg.typeName)
	}

	s, ok := targetType.Type().Underlying().(*types.Struct)
	if !ok {
		return nil, fmt.Errorf("expected struct, got %s", targetType.Type().Underlying().String())
	}

	docStringProvider := newDocStringFieldsProvider(parsingCfg.loadTests, parsingCfg.buildFlags)

	return parseStruct(s, targetType.Type().String(), docStringProvider)
}

//nolint:gocyclo
func parseStruct(s *types.Struct, typeName string, docStringProvider *docStringFieldsProvider) ([]field, error) {
	nameOwners := make(map[string]string)

	var parseRecursively func(*types.Struct, string) ([]field, error)

	parseStructField := func(t *types.Named, f *types.Var, typeName string) (field, error) {
		if !f.Embedded() {
			return field{}, parsingError{
				typeName:  typeName,
				fieldName: f.Name(),
				msg:       "structs must be embedded",
			}
		}

		nextS, ok := t.Underlying().(*types.Struct)
		if !ok {
			return field{}, parsingError{
				typeName:  typeName,
				fieldName: f.Name(),
				msg:       fmt.Sprintf("must be struct, got %s", t.Underlying().String()),
			}
		}

		if !f.Exported() {
			return field{}, parsingError{
				typeName:  typeName,
				fieldName: f.Name(),
				msg:       "must be exported",
			}
		}

		embeddedFields, err := parseRecursively(nextS, t.String())
		if err != nil {
			return field{}, parsingError{
				typeName:  typeName,
				fieldName: f.Name(),
				msg:       err.Error(),
			}
		}

		return field{
			name:                 f.Name(),
			embeddedStruct:       true,
			embeddedStructFields: embeddedFields,
		}, nil
	}

	parseBasicTypeField := func(t *types.Basic, f *types.Var, typeName string) (field, error) {
		if f.Embedded() {
			return field{}, parsingError{
				typeName:  typeName,
				fieldName: f.Name(),
				msg:       "embedded basic types are not allowed",
			}
		}
		if !f.Exported() {
			return field{}, parsingError{
				typeName:  typeName,
				fieldName: f.Name(),
				msg:       "must be exported",
			}
		}
		if _, allowed := allowedBasicKinds[t.Kind()]; !allowed {
			return field{}, parsingError{
				typeName:  typeName,
				fieldName: f.Name(),
				msg:       fmt.Sprintf("type of field must be one of %s, got %s", supportedKinds, f.Type().String()),
			}
		}

		comment, err := docStringProvider.getDocString(typeName, f.Name())
		if err != nil {
			return field{}, parsingError{
				typeName:  typeName,
				fieldName: f.Name(),
				msg:       err.Error(),
			}
		}

		return field{
			name:      f.Name(),
			fieldType: t.Kind(),
			docString: comment,
		}, nil
	}

	parseSliceField := func(t *types.Slice, f *types.Var, typeName string) (field, error) {
		// slices can't be embedded so we don't check for that here
		if !f.Exported() {
			return field{}, parsingError{
				typeName:  typeName,
				fieldName: f.Name(),
				msg:       "must be exported",
			}
		}

		elemType, ok := t.Elem().(*types.Basic)
		if !ok {
			return field{}, parsingError{
				typeName:  typeName,
				fieldName: f.Name(),
				msg:       fmt.Sprintf("type of field must be one of %s, got %s", supportedKinds, f.Type().String()),
			}
		}

		if _, allowed := allowedBasicKinds[elemType.Kind()]; !allowed {
			return field{}, parsingError{
				typeName:  typeName,
				fieldName: f.Name(),
				msg:       fmt.Sprintf("type of field must be one of %s, got %s", supportedKinds, f.Type().String()),
			}
		}

		comment, err := docStringProvider.getDocString(typeName, f.Name())
		if err != nil {
			return field{}, parsingError{
				typeName:  typeName,
				fieldName: f.Name(),
				msg:       err.Error(),
			}
		}

		return field{
			name:      f.Name(),
			fieldType: elemType.Kind(),
			slice:     true,
			docString: comment,
		}, nil
	}

	parseRecursively = func(s *types.Struct, typeName string) ([]field, error) {
		var fields []field

		for i := 0; i < s.NumFields(); i++ {
			f := s.Field(i)

			var parsedField field
			var err error

			switch t := f.Type().(type) {
			case *types.Named: // when the field is a Struct
				parsedField, err = parseStructField(t, f, typeName)
			case *types.Basic: // when the field is a basic type like int, string, etc.
				parsedField, err = parseBasicTypeField(t, f, typeName)
			case *types.Slice: // when the field is a slice of basic type like []int.
				parsedField, err = parseSliceField(t, f, typeName)
			default:
				err = parsingError{
					typeName:  typeName,
					fieldName: f.Name(),
					msg:       fmt.Sprintf("must be of embedded struct, basic type or slice of basic type, got %s", f.Type().String()),
				}
			}

			if err != nil {
				return nil, err
			}

			fields = append(fields, parsedField)

			if owner, exists := nameOwners[f.Name()]; exists {
				return nil, parsingError{
					typeName:  typeName,
					fieldName: f.Name(),
					msg:       fmt.Sprintf("already exists in %s", owner),
				}
			}

			nameOwners[f.Name()] = typeName
		}

		return fields, nil
	}

	fields, err := parseRecursively(s, typeName)
	if err != nil {
		return nil, fmt.Errorf("failed to parse struct: %w", err)
	}

	return fields, nil
}

// allowedBasicKinds is a map of allowed basic types.
// Includes all supported types from go.opentelemetry.io/otel/attribute
// except for int.
// Since int size is platform dependent and because the size is required for Avro scheme, we don't use int.
var allowedBasicKinds = map[types.BasicKind]struct{}{
	types.Int64:   {},
	types.Float64: {},
	types.String:  {},
	types.Bool:    {},
}

var supportedKinds = func() string {
	kindsToString := map[types.BasicKind]string{
		types.Int64:   "int64",
		types.Float64: "float64",
		types.String:  "string",
		types.Bool:    "bool",
	}

	kinds := make([]string, 0, len(allowedBasicKinds))

	for k := range allowedBasicKinds {
		s, exist := kindsToString[k]
		if !exist {
			panic(fmt.Sprintf("unexpected basic kind %v", k))
		}

		kinds = append(kinds, s)
	}

	sort.Strings(kinds)

	return strings.Join(kinds, ", ")
}()
