package kubeutil

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMakeDNS1123Compatible(t *testing.T) {
	examples := []struct {
		name     string
		expected string
	}{
		{
			name:     "Pinco.Pallo-kubeworld.it-clientconfig",
			expected: "pincopallo-kubeworldit-clientconfig",
		},
		{
			name:     "tOk3_?ofTHE-Year",
			expected: "tok3ofthe-year",
		},
		{
			name:     "----tOk3_?ofTHE-YEAR!",
			expected: "tok3ofthe-year",
		},
	}

	for _, example := range examples {
		t.Run(example.name, func(t *testing.T) {
			name := MakeDNS1123Compatible(example.name)

			assert.Equal(t, example.expected, name)
			assertDNS1123Compatibility(t, name)
		})
	}
}

func assertDNS1123Compatibility(t *testing.T, name string) {
	dns1123MaxLength := 63
	dns1123FormatRegexp := regexp.MustCompile("^[a-z0-9]([-a-z0-9]*[a-z0-9])?$")

	assert.True(t, len(name) <= dns1123MaxLength, "Name length needs to be shorter than %d", dns1123MaxLength)
	assert.Regexp(t, dns1123FormatRegexp, name, "Name needs to be in DNS-1123 allowed format")
}
