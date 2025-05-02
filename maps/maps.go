package maps

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

func LeafPaths(m map[string]any, prefix string) []string {
	var paths []string

	for key, value := range m {
		newPath := key
		if prefix != "" {
			newPath = prefix + "." + key
		}

		switch v := value.(type) {
		case map[string]any:
			paths = append(paths, LeafPaths(v, newPath)...)
		case []any:
			for i, item := range v {
				itemPath := fmt.Sprintf("%s[%d]", newPath, i)
				if subMap, ok := item.(map[string]any); ok {
					paths = append(paths, LeafPaths(subMap, itemPath)...)
				} else {
					paths = append(paths, itemPath)
				}
			}
		default:
			paths = append(paths, newPath)
		}
	}

	return paths
}

// ParsePath converts a "spec.containers[0].env[0].value" path to a string slice
func ParsePath(path string) []string {
	modifiedPath := strings.ReplaceAll(path, "[", ".")
	modifiedPath = strings.ReplaceAll(modifiedPath, "]", "")
	return strings.Split(modifiedPath, ".")
}

func NestedValue(obj map[string]any, fields []string) (any, bool) {
	var current any = obj

	for _, field := range fields {
		switch typedCurrent := current.(type) {
		case map[string]any:
			current, _, _ = nestedFieldNoCopy(typedCurrent, field)
		case []any:
			index, err := strconv.Atoi(field)
			if err != nil || index < 0 || index >= len(typedCurrent) {
				return nil, false
			}
			current = typedCurrent[index]
		default:
			return nil, false
		}
	}

	return current, true
}

func SetNestedValue(obj map[string]any, fields []string, newValue any) error {
	lastIndex := len(fields) - 1

	var current any = obj
	var parent any = nil
	var parentKey string

	for i, field := range fields {
		switch typedCurrent := current.(type) {
		case map[string]any:
			// if last element then update its value
			if i == lastIndex {
				typedCurrent[field] = newValue
				return nil
			}
			parent = typedCurrent
			parentKey = field
			current = typedCurrent[field]

		case []any:
			index, err := strconv.Atoi(field)
			if err != nil || index < 0 || index >= len(typedCurrent) {
				return fmt.Errorf("invalid index: %s", field)
			}

			// if last element then update the array
			if i == lastIndex {
				typedCurrent[index] = newValue
				return nil
			}

			parent = typedCurrent
			parentKey = field
			current = typedCurrent[index]

		default:
			if i == lastIndex {
				// update value in its parent
				switch typedParent := parent.(type) {
				case map[string]any:
					typedParent[parentKey] = newValue
					return nil
				case []any:
					index, err := strconv.Atoi(parentKey)
					if err == nil && index >= 0 && index < len(typedParent) {
						typedParent[index] = newValue
						return nil
					}
				}
			}

			return fmt.Errorf("data structure not navigable at path: %s", strings.Join(fields, "."))
		}
	}

	return fmt.Errorf("unable to update path: %s", strings.Join(fields, "."))
}

// NestedMapNoCopy returns a map[string]any value of a nested field.
// Returns false if value is not found and an error if not a map[string]any.
func NestedMapNoCopy(obj map[string]any, fields ...string) (map[string]any, bool, error) {
	val, found, err := nestedFieldNoCopy(obj, fields...)
	if !found || err != nil {
		return nil, found, err
	}
	m, ok := val.(map[string]interface{})
	if !ok {
		return nil, false, fmt.Errorf("%v accessor error: %v is of the type %T, expected map[string]interface{}",
			strings.Join(fields, "."), val, val)
	}
	return m, true, nil
}

// NestedMap returns a deep copy of map[string]any value of a nested field.
// Returns false if value is not found and an error if not a map[string]any.
func NestedMap(obj map[string]any, fields ...string) (map[string]any, bool, error) {
	m, found, err := NestedMapNoCopy(obj, fields...)
	if !found || err != nil {
		return nil, found, err
	}
	return DeepCopyJSON(m), true, nil
}

