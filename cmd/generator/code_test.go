//go:build generator

package main

import (
	"bytes"
	"testing"

	. "github.com/onsi/gomega"

	"github.com/nginxinc/telemetry-exporter/cmd/generator/tests"
)

func TestGenerateCode(t *testing.T) {
	g := NewGomegaWithT(t)

	cfg := parsingConfig{
		pkgName:     "tests",
		typeName:    "Data",
		loadPattern: "github.com/nginxinc/telemetry-exporter/cmd/generator/tests",
		buildFlags:  []string{"-tags=generator"},
	}

	_ = tests.Data{} // depends on the type being defined

	parsingResult, err := parse(cfg)

	g.Expect(err).ToNot(HaveOccurred())

	var buf bytes.Buffer

	codeCfg := codeGenConfig{
		packagePath: parsingResult.packagePath,
		typeName:    "Data",
		fields:      parsingResult.fields,
	}

	g.Expect(generateCode(&buf, codeCfg)).To(Succeed())

	g.Expect(buf.Bytes()).ToNot(BeEmpty())
}
