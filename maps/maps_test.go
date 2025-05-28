package maps

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNestedFieldNoCopy(t *testing.T) {
	target := map[string]any{"foo": "bar"}

	obj := map[string]any{
		"a": map[string]any{
			"b": target,
			"c": nil,
			"d": []any{"foo"},
			"e": []any{
				map[string]any{
					"f": "bar",
				},
			},
		},
	}

	// case 1: field exists and is non-nil
	res, exists, err := nestedFieldNoCopy(obj, "a", "b")
	assert.True(t, exists)
	assert.NoError(t, err)
	assert.Equal(t, target, res)
	target["foo"] = "baz"
	assert.Equal(t, target["foo"], res.(map[string]any)["foo"], "result should be a reference to the expected item")

	// case 2: field exists and is nil
	res, exists, err = nestedFieldNoCopy(obj, "a", "c")
	assert.True(t, exists)
	assert.NoError(t, err)
	assert.Nil(t, res)

	// case 3: error traversing obj
	res, exists, err = nestedFieldNoCopy(obj, "a", "d", "foo")
	assert.False(t, exists)
	assert.Error(t, err)
	assert.Nil(t, res)

	// case 4: field does not exist
	res, exists, err = nestedFieldNoCopy(obj, "a", "g")
	assert.False(t, exists)
	assert.NoError(t, err)
	assert.Nil(t, res)

	// case 5: intermediate field does not exist
	res, exists, err = nestedFieldNoCopy(obj, "a", "g", "f")
	assert.False(t, exists)
	assert.NoError(t, err)
	assert.Nil(t, res)

	// case 6: intermediate field is null
	//         (background: happens easily in YAML)
	res, exists, err = nestedFieldNoCopy(obj, "a", "c", "f")
	assert.False(t, exists)
	assert.NoError(t, err)
	assert.Nil(t, res)

	// case 7: array/slice syntax is not supported
	res, exists, err = nestedFieldNoCopy(obj, "a", "e[0]")
	assert.False(t, exists)
	assert.NoError(t, err)
	assert.Nil(t, res)
}

func TestGetNestedValue(t *testing.T) {
	data := map[string]any{
		"apiVersion": "v1",
		"kind":       "Pod",
		"metadata": map[string]any{
			"annotations": map[string]string{
				"krateo.io/verbose": "false",
			},
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
						map[string]any{"name": "ENV_VAR", "value": "${JQ_EXPRESSION}"},
					},
				},
				map[string]any{
					"name":  "nginx2",
					"image": "nginx:latest",
					"env": []any{
						map[string]any{"name": "ENV_VAR_2", "value": "${JQ_EXPRESSION_2}"},
					},
				},
			},
		},
	}

	paths := LeafPaths(data, "")
	sort.Strings(paths)

	for _, path := range paths {
		fields := ParsePath(path)

		if value, found := NestedValue(data, fields); found {
			fmt.Printf("Path: %s, Value: %v\n", path, value)

			// if the value is string, we can try to evaluate a JQ expression
			//if strValue, ok := value.(string) {
			//	fmt.Printf("evaluate JQ: %s\n", strValue)
			//}
		}
	}
}

func TestSetNestedValue(t *testing.T) {
	data := map[string]any{
		"apiVersion": "v1",
		"kind":       "Pod",
		"metadata": map[string]any{
			"name": "mypod",
			"labels": map[string]any{
				"app":     "myapp",
				"version": 2,
				"enabled": true,
			},
		},
		"spec": map[string]any{
			"replicas": 3,
			"containers": []any{
				map[string]any{
					"name":  "nginx",
					"image": "nginx:latest",
					"env": []any{
						map[string]any{"name": "ENV_VAR", "value": "${JQ_EXPRESSION}"},
					},
				},
			},
		},
	}

	updates := map[string]any{
		"metadata.name":                   "new-pod",
		"metadata.labels.version":         3,
		"spec.replicas":                   5,
		"spec.containers[0].image":        "nginx:1.23",
		"spec.containers[0].env[0].value": "resolved_value",
	}

	for path, newValue := range updates {
		fields := ParsePath(path)
		err := SetNestedValue(data, fields, newValue)
		if err != nil {
			t.Fatalf("failed to update %s: %v", path, err)
		}
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	enc.Encode(data)
}
