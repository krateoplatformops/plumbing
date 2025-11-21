package assets

import (
	"bytes"
	"embed"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

//go:embed files/*.tpl
//go:embed files/*.txt
var tmplFS embed.FS

func Render(w io.Writer, name string, data any) error {
	if !strings.HasSuffix(name, ".tpl") {
		name = fmt.Sprintf("%s.tpl", name)
	}

	eng := template.Must(template.New("").ParseFS(tmplFS, "files/*.tpl"))

	tmpl := template.Must(eng.Clone())
	tmpl = template.Must(tmpl.ParseFS(tmplFS, "files/"+name))
	return tmpl.ExecuteTemplate(w, name, data)
}

func RenderToFile(rootdir, mod string) error {
	parts := []string{rootdir}
	parts = append(parts, strings.Split(mod, "/")...)

	workdir := filepath.Join(parts...)
	err := os.MkdirAll(workdir, os.ModePerm)
	if err != nil {
		if !errors.Is(err, os.ErrExist) {
			return err
		}
	}

	buf := bytes.Buffer{}
	err = Render(&buf, "go.mod", map[string]string{"module": mod})
	if err != nil {
		return err
	}

	out, err := os.Create(filepath.Join(workdir, "go.mod"))
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, &buf)
	return err
}

func ExportBoilerPlate(rootdir, mod string) error {
	parts := []string{rootdir}
	parts = append(parts, strings.Split(mod, "/")...)
	parts = append(parts, "hack")

	workdir := filepath.Join(parts...)
	err := os.MkdirAll(workdir, os.ModePerm)
	if err != nil {
		if !errors.Is(err, os.ErrExist) {
			return err
		}
	}

	in, err := tmplFS.Open("files/boilerplate.go.txt")
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(filepath.Join(workdir, "boilerplate.go.txt"))
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)

	return err
}
