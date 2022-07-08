package cmd

import (
	"fmt"
	"os"

	"github.com/envoyproxy/gateway/api/config/v1alpha1"
	"github.com/envoyproxy/gateway/pkg/envoygateway"
	"github.com/envoyproxy/gateway/pkg/provider/kubernetes"

	"github.com/go-logr/logr"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	ctrl "sigs.k8s.io/controller-runtime"
)

var configPath string

// getServerCommand returns the server cobra command to be executed.
func getServerCommand(log logr.Logger) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "server",
		Aliases: []string{"serve"},
		Short:   "Serve Envoy Gateway",
		RunE: func(cmd *cobra.Command, args []string) error {
			return serve(log)
		},
	}
	cmd.PersistentFlags().StringVarP(&configPath, "config-path", "c", "/etc/envoy-gateway/config.yaml",
		"The path to the configuration file.")

	return cmd
}

// serve Envoy Gateway using log as the logger.
func serve(log logr.Logger) error {
	cfg := &envoygateway.Config{Logger: log}

	// Load the config file.
	data, err := os.ReadFile(configPath)
	if err != nil {
		log.Error(err, "failed to read config file", "name", configPath)
		os.Exit(1)
	} else {
		// Decode the config file.
		decoder := serializer.NewCodecFactory(envoygateway.GetScheme()).UniversalDeserializer()
		obj, gvk, err := decoder.Decode(data, nil, nil)
		if err != nil {
			log.Error(err, "failed to decode config file", "name", configPath)
			os.Exit(1)
		}
		// Figure out the resource type from the Group|Version|Kind.
		if gvk.Group != v1alpha1.GroupVersion.Group &&
			gvk.Version != v1alpha1.GroupVersion.Version &&
			gvk.Kind != v1alpha1.KindEnvoyGateway {
			log.Error(err, "invalid EnvoyGateway spec for config file", "name", configPath)
			os.Exit(1)
		} else {
			// Attempt to cast the object.
			eg, ok := obj.(*v1alpha1.EnvoyGateway)
			if !ok {
				log.Error(err, "failed to cast config file", "name", configPath)
				os.Exit(1)
			}
			// Set defaults for unset fields
			eg.SetDefaults()
			cfg.EnvoyGateway = eg
		}
	}

	switch cfg.EnvoyGateway.Provider.Type {
	case v1alpha1.ProviderTypeKubernetes:
		log.Info("Using provider", "type", v1alpha1.ProviderTypeKubernetes)
		provider, err := kubernetes.New(cfg)
		if err != nil {
			return fmt.Errorf("failed to create provider %s", v1alpha1.ProviderTypeKubernetes)
		}
		if err := provider.Start(ctrl.SetupSignalHandler()); err != nil {
			return fmt.Errorf("failed to serve provider %s", v1alpha1.ProviderTypeKubernetes)
		}
	default:
		// Unsupported provider type.
		return fmt.Errorf("unsupported provider type %v", cfg.EnvoyGateway.Provider.Type)
	}
	return nil
}
