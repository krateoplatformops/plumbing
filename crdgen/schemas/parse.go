package schemas

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

func FromJSONFile(fileName string) (*Schema, error) {
	f, err := os.Open(fileName)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}

	defer func() {
		_ = f.Close()
	}()

	return FromJSONReader(f)
}

func FromJSONReader(r io.Reader) (*Schema, error) {
	var schema Schema
	if err := json.NewDecoder(r).Decode(&schema); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	return &schema, nil
}
