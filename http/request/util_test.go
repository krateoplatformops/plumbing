package request

import (
	"net/http"
	"testing"
)

func TestCloneRequest(t *testing.T) {
	orig := &http.Request{
		Header: http.Header{
			"Test-Header": []string{"value1", "value2"},
		},
	}
	clone := cloneRequest(orig)

	if &orig == &clone {
		t.Errorf("CloneRequest did not create a new request instance")
	}
	if &orig.Header == &clone.Header {
		t.Errorf("CloneRequest did not create a deep copy of headers")
	}
	if len(clone.Header.Get("Test-Header")) == 0 {
		t.Errorf("CloneRequest did not copy headers correctly")
	}
}

func TestCloneHeader(t *testing.T) {
	orig := http.Header{
		"Test-Header": []string{"value1", "value2"},
	}
	clone := cloneHeader(orig)

	if orig.Get("Test-Header") != clone.Get("Test-Header") {
		t.Errorf("cloneHeader did not create a deep copy of header values")
	}
}

func TestIsTextResponse(t *testing.T) {
	tests := []struct {
		name        string
		contentType string
		expectText  bool
	}{
		{"empty content type", "", true},
		{"text/plain", "text/plain", true},
		{"text/html", "text/html", true},
		{"application/json", "application/json", false},
		{"invalid content type", "invalid/type", false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			resp := &http.Response{Header: http.Header{"Content-Type": []string{tc.contentType}}}
			if isTextResponse(resp) != tc.expectText {
				t.Errorf("unexpected result for %s: got %v, want %v", tc.contentType, !tc.expectText, tc.expectText)
			}
		})
	}
}
