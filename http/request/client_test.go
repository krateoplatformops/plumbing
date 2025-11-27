package request

import (
	"testing"

	"github.com/krateoplatformops/plumbing/endpoints"
)

func TestHTTPClientForEndpoint(t *testing.T) {
	tests := []struct {
		name      string
		opts      RequestOptions
		expectErr bool
	}{
		{
			name:      "valid endpoint without auth",
			opts:      RequestOptions{Endpoint: &endpoints.Endpoint{}},
			expectErr: false,
		},
		{
			name:      "valid endpoint with bearer token",
			opts:      RequestOptions{Endpoint: &endpoints.Endpoint{Token: "test-token"}},
			expectErr: false,
		},
		{
			name:      "valid endpoint with basic auth",
			opts:      RequestOptions{Endpoint: &endpoints.Endpoint{Username: "user", Password: "pass"}},
			expectErr: false,
		},
		{
			name:      "invalid endpoint with both auth methods",
			opts:      RequestOptions{Endpoint: &endpoints.Endpoint{Username: "user", Password: "pass", Token: "token"}},
			expectErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			client, err := HTTPClientForEndpoint(tc.opts.Endpoint, nil)
			if (err != nil) != tc.expectErr {
				t.Errorf("unexpected error status: got %v, expectErr %v", err, tc.expectErr)
			}
			if !tc.expectErr && client == nil {
				t.Errorf("expected client, got nil")
			}
		})
	}
}
