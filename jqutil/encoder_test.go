package jqutil

import (
	"bytes"
	"math"
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEncoder_encodeBool(t *testing.T) {
	tests := []struct {
		value    bool
		expected string
	}{
		{true, "true"},
		{false, "false"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			encoder := newEncoder(false, 0)
			err := encoder.encode(tt.value)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, encoder.w.String())
		})
	}
}

func TestEncoder_encodeInt(t *testing.T) {
	tests := []struct {
		value    int
		expected string
	}{
		{123, "123"},
		{-456, "-456"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			encoder := newEncoder(false, 0)
			err := encoder.encode(tt.value)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, encoder.w.String())
		})
	}
}

func TestEncoder_encodeFloat64(t *testing.T) {
	tests := []struct {
		value    float64
		expected string
	}{
		{3.14159, "3.14159"},
		{math.NaN(), "null"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			encoder := newEncoder(false, 0)
			encoder.encode(tt.value)
			assert.Equal(t, tt.expected, encoder.w.String())
		})
	}
}

func TestEncoder_encodeBigInt(t *testing.T) {
	tests := []struct {
		value    *big.Int
		expected string
	}{
		{big.NewInt(123456789), "123456789"},
		{big.NewInt(-987654321), "-987654321"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			encoder := newEncoder(false, 0)
			err := encoder.encode(tt.value)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, encoder.w.String())
		})
	}
}

func TestEncoder_encodeString(t *testing.T) {
	tests := []struct {
		value    string
		expected string
	}{
		{"hello", "\"hello\""},
		{"\"escaped\"", "\"\\\"escaped\\\"\""},
		{"backslash\\test", "\"backslash\\\\test\""},
		{"new\nline", "\"new\\nline\""},
		{"tab\tcharacter", "\"tab\\tcharacter\""},
		{"2025-06-03T11:37:32Z", "\"2025-06-03T11:37:32Z\""},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			encoder := newEncoder(false, 0)
			err := encoder.encode(tt.value)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, encoder.w.String())
		})
	}
}

func TestEncoder_encodeArray(t *testing.T) {
	tests := []struct {
		value    []any
		expected string
	}{
		{[]any{1, 2, 3}, "[1,2,3]"},
		{[]any{"a", "b", "c"}, "[\"a\",\"b\",\"c\"]"},
		{[]any{true, false}, "[true,false]"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			encoder := newEncoder(false, 0)
			err := encoder.encode(tt.value)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, encoder.w.String())
		})
	}
}

func TestEncoder_encodeObject(t *testing.T) {
	tests := []struct {
		value    map[string]any
		expected string
	}{
		{
			map[string]any{"name": "John", "age": 30},
			"{\"age\":30,\"name\":\"John\"}",
		},
		{
			map[string]any{"active": true, "score": 100},
			"{\"active\":true,\"score\":100}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			encoder := newEncoder(false, 0)
			err := encoder.encode(tt.value)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, encoder.w.String())
		})
	}
}

func TestEncoder_flush(t *testing.T) {
	// Simula un writer di output
	var buf bytes.Buffer
	encoder := newEncoder(false, 0)
	encoder.out = &buf

	// Scrivi qualcosa
	err := encoder.encode("Hello")
	assert.NoError(t, err)

	// Fai il flush
	err = encoder.flush()
	assert.NoError(t, err)

	// Verifica che il buffer contenga i dati corretti
	assert.Equal(t, "\"Hello\"", buf.String())
}
