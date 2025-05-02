package endpoints

import "testing"

func TestEndpoint(t *testing.T) {
	tests := []struct {
		name     string
		endpoint Endpoint
		hasCA    bool
		hasBasic bool
		hasToken bool
		hasCert  bool
	}{
		{
			name:     "Empty Endpoint",
			endpoint: Endpoint{},
			hasCA:    false, hasBasic: false, hasToken: false, hasCert: false,
		},
		{
			name:     "Has CA Data",
			endpoint: Endpoint{CertificateAuthorityData: "cert-data"},
			hasCA:    true, hasBasic: false, hasToken: false, hasCert: false,
		},
		{
			name:     "Has Basic Auth",
			endpoint: Endpoint{Username: "user", Password: "pass"},
			hasCA:    false, hasBasic: true, hasToken: false, hasCert: false,
		},
		{
			name:     "Has Token Auth",
			endpoint: Endpoint{Token: "my-token"},
			hasCA:    false, hasBasic: false, hasToken: true, hasCert: false,
		},
		{
			name:     "Has Certificate Auth",
			endpoint: Endpoint{ClientCertificateData: "cert", ClientKeyData: "key"},
			hasCA:    false, hasBasic: false, hasToken: false, hasCert: true,
		},
		{
			name: "Has All Auth Methods",
			endpoint: Endpoint{
				CertificateAuthorityData: "cert-data",
				Username:                 "user",
				Password:                 "pass",
				Token:                    "my-token",
				ClientCertificateData:    "cert",
				ClientKeyData:            "key",
			},
			hasCA: true, hasBasic: true, hasToken: true, hasCert: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.endpoint.HasCA(); got != tt.hasCA {
				t.Errorf("HasCA() = %v, want %v", got, tt.hasCA)
			}
			if got := tt.endpoint.HasBasicAuth(); got != tt.hasBasic {
				t.Errorf("HasBasicAuth() = %v, want %v", got, tt.hasBasic)
			}
			if got := tt.endpoint.HasTokenAuth(); got != tt.hasToken {
				t.Errorf("HasTokenAuth() = %v, want %v", got, tt.hasToken)
			}
			if got := tt.endpoint.HasCertAuth(); got != tt.hasCert {
				t.Errorf("HasCertAuth() = %v, want %v", got, tt.hasCert)
			}
		})
	}
}
