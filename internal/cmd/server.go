// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package cmd

import (
	"github.com/spf13/cobra"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/envoyproxy/gateway/internal/envoygateway/config"
	extensionregistry "github.com/envoyproxy/gateway/internal/extension/registry"
	gatewayapirunner "github.com/envoyproxy/gateway/internal/gatewayapi/runner"
	ratelimitrunner "github.com/envoyproxy/gateway/internal/globalratelimit/runner"
	infrarunner "github.com/envoyproxy/gateway/internal/infrastructure/runner"
	"github.com/envoyproxy/gateway/internal/message"
	providerrunner "github.com/envoyproxy/gateway/internal/provider/runner"
	xdsserverrunner "github.com/envoyproxy/gateway/internal/xds/server/runner"
	xdstranslatorrunner "github.com/envoyproxy/gateway/internal/xds/translator/runner"
)

var (
	// cfgPath is the path to the EnvoyGateway configuration file.
	cfgPath string
)

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

	log := cfg.Logger

	// Read the config file.
	if cfgPath == "" {
		// Use default config parameters
		log.Info("No config file provided, using default parameters")
	} else {
		// Load the config file.
		eg, err := config.Decode(cfgPath)
		if err != nil {
			log.Error(err, "failed to decode config file", "name", cfgPath)
			return nil, err
		}
		// Set defaults for unset fields
		eg.SetDefaults()
		cfg.EnvoyGateway = eg
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return cfg, nil
}

// setupRunners starts all the runners required for the Envoy Gateway to
// fulfill its tasks.
func setupRunners(cfg *config.Server) error {
	// TODO - Setup a Config Manager
	// https://github.com/envoyproxy/gateway/issues/43
	ctx := ctrl.SetupSignalHandler()

	// Setup the Extension Manager
	extMgr, err := extensionregistry.NewManager(cfg)
	if err != nil {
		return err
	}

	pResources := new(message.ProviderResources)
	// Start the Provider Service
	// It fetches the resources from the configured provider type
	// and publishes it
	providerRunner := providerrunner.New(&providerrunner.Config{
		Server:            *cfg,
		ProviderResources: pResources,
	})
	if err := providerRunner.Start(ctx); err != nil {
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
	if err := gwRunner.Start(ctx); err != nil {
		return err
	}

	xds := new(message.Xds)
	// Start the Xds Translator Service
	// It subscribes to the xdsIR, translates it into xds Resources and publishes it.
	xdsTranslatorRunner := xdstranslatorrunner.New(&xdstranslatorrunner.Config{
		Server:           *cfg,
		XdsIR:            xdsIR,
		Xds:              xds,
		ExtensionManager: extMgr,
	})
	if err := xdsTranslatorRunner.Start(ctx); err != nil {
		return err
	}

	rateLimitInfraIR := new(message.RateLimitInfraIR)
	// Start the Infra Manager Runner
	// It subscribes to the infraIR, translates it into Envoy Proxy infrastructure
	// resources such as K8s deployment and services.
	infraRunner := infrarunner.New(&infrarunner.Config{
		Server:           *cfg,
		InfraIR:          infraIR,
		RateLimitInfraIR: rateLimitInfraIR,
	})
	if err := infraRunner.Start(ctx); err != nil {
		return err
	}

	// Start the xDS Server
	// It subscribes to the xds Resources and configures the remote Envoy Proxy
	// via the xDS Protocol
	xdsServerRunner := xdsserverrunner.New(&xdsserverrunner.Config{
		Server: *cfg,
		Xds:    xds,
	})
	if err := xdsServerRunner.Start(ctx); err != nil {
		return err
	}

	// Start the global rateLimit runner if it has been enabled through the config
	if cfg.EnvoyGateway.RateLimit != nil {
		// Start the Global RateLimit Runner
		// It subscribes to the xds Resources and translates it to Envoy Ratelimit Service
		// infrastructure and configuration.
		rateLimitRunner := ratelimitrunner.New(&ratelimitrunner.Config{
			Server:           *cfg,
			XdsIR:            xdsIR,
			RateLimitInfraIR: rateLimitInfraIR,
		})
		if err := rateLimitRunner.Start(ctx); err != nil {
			return err
		}
	}

	// Wait until done
	<-ctx.Done()
	// Close messages
	pResources.Close()
	xdsIR.Close()
	infraIR.Close()
	rateLimitInfraIR.Close()
	xds.Close()

	cfg.Logger.Info("shutting down")

	// Close connections to extension services
	if mgr, ok := extMgr.(*extensionregistry.Manager); ok {
		mgr.CleanupHookConns()
	}

	return nil
}
