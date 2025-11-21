package coders

import (
	"bytes"
	"fmt"
	"go/format"

	gg "github.com/krateoplatformops/plumbing/codegen"
)

func newGroupVersionInfoCoder() *groupVersionInfoCoder {
	return &groupVersionInfoCoder{
		gen: gg.New(),
	}
}

type groupVersionInfoCoder struct {
	gen *gg.Generator
}

func (co *groupVersionInfoCoder) bytes(gofmt bool) ([]byte, error) {
	buf := bytes.Buffer{}
	co.gen.Write(&buf)

	if gofmt {
		return format.Source(buf.Bytes())
	}

	return buf.Bytes(), nil
}

func (co *groupVersionInfoCoder) addImports(group, version string) {
	goVer := normalizeVersion(version, '_')
	crdVer := normalizeVersion(version, '-')

	co.gen.NewGroup().
		AddLineComment("+kubebuilder:object:generate=true").
		AddLineComment("+groupName=%s", group).
		AddLineComment("+versionName=%s", crdVer).
		AddPackage(goVer).NewImport().
		AddPath("reflect").
		AddPath("k8s.io/apimachinery/pkg/runtime/schema").
		AddPath("sigs.k8s.io/controller-runtime/pkg/scheme")
}

func (co *groupVersionInfoCoder) addConst(group, version string) {
	crdVer := normalizeVersion(version, '-')

	co.gen.NewGroup().AddLineComment("Package type metadata.").
		NewConst().
		AddField("Group", fmt.Sprintf("%q", group)).
		AddField("Version", fmt.Sprintf("%q", crdVer))
}

func (co *groupVersionInfoCoder) addVars(kind string) {
	co.gen.NewGroup().
		NewVar().
		AddField("SchemeGroupVersion", "schema.GroupVersion{Group: Group, Version: Version}").
		AddField("SchemeBuilder", "&scheme.Builder{GroupVersion: SchemeGroupVersion}").
		AddField(kind+"Kind", fmt.Sprintf("reflect.TypeOf(%s{}).Name()", kind)).
		AddField(kind+"GroupKind", fmt.Sprintf("schema.GroupKind{Group: Group, Kind: %sKind}.String()", kind)).
		AddField(kind+"KindAPIVersion", kind+"Kind + \".\" + SchemeGroupVersion.String()").
		AddField(kind+"GroupVersionKind", fmt.Sprintf("SchemeGroupVersion.WithKind(%sKind)", kind))
}

func (co *groupVersionInfoCoder) initFunc(kind string) {
	co.gen.NewGroup().
		NewFunction("init").
		AddBody(gg.String("SchemeBuilder.Register(&%s{}, &%sList{})", kind, kind))
}
