package util

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"time"
)

// RetryClient wraps an http.Client to add automatic retry on 429 and 5xx errors.
type RetryClient struct {
	Client      *http.Client
	MaxRetries  int
	BaseBackoff time.Duration
	MaxBackoff  time.Duration
}

// Do executes the request with retry logic only for idempotent methods.
func (rc *RetryClient) Do(req *http.Request) (*http.Response, error) {
	if rc.Client == nil {
		rc.Client = http.DefaultClient
	}
	if rc.MaxRetries == 0 {
		rc.MaxRetries = 3
	}
	if rc.BaseBackoff == 0 {
		rc.BaseBackoff = time.Second
	}
	if rc.MaxBackoff == 0 {
		rc.MaxBackoff = 10 * time.Second
	}

	// Only retry idempotent HTTP methods
	if !isIdempotentMethod(req.Method) {
		return rc.Client.Do(req)
	}

	var lastErr error
	var resp *http.Response

	for attempt := 0; attempt <= rc.MaxRetries; attempt++ {
		// Clone the request to ensure Body can be reused
		clonedReq := req.Clone(req.Context())

		resp, lastErr = rc.Client.Do(clonedReq)
		if lastErr != nil {
			// Network error → retry
			if attempt < rc.MaxRetries {
				if err := sleepWithContext(req.Context(), backoff(attempt, rc.BaseBackoff, rc.MaxBackoff)); err != nil {
					return nil, err
				}
				continue
			}
			return nil, lastErr
		}

		// If response is successful → return immediately
		if resp.StatusCode < 500 && resp.StatusCode != http.StatusTooManyRequests {
			return resp, nil
		}

		// Handle 429 Too Many Requests with Retry-After
		if resp.StatusCode == http.StatusTooManyRequests {
			retryAfter := parseRetryAfter(resp)
			resp.Body.Close()
			if attempt < rc.MaxRetries {
				if err := sleepWithContext(req.Context(), retryAfter); err != nil {
					return nil, err
				}
				continue
			}
			return resp, errors.New("too many retries after 429 Too Many Requests")
		}

		// Handle 5xx server errors
		if resp.StatusCode >= 500 {
			resp.Body.Close()
			if attempt < rc.MaxRetries {
				if err := sleepWithContext(req.Context(), backoff(attempt, rc.BaseBackoff, rc.MaxBackoff)); err != nil {
					return nil, err
				}
				continue
			}
			return resp, fmt.Errorf("server error after %d retries: %d", rc.MaxRetries, resp.StatusCode)
		}
	}

	return resp, lastErr
}

// isIdempotentMethod returns true for HTTP methods that are safe to retry
func isIdempotentMethod(method string) bool {
	switch method {
	case http.MethodGet, http.MethodHead, http.MethodOptions, http.MethodPut, http.MethodDelete:
		return true
	default:
		return false
	}
}

// backoff calculates exponential backoff with max cap and random jitter
func backoff(attempt int, base, max time.Duration) time.Duration {
	d := base * (1 << attempt) // exponential
	if d > max {
		d = max
	}
	// Add jitter: random between 50% and 100% of d
	jitter := d/2 + time.Duration(rand.Int63n(int64(d)/2))
	return jitter
}

// sleepWithContext waits for the given duration or until context is done
func sleepWithContext(ctx context.Context, d time.Duration) error {
	select {
	case <-time.After(d):
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// parseRetryAfter parses the Retry-After header and returns a duration
func parseRetryAfter(resp *http.Response) time.Duration {
	ra := resp.Header.Get("Retry-After")
	if ra == "" {
		return 2 * time.Second
	}

	// Retry-After as seconds
	if secs, err := strconv.Atoi(ra); err == nil {
		return time.Duration(secs) * time.Second
	}

	// Retry-After as HTTP date
	if t, err := http.ParseTime(ra); err == nil {
		return time.Until(t)
	}

	// Fallback
	return 2 * time.Second
}
