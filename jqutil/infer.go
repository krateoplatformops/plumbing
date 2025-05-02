package jqutil

import (
	"encoding/json"
	"math"
	"strconv"
	"strings"
)

// InferType attempts to infer and convert a string value to its most appropriate Go type.
// It supports primitive types (bool, int32, int64, float64, string), as well as
// structured types commonly found in Kubernetes configurations (map[string]any and []any).
// The function first tries to parse the input as JSON. If that fails, it falls back to
// custom parsing logic for booleans, nil/null, integers, and floats.
// If no conversion is possible, the original string is returned.
func InferType(value string) any {
	value = strings.TrimSpace(value)
	if value == "" {
		return value
	}

	decoder := json.NewDecoder(strings.NewReader(value))
	decoder.UseNumber()

	var jsonVal any
	if err := decoder.Decode(&jsonVal); err == nil {
		switch v := jsonVal.(type) {
		case json.Number:
			if i, err := v.Int64(); err == nil {
				if i >= math.MinInt32 && i <= math.MaxInt32 {
					return int32(i)
				}
				return i
			}
			if f, err := v.Float64(); err == nil {
				return f
			}
		default:
			return jsonVal
		}
	}

	if strings.EqualFold(value, "true") {
		return true
	}
	if strings.EqualFold(value, "false") {
		return false
	}

	if strings.EqualFold(value, "null") || strings.EqualFold(value, "nil") {
		return nil
	}

	if intVal, err := strconv.ParseInt(value, 10, 64); err == nil {
		if intVal >= math.MinInt32 && intVal <= math.MaxInt32 {
			return int32(intVal)
		}
		return intVal
	}

	if floatVal, err := strconv.ParseFloat(value, 64); err == nil {
		if floatVal == math.Trunc(floatVal) {
			if floatVal >= math.MinInt64 && floatVal <= math.MaxInt64 {
				return int64(floatVal)
			}
		}
		return floatVal
	}

	return value
}
