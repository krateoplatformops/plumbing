package schemas

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

var (
	ErrCannotResolveSchema      = errors.New("cannot resolve schema")
	ErrCannotLoadSchema         = errors.New("cannot load schema")
	ErrUnsupportedContentType   = errors.New("unsupported content type")
	ErrUnsupportedFileExtension = errors.New("unsupported file extension")
	ErrUnsupportedURL           = errors.New("unsupported URL")
)

type Loader interface {
	Load(uri, parentURI string) (*Schema, error)
}

func NewCachedLoader(loader Loader, cache map[string]*Schema) *CachedLoader {
	return &CachedLoader{
		loader: loader,
		cache:  cache,
	}
}

type CachedLoader struct {
	loader Loader
	cache  map[string]*Schema
}

func (l *CachedLoader) Load(uri, parentURI string) (*Schema, error) {
	if schema, ok := l.cache[uri]; ok {
		return schema, nil
	}

	schema, err := l.loader.Load(uri, parentURI)
	if err != nil {
		return nil, errors.Join(ErrCannotLoadSchema, err)
	}

	l.cache[uri] = schema

	return schema, nil
}

func NewFileLoader(resolveExtensions []string) *FileLoader {
	return &FileLoader{
		resolveExtensions: resolveExtensions,
	}
}

type FileLoader struct {
	resolveExtensions []string
}

func (l *FileLoader) Load(fileName, parentFileName string) (*Schema, error) {
	qualified, err := QualifiedFileName(fileName, parentFileName, l.resolveExtensions)
	if err != nil {
		return nil, err
	}

	schema, err := l.parseFile(qualified)
	if err != nil {
		return nil, err
	}

	return schema, nil
}

func (l *FileLoader) parseFile(fileName string) (*Schema, error) {
	sc, err := FromJSONFile(fileName)
	if err != nil {
		return nil, fmt.Errorf("error parsing JSON file %s: %w", fileName, err)
	}

	return sc, nil
}

func NewDefaultCacheLoader(resolveExtensions []string) *CachedLoader {
	return NewCachedLoader(NewFileLoader(resolveExtensions), map[string]*Schema{})
}

func QualifiedFileName(fileName, parentFileName string, resolveExtensions []string) (string, error) {
	r, err := GetRefType(fileName)
	if err != nil {
		return "", err
	}

	if r != RefTypeFile {
		return fileName[strings.Index(fileName, "://")+3:], nil
	}

	fileName = strings.TrimPrefix(fileName, "file://")

	if !filepath.IsAbs(fileName) {
		fileName = filepath.Join(filepath.Dir(parentFileName), fileName)
	}

	exts := append([]string{""}, resolveExtensions...)
	for _, ext := range exts {
		qualified := fileName + ext

		if !fileExists(qualified) {
			continue
		}

		var err error

		qualified, err = filepath.EvalSymlinks(qualified)
		if err != nil {
			return "", fmt.Errorf("error resolving symlinks in %s: %w", qualified, err)
		}

		return qualified, nil
	}

	return "", fmt.Errorf("%w %q", ErrCannotResolveSchema, fileName)
}

func fileExists(fileName string) bool {
	_, err := os.Stat(fileName)

	return err == nil || !os.IsNotExist(err)
}

func toExtensionSet(items []string) map[string]bool {
	set := make(map[string]bool, len(items))

	for _, item := range items {
		if !strings.HasPrefix(item, ".") {
			item = "." + item
		}

		set[item] = true
	}

	return set
}
