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

	cmd.AddCommand(serve.NewCommand())

	return cmd
}
```

The serve package provides the following functionality:
- Allows the serve package to be run as a Cobra Command.
- Validates, processes, etc. command line flags and the Envoy Gateway config file. 
- Constructs an object, e.g. `GatewayServer`, based on the validated flags/config. This object encapsulates all the
necessary parameters for running Envoy Gateway.
- Runs Envoy Gateway based on the provided `GatewayServer`.

A subset of the Envoy Gateway's configuration parameters may be set by a configuration file, as a substitute for
command-line flags. Providing parameters via a config file is the recommended approach as it simplifies deployment
and configuration management. The config file is defined by the `BootstrapConfig` struct. The config file must be a YAML
representation of the parameters in this struct. For example:
```yaml
apiVersion: gateway.envoy.io/v1alpha1
kind: BootstrapConfig
foo:
  bar: baz
...
```
The `--config` flag specifies the path of the Envoy Gateway configuration file. Envoy Gateway will load its config from
this file. Command line flags which target the same value as a config file will override that value. If `--config` is
provided and values are not specified via the command line, the defaults for the config file are applied.
```go
// gateway/internal/cmd/serve/serve.go

package serve

import (
	"context"

	"github.com/oklog/run"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"

	"github.com/envoyproxy/gateway/apis/config/v1alpha1"
	kube "github.com/envoyproxy/gateway/internal/kubernetes"
)

// NewCommand creates the "serve" *cobra.Command object with default parameters.
func NewCommand() *cobra.Command {
	flags := NewBootstrapFlags()

	cfg, _ := NewBootstrapConfig()

	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Serve serves Envoy Gateway",
		RunE: func(cmd *cobra.Command, args []string) error {
			// If provided, validate flags and config file.
			// If provided, load the envoy gateway config file.
			// Construct a GatewayServer object from the BootstrapConfig type.
			// Run Envoy Gateway using the provided GatewayServer.
			return run(GatewayServer)
		},
	}

	return cmd
}

// BootstrapFlags contains Envoy Gateway bootstrap configuration flags.
type BootstrapFlags struct {
	// Add fields for command line flags, e.g. bootstrap config file, enabled services, etc.
}

// NewBootstrapFlags will create a new BootstrapFlags with default values.
func NewBootstrapFlags() *BootstrapFlags {
	return &BootstrapFlags{
		// Set flag defaults, e.g. enabled services, config file path, etc.
	}
}

// AddFlags adds flags for a specific BootstrapFlags to the specified FlagSet.
func (f *BootstrapFlags) AddFlags(fs *pflag.FlagSet) {
	// Add command line flags, e.g. config file, enabled services, etc.
}

// ValidateBootstrapFlags validates Envoy Gateway's bootstrap configuration flags.
func (f *BootstrapFlags) ValidateBootstrapFlags() error {
}

// NewBootstrapConfig will create a new BootstrapConfig with default values.
func NewBootstrapConfig() (*v1alpha1.BootstrapConfig, error) {
	// Create a Scheme that understands the types in the gateway.config API group.
	// Create an instance of the BootstrapConfig object, set defaults, and return.
}

// GatewayServer encapsulates all the necessary parameters for starting Envoy Gateway.
// These can either be set via command line or in the config file.
type GatewayServer struct {
	BootstrapFlags
	// TBD Envoy Gateway bootstrap config CRD.
	Config v1alpha1.BootstrapConfig
}

// ValidateGatewayServer validates configuration of GatewayServer and returns
// an error if the input configuration is invalid.
func (s *GatewayServer) ValidateGatewayServer() error {
	// Validate flags and config file.
}

// Run runs the provided GatewayServer.
func run(s *GatewayServer) error {
	var g run.Group

	// Set up the channels for the watcher, operator, and metrics using
	// the context provided from the controller runtime.
	signal, cancel := context.WithCancel(signals.SetupSignalHandler())
	defer cancel()

	if s.Config.Kubernetes != nil {
		// GetConfig creates a *rest.Config for talking to a Kubernetes API server.
		// If --kubeconfig is set, it will use the kubeconfig file at that location.
		// Otherwise, it will assume running in cluster and use the cluster provided kubeconfig.
		cfg, _ := config.GetConfig()

		// Set up and start the manager for managing k8s controllers.
		mgr, _ := kube.NewManager(cfg, s)

		g.Add(func() error {
			return mgr.Start(signal)
		}, func(error) {
			cancel()
		})
	}
	// Repeat for other components, e.g. Message Server, Provisioner, etc.

	return nil
}
```
The `run()` function uses `run.Group` to manage goroutine lifecycles. A zero-value `run.Group` is created and actors,
e.g. the Kubernetes Manager, are added to it.

If Kubernetes is set in BootstrapConfig, `NewManager()` and
`Manager.Start()` are called to create and run Kubernetes controllers. [Manager][mgr] creates controllers, e.g.
GatewayClass controller, and provides shared dependencies such as clients, caches, etc.
```go
// gateway/internal/kubernetes/manager.go

package kubernetes

import (
	"context"

	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	"github.com/envoyproxy/internal/cmd/serve"
	"github.com/envoyproxy/internal/kubernetes/gateway"
	"github.com/envoyproxy/internal/kubernetes/gatewayclass"
	"github.com/envoyproxy/internal/kubernetes/httproute"
)

// Manager sets-up dependencies and defines the topology of Envoy Gateway
// managed components, wiring them together.
type Manager struct {
	client client.Client

	manager.Manager
}

// NewManager creates a new Manager for managing k8s controllers.
func NewManager(cfg *rest.Config, s *serve.GatewayServer) (*Manager, error) {
	mgrOpts := manager.Options{
		// Add the scheme with types supported by Envoy Gateway.
		// Add other manager configuration options, e.g. metrics bind address,
		// leader election, etc. from GatewayServer.
	}
	mgr, _ := ctrl.NewManager(cfg, mgrOpts)

    // Create and register the kube controllers.
    gcCfg := &gatewayclass.Config{
        // GatewayClass controller config.
    }
    if _, err := gatewayclass.NewController(mgr, gcCfg); err != nil {
        return nil, err
    }
    gwCfg := &gateway.Config{
        // Gateway controller config.
    }
    if _, err := gateway.NewController(mgr, gwCfg); err != nil {
        return nil, err
    }
    httpCfg := &httproute.Config{
        // HTTPRoute controller config.
    }
    if _, err := httproute.NewController(mgr, httpCfg); err != nil {
        return nil, err
    }

	return &Manager{mgr.GetClient(), mgr}, nil
}

// Start starts Envoy Gateway manager, waiting for it to exit or an explicit stop.
func (m *Manager) Start(ctx context.Context) error {
	errChan := make(chan error)
	go func() {
		errChan <- m.Start(ctx)
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
