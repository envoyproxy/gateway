package cmd

import (
	"os"

	"github.com/envoyproxy/gateway/api/config/v1alpha1"
	"github.com/envoyproxy/gateway/pkg/envoygateway/config"
	"github.com/go-logr/logr"
	"github.com/spf13/cobra"

	"github.com/envoyproxy/gateway/pkg/provider"
)

var (
	// cfgPath is the path to the EnvoyGateway configuration file.
	cfgPath string
)

// getServerCommand returns the server cobra command to be executed.
func getServerCommand(log logr.Logger) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "server",
		Aliases: []string{"serve"},
		Short:   "Serve Envoy Gateway",
		RunE: func(cmd *cobra.Command, args []string) error {
			return server(log)
		},
	}
	cmd.PersistentFlags().StringVarP(&cfgPath, "config-path", "c", "",
		"The path to the configuration file.")

	return cmd
}

// server serves Envoy Gateway using log as the logger.
func server(log logr.Logger) error {
	cfg := new(config.Server)
	cfg.Logger = log
	// Read the config file.
	if cfgPath == "" {
		// Use default config parameters
		log.Info("No config file provided, using default parameters")
		cfg.EnvoyGateway = v1alpha1.DefaultEnvoyGateway()
	} else {
		// Load the config file.
		eg, err := config.Decode(cfgPath)
		if err != nil {
			log.Error(err, "failed to decode config file", "name", cfgPath)
			os.Exit(1)
		}
		// Set defaults for unset fields
		eg.SetDefaults()
		cfg.EnvoyGateway = eg
	}

	if err := provider.Start(cfg); err != nil {
		return err
	}
	return nil
}
