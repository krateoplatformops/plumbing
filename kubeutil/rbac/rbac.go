package rbac

import (
	"context"
	"log/slog"
	"sync"
	"sync/atomic"
	"time"

	xcontext "github.com/krateoplatformops/plumbing/context"
	"github.com/krateoplatformops/plumbing/endpoints"
	"github.com/krateoplatformops/plumbing/env"
	"github.com/krateoplatformops/plumbing/kubeconfig"
	authv1 "k8s.io/api/authorization/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes"
)

type UserCanOptions struct {
	UserConfig    endpoints.Endpoint
	Verb          string
	GroupResource schema.GroupResource
	Namespace     string
}

// UserCan evaluates whether the current user can execute the given verb on the
// target GroupResource in the provided namespace.
//
// Pipeline:
//  1. Resolves user endpoint configuration from opts.UserConfig; if empty, falls
//     back to context (xcontext.UserConfig).
//  2. Checks the in-memory TTL cache (if enabled) and returns immediately on hit.
//  3. Builds a Kubernetes REST config from the endpoint.
//  4. Creates a clientset and sends a SelfSubjectAccessReview request.
//  5. Stores the result in cache (if enabled) and returns the authorization outcome.
//
// Required context data:
//   - If opts.UserConfig is empty, a user endpoint must be attached through
//     xcontext.WithUserConfig(...) before calling this function.
//   - If both opts.UserConfig and context user config are missing, the function
//     logs an error and returns false.
//   - A logger is optional: if not present, xcontext.Logger(ctx) falls back to a default logger.
//
// Notes:
//   - opts.UserConfig allows explicit endpoint wiring, while context fallback
//     keeps existing call sites compatible.
//   - Any internal failure (missing context data, config/client creation, API call)
//     results in false.
func UserCan(ctx context.Context, opts UserCanOptions) (ok bool) {
	log := xcontext.Logger(ctx)
	ep := opts.UserConfig
	if ep == (endpoints.Endpoint{}) {
		var err error
		ep, err = xcontext.UserConfig(ctx)
		if err != nil {
			log.Error("unable to get user endpoint", slog.Any("err", err))
			return false
		}
	}

	ttlNanos := userCanCacheTTLNanos.Load()
	if ttlNanos > 0 {
		now := time.Now().UnixNano()
		cacheKey := newUserCanCacheKey(ep, opts)
		if allowed, hit := getUserCanCache(cacheKey, now); hit {
			log.Debug("UserCan result from cache",
				slog.String("source", "cache"),
				slog.Bool("allowed", allowed))
			return allowed
		}
	}

	log.Debug("UserCan requesting SelfSubjectAccessReview",
		slog.String("source", "k8s-api"))

	rc, err := kubeconfig.NewClientConfig(ctx, ep)
	if err != nil {
		log.Error("unable to create user client config", slog.Any("err", err))
		return false
	}

	clientset, err := kubernetes.NewForConfig(rc)
	if err != nil {
		log.Error("unable to create kubernetes clientset", slog.Any("err", err))
		return false
	}

	selfCheck := authv1.SelfSubjectAccessReview{
		Spec: authv1.SelfSubjectAccessReviewSpec{
			ResourceAttributes: &authv1.ResourceAttributes{
				Group:     opts.GroupResource.Group,
				Resource:  opts.GroupResource.Resource,
				Namespace: opts.Namespace,
				Verb:      opts.Verb,
			},
		},
	}

	resp, err := clientset.AuthorizationV1().SelfSubjectAccessReviews().
		Create(context.TODO(), &selfCheck, metav1.CreateOptions{})
	if err != nil {
		log.Error("unable to perform SelfSubjectAccessReviews",
			slog.Any("selfCheck", selfCheck), slog.Any("err", err))
		return false
	}

	log.Debug("SelfSubjectAccessReviews result", slog.Any("response", resp))

	if ttlNanos > 0 {
		setUserCanCache(
			newUserCanCacheKey(ep, opts), resp.Status.Allowed, time.Now().UnixNano()+ttlNanos)
	}

	return resp.Status.Allowed
}

const defaultUserCanCacheTTL = 10 * time.Second

var (
	userCanCacheTTLNanos atomic.Int64
	userCanCacheMu       sync.RWMutex
	userCanCache         = make(map[userCanCacheKey]userCanCacheEntry, 256)
)

func init() {
	userCanCacheTTLNanos.Store(
		env.Duration("RBAC_USERCAN_CACHE_TTL", defaultUserCanCacheTTL).Nanoseconds(),
	)
}

type userCanCacheKey struct {
	Endpoint  endpoints.Endpoint
	Verb      string
	Group     string
	Resource  string
	Namespace string
}

type userCanCacheEntry struct {
	Allowed   bool
	ExpiresAt int64
}

func SetUserCanCacheTTL(ttl time.Duration) {
	userCanCacheTTLNanos.Store(ttl.Nanoseconds())
	if ttl <= 0 {
		userCanCacheMu.Lock()
		clear(userCanCache)
		userCanCacheMu.Unlock()
	}
}

func newUserCanCacheKey(ep endpoints.Endpoint, opts UserCanOptions) userCanCacheKey {
	return userCanCacheKey{
		Endpoint:  ep,
		Verb:      opts.Verb,
		Group:     opts.GroupResource.Group,
		Resource:  opts.GroupResource.Resource,
		Namespace: opts.Namespace,
	}
}

func getUserCanCache(key userCanCacheKey, now int64) (bool, bool) {
	userCanCacheMu.RLock()
	entry, ok := userCanCache[key]
	userCanCacheMu.RUnlock()
	if !ok {
		return false, false
	}
	if now < entry.ExpiresAt {
		return entry.Allowed, true
	}

	userCanCacheMu.Lock()
	entry, ok = userCanCache[key]
	if ok && now >= entry.ExpiresAt {
		delete(userCanCache, key)
	}
	userCanCacheMu.Unlock()

	return false, false
}

func setUserCanCache(key userCanCacheKey, allowed bool, expiresAt int64) {
	userCanCacheMu.Lock()
	userCanCache[key] = userCanCacheEntry{
		Allowed:   allowed,
		ExpiresAt: expiresAt,
	}
	userCanCacheMu.Unlock()
}
