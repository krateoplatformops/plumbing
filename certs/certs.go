package certs

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"time"

	"k8s.io/client-go/kubernetes"

	certv1 "k8s.io/api/certificates/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
)

const (
	certificateWaitTimeout       = 5 * time.Minute
	certificateWaitPollInternval = 3 * time.Second
	resourceAnnotationKey        = "krateo.user.id"
)

func NewPrivateKey() (*rsa.PrivateKey, error) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, fmt.Errorf("generating private key: %w", err)
	}

	return key, nil
}

type CertificateRequestOptions struct {
	Key             *rsa.PrivateKey
	Username        string
	Groups          []string
	ExtraExtensions []pkix.Extension
}

func NewCertificateRequest(opts CertificateRequestOptions) ([]byte, error) {
	req := x509.CertificateRequest{
		Subject: pkix.Name{
			CommonName:   opts.Username,
			Organization: opts.Groups,
		},
		SignatureAlgorithm: x509.SHA256WithRSA,
		ExtraExtensions:    []pkix.Extension{},
	}
	if len(opts.ExtraExtensions) > 0 {
		req.ExtraExtensions = opts.ExtraExtensions
	}

	dat, err := x509.CreateCertificateRequest(rand.Reader, &req, opts.Key)
	if err != nil {
		return nil, fmt.Errorf("creating certificate request: %w", err)
	}

	enc := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE REQUEST", Bytes: dat})

	return enc, nil
}

func DeleteCertificateSigningRequest(client kubernetes.Interface, name string) error {
	err := client.CertificatesV1().CertificateSigningRequests().
		Delete(context.Background(), name, metav1.DeleteOptions{})
	if err != nil {
		return err
	}

	condFunc := func(ctx context.Context) (bool, error) {
		_, err := client.CertificatesV1().CertificateSigningRequests().
			Get(ctx, name, metav1.GetOptions{})
		if errors.IsNotFound(err) {
			return true, nil
		}

		return false, nil
	}

	ctx := context.Background()
	err = wait.PollUntilContextTimeout(ctx, certificateWaitPollInternval, certificateWaitTimeout, true, condFunc)
	if err != nil {
		return fmt.Errorf("waiting for CSR certificate to be deleted: %w", err)
	}
	return nil
}

func CreateCertificateSigningRequests(client kubernetes.Interface, csr *certv1.CertificateSigningRequest) error {
	_, err := client.CertificatesV1().CertificateSigningRequests().
		Create(context.Background(), csr, metav1.CreateOptions{})
	return err
}

func ApproveCertificateSigningRequest(client kubernetes.Interface, csr *certv1.CertificateSigningRequest, approver string) error {
	cond := certv1.CertificateSigningRequestCondition{
		Type:           certv1.CertificateApproved,
		Status:         corev1.ConditionTrue,
		Reason:         "CertificateApproved",
		Message:        fmt.Sprintf("Certificate was approved by %s", approver),
		LastUpdateTime: metav1.Now(),
	}

	csr.Status.Conditions = append(csr.Status.Conditions, cond)

	ctx := context.Background()
	_, err := client.CertificatesV1().CertificateSigningRequests().
		UpdateApproval(ctx, csr.Name, csr, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("approving CertificateSigningRequest: %w", err)
	}
	return nil
}

func WaitForCertificate(client kubernetes.Interface, name string) error {
	ctx := context.Background()
	err := wait.PollUntilContextTimeout(ctx, certificateWaitPollInternval,
		certificateWaitTimeout, false, CertificateExistsFunc(client, name))
	if err != nil {
		return fmt.Errorf("waiting for CSR certificate to be generated: %w", err)
	}

	return nil
}

func CertificateExistsFunc(client kubernetes.Interface, name string) wait.ConditionWithContextFunc {
	return func(ctx context.Context) (bool, error) {
		obj, err := client.CertificatesV1().CertificateSigningRequests().
			Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			return false, err
		}

		if len(obj.Status.Certificate) > 0 {
			return true, nil
		}

		return false, nil
	}
}

func Certificate(client kubernetes.Interface, name string) ([]byte, error) {
	csr, err := GetCertificateSigningRequest(client, name)
	if err != nil {
		return nil, err
	}

	return csr.Status.Certificate, nil
}

func GetCertificateSigningRequest(client kubernetes.Interface, name string) (*certv1.CertificateSigningRequest, error) {
	csr, err := client.CertificatesV1().CertificateSigningRequests().
		Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("getting CSR '%s': %w", name, err)
	}

	return csr, nil
}

type CertificateSigningRequestOptions struct {
	CSR        []byte
	Duration   time.Duration
	UserID     string
	Username   string
	SignerName string
	Usages     []string
}

func NewCertificateSigningRequest(opts CertificateSigningRequestOptions) *certv1.CertificateSigningRequest {
	if opts.SignerName == "" {
		opts.SignerName = certv1.KubeAPIServerClientSignerName
	}

	if len(opts.Usages) == 0 {
		opts.Usages = []string{string(certv1.UsageClientAuth)}
	}

	durationSeconds := int32(opts.Duration.Seconds())
	csrObject := &certv1.CertificateSigningRequest{
		ObjectMeta: metav1.ObjectMeta{
			Name: opts.Username,
		},
		Spec: certv1.CertificateSigningRequestSpec{
			Request:           opts.CSR,
			SignerName:        opts.SignerName,
			ExpirationSeconds: &durationSeconds,
		},
	}

	csrObject.Spec.Usages = make([]certv1.KeyUsage, 0, len(opts.Usages))
	for _, el := range opts.Usages {
		csrObject.Spec.Usages = append(csrObject.Spec.Usages, certv1.KeyUsage(el))
	}

	if opts.UserID != "" {
		csrObject.Annotations = map[string]string{
			resourceAnnotationKey: opts.UserID,
		}
	}

	return csrObject
}
