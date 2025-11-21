package coders

import (
	"fmt"
	"strings"
	"testing"

	"github.com/krateoplatformops/plumbing/crdgen/schemas"
)

func TestNormalizeVersion(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		// Basic cases
		{"simple semantic version", "v1.0.2", "v1-0-2"},
		{"numeric only", "1.2.3", "v1-2-3"},
		{"beta with dash", "v1-beta.2", "v1-beta-2"},
		{"release path", "release/1.0.0", "release-1-0-0"},
		{"mixed symbols", "RC_2024-05", "rc-2024-05"},

		// Edge cases
		{"empty string", "", ""},
		{"starts with non-alphanumeric", "-v1.0.0", "v1-0-0"},
		{"multiple separators", "v1..0--2", "v1-0-2"},
		{"trailing symbols", "v1.0.0-", "v1-0-0"},
		{"leading digit", "9alpha", "v9alpha"},
		{"only digits", "123", "v123"},
		{"uppercase letters", "V2.BETA", "v2-beta"},
		{"multiple slashes", "release/v2/1.0", "release-v2-1-0"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := normalizeVersion(tt.input, '-')
			if got != tt.expected {
				t.Errorf("normalizeVersion(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestParseAdditionalProperties(t *testing.T) {
	const (
		js = `{
   "type": "object", 
   "additionalProperties": false,
   "x-kubernetes-preserve-unknown-fields": false
}`
	)

	sch, err := schemas.FromJSONReader(strings.NewReader(js))
	if err != nil {
		t.Fatal(err)
	}

	ok := mustPreserveUnknownFields((*schemas.Type)(sch.ObjectAsType))
	fmt.Println(ok)
}
