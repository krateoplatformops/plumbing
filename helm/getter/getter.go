package getter

import (
	"bytes"
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const MaxResponseSize = 100 * 1024 * 1024 // 100MB

// Getter is an interface to support GET to the specified URI.
type Getter interface {
	// Get file content by url string. It returns the content as a byte slice, the absolute URI (repo+chart+version), and error if any.
	Get(ctx context.Context, opts GetOptions) (io.Reader, string, error)
}

func Get(ctx context.Context, uri string, opts ...Option) (io.Reader, string, error) {
	o := GetOptions{
		URI:     uri,
		Timeout: 60 * time.Second,
	}
	for _, opt := range opts {
		opt(&o)
	}

	// Check cache first
	if o.Cache != nil {
		if data, ok := o.Cache.Get(o.URI, o.Version); ok {
			return data, o.Version, nil
		}
	}

	if o.URI == "" {
		return nil, "", errors.New("URI is required")
	}
	var err error
	var g Getter
	if isOCI(o.URI) {
		g = &ociGetter{}
	} else if isTGZ(o.URI) {
		g = &tgzGetter{}
	} else if isHTTP(o.URI) {
		g = &repoGetter{}
	} else {
		return nil, "", fmt.Errorf("%w: uri '%s'", ErrNoHandler, o.URI)
	}

	b, uri, err := g.Get(ctx, o)
	if err != nil {
		return nil, "", fmt.Errorf("failed to get %s: %w", o.URI, err)
	}
	// Store in cache
	if o.Cache != nil {
		err = o.Cache.Set(o.URI, o.Version, b)
		if err != nil {
			return nil, "", fmt.Errorf("failed to cache %s: %w", o.URI, err)
		}
		// Reset the pointer to the beginning so the caller can read it
		if seeker, ok := b.(io.Seeker); ok {
			seeker.Seek(0, io.SeekStart)
		}
	}

	return b, uri, nil
}

func fetch(ctx context.Context, opts GetOptions) (io.Reader, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, opts.URI, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create http request for uri %s: %w", opts.URI, err)
	}

	// Always set credentials on initial request
	if opts.Username != "" && opts.Password != "" {
		req.SetBasicAuth(opts.Username, opts.Password)
	}

	client := newHTTPClient(opts)

	// Strip credentials on cross-domain redirects unless PassCredentialsAll is true
	if !opts.PassCredentialsAll {
		client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			if len(via) > 0 && req.URL.Host != via[0].URL.Host {
				req.Header.Del("Authorization")
			}
			return nil
		}
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch %s: %w", opts.URI, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch %s : %s", opts.URI, resp.Status)
	}

	size := int64(512)
	if resp.ContentLength > 0 {
		size = resp.ContentLength
	}

	data := make([]byte, 0, size)
	buf := bytes.NewBuffer(data)

	// Limit the reading to avoid memory exhaustion attacks
	limitedReader := io.LimitReader(resp.Body, MaxResponseSize)

	_, err = buf.ReadFrom(limitedReader)
	if err != nil {
		return nil, fmt.Errorf("failed to read body: %w", err)
	}

	// Convert to *bytes.Reader to allow Seeking (ReadAt, Seek)
	return bytes.NewReader(buf.Bytes()), nil
}

func newHTTPClient(opts GetOptions) *http.Client {
	transport := &http.Transport{
		DisableCompression: true,
		Proxy:              http.ProxyFromEnvironment,
	}

	if opts.InsecureSkipVerifyTLS {
		transport.TLSClientConfig = &tls.Config{
			InsecureSkipVerify: true,
		}
	}

	return &http.Client{
		Transport: transport,
		Timeout:   opts.Timeout,
	}
}
func isOCI(url string) bool {
	return strings.HasPrefix(url, "oci://")
}
