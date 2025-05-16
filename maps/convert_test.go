package maps

import (
	"reflect"
	"testing"
)

func TestToMapSlice(t *testing.T) {
	tests := []struct {
		name    string
		input   []any
		want    []map[string]any
		wantErr bool
	}{
		{
			name: "valid input",
			input: []any{
				map[string]any{"foo": "bar"},
				map[string]any{"baz": 123},
			},
			want: []map[string]any{
				{"foo": "bar"},
				{"baz": 123},
			},
			wantErr: false,
		},
		{
			name:    "invalid input type",
			input:   []any{"not-a-map"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ToMapSlice(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("toMapSlice() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("toMapSlice() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFromUnstructuredMapSlice(t *testing.T) {
	tests := []struct {
		name    string
		input   []map[string]any
		want    []example
		wantErr bool
	}{
		{
			name: "valid input",
			input: []map[string]any{
				{"name": "Bob", "age": 42},
			},
			want:    []example{{Name: "Bob", Age: 42}},
			wantErr: false,
		},
		{
			name: "invalid field type",
			input: []map[string]any{
				{"name": "Charlie", "age": "not-a-number"},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := MapSliceToStructSlice[example](tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("fromUnstructuredMapSlice() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("fromUnstructuredMapSlice() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestToUnstructuredMapSlice(t *testing.T) {
	tests := []struct {
		name    string
		input   []example
		want    []map[string]any
		wantErr bool
	}{
		{
			name:  "valid input",
			input: []example{{Name: "Alice", Age: 30}},
			want: []map[string]any{
				{"name": "Alice", "age": float64(30)},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := StructSliceToMapSlice(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("toUnstructuredMapSlice() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("toUnstructuredMapSlice() = %v, want %v", got, tt.want)
			}
		})
	}
}

type example struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}
