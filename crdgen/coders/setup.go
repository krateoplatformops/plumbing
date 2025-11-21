package coders

import (
	"bytes"
	"fmt"
	"go/format"
	"strings"

	gg "github.com/krateoplatformops/plumbing/codegen"
)

func newSetupCoder() *setupCoder {
	return &setupCoder{
		gen: gg.New(),
	}
}

type setupCoder struct {
	gen *gg.Generator
}

func (co *setupCoder) bytes(gofmt bool) ([]byte, error) {
	buf := bytes.Buffer{}
	co.gen.Write(&buf)

	if gofmt {
		return format.Source(buf.Bytes())
	}

	return buf.Bytes(), nil
}

func (co *setupCoder) addImports(kind, version string) {
	goVer := normalizeVersion(version, '_')

	co.gen.NewGroup().
		AddPackage("apis").NewImport().
		AddPath("k8s.io/apimachinery/pkg/runtime").
		AddPath(fmt.Sprintf("%s/apis/%s/%s",
			ModuleName(kind), strings.ToLower(kind), goVer))
}

func (co *setupCoder) addVar() {
	co.gen.NewGroup().
		AddLineComment("AddToSchemes may be used to add all resources defined in the project to a Scheme").
		NewVar().
		AddDecl("AddToSchemes", "runtime.SchemeBuilder")
}

func (co *setupCoder) addFuncs() {
	co.gen.NewGroup().
		AddLineComment("Register the types with the Scheme so the components can map objects to GroupVersionKinds and back").
		NewFunction("init").
		AddBody(gg.String("AddToSchemes = append(AddToSchemes, v1alpha1.SchemeBuilder.AddToScheme)"))

	co.gen.NewGroup().
		AddLineComment("AddToScheme adds all Resources to the Scheme").
		NewFunction("AddToScheme").
		AddParameter("s", "*runtime.Scheme").
		AddResult("", "error").
		AddBody(gg.String("return AddToSchemes.AddToScheme(s)"))
}
