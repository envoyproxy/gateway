// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package cmd

import (
	"github.com/spf13/cobra"
	ctrl "sigs.k8s.io/controller-runtime"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/admin"
	"github.com/envoyproxy/gateway/internal/envoygateway/config"
	extensionregistry "github.com/envoyproxy/gateway/internal/extension/registry"
	"github.com/envoyproxy/gateway/internal/extension/types"
	gatewayapirunner "github.com/envoyproxy/gateway/internal/gatewayapi/runner"
	ratelimitrunner "github.com/envoyproxy/gateway/internal/globalratelimit/runner"
	infrarunner "github.com/envoyproxy/gateway/internal/infrastructure/runner"
	"github.com/envoyproxy/gateway/internal/logging"
	"github.com/envoyproxy/gateway/internal/message"
	"github.com/envoyproxy/gateway/internal/metrics"
	providerrunner "github.com/envoyproxy/gateway/internal/provider/runner"
	xdsrunner "github.com/envoyproxy/gateway/internal/xds/runner"
)

// cfgPath is the path to the EnvoyGateway configuration file.
var cfgPath string

// getServerCommand returns the server cobra command to be executed.
func getServerCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "server",
		Aliases: []string{"serve"},
		Short:   "Serve Envoy Gateway",
		RunE: func(cmd *cobra.Command, args []string) error {
			return server()
		},
	}
	cmd.PersistentFlags().StringVarP(&cfgPath, "config-path", "c", "",
		"The path to the configuration file.")

	return cmd
}

// server serves Envoy Gateway.
func server() error {
	cfg, err := getConfig()
	if err != nil {
		return err
	}

	// Init eg admin servers.
	if err := admin.Init(cfg); err != nil {
		return err
	}
	// Init eg metrics servers.
	if err := metrics.Init(cfg); err != nil {
		return err
	}

	// init eg runners.
	if err := setupRunners(cfg); err != nil {
		return err
	}

	return nil
}

// getConfig gets the Server configuration
func getConfig() (*config.Server, error) {
	return getConfigByPath(cfgPath)
}

// make `cfgPath` an argument to test it without polluting the global var
func getConfigByPath(cfgPath string) (*config.Server, error) {
	// Initialize with default config parameters.
	cfg, err := config.New()
	if err != nil {
		return nil, err
	}

	logger := cfg.Logger

	// Read the config file.
	if cfgPath == "" {
		// Use default config parameters
		logger.Info("No config file provided, using default parameters")
	} else {
		// Load the config file.
		eg, err := config.Decode(cfgPath)
		if err != nil {
			logger.Error(err, "failed to decode config file", "name", cfgPath)
			return nil, err
		}
		// Set defaults for unset fields
		eg.SetEnvoyGatewayDefaults()
		cfg.EnvoyGateway = eg
		// update cfg logger
		eg.Logging.SetEnvoyGatewayLoggingDefaults()
		cfg.Logger = logging.NewLogger(eg.Logging)
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return cfg, nil
}

// setupRunners starts all the runners required for the Envoy Gateway to
// fulfill its tasks.
func setupRunners(cfg *config.Server) (err error) {
	// TODO - Setup a Config Manager
	// https://github.com/envoyproxy/gateway/issues/43
	ctx := ctrl.SetupSignalHandler()

	// Setup the Extension Manager
	var extMgr types.Manager
	if cfg.EnvoyGateway.Provider.Type == egv1a1.ProviderTypeKubernetes {
		extMgr, err = extensionregistry.NewManager(cfg)
		if err != nil {
			return err
		}
	}

	pResources := new(message.ProviderResources)
	// Start the Provider Service
	// It fetches the resources from the configured provider type
	// and publishes it.
	// It also subscribes to status resources and once it receives
	// a status resource back, it writes it out.
	providerRunner := providerrunner.New(&providerrunner.Config{
		Server:            *cfg,
		ProviderResources: pResources,
	})
	if err = providerRunner.Start(ctx); err != nil {
		return err
	}

	xdsIR := new(message.XdsIR)
	infraIR := new(message.InfraIR)
	// Start the GatewayAPI Translator Runner
	// It subscribes to the provider resources, translates it to xDS IR
	// and infra IR resources and publishes them.
	gwRunner := gatewayapirunner.New(&gatewayapirunner.Config{
		Server:            *cfg,
		ProviderResources: pResources,
		XdsIR:             xdsIR,
		InfraIR:           infraIR,
		ExtensionManager:  extMgr,
	})
	if err = gwRunner.Start(ctx); err != nil {
		return err
	}

	// Start the Xds Service
	// It subscribes to the xdsIR, translates it into xds Resources and
	// updates the xds control plane cache.
	// It also computes the EnvoyPatchPolicy statuses and publishes it.
	xdsRunner := xdsrunner.New(&xdsrunner.Config{
		Server:            *cfg,
		XdsIR:             xdsIR,
		ExtensionManager:  extMgr,
		ProviderResources: pResources,
	})
	if err = xdsRunner.Start(ctx); err != nil {
		return err
	}

	// Start the Infra Manager Runner
	// It subscribes to the infraIR, translates it into Envoy Proxy infrastructure
	// resources such as K8s deployment and services.
	infraRunner := infrarunner.New(&infrarunner.Config{
		Server:  *cfg,
		InfraIR: infraIR,
	})
	if err = infraRunner.Start(ctx); err != nil {
		return err
	}

	// Start the global rateLimit if it has been enabled through the config
	if cfg.EnvoyGateway.RateLimit != nil {
		// Start the Global RateLimit xDS Server
		// It subscribes to the xds Resources and translates it to Envoy Ratelimit configuration.
		rateLimitRunner := ratelimitrunner.New(&ratelimitrunner.Config{
			Server: *cfg,
			XdsIR:  xdsIR,
		})
		if err = rateLimitRunner.Start(ctx); err != nil {
			return err
		}
	}

	// Wait until done
	<-ctx.Done()
	// Close messages
	pResources.Close()
	xdsIR.Close()
	infraIR.Close()

	cfg.Logger.Info("shutting down")

	if extMgr != nil {
		// Close connections to extension services
		if mgr, ok := extMgr.(*extensionregistry.Manager); ok {
			mgr.CleanupHookConns()
		}
	}

	return nil
}