func NestedSliceNoCopy(obj map[string]any, fields ...string) ([]any, bool, error) {
	val, found, err := nestedFieldNoCopy(obj, fields...)
	if !found || err != nil {
		return nil, found, err
	}
	_, ok := val.([]any)
	if !ok {
		return nil, false, fmt.Errorf("%v accessor error: %v is of the type %T, expected []any",
			strings.Join(fields, "."), val, val)
	}
	return val.([]any), true, nil
}

// NestedSlice returns a deep copy of []any value of a nested field.
// Returns false if value is not found and an error if not a []any.
func NestedSlice(obj map[string]any, fields ...string) ([]any, bool, error) {
	val, found, err := nestedFieldNoCopy(obj, fields...)
	if !found || err != nil {
		return nil, found, err
	}
	_, ok := val.([]any)
	if !ok {
		return nil, false, fmt.Errorf("%v accessor error: %v is of the type %T, expected []any",
			strings.Join(fields, "."), val, val)
	}
	return deepCopyJSONValue(val).([]any), true, nil
}

// SetNestedField sets the value of a nested field to a deep copy of the value provided.
// Returns an error if value cannot be set because one of the nesting levels is not a map[string]any.
func SetNestedField(obj map[string]any, value any, fields ...string) error {
	return setNestedFieldNoCopy(obj, deepCopyJSONValue(value), fields...)
}

func setNestedFieldNoCopy(obj map[string]any, value any, fields ...string) error {
	m := obj

	for i, field := range fields[:len(fields)-1] {
		if val, ok := m[field]; ok {
			if valMap, ok := val.(map[string]any); ok {
				m = valMap
			} else {
				return fmt.Errorf("value cannot be set because %v is not a map[string]any", strings.Join(fields[:i+1], "."))
			}
		} else {
			newVal := make(map[string]any)
			m[field] = newVal
			m = newVal
		}
	}
	m[fields[len(fields)-1]] = value
	return nil
}

// nestedFieldNoCopy returns a reference to a nested field.
// Returns false if value is not found and an error if unable
// to traverse obj.
//
// Note: fields passed to this function are treated as keys within the passed
// object; no array/slice syntax is supported.
func nestedFieldNoCopy(obj map[string]any, fields ...string) (any, bool, error) {
	var val any = obj

	for i, field := range fields {
		if val == nil {
			return nil, false, nil
		}
		if m, ok := val.(map[string]any); ok {
			val, ok = m[field]
			if !ok {
				return nil, false, nil
			}
		} else {
			return nil, false, fmt.Errorf("%v accessor error: %v is of the type %T, expected map[string]any",
				strings.Join(fields[:i+1], "."), val, val)
		}
	}
	return val, true, nil
}

// deepCopyJSONValue deep copies the passed value, assuming it is a valid JSON representation i.e. only contains
// types produced by json.Unmarshal() and also int64.
// bool, int64, float64, string, []any, map[string]any, json.Number and nil
func deepCopyJSONValue(x any) any {
	switch x := x.(type) {
	case map[string]any:
		if x == nil {
			// Typed nil - an any that contains a type map[string]any with a value of nil
			return x
		}
		clone := make(map[string]any, len(x))
		for k, v := range x {
			clone[k] = deepCopyJSONValue(v)
		}
		return clone
	case []any:
		if x == nil {
			// Typed nil - an any that contains a type []any with a value of nil
			return x
		}
		clone := make([]any, len(x))
		for i, v := range x {
			clone[i] = deepCopyJSONValue(v)
		}
		return clone
	case string, int64, bool, float64, nil, json.Number:
		return x
	default:
		panic(fmt.Errorf("cannot deep copy %T", x))
	}
}

// DeepCopyJSON deep copies the passed value, assuming it is a valid JSON representation i.e. only contains
// types produced by json.Unmarshal() and also int64.
// bool, int64, float64, string, []any, map[string]any, json.Number and nil
func DeepCopyJSON(x map[string]any) map[string]any {
	return deepCopyJSONValue(x).(map[string]any)
}
