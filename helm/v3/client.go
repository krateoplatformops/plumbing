package helm

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"

	helmconfig "github.com/krateoplatformops/plumbing/helm"
	"github.com/krateoplatformops/plumbing/helm/getter/cache"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/storage/driver"
	"k8s.io/client-go/rest"
)

var _ helmconfig.Client = (*client)(nil)

type client struct {
	settings     *cli.EnvSettings
	actionConfig *action.Configuration
	namespace    string
	cache        *cache.DiskCache
	debugLog     action.DebugLog
}

type ClientOption func(*client) error

func NewClient(cfg *rest.Config, opts ...ClientOption) (*client, error) {
	settings := cli.New()

	// Default configuration
	c := &client{
		settings:  settings,
		namespace: settings.Namespace(),
		debugLog: func(format string, v ...interface{}) {
			// Default to discard if no logger is provided
		},
	}

	// Apply functional options
	for _, opt := range opts {
		if err := opt(c); err != nil {
			return nil, err
		}
	}

	// Initialize Helm action configuration
	actionConfig := new(action.Configuration)
	driver := os.Getenv("HELM_DRIVER")

	// Discard logger by default
	debugLog := func(format string, v ...interface{}) {
		slog.Debug(fmt.Sprintf(format, v...))
	}

	clientGetter := NewRESTClientGetter(c.namespace, nil, cfg)
	if err := actionConfig.Init(clientGetter, c.namespace, driver, debugLog); err != nil {
		return nil, fmt.Errorf("failed to init action config: %w", err)
	}

	c.actionConfig = actionConfig
	return c, nil
}

// WithNamespace sets a custom namespace for the client.
func WithNamespace(ns string) ClientOption {
	return func(c *client) error {
		c.namespace = ns
		c.settings.SetNamespace(ns)
		return nil
	}
}

// WithCache initializes a disk cache with the provided options.
func WithCache(opts ...cache.Option) ClientOption {
	return func(c *client) error {
		dCache, err := cache.NewDiskCache(opts...)
		if err != nil {
			return fmt.Errorf("failed to create cache: %w", err)
		}
		c.cache = dCache
		return nil
	}
}

// WithLogger configures a custom slog handler for the client.
func WithLogger(handler slog.Handler) ClientOption {
	return func(c *client) error {
		if handler == nil {
			// Fallback to discard if nil is passed
			handler = slog.DiscardHandler
		}

		logger := slog.New(handler)

		// Map the slog.Logger to the Helm debug function signature
		c.debugLog = func(format string, v ...interface{}) {
			logger.Debug(fmt.Sprintf(format, v...))
		}

		return nil
	}
}

func (c *client) Close() error {
	if c.cache != nil {
		c.cache.Stop()
	}
	return nil
}

func (c *client) Install(ctx context.Context, releaseName string, chartRef string, cfg *helmconfig.InstallConfig) (*helmconfig.Release, error) {
	installClient := action.NewInstall(c.actionConfig)
	applyInstallConfig(installClient, releaseName, c.namespace, cfg)

	chart, err := c.loadChart(ctx, chartRef, c.buildGetterOpts(cfg.ActionConfig))
	if err != nil {
		return nil, fmt.Errorf("failed to load chart: %w", err)
	}

	if err := checkChartType(chart); err != nil {
		return nil, fmt.Errorf("chart type check failed: %w", err)
	}

	chart, err = c.checkDependencies(ctx, chart, chartRef, cfg.ActionConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to check dependencies: %w", err)
	}

	rel, err := installClient.RunWithContext(ctx, chart, cfg.Values)
	if err != nil {
		return nil, fmt.Errorf("install failed: %w", err)
	}

	return toWrapperRelease(rel), nil
}

func (c *client) Upgrade(ctx context.Context, releaseName, chartRef string, cfg *helmconfig.UpgradeConfig) (*helmconfig.Release, error) {
	upgradeClient := action.NewUpgrade(c.actionConfig)
	applyUpgradeConfig(upgradeClient, c.namespace, cfg)

	chart, err := c.loadChart(ctx, chartRef, c.buildGetterOpts(cfg.ActionConfig))
	if err != nil {
		return nil, fmt.Errorf("failed to load chart: %w", err)
	}

	if err := checkChartType(chart); err != nil {
		return nil, fmt.Errorf("chart type check failed: %w", err)
	}

	chart, err = c.checkDependencies(ctx, chart, chartRef, cfg.ActionConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to check dependencies: %w", err)
	}

	rel, err := upgradeClient.RunWithContext(ctx, releaseName, chart, cfg.Values)
	if err != nil {
		return nil, fmt.Errorf("upgrade failed: %w", err)
	}
	return toWrapperRelease(rel), nil
}

func (c *client) Uninstall(ctx context.Context, releaseName string, cfg *helmconfig.UninstallConfig) error {
	cmd := action.NewUninstall(c.actionConfig)
	applyUninstallConfig(cmd, cfg)

	_, err := cmd.Run(releaseName)
	if err != nil {
		return fmt.Errorf("uninstall failed: %w", err)
	}

	return nil
}

func (c *client) Rollback(ctx context.Context, releaseName string, cfg *helmconfig.RollbackConfig) (*helmconfig.Release, error) {
	rollbackClient := action.NewRollback(c.actionConfig)
	applyRollbackConfig(rollbackClient, cfg)

	err := rollbackClient.Run(releaseName)
	if err != nil {
		return nil, fmt.Errorf("rollback failed: %w", err)
	}

	rel, err := c.GetRelease(ctx, releaseName, &helmconfig.GetConfig{})
	if err != nil {
		return nil, fmt.Errorf("failed to get release after rollback: %w", err)
	}

	return rel, nil
}

func (c *client) GetRelease(ctx context.Context, releaseName string, cfg *helmconfig.GetConfig) (*helmconfig.Release, error) {
	getClient := action.NewGet(c.actionConfig)
	applyGetConfig(getClient, cfg)

	rel, err := getClient.Run(releaseName)
	if err != nil {
		if errors.Is(err, driver.ErrReleaseNotFound) {
			return nil, nil // Not found is not an error for GetRelease
		}
		return nil, fmt.Errorf("get release failed: %w", err)
	}

	return toWrapperRelease(rel), nil
}

func (c *client) ListReleases(ctx context.Context, cfg *helmconfig.ListConfig) ([]*helmconfig.Release, error) {
	listClient := action.NewList(c.actionConfig)
	applyListConfig(listClient, cfg)

	helmReleases, err := listClient.Run()
	if err != nil {
		return nil, fmt.Errorf("list releases failed: %w", err)
	}

	var releases []*helmconfig.Release
	for _, rel := range helmReleases {
		releases = append(releases, toWrapperRelease(rel))
	}

	return releases, nil
}
