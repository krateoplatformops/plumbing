package schemas

import (
	"testing"
)

func TestCollectAllDefinitions(t *testing.T) {
	src, err := FromJSONFile("../../testdata/git.spec.schema.json")
	if err != nil {
		t.Fatal(err)
	}

	defs := CollectAllDefinitions(src)

	// Controlli di base
	expectedDefs := []string{
		"Reference",
		"EnvSelector",
		"SecretKeySelector",
		"ConfigMapKeySelector",
	}

	for _, name := range expectedDefs {
		if _, ok := defs[name]; !ok {
			t.Errorf("definition %s not found in collected definitions", name)
		}
	}

	// Verifica che non ci siano definizioni extra
	if len(defs) != len(expectedDefs) {
		t.Errorf("expected %d definitions, got %d", len(expectedDefs), len(defs))
	}

}
