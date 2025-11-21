//go:build ignore
// +build ignore

//go:generate go run generate.go

// This file is used to generate CRD golden files from JSON Schemas.
// Run with `go generate` or `go run generate.go`.

package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/krateoplatformops/plumbing/crdgen"
	"github.com/krateoplatformops/plumbing/crdgen/coders"
)

func main() {
	const (
		testDir       = "testdata"
		defaultStatus = `{"type": "object", "additionalProperties": true,"x-kubernetes-preserve-unknown-fields": true}`
	)

	keepCodeFor := ""

	files, err := os.ReadDir(testDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to read testdir: %v\n", err)
		os.Exit(1)
	}

	for _, f := range files {
		if !strings.HasSuffix(f.Name(), ".schema.json") {
			continue
		}

		if keepCodeFor != "" && strings.HasPrefix(f.Name(), keepCodeFor) {
			os.Setenv(coders.EnvKeepCode, "1")
		} else {
			os.Setenv(coders.EnvKeepCode, "0")
		}

		name := strings.TrimSuffix(f.Name(), ".schema.json")
		schemaPath := filepath.Join(testDir, f.Name())
		crdPath := filepath.Join(testDir, name+".crd.yaml")

		fmt.Printf("Generating CRD for %s → %s\n", f.Name(), name+".crd.yaml")

		input, err := os.ReadFile(schemaPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to read schema %s: %v\n", schemaPath, err)
			continue
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
			fmt.Fprintf(os.Stderr, "failed to generate CRD for %s: %v\n", schemaPath, err)
			continue
		}

		// Check if file exists and content is unchanged
		if existing, err := os.ReadFile(crdPath); err == nil && bytes.Equal(existing, got) {
			fmt.Printf("CRD %s is up-to-date, skipping write.\n", crdPath)
			continue
		}

		out, err := os.Create(crdPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to create CRD file %s: %v\n", crdPath, err)
			continue
		}

		_, err = io.Copy(out, bytes.NewReader(got))
		out.Close()
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to write CRD file %s: %v\n", crdPath, err)
			continue
		}

		fmt.Printf("CRD %s generated successfully.\n", crdPath)

		if strings.HasPrefix("preserve", f.Name()) {
			os.Setenv("KEEP_CODE", "1")
		}
	}
}
