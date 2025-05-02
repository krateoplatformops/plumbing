package kubeutil

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"strings"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// ErrNoNamespace indicates that a namespace could not
// be found for the current environment
var (
	ErrNoNamespace = fmt.Errorf("namespace not found for current environment")
)

func ServiceAccountNamespace() (string, error) {
	nsBytes, err := os.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace")
	if err != nil {
		if os.IsNotExist(err) {
			return "", ErrNoNamespace
		}
		return "", err
	}

	ns := strings.TrimSpace(string(nsBytes))
	return ns, nil
}

func CACrt(ctx context.Context, rc *rest.Config) (string, error) {
	const (
		name = "kube-root-ca.crt"
	)

	namespace, err := ServiceAccountNamespace()
	if err != nil {
		return "", err
	}

	data, err := ConfigMapData(ctx, rc, name, namespace)
	if err != nil {
		return "", err
	}

	crt, ok := data["ca.crt"]
	if !ok {
		return "", fmt.Errorf("ca.crt key not found in configmaps '%s' (namespace: %s)", name, namespace)
	}

	enc := base64.StdEncoding.EncodeToString([]byte(crt))
	return enc, err
}

func ConfigMapData(ctx context.Context, rc *rest.Config, name, namespace string) (map[string]string, error) {
	cli, err := kubernetes.NewForConfig(rc)
	if err != nil {
		return nil, err
	}

	res, err := cli.CoreV1().ConfigMaps(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return nil, fmt.Errorf("configmaps '%s' not found (namespace: %s)", name, namespace)
		}
		return nil, err
	}
	return res.Data, nil
}
