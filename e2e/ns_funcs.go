package e2e

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"sigs.k8s.io/e2e-framework/pkg/env"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/envfuncs"
)

// CreateNamespace provides an Environment.Func that
// creates a new namespace API object and stores it the context
// using its name as key.
func CreateNamespace(name string, opts ...envfuncs.CreateNamespaceOpts) env.Func {
	return func(ctx context.Context, cfg *envconf.Config) (context.Context, error) {
		namespace := corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: name}}
		client, err := cfg.NewClient()
		if err != nil {
			return ctx, fmt.Errorf("create namespace func: %w", err)
		}
		for _, opt := range opts {
			opt(client, &namespace)
		}
		if err := client.Resources().Create(ctx, &namespace); err != nil {
			if !errors.IsAlreadyExists(err) {
				return ctx, fmt.Errorf("create namespace func: %w", err)
			}
		}
		cfg.WithNamespace(name) // set env config default namespace
		return context.WithValue(ctx, envfuncs.NamespaceContextKey(name), namespace), nil
	}
}
