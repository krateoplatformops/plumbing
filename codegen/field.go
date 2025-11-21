package codegen

import (
	"bytes"
	"io"
)

// ifield is used to represent a key-value pair.
//
// It could be used in:
// - struct decl
// - struct value
// - method receiver
// - function parameter
// - function result
// - ...
type ifield struct {
	name      Node
	typ       Node
	value     Node
	separator string
	tags      map[string]string
}

func field(name, value any, sep string, tags map[string]string) *ifield {
	return &ifield{
		name:      parseNode(name),
		value:     parseNode(value),
		separator: sep,
		tags:      tags,
	}
}

func typedField(name, typ, value any, sep string, tags map[string]string) *ifield {
	return &ifield{
		name:      parseNode(name),
		typ:       parseNode(typ),
		value:     parseNode(value),
		separator: sep,
		tags:      tags,
	}
}

func (f *ifield) render(w io.Writer) {
	f.name.render(w)
	if f.typ != nil {
		writeString(w, " ")
		f.typ.render(w)
	}

	// separator + value solo se c'è davvero un value
	if f.value != nil {
		str := renderToString(f.value)
		if str != "" {
			writeString(w, f.separator)
			writeString(w, str)
		}
	}

	if len(f.tags) > 0 {
		writeString(w, " `")
		first := true
		for k, v := range f.tags {
			if !first {
				writeString(w, " ")
			}
			writeString(w, k+`:"`+v+`"`)
			first = false
		}
		writeString(w, "`")
	}
}

func renderToString(n Node) string {
	if n == nil {
		return ""
	}
	var b bytes.Buffer
	n.render(&b)
	return b.String()
}
