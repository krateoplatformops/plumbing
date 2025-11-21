package schemas

func CollectAllDefinitions(s *Schema) map[string]*Type {
	if s == nil {
		return nil
	}

	result := make(map[string]*Type)

	// 1️⃣ Aggiungi le definitions dello schema stesso
	for name, def := range s.Definitions {
		if def != nil {
			result[name] = def
		}
	}

	// 2️⃣ Ricorsione sulle proprietà
	if s.ObjectAsType != nil {
		collectFromType((*Type)(s.ObjectAsType), result)
	}

	return result
}

// Funzione di supporto ricorsiva
func collectFromType(t *Type, out map[string]*Type) {
	if t == nil {
		return
	}

	// Prima risolviamo gli allOf, così t diventa già mergiato
	if len(t.AllOf) > 0 {
		merged, err := AllOf(t.AllOf, out) // `out` contiene tutte le defs
		if err == nil {
			t = merged
		}
	}

	// Proprietà ricorsive
	for _, p := range t.Properties {
		collectFromType(p, out)
	}
	for _, p := range t.PatternProperties {
		collectFromType(p, out)
	}

	if t.Items != nil {
		collectFromType(t.Items, out)
	}
	if t.AdditionalItems != nil {
		collectFromType(t.AdditionalItems, out)
	}
	for _, sub := range t.AllOf {
		collectFromType(sub, out)
	}
	for _, sub := range t.AnyOf {
		collectFromType(sub, out)
	}
	for _, sub := range t.OneOf {
		collectFromType(sub, out)
	}
	if t.Not != nil {
		collectFromType(t.Not, out)
	}

	// Aggiungi le definitions
	for name, def := range t.Definitions {
		if def != nil {
			// Se dentro la definition c’è un allOf, risolvilo prima di aggiungerlo
			if len(t.AllOf) > 0 {
				merged, err := AllOf(t.AllOf, out) // `out` contiene tutte le defs
				if err == nil {
					def = merged
				}
			}
			out[name] = def
			collectFromType(def, out) // continua ricorsione
		}
	}
}
