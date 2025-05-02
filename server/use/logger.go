package use

import (
	"net/http"

	"log/slog"

	xcontext "github.com/krateoplatformops/plumbing/context"
	"github.com/krateoplatformops/plumbing/shortid"
)

func Logger(root *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(wri http.ResponseWriter, req *http.Request) {
			traceId := req.Header.Get("X-Krateo-TraceId")
			if len(traceId) == 0 {
				traceId = shortid.MustGenerate()
			}

			sub := req.Header.Get("X-Krateo-User")
			orgs := req.Header.Get("X-Krateo-Groups")

			log := root
			if len(sub) > 0 {
				log = root.With("traceId", traceId,
					slog.Group("user",
						slog.String("name", sub),
						slog.String("groups", orgs)),
				)
			}

			ctx := xcontext.BuildContext(req.Context(),
				xcontext.WithTraceId(traceId),
				xcontext.WithLogger(log),
			)

			next.ServeHTTP(wri, req.WithContext(ctx))
		}

		return http.HandlerFunc(fn)
	}
}
