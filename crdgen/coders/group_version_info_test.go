package coders

import (
	"bytes"
	"io"
	"os"
	"testing"
)

func TestGenGroupVersionInfo(t *testing.T) {
	os.Setenv(EnvFormatCode, "1")

	specSchemaBytes, err := os.ReadFile("../../testdata/git.spec.schema.json")
	if err != nil {
		t.Fatal(err)
	}

	statusSchemaBytes, err := os.ReadFile("../../testdata/git.status.schema.json")
	if err != nil {
		t.Fatal(err)
	}

	opts := Options{
		Group:        "git.krateo.io",
		Version:      "v1alpha1",
		Kind:         "Repo",
		Categories:   []string{"krateo", "git", "repo"},
		SpecSchema:   specSchemaBytes,
		StatusSchema: statusSchemaBytes,
		Managed:      true,
	}

	dat, err := GenGroupVersionInfo(&opts)
	if err != nil {
		t.Fatal(err)
	}

	io.Copy(os.Stdout, bytes.NewReader(dat))
}
