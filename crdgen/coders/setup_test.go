package coders

import (
	"bytes"
	"io"
	"os"
	"testing"
)

func TestGenSetup(t *testing.T) {
	os.Setenv(EnvFormatCode, "1")

	opts := Options{
		Group:   "git.krateo.io",
		Version: "v1alpha1",
		Kind:    "Repo",
	}

	dat, err := GenSetup(&opts)
	if err != nil {
		t.Fatal(err)
	}

	io.Copy(os.Stdout, bytes.NewReader(dat))
}
