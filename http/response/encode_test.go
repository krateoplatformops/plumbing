package response

import (
	"bytes"
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUnauthorized(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		expected   string
		statusCode int
	}{
		{
			name:       "Unauthorized error",
			err:        errors.New("unauthorized access"),
			expected:   `{"apiVersion":"v1", "code":401, "kind":"Status", "message":"unauthorized access", "reason":"Unauthorized", "status":"Failure"}`,
			statusCode: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Mock ResponseWriter
			var buf bytes.Buffer
			w := &mockResponseWriter{buf: &buf}

			// Call the Unauthorized function
			err := Unauthorized(w, tt.err)
			assert.NoError(t, err)

			// Check the response
			assert.Equal(t, tt.statusCode, w.statusCode)
			assert.JSONEq(t, tt.expected, buf.String())
		})
	}
}

func TestInternalError(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		expected   string
		statusCode int
	}{
		{
			name:       "Internal server error",
			err:        errors.New("internal server error"),
			expected:   `{"apiVersion":"v1", "code":500, "kind":"Status", "message":"internal server error", "reason":"InternalError", "status":"Failure"}`,
			statusCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Mock ResponseWriter
			var buf bytes.Buffer
			w := &mockResponseWriter{buf: &buf}

			// Call the InternalError function
			err := InternalError(w, tt.err)
			assert.NoError(t, err)

			// Check the response
			assert.Equal(t, tt.statusCode, w.statusCode)
			assert.JSONEq(t, tt.expected, buf.String())
		})
	}
}

func TestServiceUnavailable(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		expected   string
		statusCode int
	}{
		{
			name:       "Service unavailable error",
			err:        errors.New("service unavailable"),
			expected:   `{"apiVersion":"v1", "code":503, "kind":"Status", "message":"service unavailable", "reason":"ServiceUnavailable", "status":"Failure"}`,
			statusCode: http.StatusServiceUnavailable,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Mock ResponseWriter
			var buf bytes.Buffer
			w := &mockResponseWriter{buf: &buf}

			// Call the ServiceUnavailable function
			err := ServiceUnavailable(w, tt.err)
			assert.NoError(t, err)

			// Check the response
			assert.Equal(t, tt.statusCode, w.statusCode)
			assert.JSONEq(t, tt.expected, buf.String())
		})
	}
}

func TestBadRequest(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		expected   string
		statusCode int
	}{
		{
			name:       "Bad request error",
			err:        errors.New("bad request"),
			expected:   `{"apiVersion":"v1", "code":400, "kind":"Status", "message":"bad request", "reason":"BadRequest", "status":"Failure"}`,
			statusCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Mock ResponseWriter
			var buf bytes.Buffer
			w := &mockResponseWriter{buf: &buf}

			// Call the BadRequest function
			err := BadRequest(w, tt.err)
			assert.NoError(t, err)

			// Check the response
			assert.Equal(t, tt.statusCode, w.statusCode)
			assert.JSONEq(t, tt.expected, buf.String())
		})
	}
}

func TestNotAcceptable(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		expected   string
		statusCode int
	}{
		{
			name:       "Not acceptable error",
			err:        errors.New("not acceptable"),
			expected:   `{"apiVersion":"v1", "code":406, "kind":"Status", "message":"not acceptable", "reason":"NotAcceptable", "status":"Failure"}`,
			statusCode: http.StatusNotAcceptable,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Mock ResponseWriter
			var buf bytes.Buffer
			w := &mockResponseWriter{buf: &buf}

			// Call the NotAcceptable function
			err := NotAcceptable(w, tt.err)
			assert.NoError(t, err)

			// Check the response
			assert.Equal(t, tt.statusCode, w.statusCode)
			assert.JSONEq(t, tt.expected, buf.String())
		})
	}
}

func TestMethodNotAllowed(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		expected   string
		statusCode int
	}{
		{
			name:       "Method not allowed error",
			err:        errors.New("method not allowed"),
			expected:   `{"apiVersion":"v1", "code":405, "kind":"Status", "message":"method not allowed", "reason":"MethodNotAllowed", "status":"Failure"}`,
			statusCode: http.StatusMethodNotAllowed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Mock ResponseWriter
			var buf bytes.Buffer
			w := &mockResponseWriter{buf: &buf}

			// Call the MethodNotAllowed function
			err := MethodNotAllowed(w, tt.err)
			assert.NoError(t, err)

			// Check the response
			assert.Equal(t, tt.statusCode, w.statusCode)
			assert.JSONEq(t, tt.expected, buf.String())
		})
	}
}

func TestNotFound(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		expected   string
		statusCode int
	}{
		{
			name:       "Not found error",
			err:        errors.New("not found"),
			expected:   `{"apiVersion":"v1", "code":404, "kind":"Status", "message":"not found", "reason":"NotFound", "status":"Failure"}`,
			statusCode: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Mock ResponseWriter
			var buf bytes.Buffer
			w := &mockResponseWriter{buf: &buf}

			// Call the NotFound function
			err := NotFound(w, tt.err)
			assert.NoError(t, err)

			// Check the response
			assert.Equal(t, tt.statusCode, w.statusCode)
			assert.JSONEq(t, tt.expected, buf.String())
		})
	}
}

func TestForbidden(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		expected   string
		statusCode int
	}{
		{
			name:       "Forbidden error",
			err:        errors.New("forbidden"),
			expected:   `{"apiVersion":"v1", "code":403, "kind":"Status", "message":"forbidden", "reason":"Forbidden", "status":"Failure"}`,
			statusCode: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Mock ResponseWriter
			var buf bytes.Buffer
			w := &mockResponseWriter{buf: &buf}

			// Call the Forbidden function
			err := Forbidden(w, tt.err)
			assert.NoError(t, err)

			// Check the response
			assert.Equal(t, tt.statusCode, w.statusCode)
			assert.JSONEq(t, tt.expected, buf.String())
		})
	}
}

// Mock ResponseWriter
type mockResponseWriter struct {
	headers    http.Header
	statusCode int
	buf        *bytes.Buffer
}

func (m *mockResponseWriter) Header() http.Header {
	if m.headers == nil {
		m.headers = make(http.Header)
	}
	return m.headers
}

func (m *mockResponseWriter) WriteHeader(statusCode int) {
	m.statusCode = statusCode
}

func (m *mockResponseWriter) Write(p []byte) (n int, err error) {
	return m.buf.Write(p)
}
