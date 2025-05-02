package use

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	xcontext "github.com/krateoplatformops/plumbing/context"
)

func TestTraceIdMiddleware(t *testing.T) {
	buf := bytes.Buffer{}

	// Create a simple handler that uses the logger.
	sillyHandler := http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			log := xcontext.Logger(r.Context())
			log.Info("Processing a lot...")
			log.Debug("for devs only")
			w.Write([]byte("Hello, World!"))
			log.Info("Done!")
		})

	route := NewChain(TraceId()).Then(sillyHandler)

	// Create a test request.
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set(xcontext.LabelKrateoUser, "cyberjoker")
	req.Header.Set(xcontext.LabelKrateoGroups, "devs,testers")
	rec := httptest.NewRecorder()

	// Serve the request.
	route.ServeHTTP(rec, req)

	// Check the log output.
	fmt.Println(buf.String())
}
