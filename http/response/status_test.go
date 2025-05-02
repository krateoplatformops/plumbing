package response

import (
	"reflect"
	"testing"
)

func TestAsMap(t *testing.T) {
	tests := []struct {
		name   string
		input  Status
		expect map[string]any
	}{
		{
			name: "Complete struct",
			input: Status{
				Kind:       "Example",
				APIVersion: "v1",
				Status:     "Success",
				Message:    "Operation completed successfully",
				Reason:     "None",
				Code:       200,
			},
			expect: map[string]any{
				"status":  "Success",
				"message": "Operation completed successfully",
				"reason":  "None",
				"code":    float64(200),
			},
		},
		{
			name:   "Empty struct",
			input:  Status{},
			expect: map[string]any{},
		},
		{
			name: "Partial struct",
			input: Status{
				Status: "Failure",
				Code:   500,
			},
			expect: map[string]any{
				"status": "Failure",
				"code":   float64(500),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := AsMap(&tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			// Confrontiamo le mappe
			if !reflect.DeepEqual(result, tt.expect) {
				t.Errorf("expected: %v, got: %v", tt.expect, result)
			}
		})
	}
}
