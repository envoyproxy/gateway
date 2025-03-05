// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package cmd

import (
	"context"

	"github.com/spf13/cobra"
	ctrl "sigs.k8s.io/controller-runtime"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/admin"
	"github.com/envoyproxy/gateway/internal/envoygateway/config"
	"github.com/envoyproxy/gateway/internal/envoygateway/config/loader"
	extensionregistry "github.com/envoyproxy/gateway/internal/extension/registry"
	"github.com/envoyproxy/gateway/internal/extension/types"
	gatewayapirunner "github.com/envoyproxy/gateway/internal/gatewayapi/runner"
	ratelimitrunner "github.com/envoyproxy/gateway/internal/globalratelimit/runner"
	infrarunner "github.com/envoyproxy/gateway/internal/infrastructure/runner"
	"github.com/envoyproxy/gateway/internal/logging"
	"github.com/envoyproxy/gateway/internal/message"
	"github.com/envoyproxy/gateway/internal/metrics"
	providerrunner "github.com/envoyproxy/gateway/internal/provider/runner"
	xdsserverrunner "github.com/envoyproxy/gateway/internal/xds/server/runner"
	xdstranslatorrunner "github.com/envoyproxy/gateway/internal/xds/translator/runner"
)

// cfgPath is the path to the EnvoyGateway configuration file.
var cfgPath string

// GetServerCommand returns the server cobra command to be executed.
func GetServerCommand() *cobra.Command {
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

	ctx := ctrl.SetupSignalHandler()
	hook := func(c context.Context, cfg *config.Server) error {
		cfg.Logger.Info("Setup runners")
		if err := setupRunners(c, cfg); err != nil {
			cfg.Logger.Error(err, "failed to setup runners")
			return err
		}
		return nil
	}
	l := loader.New(cfgPath, cfg, hook)
	if err := l.Start(ctx); err != nil {
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

	// Wait exit signal
	<-ctx.Done()

	cfg.Logger.Info("shutting down")

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
func setupRunners(ctx context.Context, cfg *config.Server) (err error) {
	// The Elected channel is used to block the tasks that are waiting for the leader to be elected.
	// It will be closed once the leader is elected in the controller manager.
	cfg.Elected = make(chan struct{})

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

	xds := new(message.Xds)
	// Start the Xds Translator Service
	// It subscribes to the xdsIR, translates it into xds Resources and publishes it.
	// It also computes the EnvoyPatchPolicy statuses and publishes it.
	xdsTranslatorRunner := xdstranslatorrunner.New(&xdstranslatorrunner.Config{
		Server:            *cfg,
		XdsIR:             xdsIR,
		Xds:               xds,
		ExtensionManager:  extMgr,
		ProviderResources: pResources,
	})
	if err = xdsTranslatorRunner.Start(ctx); err != nil {
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

	// Start the xDS Server
	// It subscribes to the xds Resources and configures the remote Envoy Proxy
	// via the xDS Protocol.
	xdsServerRunner := xdsserverrunner.New(&xdsserverrunner.Config{
		Server: *cfg,
		Xds:    xds,
	})
	if err = xdsServerRunner.Start(ctx); err != nil {
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
	xds.Close()

	cfg.Logger.Info("runners are shutting down")

	if extMgr != nil {
		// Close connections to extension services
		if mgr, ok := extMgr.(*extensionregistry.Manager); ok {
			mgr.CleanupHookConns()
		}
	}

	return nil
}
