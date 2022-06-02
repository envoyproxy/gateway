Config Management Design
===================

## Motivation

[Issue 43][issue_43] specifies the need for lifecycle management of Envoy Gateway system components that run as Go
routines.

## Goals

* Gracefully manage the services that comprise Envoy Gateway.

## Non-Goals

* Provide a detailed Envoy Gateway configuration specification.

## Proposal

Introduce a `serve` argument to the `envoy-gateway` command. The `serve` argument encapsulates the management of Envoy
Gateway.
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

	cmd.AddCommand(serve.AddCommand())

	return cmd
}
```

The serve package provides the following functionality:
- Allows the serve package to be run as a Cobra Command.
- Validates, processes, etc. command line flags and the Envoy Gateway config file. 
- Constructs a object, e.g. `GatewayServer`, based on the validated flags/config. This object encapsulates all the
necessary parameters for running Envoy Gateway.
- Runs Envoy Gateway based on the provided `GatewayServer`.
```go
// gateway/internal/cmd/serve/serve.go

package serve

import (
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/envoyproxy/gateway/apis/config/v1alpha1"
)

// AddCommand allows the serve package to be run as a Cobra Command.
func AddCommand() *cobra.Command {
	flags := NewGatewayFlags()

	gatewayCfg, _ := NewGatewayConfig()

	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Serve serves Envoy Gateway",
		RunE: func(cmd *cobra.Command, args []string) error {
			// If provided, validate flags and config file.
			// If provided, load the envoy gateway config file.
			// Construct a GatewayServer object from the GatewayConfig type.
			// Run Envoy Gateway using the provided GatewayServer.
		},
	}

	return cmd
}

// GatewayFlags contains Envoy Gateway configuration flags.
type GatewayFlags struct {
	// Add fields for command line flags, e.g. config file, enabled services, etc.
}

// NewGatewayFlags will create a new GatewayFlags with default values.
func NewGatewayFlags() *GatewayFlags {
	return &GatewayFlags{
		// Set flag defaults, e.g. enabled services, config file path, etc.
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
	// TBD Envoy Gateway config CRD.
	Config v1alpha1.GatewayConfig
}

// ValidateGatewayServer validates configuration of GatewayServer and returns
// an error if the input configuration is invalid.
func (s *GatewayServer) ValidateGatewayServer() error {
	// Validate flags and config file.
}

// AddFlags adds flags for a specific GatewayFlags to the specified FlagSet.
func (f *GatewayFlags) AddFlags(fs *pflag.FlagSet) {
	// Add command line flags, e.g. config file, enabled services, etc.
}
```
A subset of the Envoy Gateway's configuration parameters may be set by a configuration file, as a substitute for
command-line flags. Providing parameters via a config file is the recommended approach as it simplifies deployment
and configuration management. The config file is defined by the `GatewayConfig` struct. The config file must be a YAML
representation of the parameters in this struct. For example:
```yaml
apiVersion: gateway.envoy.io/v1alpha1
kind: GatewayConfig
foo:
  bar: baz
...
```
The `--config` flag specifies the path of the Envoy Gateway configuration file. Envoy Gateway will load its config from
this file. Command line flags which target the same value as a config file will override that value. If `--config` is
provided and values are not specified via the command line, the defaults for the config file are applied.

The `serve` command calls `New()` and `Start()` to create and run Envoy Gateway. [Manager][mgr] creates controllers,
e.g. GatewayClass controller, and provides shared dependencies such as clients, caches, etc. Non-controller
processes, e.g. xDS server, implement the [Runnable][runnable] interface. This allows Manager to manage these processes
in the same manner as controller-based processes.
```go
package manager

import (
	"context"

	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	"github.com/envoyproxy/gateway/internal/kubernetes/gateway"
	"github.com/envoyproxy/gateway/internal/kubernetes/gatewayclass"
	"github.com/envoyproxy/gateway/internal/kubernetes/httproute"
)

// GatewayManager sets-up dependencies and defines the topology of Envoy Gateway
// managed components, wiring them together.
type GatewayManager struct {
	client  client.Client
	manager manager.Manager
	config  *v1alpha1.GatewayConfig
}

// New creates a new Envoy Gateway using the provided configs.
// TODO: A Manager requires a rest config to construct a clutser. This means
// using Manager to manage services requires access to a k8s cluster.
// xref: https://github.com/kubernetes-sigs/controller-runtime/blob/v0.12.1/pkg/cluster/cluster.go#L145-L205
func New(cliCfg *rest.Config, s *GatewayServer) (*GatewayManager, error) {
	mgrOpts := manager.Options{
		// Add the scheme with types supported by Envoy Gateway.
		// Add other manager configuration options, e.g. metrics bind address,
		// leader election, etc. from GatewayServer.
	}
	mgr, _ := ctrl.NewManager(cliCfg, mgrOpts)

	// Create and register components with the manager.
	if !s.Config.StandAlone {
		// Create and register the kube controllers.
		gcCfg := &gatewayclass.ControllerConfig{
			// GatewayClass controller config.
		}
		if _, err := gatewayclass.NewController(mgr, gcCfg); err != nil {
			return nil, err
		}
		gwCfg := &gateway.ControllerConfig{
			// Gateway controller config.
		}
		if _, err := gateway.NewController(mgr, gwCfg); err != nil {
			return nil, err
		}
		httpCfg := &controller.HttpRouteConfig{
			// HTTPRoute controller config.
		}
		if _, err := httproute.NewController(mgr, httpCfg); err != nil {
			return nil, err
		}
    }
	if s.Config.XdsService {
		// The xds service will implement the Runnable interface, so it can be added to manager.
		x, _ := xds.NewServer()
		if err := mgr.Add(x); err != nil {
			return nil, err
		}
	}
	if s.Config.ProvisionerService {
		// The provisioner service will implement the Runnable interface, so it can be added to manager.
		p, _ := provisioner.NewServer()
		if err := mgr.Add(p); err != nil {
			return nil, err
		}
	}
}

// Start starts Envoy Gateway manager, waiting for it to exit or an explicit stop.
func (g *GatewayManager) Start(ctx context.Context) error {
	errChan := make(chan error)
	go func() {
		errChan <- g.manager.Start(ctx)
	}()

	// Wait for the manager to exit or an explicit stop.
	select {
	case <-ctx.Done():
		return nil
	case err := <-errChan:
		return err
	}
}
```

[issue_43]: https://github.com/envoyproxy/gateway/issues/43
[issue_60]: https://github.com/envoyproxy/gateway/issues/60
[mgr]: https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.12.1/pkg/manager#Manager
[runnable]: https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.12.1/pkg/manager#Runnable
