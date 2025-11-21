//go:build integration
// +build integration

package assets_test

import (
	"bytes"
	"fmt"
	"os"
	"testing"

	"github.com/krateoplatformops/plumbing/crdgen/assets"
)

func TestRender(t *testing.T) {
	ds := map[string]string{
		"module": "github.com/krateoplatformops/form1",
	}

	err := assets.Render(os.Stdout, "go.mod", ds)
	if err != nil {
		t.Fatal(err)
	}
}

func TestExport(t *testing.T) {
	ds := map[string]string{
		"module": "github.com/krateoplatformops/form1",
	}

	buf := bytes.Buffer{}
	err := assets.Render(&buf, "go.mod", ds)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(buf.String())
}
