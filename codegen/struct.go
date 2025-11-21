package codegen

import (
	"io"
)

type IStruct struct {
	name  string
	items *group
}

// Struct will insert a new struct.
func Struct(name string) *IStruct {
	return &IStruct{
		name: name,
		// We will insert new line before closing the struct to avoid being affect
		// by line comments.
		items: newGroup("{\n", "\n}", "\n"),
	}
}

func (i *IStruct) render(w io.Writer) {
	writeStringF(w, "type %s struct ", i.name)
	i.items.render(w)
}

// AddLine will insert an empty line.
func (i *IStruct) AddLine() *IStruct {
	i.items.append(Line())
	return i
}

// AddLineComment will insert a new line comment.
func (i *IStruct) AddLineComment(content string, args ...any) *IStruct {
	i.items.append(LineComment(content, args...))
	return i
}

func (i *IStruct) AddField(name, typ any, tags map[string]string) *IStruct {
	i.items.append(field(name, typ, " ", tags))
	return i
}
