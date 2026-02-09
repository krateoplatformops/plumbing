//go:build integration
// +build integration

package helm

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"

	xenv "github.com/krateoplatformops/plumbing/env"
	helmconfig "github.com/krateoplatformops/plumbing/helm"
	"github.com/stretchr/testify/assert"

	"sigs.k8s.io/e2e-framework/pkg/env"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/envfuncs"
	"sigs.k8s.io/e2e-framework/pkg/features"
	"sigs.k8s.io/e2e-framework/support/kind"
)

var (
	testenv     env.Environment
	clusterName string
	namespace   string
)

func TestMain(m *testing.M) {
	xenv.SetTestMode(true)

	namespace = "helm-test-system"
	clusterName = "helm-test"
	testenv = env.New()

	testenv.Setup(
		envfuncs.CreateCluster(kind.NewProvider(), clusterName),
		envfuncs.CreateNamespace(namespace),
	).Finish(
		envfuncs.DeleteNamespace(namespace),
		envfuncs.DestroyCluster(clusterName),
	)

	os.Exit(testenv.Run(m))
}

func TestInstallAndUninstall_OCI(t *testing.T) {
	os.Setenv("DEBUG", "0")

	f := features.New("Install and Uninstall OCI Chart").
		Assess("Install and Uninstall nginx from OCI", func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
			// 1. Setup Logger
			logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

			// 2. Initialize Client
			cli, err := NewClient(c.Client().RESTConfig(), namespace, logger.Handler())
			assert.Nil(t, err)

			cli, err = cli.WithCache()
			assert.Nil(t, err)
			defer cli.Close()

			// 3. Define Chart Details - Using official Bitnami OCI registry
			chartRef := "oci://registry-1.docker.io/bitnamicharts/nginx"
			releaseName := "test-release-oci"

			t.Logf("Attempting to install chart: %s", chartRef)

			// 4. Run Install
			rel, err := cli.Install(
				ctx,
				releaseName,
				chartRef,
				&helmconfig.InstallConfig{
					ActionConfig: &helmconfig.ActionConfig{
						TakeOwnership: true,
					},
				},
			)
			assert.Nil(t, err)
			assert.Equal(t, releaseName, rel.Name)
			assert.Equal(t, "deployed", string(rel.Status))

			t.Logf("Successfully installed release: %s (v%d)", rel.Name, rel.Revision)

			// 5. Run Uninstall
			t.Logf("Attempting to uninstall release: %s", releaseName)
			err = cli.Uninstall(ctx, releaseName, &helmconfig.UninstallConfig{})
			assert.Nil(t, err)

			t.Log("Successfully uninstalled release")

			return ctx
		}).
		Feature()

	testenv.Test(t, f)
}

func TestUpgrade(t *testing.T) {
	os.Setenv("DEBUG", "0")

	f := features.New("Upgrade Release").
		Assess("Install and Upgrade nginx chart", func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
			logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

			cli, err := NewClient(c.Client().RESTConfig(), namespace, logger.Handler())
			assert.Nil(t, err)

			cli, err = cli.WithCache()
			assert.Nil(t, err)
			defer cli.Close()

			chartRef := "oci://registry-1.docker.io/bitnamicharts/nginx"
			releaseName := "test-upgrade"

			// Install initial version
			t.Log("Installing initial release")
			rel, err := cli.Install(
				ctx,
				releaseName,
				chartRef,
				&helmconfig.InstallConfig{
					ActionConfig: &helmconfig.ActionConfig{
						TakeOwnership: true,
						Values: map[string]interface{}{
							"replicaCount": 1,
						},
					},
				},
			)
			assert.Nil(t, err)
			assert.Equal(t, releaseName, rel.Name)
			assert.Equal(t, 1, rel.Revision)

			// Upgrade the release
			t.Log("Upgrading release")
			rel, err = cli.Upgrade(
				ctx,
				releaseName,
				chartRef,
				&helmconfig.UpgradeConfig{
					ActionConfig: &helmconfig.ActionConfig{
						TakeOwnership: true,
						Values: map[string]interface{}{
							"replicaCount": 2,
						},
					},
				},
			)
			assert.Nil(t, err)
			assert.Equal(t, releaseName, rel.Name)
			assert.Equal(t, 2, rel.Revision)

			// Cleanup
			err = cli.Uninstall(ctx, releaseName, &helmconfig.UninstallConfig{})
			assert.Nil(t, err)

			return ctx
		}).
		Feature()

	testenv.Test(t, f)
}

