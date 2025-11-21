package tools

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/krateoplatformops/plumbing/env"
)

func Tidy(dir string) error {
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return fmt.Errorf("unable to resolve path '%s': %w", dir, err)
	}

	cmd := exec.Command(
		"go", "mod", "tidy",
	)

	cmd.Dir = absDir

	if env.True("VERBOSE") {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	} else {
		cmd.Stdout = io.Discard
		cmd.Stderr = io.Discard
	}

	return cmd.Run()
}

func GenerateCRDs(dir string) (dat []byte, err error) {
	absDir, err := filepath.Abs(dir)
	if err != nil {
		err = fmt.Errorf("unable to resolve path '%s': %w", dir, err)
		return
	}

	cmd := exec.Command(
		"go", "run", "-tags", "generate",
		"sigs.k8s.io/controller-tools/cmd/controller-gen",
		"object:headerFile=./hack/boilerplate.go.txt",
		"paths=./...",
		"crd:crdVersions=v1",
		"output:artifacts:config=./crds",
	)

	cmd.Dir = absDir

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Run()
	if err != nil {
		return
	}

	// Trova il file YAML generato
	var files []string
	crdDir := filepath.Join(absDir, "crds")
	files, err = filepath.Glob(filepath.Join(crdDir, "*.yaml"))
	if err != nil {
		return
	}
	if len(files) == 0 {
		err = fmt.Errorf("no CRD found in: %s", crdDir)
		return
	}

	return os.ReadFile(files[0])
}
