package maps

type pathSet map[string]struct{}

func (ps pathSet) add(path string) {
	ps[path] = struct{}{}
}

func (ps pathSet) contains(path string) bool {
	_, ok := ps[path]
	return ok
}

func Reorder(input map[string]any, paths []string) map[string]any {
	output := map[string]any{}
	included := pathSet{}

	// Step 1: add the ordered fields
	for _, p := range paths {
		parsed := ParsePath(p)

		val, found, err := NestedFieldCopy(input, parsed...)
		if err != nil || !found {
			continue
		}
		_ = SetNestedField(output, val, parsed...)
		included.add(p)
	}

	// Step 2: add everithing else
	allPaths := LeafPaths(input, "")
	for _, fullPath := range allPaths {
		if included.contains(fullPath) {
			continue
		}
		parsed := ParsePath(fullPath)
		val, found, err := NestedFieldCopy(input, parsed...)
		if err != nil || !found {
			continue
		}
		_ = SetNestedField(output, val, parsed...)
	}

	return output
}
