Component Manager Design
===================

## Motivation

[Issue 43][issue_43] specifies the need for lifecycle management of Envoy Gateway system components that run as Go
routines.

???
The Component Manager can decide which services, e.g. xDS server to instantiate based on a provided
configuration. For example, if the xDs server configuration parameter is not included in the provided configuration, the
xDS server component will not be instantiated). This approach adds optionality of running all components as one process
if needed (mainly useful for standalone VM controller cases or app dev local dev workflow testing) or as separate
processes/containers in production.
???

## Goals

* Gracefully manage the software services that comprise Envoy Gateway.

## Non-Goals

* TODO

## Proposal

It was [decided][issue_60] to use the Runnable interface for providing goroutine management. TODO

A `Manager` is created....

A `serve` argument is introduced to the `envoy-gateway` command. The `serve` argument encapsulates the management of
supported Envoy Gateway services, e.g. xDS Server, Provisioner, etc.
```go
// gateway/internal/cmd/root.go

package cmd

import (
	"github.com/spf13/cobra"

	"github.com/envoyproxy/gateway/internal/cmd/serve"
)

// GetRootCommand returns the root cobra command to be executed
// by main.
func GetRootCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "envoy-gateway",
		Short: "Manages Envoy Proxy as a standalone or Kubernetes-based application gateway",
	}

	cmd.AddCommand(serve.NewCommand())

	return cmd
}
```

The serve package:
- Exposes EG config.
- Processes, validates, etc. command line flags and config file.
- Starts EG services.

```go
// gateway/internal/cmd/server/server.go

package serve

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	restclient "k8s.io/client-go/rest"

	"github.com/envoyproxy/gateway/apis/config/v1alpha1"
	"github.com/envoyproxy/gateway/apis/config/v1alpha1/validation"
)

// NewCommand allows the serve package to be run as a Cobra Command.
func NewCommand() *cobra.Command {
	flags := NewGatewayFlags()

	gatewayCfg, err := NewGatewayConfig()
	if err != nil {
		os.Exit(1)
	}

	cmd := &cobra.Command{
		Use:   "server",
		Short: "Server serves Envoy Gateway services, e.g. xDS Server, Provisioner, etc.",
		RunE: func(cmd *cobra.Command, args []string) error {
			// validate the initial GatewayFlags
			if err := ValidateGatewayFlags(flags); err != nil {
				return fmt.Errorf("failed to validate envoy gateway flags: %w", err)
			}

			// load envoy gateway config file, if provided.
			if cfgFile := flags.ConfigFile; len(cfgFile) > 0 {
				gatewayCfg, err = loadConfigFile(cfgFile)
				if err != nil {
					return fmt.Errorf("failed to load envoy gateway config file, error: %w, path: %s", err, cfgFile)
				}
			}

			// Validate the local configuration (command line flags + config file).
			if err := validation.ValidateGatewayConfig(gatewayCfg); err != nil {
				return fmt.Errorf("failed to validate envoy gateway configuration, error: %w, path: %s",
					err, gatewayCfg)
			}

			// construct a GatewayServer from flags + config.
			gatewayServer := &GatewayServer{
				GatewayFlags:  *flags,
				GatewayConfig: *gatewayCfg,
			}

			// run envoy gateway
			return run(context.TODO, gatewayServer)
		},
	}

	addFlags(flags)

	return cmd
}

// Load the config file specified by name.
func loadConfigFile(name string) (*v1alpha1.GatewayConfig, error) {
	// Read config file from disk.
	// Load and return the GatewayConfig or an error if a configuration can't be loaded.
}

// Run the specified GatewayServer with the provided context.
func run(ctx context.Context, s *GatewayServer, gatewayDeps *gateway.Dependencies) error {
	// validate the initial GatewayServer.
	if err := options.ValidateGatewayServer(s); err != nil {
		return err
	}

	// If KubeConfig is nil then consider Envoy Gateway in "standalone" mode, e.g. virtual machine
	// and don't create k8s client.

	clientConfig, err := buildGatewayClientConfig(ctx, s)
	if err != nil {
		return err
	}	
	
	return nil
}

// buildGatewayClientConfig constructs the appropriate client config for Envoy Gateway.
func buildGatewayClientConfig(ctx context.Context, s *GatewayServer) (*restclient.Config, func(), error) {
	if s.KubeConfig != "" {
		return clientcmd.BuildConfigFromFlags("", s.KubeConfig)
	}
	return rest.InClusterConfig()
}
```

