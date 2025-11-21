package codegen

import (
	"bytes"
	"testing"
)

func TestPackageRender(t *testing.T) {
	tests := []struct {
		name     string
		pkgName  string
		expected string
	}{
		{
			name:     "simple package",
			pkgName:  "main",
			expected: "package main\n",
		},
		{
			name:     "custom package",
			pkgName:  "mylib",
			expected: "package mylib\n",
		},
		{
			name:     "empty package name",
			pkgName:  "",
			expected: "package \n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			p := Package(tt.pkgName)
			p.render(buf)

			got := buf.String()
			if got != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, got)
			}
		})
	}
}
