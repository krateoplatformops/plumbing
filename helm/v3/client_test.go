package helm

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"

	helmconfig "github.com/krateoplatformops/plumbing/helm"
	"k8s.io/client-go/rest"
)

func TestInstallAndUninstall_OCI(t *testing.T) {
	// 1. Setup Logger
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	cfg, err := rest.InClusterConfig()
	if err != nil {
		t.Fatalf("Failed to get in-cluster config: %v", err)
	}
	// 2. Initialize Client
	cli, err := NewClient(cfg, "default", logger.Handler())
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	cli, err = cli.WithCache()
	if err != nil {
		t.Fatalf("Failed to setup cache: %v", err)
	}

	// 3. Define Chart Details - Using official Bitnami OCI registry
	chartRef := "oci://registry-1.docker.io/bitnamicharts/nginx"
	releaseName := "test-release-oci"

	t.Logf("Attempting to install chart: %s", chartRef)

	// 4. Run Install
	rel, err := cli.Install(
		context.TODO(),
		releaseName,
		chartRef,
		&helmconfig.InstallConfig{
			ActionConfig: &helmconfig.ActionConfig{
				// Version:       "13.0.0",
				TakeOwnership: true,
			},
		},
	)
	if err != nil {
		t.Fatalf("Install failed: %v", err)
	}

	if rel.Name != releaseName {
		t.Errorf("Expected release name %s, got %s", releaseName, rel.Name)
	}
	if rel.Status != "deployed" {
		t.Errorf("Expected status deployed, got %s", rel.Status)
	}

	t.Logf("Successfully installed release: %s (v%d)", rel.Name, rel.Revision)

	// 5. Run Uninstall
	t.Logf("Attempting to uninstall release: %s", releaseName)
	if err := cli.Uninstall(context.TODO(), releaseName, &helmconfig.UninstallConfig{}); err != nil {
		t.Fatalf("Uninstall failed: %v", err)
	}

	t.Log("Successfully uninstalled release")
}

func TestInstallAndUninstall_Repo(t *testing.T) {
	// 1. Setup Logger
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	cfg, err := rest.InClusterConfig()
	if err != nil {
		t.Fatalf("Failed to get in-cluster config: %v", err)
	}
	// 2. Initialize Client
	cli, err := NewClient(cfg, "default", logger.Handler())
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	cli, err = cli.WithCache()
	if err != nil {
		t.Fatalf("Failed to setup cache: %v", err)
	}

	// 3. Define Chart Details - Using official Kubernetes ingress-nginx repo
	chartRef := "https://kubernetes.github.io/ingress-nginx"
	releaseName := "test-release-repo"
	chartName := "ingress-nginx"

	t.Logf("Attempting to install chart %s from repo: %s", chartName, chartRef)

	// 4. Run Install
	rel, err := cli.Install(
		context.TODO(),
		releaseName,
		chartRef,
		&helmconfig.InstallConfig{
			ActionConfig: &helmconfig.ActionConfig{
				ChartName:     chartName,
				ChartVersion:  "4.10.0", // Pin version for stability
				Timeout:       2 * time.Minute,
				TakeOwnership: true,
			},
		},
	)
	if err != nil {
		t.Fatalf("Install failed: %v", err)
	}

	if rel.Name != releaseName {
		t.Errorf("Expected release name %s, got %s", releaseName, rel.Name)
	}
	if rel.Status != "deployed" {
		t.Errorf("Expected status deployed, got %s", rel.Status)
	}

	t.Logf("Successfully installed release: %s (v%d)", rel.Name, rel.Revision)

	// 5. Run Uninstall
	t.Logf("Attempting to uninstall release: %s", releaseName)
	if err := cli.Uninstall(context.TODO(), releaseName, &helmconfig.UninstallConfig{}); err != nil {
		t.Fatalf("Uninstall failed: %v", err)
	}

	t.Log("Successfully uninstalled release")
}

func TestInstallAndUninstall_TGZ(t *testing.T) {
	// 1. Setup Logger
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	cfg, err := rest.InClusterConfig()
	if err != nil {
		t.Fatalf("Failed to get in-cluster config: %v", err)
	}
	// 2. Initialize Client
	cli, err := NewClient(cfg, "default", logger.Handler())
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	cli, err = cli.WithCache()
	if err != nil {
		t.Fatalf("Failed to setup cache: %v", err)
	}

	// 3. Define Chart Details - Direct .tgz URL
	// Using official Kubernetes ingress-nginx chart archive
	chartRef := "https://github.com/kubernetes/ingress-nginx/releases/download/helm-chart-4.10.0/ingress-nginx-4.10.0.tgz"
	releaseName := "test-release-tgz"

	t.Logf("Attempting to install chart from tgz: %s", chartRef)

	// 4. Run Install
	rel, err := cli.Install(
		context.TODO(),
		releaseName,
		chartRef,
		&helmconfig.InstallConfig{
			ActionConfig: &helmconfig.ActionConfig{
				ChartVersion:  "4.10.0",
				Timeout:       2 * time.Minute,
				TakeOwnership: true,
			},
		},
	)
	if err != nil {
		t.Fatalf("Install failed: %v", err)
	}

	if rel.Name != releaseName {
		t.Errorf("Expected release name %s, got %s", releaseName, rel.Name)
	}
	if rel.Status != "deployed" {
		t.Errorf("Expected status deployed, got %s", rel.Status)
	}

	t.Logf("Successfully installed release: %s (v%d)", rel.Name, rel.Revision)

	// 5. Run Uninstall
	t.Logf("Attempting to uninstall release: %s", releaseName)
	if err := cli.Uninstall(context.TODO(), releaseName, &helmconfig.UninstallConfig{}); err != nil {
		t.Fatalf("Uninstall failed: %v", err)
	}

	t.Log("Successfully uninstalled release")
}