func TestRollback(t *testing.T) {
	os.Setenv("DEBUG", "0")

	f := features.New("Rollback Release").
		Assess("Install, Upgrade, and Rollback nginx chart", func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
			logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

			cli, err := NewClient(c.Client().RESTConfig(), namespace, logger.Handler())
			assert.Nil(t, err)

			cli, err = cli.WithCache()
			assert.Nil(t, err)
			defer cli.Close()

			chartRef := "oci://registry-1.docker.io/bitnamicharts/nginx"
			releaseName := "test-rollback"

			// Install initial version
			t.Log("Installing initial release")
			rel, err := cli.Install(
				ctx,
				releaseName,
				chartRef,
				&helmconfig.InstallConfig{
					ActionConfig: &helmconfig.ActionConfig{
						TakeOwnership: true,
						Values: map[string]interface{}{
							"replicaCount": 1,
						},
					},
				},
			)
			assert.Nil(t, err)
			assert.Equal(t, 1, rel.Revision)

			// Upgrade the release
			t.Log("Upgrading release")
			rel, err = cli.Upgrade(
				ctx,
				releaseName,
				chartRef,
				&helmconfig.UpgradeConfig{
					ActionConfig: &helmconfig.ActionConfig{
						TakeOwnership: true,
						Values: map[string]interface{}{
							"replicaCount": 3,
						},
					},
				},
			)
			assert.Nil(t, err)
			assert.Equal(t, 2, rel.Revision)

			// Rollback to revision 1
			t.Log("Rolling back to revision 1")
			rel, err = cli.Rollback(
				ctx,
				releaseName,
				&helmconfig.RollbackConfig{
					ReleaseVersion: 1,
				},
			)
			assert.Nil(t, err)
			assert.Equal(t, 3, rel.Revision) // Rollback creates a new revision

			// Cleanup
			err = cli.Uninstall(ctx, releaseName, &helmconfig.UninstallConfig{})
			assert.Nil(t, err)

			return ctx
		}).
		Feature()

	testenv.Test(t, f)
}

func TestGetRelease(t *testing.T) {
	os.Setenv("DEBUG", "0")

	f := features.New("Get Release").
		Assess("Install and Get nginx release", func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
			logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

			cli, err := NewClient(c.Client().RESTConfig(), namespace, logger.Handler())
			assert.Nil(t, err)

			cli, err = cli.WithCache()
			assert.Nil(t, err)
			defer cli.Close()

			chartRef := "oci://registry-1.docker.io/bitnamicharts/nginx"
			releaseName := "test-get-release"

			// Install a release
			t.Log("Installing release")
			rel, err := cli.Install(
				ctx,
				releaseName,
				chartRef,
				&helmconfig.InstallConfig{
					ActionConfig: &helmconfig.ActionConfig{
						TakeOwnership: true,
					},
				},
			)
			assert.Nil(t, err)

			// Get the release
			t.Log("Getting release")
			gotRel, err := cli.GetRelease(ctx, releaseName, &helmconfig.GetConfig{})
			assert.Nil(t, err)
			assert.NotNil(t, gotRel)
			assert.Equal(t, releaseName, gotRel.Name)
			assert.Equal(t, rel.Revision, gotRel.Revision)
			assert.Equal(t, "deployed", string(gotRel.Status))

			// Test getting non-existent release
			t.Log("Getting non-existent release")
			notFound, err := cli.GetRelease(ctx, "does-not-exist", &helmconfig.GetConfig{})
			assert.Nil(t, err)
			assert.Nil(t, notFound)

			// Cleanup
			err = cli.Uninstall(ctx, releaseName, &helmconfig.UninstallConfig{})
			assert.Nil(t, err)

			return ctx
		}).
		Feature()

	testenv.Test(t, f)
}

