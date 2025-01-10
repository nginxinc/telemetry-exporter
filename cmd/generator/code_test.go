//go:build generator

package main

import (
	"bytes"
	"testing"

	. "github.com/onsi/gomega"

	"github.com/nginx/telemetry-exporter/cmd/generator/tests"
)

func TestGenerateCode(t *testing.T) {
	t.Parallel()
	g := NewGomegaWithT(t)

	cfg := parsingConfig{
		pkgName:     "tests",
		typeName:    "Data",
		loadPattern: "github.com/nginx/telemetry-exporter/cmd/generator/tests",
		buildFlags:  []string{"-tags=generator"},
	}

	_ = tests.Data{} // depends on the type being defined

	pResult, err := parse(cfg)

	g.Expect(err).ToNot(HaveOccurred())

	var buf bytes.Buffer

	codeCfg := codeGenConfig{
		packagePath: pResult.packagePath,
		typeName:    "Data",
		fields:      pResult.fields,
	}

	g.Expect(generateCode(&buf, codeCfg)).To(Succeed())

	g.Expect(buf.Bytes()).ToNot(BeEmpty())
}
