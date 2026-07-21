package use

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	xcontext "github.com/krateoplatformops/plumbing/context"
	"github.com/krateoplatformops/plumbing/endpoints"
	"github.com/krateoplatformops/plumbing/http/response"
	"github.com/krateoplatformops/plumbing/jwtutil"
	"github.com/krateoplatformops/plumbing/kubeutil"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/rest"
)

// UserConfig builds a middleware that validates the incoming bearer token.
// publicKeyPEM is the PEM-encoded RSA public key (the public half of the
// keypair authn signs with) used to verify token signatures. The key is parsed
// once here, not per request; a parse failure is surfaced on every request as
// an internal error so the misconfiguration is obvious.
func UserConfig(publicKeyPEM, authnNS string) func(http.Handler) http.Handler {
	publicKey, keyErr := jwtutil.ParseRSAPublicKeyFromPEM([]byte(publicKeyPEM))

	return func(next http.Handler) http.Handler {
		fn := func(wri http.ResponseWriter, req *http.Request) {
			if keyErr != nil {
				response.InternalError(wri, fmt.Errorf("unable to parse JWT public key: %w", keyErr))
				return
			}

			authHeader := req.Header.Get("Authorization")
			if authHeader == "" {
				response.Unauthorized(wri, fmt.Errorf("missing authorization header"))
				return
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
				response.Unauthorized(wri, fmt.Errorf("invalid authorization header format"))
				return
			}

			userInfo, err := jwtutil.Validate(publicKey, parts[1])
			if err != nil {
				if errors.Is(err, jwtutil.ErrTokenExpired) {
					response.Unauthorized(wri, err)
				} else {
					response.Unauthorized(wri, err)
				}
				return
			}

			sarc, err := rest.InClusterConfig()
			if err != nil {
				response.InternalError(wri, fmt.Errorf("unable to create in cluster config: %w", err))
				return
			}

			ep, err := endpoints.FromSecret(context.Background(), sarc,
				fmt.Sprintf("%s-clientconfig",
					kubeutil.MakeDNS1123Compatible(userInfo.Username)), authnNS)
			if err != nil {
				if apierrors.IsNotFound(err) {
					response.Unauthorized(wri, err)
					return
				}
				response.InternalError(wri, err)
				return
			}

			ctx := xcontext.BuildContext(req.Context(),
				xcontext.WithAccessToken(parts[1]),
				xcontext.WithUserInfo(userInfo),
				xcontext.WithUserConfig(ep),
			)

			next.ServeHTTP(wri, req.WithContext(ctx))
		}

		return http.HandlerFunc(fn)
	}
}
