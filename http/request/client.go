package request

import (
	"fmt"
	"net/http"

	"github.com/krateoplatformops/plumbing/endpoints"
)

func HTTPClientForEndpoint(ep *endpoints.Endpoint) (*http.Client, error) {
	rt, err := tlsConfigFor(ep)
	if err != nil {
		return &http.Client{
			Transport: &traceIdRoundTripper{defaultTransport()},
		}, err
	}
	rt = &traceIdRoundTripper{rt}

	if ep.Debug {
		rt = &debuggingRoundTripper{
			delegatedRoundTripper: rt,
		}
	}

	// Set authentication wrappers
	switch {
	case ep.HasBasicAuth() && ep.HasTokenAuth():
		return nil, fmt.Errorf("username/password or bearer token may be set, but not both")

	case ep.HasTokenAuth():
		rt = &bearerAuthRoundTripper{
			bearer: ep.Token,
			rt:     rt,
		}

	case ep.HasBasicAuth():
		rt = &basicAuthRoundTripper{
			username: ep.Username,
			password: ep.Password,
			rt:       rt,
		}
	}

	return &http.Client{Transport: rt}, nil
}
