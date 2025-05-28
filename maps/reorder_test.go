package maps

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestReorder(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]any
		paths    []string
		expected map[string]any
	}{
		{
			name: "simple map reorder",
			input: map[string]any{
				"user": map[string]any{
					"name": "Alice",
					"age":  30,
				},
				"meta": map[string]any{
					"timestamp": "now",
				},
			},
			paths: []string{
				"meta.timestamp",
				"user.age",
				"user.name",
			},
			expected: map[string]any{
				"meta": map[string]any{
					"timestamp": "now",
				},
				"user": map[string]any{
					"age":  30,
					"name": "Alice",
				},
			},
		},
		{
			name: "ignores missing path",
			input: map[string]any{
				"user": map[string]any{
					"name": "Alice",
				},
			},
			paths: []string{
				"user.name",
				"user.age", // non esiste
			},
			expected: map[string]any{
				"user": map[string]any{
					"name": "Alice",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Reorder(tt.input, tt.paths)

			// Marshal/unmarshal to normalize nil vs omitted
			resultJSON, _ := json.Marshal(result)
			expectedJSON, _ := json.Marshal(tt.expected)

			var normalizedResult, normalizedExpected map[string]any
			_ = json.Unmarshal(resultJSON, &normalizedResult)
			_ = json.Unmarshal(expectedJSON, &normalizedExpected)

			if !reflect.DeepEqual(normalizedResult, normalizedExpected) {
				t.Errorf("got:\n%v\nwant:\n%v", string(resultJSON), string(expectedJSON))
			}
		})
	}
}
