package rbac

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	xcontext "github.com/krateoplatformops/plumbing/context"
	"github.com/krateoplatformops/plumbing/endpoints"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func TestUserCanUsesCache(t *testing.T) {
	t.Setenv("TEST_MODE", "true")
	SetUserCanCacheTTL(0)
	SetUserCanCacheTTL(30 * time.Second)
	defer SetUserCanCacheTTL(defaultUserCanCacheTTL)

	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("unexpected method: %s", r.Method)
		}
		if r.URL.Path != "/apis/authorization.k8s.io/v1/selfsubjectaccessreviews" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		calls.Add(1)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"apiVersion":"authorization.k8s.io/v1","kind":"SelfSubjectAccessReview","status":{"allowed":true}}`))
	}))
	defer server.Close()

	opts := UserCanOptions{
		UserConfig: endpoints.Endpoint{
			ServerURL: server.URL,
		},
		Verb: "get",
		GroupResource: schema.GroupResource{
			Group:    "",
			Resource: "pods",
		},
		Namespace: "default",
	}

	if ok := UserCan(context.TODO(), opts); !ok {
		t.Fatal("expected first call to be allowed")
	}
	if ok := UserCan(context.TODO(), opts); !ok {
		t.Fatal("expected second call to be allowed")
	}

	if got := calls.Load(); got != 1 {
		t.Fatalf("expected 1 API call due to cache hit, got %d", got)
	}
}

func TestUserCanWithoutCacheCallsAPIEachTime(t *testing.T) {
	t.Setenv("TEST_MODE", "true")
	SetUserCanCacheTTL(0)
	defer SetUserCanCacheTTL(defaultUserCanCacheTTL)

	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"apiVersion":"authorization.k8s.io/v1","kind":"SelfSubjectAccessReview","status":{"allowed":true}}`))
	}))
	defer server.Close()

	ctx := xcontext.BuildContext(context.Background(),
		xcontext.WithUserConfig(endpoints.Endpoint{
			ServerURL: server.URL,
		}))

	opts := UserCanOptions{
		Verb: "list",
		GroupResource: schema.GroupResource{
			Group:    "",
			Resource: "pods",
		},
		Namespace: "default",
	}

	if ok := UserCan(ctx, opts); !ok {
		t.Fatal("expected first call to be allowed")
	}
	if ok := UserCan(ctx, opts); !ok {
		t.Fatal("expected second call to be allowed")
	}

	if got := calls.Load(); got != 2 {
		t.Fatalf("expected 2 API calls with cache disabled, got %d", got)
	}
}
