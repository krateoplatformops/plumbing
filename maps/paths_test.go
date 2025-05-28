package maps

import (
	"reflect"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLeafPaths(t *testing.T) {
	data := map[string]any{
		"apiVersion": "v1",
		"kind":       "Pod",
		"metadata": map[string]any{
			"name": "mypod",
			"labels": map[string]any{
				"app": "myapp",
			},
		},
		"spec": map[string]any{
			"containers": []any{
				map[string]any{
					"name":  "nginx",
					"image": "nginx:latest",
					"env": []any{
						map[string]any{"name": "ENV_VAR", "value": "$(JQ_EXPRESSION)"},
					},
				},
				map[string]any{
					"name":  "nginx2",
					"image": "nginx:latest",
					"env": []any{
						map[string]any{"name": "ENV_VAR_2", "value": "$(JQ_EXPRESSION)"},
					},
				},
			},
		},
	}

	paths := LeafPaths(data, "")
	sort.Strings(paths)

	assert.EqualValues(t, paths, []string{
		"apiVersion",
		"kind",
		"metadata.labels.app",
		"metadata.name",
		"spec.containers[0].env[0].name",
		"spec.containers[0].env[0].value",
		"spec.containers[0].image",
		"spec.containers[0].name",
		"spec.containers[1].env[0].name",
		"spec.containers[1].env[0].value",
		"spec.containers[1].image",
		"spec.containers[1].name",
	})
}

func TestParsePath(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "simple path",
			input:    "metadata.name",
			expected: []string{"metadata", "name"},
		},
		{
			name:     "path with one index",
			input:    "spec.containers[0].name",
			expected: []string{"spec", "containers", "0", "name"},
		},
		{
			name:     "path with multiple indices",
			input:    "spec.containers[0].env[1].value",
			expected: []string{"spec", "containers", "0", "env", "1", "value"},
		},
		{
			name:     "nested maps without arrays",
			input:    "status.conditions.ready",
			expected: []string{"status", "conditions", "ready"},
		},
		{
			name:     "only array indices",
			input:    "[0][1][2]",
			expected: []string{"0", "1", "2"},
		},
		{
			name:     "empty string",
			input:    "",
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParsePath(tt.input)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("ParsePath(%q) = %v; want %v", tt.input, result, tt.expected)
			}
		})
	}
}
