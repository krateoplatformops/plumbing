package schemas

import "strings"

func AllOf(types []*Type, defs Definitions) (*Type, error) {
	resolved := resolveRefs(types, defs)

	typ, err := MergeTypes(resolved)
	if err != nil {
		return nil, err
	}
	typ.subSchemaType = SubSchemaTypeAllOf
	return typ, nil
}

// resolveRefs sostituisce i $ref con i Type reali presi dalle Definitions
func resolveRefs(types []*Type, defs Definitions) []*Type {
	out := make([]*Type, 0, len(types))
	for _, t := range types {
		if t == nil {
			continue
		}
		if t.Ref != "" {
			name := extractDefNameFromRef(t.Ref)
			if def, ok := defs[name]; ok {
				out = append(out, def)
				continue
			}
		}
		out = append(out, t)
	}
	return out
}

// estrae il nome dalla stringa tipo "#/$defs/Reference"
func extractDefNameFromRef(ref string) string {
	parts := strings.Split(ref, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return ref
}
