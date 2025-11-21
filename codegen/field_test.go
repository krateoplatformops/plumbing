package codegen

import (
	"bytes"
	"testing"
)

func TestFieldRender(t *testing.T) {
	buf := &bytes.Buffer{}
	f := field("foo", "bar", ":", nil)
	f.render(buf)

	got := buf.String()
	want := "foo:bar"

	if got != want {
		t.Errorf("expected %q, got %q", want, got)
	}
}

func TestTypedFieldRender(t *testing.T) {
	buf := &bytes.Buffer{}
	f := typedField("foo", "int", "42", "=", nil)
	f.render(buf)

	got := buf.String()
	want := "foo int=42"

	if got != want {
		t.Errorf("expected %q, got %q", want, got)
	}
}

func TestTypedFieldRender_NoValue(t *testing.T) {
	buf := &bytes.Buffer{}
	f := typedField("foo", "string", "", " ", map[string]string{"json": "name,omitempty"})
	f.render(buf)

	got := buf.String()
	want := "foo string `json:\"name,omitempty\"`"

	if got != want {
		t.Errorf("expected %q, got %q", want, got)
	}
}
