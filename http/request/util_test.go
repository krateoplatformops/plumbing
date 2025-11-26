package request

import (
	"crypto/sha256"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/krateoplatformops/plumbing/endpoints"
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

func TestComputeAwsHeader(t *testing.T) {
	// Data from documentation example
	// https://docs.aws.amazon.com/AmazonS3/latest/API/sig-v4-header-based-auth.html
	payload := ""
	verb := "GET"

	opts := RequestOptions{
		Endpoint: &endpoints.Endpoint{
			ServerURL:    "examplebucket.s3.amazonaws.com",
			AwsAccessKey: "AKIAIOSFODNN7EXAMPLE",
			AwsSecretKey: "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
			AwsRegion:    "us-east-1",
			AwsService:   "s3",
			AwsTime:      "20130524",
		},
		Headers: []string{
			"range: bytes=0-9",
		},
		Path:    "/test.txt",
		Payload: &payload,
		Verb:    &verb,
	}

	var amzDate string
	if len(opts.Endpoint.AwsTime) != 0 {
		// If AwsTime is just YYYYMMDD, construct full timestamp
		if len(opts.Endpoint.AwsTime) == 8 {
			amzDate = opts.Endpoint.AwsTime + "T000000Z"
		} else {
			amzDate = opts.Endpoint.AwsTime
		}
	} else {
		t := time.Now().UTC()
		amzDate = t.Format("20060102T150405Z")
	}
	host := opts.Endpoint.ServerURL
	if host == "" {
		host = "localhost"
	}

	canonicalURI := opts.Path
	if canonicalURI == "" {
		canonicalURI = "/"
	}

	// Step 2: Build headers
	// Empty payload hash (SHA256 of empty string): "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
	payloadHash := fmt.Sprintf("%x", sha256.Sum256([]byte(*opts.Payload)))

	headers := []string{
		"host:" + host,
		"x-amz-content-sha256:" + payloadHash,
		"x-amz-date:" + amzDate,
	}

	opts.Headers = append(opts.Headers, headers...)

	expected := "AWS4-HMAC-SHA256 Credential=AKIAIOSFODNN7EXAMPLE/20130524/us-east-1/s3/aws4_request,SignedHeaders=host;range;x-amz-content-sha256;x-amz-date,Signature=f0e8bdb87c964420e857bd35b5d6ed310bd44f0170aba48dd91039c6036bdb41"

	actual := ComputeAwsHeader(opts)
	if actual != expected {
		t.Errorf("unexpected result in AWS signature computation: expected %s\ngot %s", expected, actual)
	}

}
