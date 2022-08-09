package cmd

import (
	"github.com/spf13/cobra"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/envoyproxy/gateway/internal/envoygateway/config"
	gatewayapisvc "github.com/envoyproxy/gateway/internal/gatewayapi/service"
	infrasvc "github.com/envoyproxy/gateway/internal/infrastructure/service"
	"github.com/envoyproxy/gateway/internal/message"
	providersvc "github.com/envoyproxy/gateway/internal/provider/service"
	xdsserversvc "github.com/envoyproxy/gateway/internal/xds/server/service"
	xdstranslatorsvc "github.com/envoyproxy/gateway/internal/xds/translator/service"
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

	if err := setupServices(cfg); err != nil {
		return err
	}

	return nil
}

// getConfig gets the Server configuration
func getConfig() (*config.Server, error) {
	// Initialize with default config parameters.
	cfg, err := config.NewDefaultServer()
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
	return cfg, nil
}

// setupServices starts all the services required for the Envoy Gateway to
// fulfill its tasks.
func setupServices(cfg *config.Server) error {
	// TODO - Setup a Config Manager
	// https://github.com/envoyproxy/gateway/issues/43
	ctx := ctrl.SetupSignalHandler()

	pResources := new(message.ProviderResources)
	// Start the Provider Service
	// It fetches the resources from the configured provider type
	// and publishes it
	providerSvc := &providersvc.Service{
		Server:            *cfg,
		ProviderResources: pResources,
	}
	if err := providerSvc.Start(ctx); err != nil {
		return err
	}

	xdsIR := new(message.XdsIR)
	infraIR := new(message.InfraIR)
	// Start the GatewayAPI Translator Service
	// It subscribes to the provider resources, translates it to xDS IR
	// and infra IR resources and publishes them.
	gwSvc := &gatewayapisvc.Service{
		Server:            *cfg,
		ProviderResources: pResources,
		XdsIR:             xdsIR,
		InfraIR:           infraIR,
	}
	if err := gwSvc.Start(ctx); err != nil {
		return err
	}

	xResources := new(message.XdsResources)
	// Start the Xds Translator Service
	// It subscribes to the xdsIR, translates it into xds Resources and publishes it.
	xdsTranslatorSvc := &xdstranslatorsvc.Service{
		Server:       *cfg,
		XdsIR:        xdsIR,
		XdsResources: xResources,
	}
	if err := xdsTranslatorSvc.Start(ctx); err != nil {
		return err
	}

	// Start the Infra Manager Service
	// It subscribes to the infraIR, translates it into Envoy Proxy infrastructure
	// resources such as K8s deployment and services.
	infraSvc := &infrasvc.Service{
		Server:  *cfg,
		InfraIR: infraIR,
	}
	if err := infraSvc.Start(ctx); err != nil {
		return err
	}

	// Start the xDS Server
	// It subscribes to the xds Resources and configures the remote Envoy Proxy
	// via the xDS Protocol
	xdsServerSvc := &xdsserversvc.Service{
		Server:       *cfg,
		XdsResources: xResources,
	}
	if err := xdsServerSvc.Start(ctx); err != nil {
		return err
	}

	return nil
}
