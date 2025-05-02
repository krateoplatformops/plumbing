package use

import (
	"bytes"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	xcontext "github.com/krateoplatformops/plumbing/context"
)

func TestLoggerMiddleware(t *testing.T) {
	buf := bytes.Buffer{}

	log := slog.New(slog.NewJSONHandler(&buf,
		&slog.HandlerOptions{Level: slog.LevelDebug}))

	// Create a simple handler that uses the logger.
	sillyHandler := http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			log := xcontext.Logger(r.Context())
			log.Info("Processing a lot...")
			log.Debug("for devs only")
			time.Sleep(1 * time.Second)
			w.Write([]byte("Hello, World!"))

			log.Info("Done!")
		})

	route := NewChain(Logger(log)).Then(sillyHandler)

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
