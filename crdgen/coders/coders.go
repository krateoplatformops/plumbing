package coders

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/krateoplatformops/plumbing/crdgen/assets"
	"github.com/krateoplatformops/plumbing/env"
)

const (
	EnvFormatCode = "FORMAT_CODE"
	EnvKeepCode   = "KEEP_CODE"
	EnvVerbose    = "VERBOSE"
)

type Options struct {
	Group        string
	Version      string
	Kind         string
	Categories   []string
	SpecSchema   []byte
	StatusSchema []byte
	Managed      bool
}

func ModuleName(kind string) string {
	return fmt.Sprintf("github.com/krateoplatformops/%s-crdgen", strings.ToLower(kind))
}

func SourceDir(rootdir, kind string) string {
	mod := ModuleName(kind)

	parts := []string{rootdir}
	parts = append(parts, strings.Split(mod, "/")...)

	return filepath.Join(parts...)
}

func GenAll(rootdir string, opts *Options) error {
	mod := ModuleName(opts.Kind)

	err := WriteTypesToFile(rootdir, opts)
	if err != nil {
		return err
	}

	err = WriteGroupVersionInfoToFile(rootdir, opts)
	if err != nil {
		return err
	}

	err = WriteGenerateToFile(rootdir, opts)
	if err != nil {
		return err
	}

	err = WriteSetupToFile(rootdir, opts)
	if err != nil {
		return err
	}

	err = assets.RenderToFile(rootdir, mod)
	if err != nil {
		return err
	}

	err = assets.ExportBoilerPlate(rootdir, mod)
	return err
}

func WriteSetupToFile(rootdir string, opts *Options) error {
	mod := ModuleName(opts.Kind)

	parts := []string{rootdir}
	parts = append(parts, strings.Split(mod, "/")...)
	parts = append(parts, "apis")

	workdir := filepath.Join(parts...)
	err := os.MkdirAll(workdir, os.ModePerm)
	if err != nil {
		if !errors.Is(err, os.ErrExist) {
			return err
		}
	}

	bin, err := GenSetup(opts)
	if err != nil {
		return err
	}

	out, err := os.Create(filepath.Join(workdir, "apis.go"))
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, bytes.NewReader(bin))
	return err
}

func WriteGenerateToFile(rootdir string, opts *Options) error {
	mod := ModuleName(opts.Kind)

	parts := []string{rootdir}
	parts = append(parts, strings.Split(mod, "/")...)
	parts = append(parts, "apis")

	workdir := filepath.Join(parts...)
	err := os.MkdirAll(workdir, os.ModePerm)
	if err != nil {
		if !errors.Is(err, os.ErrExist) {
			return err
		}
	}

	bin, err := GenGenerate(opts)
	if err != nil {
		return err
	}

	out, err := os.Create(filepath.Join(workdir, "generate.go"))
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, bytes.NewReader(bin))
	return err
}

func WriteGroupVersionInfoToFile(rootdir string, opts *Options) error {
	mod := ModuleName(opts.Kind)

	goVer := normalizeVersion(opts.Version, '_')

	parts := []string{rootdir}
	parts = append(parts, strings.Split(mod, "/")...)
	parts = append(parts, "apis", strings.ToLower(opts.Kind), goVer)

	workdir := filepath.Join(parts...)
	err := os.MkdirAll(workdir, os.ModePerm)
	if err != nil {
		if !errors.Is(err, os.ErrExist) {
			return err
		}
	}

	bin, err := GenGroupVersionInfo(opts)
	if err != nil {
		return err
	}

	out, err := os.Create(filepath.Join(workdir, "group_version_info.go"))
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, bytes.NewReader(bin))
	return err
}

func WriteTypesToFile(rootdir string, opts *Options) error {
	mod := ModuleName(opts.Kind)

	goVer := normalizeVersion(opts.Version, '_')

	parts := []string{rootdir}
	parts = append(parts, strings.Split(mod, "/")...)
	parts = append(parts, "apis", strings.ToLower(opts.Kind), goVer)

	workdir := filepath.Join(parts...)
	err := os.MkdirAll(workdir, os.ModePerm)
	if err != nil {
		if !errors.Is(err, os.ErrExist) {
			return err
		}
	}

	bin, err := GenTypes(opts)
	if err != nil {
		return err
	}

	out, err := os.Create(filepath.Join(workdir, "types.go"))
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, bytes.NewReader(bin))
	return err
}

func GenTypes(opts *Options) (dat []byte, err error) {
	co := newTypesCoder()

	err = co.parseSchemaForSpec(opts.SpecSchema)
	if err != nil {
		return nil, err
	}

	err = co.parseSchemaForStatus(opts.StatusSchema)
	if err != nil {
		return nil, err
	}

	co.addImports(opts.Version, opts.Managed)

	err = co.buildStructForDefs()
	if err != nil {
		return nil, err
	}

	err = co.buildStructForSpec(opts.Kind)
	if err != nil {
		return nil, err
	}

	err = co.buildStructForStatus(opts.Kind, opts.Managed)
	if err != nil {
		return nil, err
	}

	co.buildEntryItemStructs(opts.Kind, opts.Categories, opts.Managed)
	co.buildEntryListStructs(opts.Kind, opts.Managed)

	return co.bytes(env.True(EnvFormatCode))
}

func GenGroupVersionInfo(opts *Options) (dat []byte, err error) {
	co := newGroupVersionInfoCoder()

	co.addImports(opts.Group, opts.Version)
	co.addConst(opts.Group, opts.Version)

	co.addVars(opts.Kind)

	co.initFunc(opts.Kind)

	return co.bytes(env.True(EnvFormatCode))
}

func GenGenerate(opts *Options) (dat []byte, err error) {
	co := newGenerateCoder()

	co.generate()

	return co.bytes(env.True(EnvFormatCode))
}

func GenSetup(opts *Options) (dat []byte, err error) {
	co := newSetupCoder()

	co.addImports(opts.Kind, opts.Version)
	co.addVar()
	co.addFuncs()

	return co.bytes(env.True(EnvFormatCode))
}