A subset of the Envoy Gateway's configuration parameters may be set by a configuration file, as a substitute for
command-line flags. Providing parameters via a config file is the recommended approach as it simplifies deployment
and configuration management.
```go
// gateway/internal/cmd/serve/options.go

package serve

import (
	"github.com/spf13/pflag"

	"github.com/envoyproxy/gateway/apis/config/v1alpha1"
	"github.com/envoyproxy/gateway/apis/config/validation"
	"github.com/envoyproxy/gateway/internal/cmd/serve"
)

// GatewayFlags contains Envoy Gateway configuration flags.
type GatewayFlags struct {
	// KubeConfig is the path to the kubeconfig file.
	KubeConfig string
	// ConfigFile is the name of the Envoy Gateway configuration file.
	ConfigFile string
	// XdsService should be set to true to enable the xDS Server.
	XdsService bool
	// ProvisionerService should be set to true to enable the Provisioner.
	ProvisionerService bool
	// TODO: Expose other serve start flags.
}

// NewGatewayFlags will create a new GatewayFlags with default values.
func NewGatewayFlags() *GatewayFlags {
	return &GatewayFlags{
		XdsService:         true,
		ProvisionerService: true,
		// TODO: Set other default serve flags.
	}
}

// ValidateGatewayFlags validates Envoy Gateway's configuration flags.
func ValidateGatewayFlags(f *GatewayFlags) error {
}

// NewGatewayConfig will create a new GatewayConfig with default values.
func NewGatewayConfig() (*v1alpha1.GatewayConfig, error) {
	// Create a Scheme that understands the types in the gateway.config API group.
	// Create an instance of the GatewayConfig object, set defaults, and return.
}

// GatewayServer encapsulates all the necessary parameters for starting Envoy Gateway.
// These can either be set via command line or in the config file.
type GatewayServer struct {
	GatewayFlags
	v1alpha1.GatewayConfig
}

// ValidateGatewayServer validates configuration of GatewayServer and returns
// an error if the input configuration is invalid.
func ValidateGatewayServer(s *GatewayServer) error {
	// Validate flags and config file.
}

// AddFlags adds flags for a specific GatewayFlags to the specified FlagSet.
func (f *GatewayFlags) AddFlags(fs *pflag.FlagSet) {
	fs := pflag.NewFlagSet("", pflag.ExitOnError)
	// Providing --kubeconfig enables Kubernetes API server mode. Omitting --kubeconfig enables standalone mode.
	fs.StringVar(&f.KubeConfig, "kubeconfig", f.KubeConfig, "Path to a kubeconfig file for connecting to Kube API server.")
	// Omit the --config flag to use the built-in default configuration values.
	// Command-line flags override configuration from this file.
	fs.StringVar(&f.ConfigFile, "config", f.ConfigFile, "The Envoy Gateway configuration file.")
	fs.BoolVar(&f.XdsService, "xds-service", f.XdsService, "Set to false to disable the xds server.")
	fs.BoolVar(&f.ProvisionerService, "provisioner-service", f.ProvisionerService,
		"Set to false to disable the provisioner server.")
}
```

The --config flag specifies the path of the Envoy Gateway configuration file. Envoy Gateway will load its config from
this file. Command line flags which target the same value as a config file will override that value. If --config is
provided and values are not specified via the command line, the defaults for the config file apply.

In `gateway/internal/componentmanager/manager.go`:
```go
package componentmanager

import (
	"fmt"

	"github.com/envoyproxy/gateway/internal/provisioner"
	"github.com/envoyproxy/gateway/internal/xds"

	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

const (
	// DefaultEnvoyImage is the default Envoy image to use when unspecified.
	DefaultEnvoyImage = "docker.io/envoyproxy/envoy:v1.23.0"
)

// ComponentManager manages Envoy Gateway components. It sets up dependencies and defines the topology
// of Envoy Gateway and its managed components, wiring them together, properly shutting them down, etc.
type ComponentManager struct {
	manager manager.Manager
	config  *Config
}

// Config is configuration of a ComponentManager.
type Config struct {
	// LeaderElection determines whether to use leader election.
	LeaderElection bool
	// Provisioner
	Provisioner provisioner.Config
	// XdsServer
	XdsServer xds.ServerConfig
}

// DefaultConfig returns a ComponentManager config using default values.
func DefaultConfig() *Config {
	return &Config{
		EnvoyImage: DefaultEnvoyImage,
	}
}

// New creates a new ComponentManager from cliCfg and operatorConfig.
func New(cliCfg *rest.Config, mgrCfg *Config) (*ComponentManager, error) {
	mgrOpts := manager.Options{
		LeaderElection: mgrCfg.LeaderElection,
	}
	mgr, err := ctrl.NewManager(cliCfg, mgrOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to create manager: %w", err)
	}

	// Create and register components.
	if _, err := provisioner.New(mgr, provisioner.Config{
		EnvoyImage: mgrCfg.EnvoyImage,
	}); err != nil {
		return nil, fmt.Errorf("failed to create provisioner: %w", err)
	}
	if _, err := xds.NewServer(mgr, xds.ServerConfig{
		BindAddress: mgrCfg.EnvoyImage,
	}); err != nil {
		return nil, fmt.Errorf("failed to create contour controller: %w", err)
	}
}
```

[issue_43]: https://github.com/envoyproxy/gateway/issues/43
[issue_60]: https://github.com/envoyproxy/gateway/issues/60

