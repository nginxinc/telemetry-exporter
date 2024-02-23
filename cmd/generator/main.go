//go:build generator

package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
)

var (
	code                     = flag.Bool("code", true, "Generate code")
	buildTags                = flag.String("build-tags", "", "Comma separated list of build tags expected in the source files and that will be added to the generated code") //nolint:lll
	scheme                   = flag.Bool("scheme", false, "Generate Avro scheme")
	schemeNamespace          = flag.String("scheme-namespace", "gateway.nginx.org", "Scheme namespace; required when -scheme is set") //nolint:lll
	schemeProtocol           = flag.String("scheme-protocol", "", "Scheme protocol; required when -scheme is set")
	schemeDataFabricDataType = flag.String("scheme-df-datatype", "", "Scheme data fabric data type; required when -scheme is set") //nolint:lll
	typeName                 = flag.String("type", "", "Type to generate; required")
)

func exitWithError(err error) {
	fmt.Fprintln(os.Stderr, "error: "+err.Error())
	os.Exit(1)
}

func existWithUsage() {
	flag.Usage()
	os.Exit(1)
}

func validateFlags() {
	if *typeName == "" {
		existWithUsage()
	}

	if *scheme {
		if *schemeNamespace == "" {
			existWithUsage()
		}
		if *schemeProtocol == "" {
			existWithUsage()
		}
		if *schemeDataFabricDataType == "" {
			existWithUsage()
		}
	}
}

func main() {
	flag.Parse()

	validateFlags()

	pkgName := os.Getenv("GOPACKAGE")
	if pkgName == "" {
		exitWithError(fmt.Errorf("GOPACKAGE is not set"))
	}

	var buildFlags []string
	if *buildTags != "" {
		buildFlags = []string{"-tags=" + *buildTags}
	}

	cfg := parsingConfig{
		pkgName:    pkgName,
		typeName:   *typeName,
		buildFlags: buildFlags,
	}

	fields, err := parse(cfg)
	if err != nil {
		exitWithError(fmt.Errorf("failed to parse struct: %w", err))
	}

	fmt.Printf("Successfully parsed struct %s\n", *typeName)

	if *code {
		fmt.Println("Generating code")

		fileName := fmt.Sprintf("%s_attributes_generated.go", strings.ToLower(*typeName))

		file, err := os.Create(fileName)
		if err != nil {
			exitWithError(fmt.Errorf("failed to create file: %w", err))
		}
		defer file.Close()

		var codeGenBuildTags string
		if *buildTags != "" {
			codeGenBuildTags = strings.ReplaceAll(*buildTags, ",", " && ")
		}

		codeCfg := codeGenConfig{
			packageName: pkgName,
			typeName:    *typeName,
			fields:      fields,
			buildTags:   codeGenBuildTags,
		}

		if err := generateCode(file, codeCfg); err != nil {
			exitWithError(fmt.Errorf("failed to generate code: %w", err))
		}
	}

	if *scheme {
		fmt.Println("Generating scheme")

		fileName := fmt.Sprintf("%s.avdl", strings.ToLower(*typeName))

		file, err := os.Create(fileName)
		if err != nil {
			exitWithError(fmt.Errorf("failed to create file: %w", err))
		}
		defer file.Close()

		schemeCfg := schemeGenConfig{
			namespace:          *schemeNamespace,
			protocol:           *schemeProtocol,
			dataFabricDataType: *schemeDataFabricDataType,
			record:             *typeName,
			fields:             fields,
		}

		if err := generateScheme(file, schemeCfg); err != nil {
			exitWithError(fmt.Errorf("failed to generate scheme: %w", err))
		}
	}
}
