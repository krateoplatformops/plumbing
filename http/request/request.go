package request

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	xcontext "github.com/krateoplatformops/plumbing/context"
	"github.com/krateoplatformops/plumbing/endpoints"
	"github.com/krateoplatformops/plumbing/http/response"
	"github.com/krateoplatformops/plumbing/http/util"
	"github.com/krateoplatformops/plumbing/ptr"
)

const maxUnstructuredResponseTextBytes = 2048

type RequestOptions struct {
	Path            string
	Verb            *string
	Headers         []string
	Payload         *string
	Endpoint        *endpoints.Endpoint
	ResponseHandler func(io.ReadCloser) error
	ErrorKey        string
	ContinueOnError bool
}

func Do(ctx context.Context, opts RequestOptions) *response.Status {
	uri := strings.TrimSuffix(opts.Endpoint.ServerURL, "/")
	if len(opts.Path) > 0 {
		uri = fmt.Sprintf("%s/%s", uri, strings.TrimPrefix(opts.Path, "/"))
	}

	u, err := url.Parse(uri)
	if err != nil {
		return response.New(http.StatusInternalServerError, err)
	}

	verb := ptr.Deref(opts.Verb, http.MethodGet)

	var body io.Reader
	if s := ptr.Deref(opts.Payload, ""); len(s) > 0 {
		body = strings.NewReader(s)
	}

	call, err := http.NewRequestWithContext(ctx, verb, u.String(), body)
	if err != nil {
		return response.New(http.StatusInternalServerError, err)
	}
	// Additional headers for AWS Signature 4 algorithm
	if opts.Endpoint.HasAwsAuth() {
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
		url, _ := url.Parse(opts.Endpoint.ServerURL)
		host := url.Host
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
		opts.Headers = append(opts.Headers, strings.ToLower(xcontext.LabelKrateoTraceId)+":"+xcontext.TraceId(ctx, true))
		sort.Strings(opts.Headers)
	}

	if len(opts.Headers) > 0 {
		for _, el := range opts.Headers {
			idx := strings.Index(el, ":")
			if idx <= 0 {
				continue
			}
			key := el[:idx]
			val := strings.TrimSpace(el[idx+1:])
			call.Header.Set(key, val)
		}
	}

	cli, err := HTTPClientForEndpoint(opts)
	if err != nil {
		return response.New(http.StatusInternalServerError,
			fmt.Errorf("unable to create HTTP Client for endpoint: %w", err))
	}

	// Wrap the existing client in a RetryClient
	retryCli := util.NewRetryClient(cli)

	// Use RetryClient instead of the raw client  cli.Do(call)
	respo, err := retryCli.Do(call)
	if err != nil {
		return response.New(http.StatusInternalServerError, err)
	}
	defer respo.Body.Close()

	statusOK := respo.StatusCode >= 200 && respo.StatusCode < 300
	if !statusOK {
		dat, err := io.ReadAll(io.LimitReader(respo.Body, maxUnstructuredResponseTextBytes))
		if err != nil {
			return response.New(http.StatusInternalServerError, err)
		}

		res := &response.Status{}
		if err := json.Unmarshal(dat, res); err != nil {
			res = response.New(respo.StatusCode, fmt.Errorf("%s", string(dat)))
			return res
		}

		return res
	}

	if ct := respo.Header.Get("Content-Type"); !strings.Contains(ct, "json") {
		return response.New(http.StatusNotAcceptable, fmt.Errorf("content type %q is not allowed", ct))
	}

	if opts.ResponseHandler != nil {
		if err := opts.ResponseHandler(respo.Body); err != nil {
			return response.New(http.StatusInternalServerError, err)
		}
		return response.New(http.StatusOK, nil)
	}

	return response.New(http.StatusNoContent, nil)
}
