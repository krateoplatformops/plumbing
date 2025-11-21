package coders

import (
	"fmt"
	"os"
	"testing"

	"github.com/krateoplatformops/plumbing/crdgen/tools"
)

func TestGenAll(t *testing.T) {
	//os.Setenv(EnvKeepCode, "1")
	os.Setenv(EnvFormatCode, "1")

	rootdir := os.TempDir()

	specSchemaBytes, err := os.ReadFile("../../testdata/subnet.spec.schema.json")
	if err != nil {
		t.Fatal(err)
	}

	var statusSchemaBytes []byte
	statusSchemaBytes, err = os.ReadFile("../../testdata/subnet.status.schema.json")
	if err != nil {
		t.Fatal(err)
	}

	opts := Options{
		Group:        "vclusters.subnet.krateo.io",
		Version:      "v1alpha1",
		Kind:         "Subnet",
		Categories:   []string{"krateo", "vcluster"},
		SpecSchema:   specSchemaBytes,
		StatusSchema: statusSchemaBytes,
		Managed:      false,
	}

	err = GenAll(rootdir, &opts)
	if err != nil {
		t.Fatal(err)
	}

	srcdir := SourceDir(rootdir, opts.Kind)
	defer os.RemoveAll(srcdir)

	err = tools.Tidy(srcdir)
	if err != nil {
		t.Fatal(err)
	}
	yml, err := tools.GenerateCRDs(srcdir)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(string(yml))
}
