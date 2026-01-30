package getter

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"oras.land/oras-go/v2/content"
	"oras.land/oras-go/v2/registry/remote"
	"oras.land/oras-go/v2/registry/remote/auth"
	"oras.land/oras-go/v2/registry/remote/retry"
)

var _ Getter = (*ociGetter)(nil)

// Constants defined by Helm OCI support
const (
	ChartLayerMediaType  = "application/vnd.cncf.helm.chart.content.v1.tar+gzip"
	LegacyLayerMediaType = "application/tar+gzip" // Fallback for legacy registries
)

// Shared Transport to enable Connection Pooling
var sharedTransport = &http.Transport{
	Proxy: http.ProxyFromEnvironment,
	DialContext: (&net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
	}).DialContext,
	ForceAttemptHTTP2:     true,
	MaxIdleConns:          100,
	IdleConnTimeout:       90 * time.Second,
	TLSHandshakeTimeout:   10 * time.Second,
	ExpectContinueTimeout: 1 * time.Second,
}

type ociGetter struct{}

func NewOCIGetter() Getter {
	return &ociGetter{}
}

func (g *ociGetter) Get(ctx context.Context, opts GetOptions) (io.Reader, string, error) {
	if !isOCI(opts.URI) {
		return nil, "", fmt.Errorf("uri '%s' is not a valid OCI ref", opts.URI)
	}

	// 1. Prepare URI (remove oci:// prefix)
	refString := strings.TrimPrefix(opts.URI, "oci://")
	if opts.Repo != "" {
		refString = fmt.Sprintf("%s/%s", refString, opts.Repo)
	}

	// Add tag/version if missing and not using digest
	// Check for tag by looking for : after the last /
	if opts.Version != "" && !strings.Contains(refString, "@") {
		lastSlash := strings.LastIndex(refString, "/")
		afterLastSlash := refString
		if lastSlash != -1 {
			afterLastSlash = refString[lastSlash:]
		}
		if !strings.Contains(afterLastSlash, ":") {
			refString = fmt.Sprintf("%s:%s", refString, opts.Version)
		}
	}

	// 2. Create Remote Repository (ORAS v2)
	repo, err := remote.NewRepository(refString)
	if err != nil {
		return nil, "", fmt.Errorf("invalid repository reference: %w", err)
	}

	// 3. Configure HTTP Client and Auth
	// Create an auth client using the shared transport for efficiency
	authClient := &auth.Client{
		Client: &http.Client{
			// Select secure or insecure transport based on options
			Transport: getTransport(opts.InsecureSkipVerifyTLS),
			Timeout:   opts.Timeout,
		},
		Cache: auth.NewCache(), // Auth token cache
	}

	// Set static credentials if provided
	if opts.Username != "" && opts.Password != "" {
		authClient.Credential = auth.StaticCredential(repo.Reference.Registry, auth.Credential{
			Username: opts.Username,
			Password: opts.Password,
		})
	}

	repo.Client = authClient
	repo.PlainHTTP = opts.InsecureSkipVerifyTLS // Support for HTTP registries (non-HTTPS)

	// 4. Resolve Tag -> Descriptor
	// This performs a HEAD/GET request to fetch the manifest digest
	desc, err := repo.Resolve(ctx, repo.Reference.Reference)
	if err != nil {
		return nil, "", fmt.Errorf("failed to resolve reference %s: %w", refString, err)
	}

	// 5. Download Manifest
	// FetchAll is acceptable here as the manifest is small (KB), so loading it into RAM is safe.
	manifestBytes, err := content.FetchAll(ctx, repo, desc)
	if err != nil {
		return nil, "", fmt.Errorf("failed to fetch manifest: %w", err)
	}

	var manifest ocispec.Manifest
	if err := json.Unmarshal(manifestBytes, &manifest); err != nil {
		return nil, "", fmt.Errorf("failed to parse manifest: %w", err)
	}

	// 6. Identify the Chart Layer
	var chartLayerDesc *ocispec.Descriptor
	for _, layer := range manifest.Layers {
		if layer.MediaType == ChartLayerMediaType || layer.MediaType == LegacyLayerMediaType {
			l := layer // Create local copy for pointer
			chartLayerDesc = &l
			break
		}
	}

	if chartLayerDesc == nil {
		// Fallback: if there is only one layer, assume it is the chart
		if len(manifest.Layers) == 1 {
			chartLayerDesc = &manifest.Layers[0]
		} else {
			return nil, "", fmt.Errorf("chart layer not found in manifest")
		}
	}

	// 7. Stream the Layer
	// repo.Fetch returns an io.ReadCloser connected directly to the HTTP response.
	// This does NOT load the chart data into memory.
	rc, err := repo.Fetch(ctx, *chartLayerDesc)
	if err != nil {
		return nil, "", fmt.Errorf("failed to start stream for layer %s: %w", chartLayerDesc.Digest, err)
	}

	return rc, "oci://" + refString, nil
}

// getTransport handles secure/insecure TLS while reusing the shared connection pool
func getTransport(insecure bool) http.RoundTripper {
	if insecure {
		// Shallow clone and modify TLS config
		t := sharedTransport.Clone()
		t.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
		return retry.NewTransport(t) // ORAS retry wrapper
	}
	return retry.NewTransport(sharedTransport)
}
