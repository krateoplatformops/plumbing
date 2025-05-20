package maps

import (
	"encoding/json"
	"fmt"
)

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
	case []map[string]any:
		if x == nil {
			return x
		}
		clone := make([]any, len(x))
		for i, v := range x {
			clone[i] = deepCopyJSONValue(v)
		}
		return clone
	case string, int64, bool, float64, nil, json.Number:
		return x
	case int:
		return int64(x)
	case int32:
		return int64(x)
	case float32:
		return float64(x)
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
