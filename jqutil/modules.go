package jqutil

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"

	"github.com/itchyny/gojq"
)

func DirModuleLoader(basePath string) gojq.ModuleLoader {
	return &dirModuleLoader{
		basePath: basePath,
	}
}

type dirModuleLoader struct {
	basePath string // es: "/app/modules"
}

func (l *dirModuleLoader) LoadModule(name string) (*gojq.Query, error) {
	fullPath := filepath.Join(l.basePath, name+".jq")

	content, err := os.ReadFile(fullPath)
	if err != nil {
		return nil, fmt.Errorf("module %q not found: %w", name, err)
	}

	parsed, err := gojq.Parse(string(content))
	if err != nil {
		return nil, fmt.Errorf("error parsing module %q: %w", name, err)
	}

	return parsed, nil
}

func newEmbedModuleLoader() gojq.ModuleLoader {
	return &embedModuleLoader{
		fsys: modulesFS,
	}
}

var (
	//go:embed assets/*.jq
	modulesFS embed.FS
)

type embedModuleLoader struct {
	fsys embed.FS
}

func (l *embedModuleLoader) LoadModule(name string) (*gojq.Query, error) {
	path := "assets/" + name + ".jq"
	content, err := l.fsys.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("module %q not found: %w", name, err)
	}

	code, err := gojq.Parse(string(content))
	if err != nil {
		return nil, fmt.Errorf("failed to parse module %q: %w", name, err)
	}

	return code, nil
}
