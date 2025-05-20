package maps

import (
	"fmt"
	"strings"
)

func NestedString(obj map[string]any, fields ...string) (string, error) {
	val, found := NestedValue(obj, fields)
	if !found {
		return "", nil
	}
	s, ok := val.(string)
	if !ok {
		return "", fmt.Errorf("%v access error: %v is of the type %T, expected string",
			strings.Join(fields, "."), val, val)
	}
	return s, nil
}
