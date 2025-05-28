package maps

import (
	"fmt"
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

	parts := strings.Split(modifiedPath, ".")

	all := make([]string, 0, len(parts))
	for _, el := range parts {
		if el != "" {
			all = append(all, el)
		}
	}
	return all
}
