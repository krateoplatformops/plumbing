package request

import (
	"fmt"
	"net/http"
)

func HTTPClientForEndpoint(opts RequestOptions) (*http.Client, error) {
	ep := opts.Endpoint
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

	// Set authentication wrapper
	// Only one can be active at a time
	countTrue := 0
	if ep.HasBasicAuth() {
		countTrue += 1
	}
	if ep.HasTokenAuth() {
		countTrue += 1
	}
	if ep.HasAwsAuth() {
		countTrue += 1
	}

	switch {
	case countTrue > 1:
		return nil, fmt.Errorf("only one of username/password, bearer token and AWS must be set")

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

	case ep.HasAwsAuth():
		rt = &awsAuthRoundTripper{
			headerPayload: ComputeAwsHeader(opts),
			rt:            rt,
		}
	}

	return &http.Client{Transport: rt}, nil
}
