package maps

import (
	"encoding/json"
	"fmt"
)

func MapSliceToStructSlice[T any](in []map[string]any) ([]T, error) {
	var out []T

	for _, m := range in {
		bytes, err := json.Marshal(m)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal map: %w", err)
		}

		var elem T
		if err := json.Unmarshal(bytes, &elem); err != nil {
			return nil, fmt.Errorf("failed to unmarshal to %T: %w", elem, err)
		}

		out = append(out, elem)
	}

	return out, nil
}

func StructSliceToMapSlice[T any](results []T) ([]map[string]any, error) {
	var list []map[string]any

	for _, r := range results {
		bytes, err := json.Marshal(r)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal %T: %w", r, err)
		}

		var m map[string]any
		if err := json.Unmarshal(bytes, &m); err != nil {
			return nil, fmt.Errorf("failed to unmarshal %T into map: %w", r, err)
		}

		list = append(list, m)
	}

	return list, nil
}

func ToMapSlice(in []any) ([]map[string]any, error) {
	out := make([]map[string]any, len(in))

	for i, v := range in {
		m, ok := v.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("element at index %d has type %T, expected map[string]any", i, v)
		}
		out[i] = m
	}

	return out, nil
}
