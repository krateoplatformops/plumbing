package jqutil

import (
	"reflect"
	"testing"
)

func TestInferType(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected any
	}{
		{
			name:     "Empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "Boolean true",
			input:    "true",
			expected: true,
		},
		{
			name:     "Boolean false",
			input:    "false",
			expected: false,
		},
		{
			name:     "Null value",
			input:    "null",
			expected: nil,
		},
		{
			name:     "Nil value",
			input:    "nil",
			expected: nil,
		},
		{
			name:     "Integer within int32 range",
			input:    "123",
			expected: int32(123),
		},
		{
			name:     "Integer outside int32 range",
			input:    "2147483648",
			expected: int64(2147483648),
		},
		{
			name:     "Floating point number",
			input:    "3.14",
			expected: 3.14,
		},
		{
			name:     "Map with string keys and values",
			input:    "{\"foo\": \"bar\"}",
			expected: map[string]any{"foo": "bar"},
		},
		{
			name:     "Unquoted string",
			input:    "hello world",
			expected: "hello world",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := InferType(tt.input)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("InferType(%q) = %v (%T), want %v (%T)", tt.input, result, result, tt.expected, tt.expected)
			}
		})
	}
}
