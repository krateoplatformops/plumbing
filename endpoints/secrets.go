package endpoints

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/rest"
)

type clientOptions struct {
	cli       *rest.RESTClient
	name      string
	namespace string
}

func newSecretsRESTClient(rc *rest.Config) (*rest.RESTClient, error) {
	gv := schema.GroupVersion{
		Group:   "",
		Version: "v1",
	}

	sb := runtime.NewSchemeBuilder(
		func(reg *runtime.Scheme) error {
			reg.AddKnownTypes(
				gv,
				&corev1.Secret{},
				&corev1.SecretList{},
				&metav1.ListOptions{},
				&metav1.GetOptions{},
				&metav1.DeleteOptions{},
				&metav1.CreateOptions{},
				&metav1.UpdateOptions{},
				&metav1.PatchOptions{},
				&metav1.Status{},
			)
			return nil
		})

	s := runtime.NewScheme()
	sb.AddToScheme(s)

	config := *rc
	config.APIPath = "/api"
	config.GroupVersion = &gv
	config.NegotiatedSerializer = serializer.NewCodecFactory(s).
		WithoutConversion()
	config.UserAgent = rest.DefaultKubernetesUserAgent()

	cli, err := rest.RESTClientFor(&config)
	if err != nil {
		return nil, err
	}

	//pc := runtime.NewParameterCodec(s)

	return cli, nil
}

func getSecret(ctx context.Context, opts clientOptions) (result *corev1.Secret, err error) {
	result = &corev1.Secret{}
	err = opts.cli.Get().
		Namespace(opts.namespace).
		Resource("secrets").
		Name(opts.name).
		Do(ctx).
		Into(result)
	return
}

func createSecret(ctx context.Context, cli *rest.RESTClient, in *corev1.Secret) error {
	return cli.Post().
		Namespace(in.GetNamespace()).
		Resource("secrets").
		Body(in).
		Do(ctx).
		Error()
}

func updateSecret(ctx context.Context, cli *rest.RESTClient, in *corev1.Secret) error {
	return cli.Put().
		Namespace(in.GetNamespace()).
		Resource("secrets").
		Name(in.Name).
		Body(in).
		Do(ctx).
		Error()
}

func deleteSecret(ctx context.Context, opts clientOptions) error {
	return opts.cli.Delete().
		Namespace(opts.namespace).
		Resource("secrets").
		Name(opts.name).
		Do(ctx).
		Error()
}
