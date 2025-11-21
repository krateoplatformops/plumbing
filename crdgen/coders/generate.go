package coders

import (
	"bytes"
	"go/format"
	"strings"

	gg "github.com/krateoplatformops/plumbing/codegen"
)

func newGenerateCoder() *generateCoder {
	return &generateCoder{
		gen: gg.New(),
	}
}

type generateCoder struct {
	gen *gg.Generator
}

func (co *generateCoder) bytes(gofmt bool) ([]byte, error) {
	buf := bytes.Buffer{}
	co.gen.Write(&buf)

	if gofmt {
		return format.Source(buf.Bytes())
	}

	return buf.Bytes(), nil
}

func (co *generateCoder) generate() {
	co.gen.NewGroup().
		AddLineComment("+build generate").
		AddLine().
		AddLineComment("Remove existing CRDs").
		AddLineComment("go:generate rm -rf ../crds").
		AddLine().
		AddLineComment("Generate deepcopy methodsets and CRD manifests").
		AddLine().
		AddLineComment(strings.Join([]string{
			"go:generate go run -tags generate sigs.k8s.io/controller-tools/cmd/controller-gen",
			"object:headerFile=../hack/boilerplate.go.txt",
			"paths=./...",
			"crd:crdVersions=v1 output:artifacts:config=../crds",
		}, " ")).
		AddLine().
		AddPackage("apis").AddLine().
		NewImport().
		//AddLineComment("nolint:typecheck").
		AddAlias("sigs.k8s.io/controller-tools/cmd/controller-gen", "_")
}
