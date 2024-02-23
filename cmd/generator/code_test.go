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

	fields, err := parse(cfg)

	g.Expect(err).ToNot(HaveOccurred())

	var buf bytes.Buffer

	codeCfg := codeGenConfig{
		packageName: "tests",
		typeName:    "Data",
		fields:      fields,
	}

	err = generateCode(&buf, codeCfg)

	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(buf.Bytes()).ToNot(BeEmpty())
}
