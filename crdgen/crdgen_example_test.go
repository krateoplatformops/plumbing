//go:build integration
// +build integration

package crdgen_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/krateoplatformops/plumbing/crdgen"
)

func TestGoldenFiles(t *testing.T) {
	//os.Setenv("KEEP_CODE", "1")
	const (
		testDir       = "testdata"
		defaultStatus = `{"type": "object", "additionalProperties": true,"x-kubernetes-preserve-unknown-fields": true}`
	)

	files, err := os.ReadDir(testDir)
	if err != nil {
		t.Fatal(err)
	}

	for _, f := range files {
		if strings.HasSuffix(f.Name(), ".schema.json") {
			name := strings.TrimSuffix(f.Name(), ".schema.json")
			t.Run(name, func(t *testing.T) {
				schemaPath := filepath.Join(testDir, name+".schema.json")
				crdExpectedPath := filepath.Join(testDir, name+".crd.yaml")

				input, err := os.ReadFile(schemaPath)
				if err != nil {
					t.Fatal(err)
				}

				expected, err := os.ReadFile(crdExpectedPath)
				if err != nil {
					t.Fatal(err)
				}

				got, err := crdgen.Generate(crdgen.Options{
					Group:        "test.krateo.io",
					Version:      "v0.0.0",
					Kind:         "Hello",
					Categories:   []string{"krateo", "test", "hello"},
					SpecSchema:   input,
					StatusSchema: []byte(defaultStatus),
				})
				if err != nil {
					t.Fatal(err)
				}

				if diff := cmp.Diff(string(expected), string(got)); diff != "" {
					t.Errorf("CRD mismatch (-want +got):\n%s", diff)
				}
			})
		}
	}
}