func TestListReleases(t *testing.T) {
	os.Setenv("DEBUG", "0")

	f := features.New("List Releases").
		Assess("Install multiple releases and list them", func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
			logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

			cli, err := NewClient(c.Client().RESTConfig(), namespace, logger.Handler())
			assert.Nil(t, err)

			cli, err = cli.WithCache()
			assert.Nil(t, err)
			defer cli.Close()

			chartRef := "oci://registry-1.docker.io/bitnamicharts/nginx"

			// Install multiple releases
			releases := []string{"test-list-1", "test-list-2"}
			for _, releaseName := range releases {
				t.Logf("Installing release: %s", releaseName)
				_, err := cli.Install(
					ctx,
					releaseName,
					chartRef,
					&helmconfig.InstallConfig{
						ActionConfig: &helmconfig.ActionConfig{
							TakeOwnership: true,
						},
					},
				)
				assert.Nil(t, err)
			}

			time.Sleep(5 * time.Second) // Wait for releases to be fully registered

			// List all releases
			t.Log("Listing releases")
			list, err := cli.ListReleases(ctx, &helmconfig.ListConfig{
				All:       true, // List all releases regardless of limit/offset
				StateMask: helmconfig.ListDeployed,
			})
			assert.Nil(t, err)
			t.Logf("Found %d releases", len(list))
			for _, rel := range list {
				t.Logf("  - %s (namespace: %s, status: %s)", rel.Name, rel.Namespace, rel.Status)
			}
			assert.GreaterOrEqual(t, len(list), 2)

			// Verify our releases are in the list
			releaseNames := make(map[string]bool)
			for _, rel := range list {
				releaseNames[rel.Name] = true
			}
			for _, name := range releases {
				assert.True(t, releaseNames[name], "Release %s should be in the list", name)
			}

			// Cleanup
			for _, releaseName := range releases {
				err = cli.Uninstall(ctx, releaseName, &helmconfig.UninstallConfig{})
				assert.Nil(t, err)
			}

			return ctx
		}).
		Feature()

	testenv.Test(t, f)
}

func TestInstallAndUninstall_Repo(t *testing.T) {
	os.Setenv("DEBUG", "0")

	f := features.New("Install and Uninstall from Repo").
		Assess("Install and Uninstall ingress-nginx from repo", func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
			// 1. Setup Logger
			logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

			// 2. Initialize Client
			cli, err := NewClient(c.Client().RESTConfig(), namespace, logger.Handler())
			assert.Nil(t, err)

			cli, err = cli.WithCache()
			assert.Nil(t, err)
			defer cli.Close()

			// 3. Define Chart Details - Using official Kubernetes ingress-nginx repo
			chartRef := "https://kubernetes.github.io/ingress-nginx"
			releaseName := "test-release-repo"
			chartName := "ingress-nginx"

			t.Logf("Attempting to install chart %s from repo: %s", chartName, chartRef)

			// 4. Run Install
			rel, err := cli.Install(
				ctx,
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
			assert.Nil(t, err)
			assert.Equal(t, releaseName, rel.Name)
			assert.Equal(t, "deployed", string(rel.Status))

			t.Logf("Successfully installed release: %s (v%d)", rel.Name, rel.Revision)

			// 5. Run Uninstall
			t.Logf("Attempting to uninstall release: %s", releaseName)
			err = cli.Uninstall(ctx, releaseName, &helmconfig.UninstallConfig{})
			assert.Nil(t, err)

			t.Log("Successfully uninstalled release")

			return ctx
		}).
		Feature()

	testenv.Test(t, f)
}

func TestInstallAndUninstall_TGZ(t *testing.T) {
	os.Setenv("DEBUG", "0")

	f := features.New("Install and Uninstall TGZ Chart").
		Assess("Install and Uninstall from .tgz archive", func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
			// 1. Setup Logger
			logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

			// 2. Initialize Client
			cli, err := NewClient(c.Client().RESTConfig(), namespace, logger.Handler())
			assert.Nil(t, err)

			cli, err = cli.WithCache()
			assert.Nil(t, err)
			defer cli.Close()

			// 3. Define Chart Details - Direct .tgz URL
			// Using official Kubernetes ingress-nginx chart archive
			chartRef := "https://github.com/kubernetes/ingress-nginx/releases/download/helm-chart-4.10.0/ingress-nginx-4.10.0.tgz"
			releaseName := "test-release-tgz"

			t.Logf("Attempting to install chart from tgz: %s", chartRef)

			// 4. Run Install
			rel, err := cli.Install(
				ctx,
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
			assert.Nil(t, err)
			assert.Equal(t, releaseName, rel.Name)
			assert.Equal(t, "deployed", string(rel.Status))

			t.Logf("Successfully installed release: %s (v%d)", rel.Name, rel.Revision)

			// 5. Run Uninstall
			t.Logf("Attempting to uninstall release: %s", releaseName)
			err = cli.Uninstall(ctx, releaseName, &helmconfig.UninstallConfig{})
			assert.Nil(t, err)

			t.Log("Successfully uninstalled release")

			return ctx
		}).
		Feature()

	testenv.Test(t, f)
}
