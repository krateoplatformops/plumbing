package request

import (
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseProxyURL(t *testing.T) {
	tests := []struct {
		name      string
		proxyURL  string
		expectErr bool
	}{
		{"Valid HTTP proxy", "http://proxy.example.com", false},
		{"Valid HTTPS proxy", "https://secure-proxy.example.com", false},
		{"Invalid scheme", "ftp://invalid-proxy.com", true},
		{"Malformed URL", ":://bad-url", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u, err := parseProxyURL(tt.proxyURL)
			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, u)
			}
		})
	}
}

func TestBasicAuthRoundTripper(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "Basic "+base64.StdEncoding.EncodeToString([]byte("user:pass")), r.Header.Get("Authorization"))
		w.WriteHeader(http.StatusOK)
	})

	mockRT := &mockRoundTripper{mux: mux}

	client := &basicAuthRoundTripper{
		username: "user",
		password: "pass",
		rt:       mockRT,
	}

	req, _ := http.NewRequest("GET", "https://httpbin.org", nil)
	_, err := client.RoundTrip(req)
	assert.NoError(t, err)
}

func TestBearerAuthRoundTripper(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "Bearer mytoken", r.Header.Get("Authorization"))
		w.WriteHeader(http.StatusOK)
	})

	mockRT := &mockRoundTripper{mux: mux}

	client := &bearerAuthRoundTripper{
		bearer: "mytoken",
		rt:     mockRT,
	}

	req, _ := http.NewRequest("GET", "http://example.com", nil)
	_, err := client.RoundTrip(req)
	assert.NoError(t, err)
}

// mockRoundTripper è una struttura che incapsula http.ServeMux e implementa http.RoundTripper.
type mockRoundTripper struct {
	mux *http.ServeMux
}

// RoundTrip esegue una risposta HTTP con il ServeMux.
func (m *mockRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	// Esegui il ServeMux per gestire la richiesta
	// Crea un ResponseRecorder per catturare la risposta
	rr := httptest.NewRecorder()
	m.mux.ServeHTTP(rr, req)

	return rr.Result(), nil
}
