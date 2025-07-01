// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package egctl

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"runtime"

	adminv3 "github.com/envoyproxy/go-control-plane/envoy/admin/v3"
	"google.golang.org/protobuf/reflect/protoreflect"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"

	"github.com/envoyproxy/gateway/internal/envoygateway"
	kube "github.com/envoyproxy/gateway/internal/kubernetes"
)

type envoyConfigType string

var (
	BootstrapEnvoyConfigType envoyConfigType = "bootstrap"
	ClusterEnvoyConfigType   envoyConfigType = "cluster"
	EndpointEnvoyConfigType  envoyConfigType = "endpoint"
	ListenerEnvoyConfigType  envoyConfigType = "listener"
	RouteEnvoyConfigType     envoyConfigType = "route"
	AllEnvoyConfigType       envoyConfigType = "all"
)

func findXDSResourceFromConfigDump(resourceType envoyConfigType, globalConfigs *adminv3.ConfigDump) (protoreflect.ProtoMessage, error) {
	switch resourceType {
	case BootstrapEnvoyConfigType:
		for _, cfg := range globalConfigs.Configs {
			if cfg.GetTypeUrl() == "type.googleapis.com/envoy.admin.v3.BootstrapConfigDump" {
				return cfg, nil
			}
		}
	case EndpointEnvoyConfigType:
		for _, cfg := range globalConfigs.Configs {
			if cfg.GetTypeUrl() == "type.googleapis.com/envoy.admin.v3.EndpointsConfigDump" {
				return cfg, nil
			}
		}

	case ClusterEnvoyConfigType:
		for _, cfg := range globalConfigs.Configs {
			if cfg.GetTypeUrl() == "type.googleapis.com/envoy.admin.v3.ClustersConfigDump" {
				return cfg, nil
			}
		}
	case ListenerEnvoyConfigType:
		for _, cfg := range globalConfigs.Configs {
			if cfg.GetTypeUrl() == "type.googleapis.com/envoy.admin.v3.ListenersConfigDump" {
				return cfg, nil
			}
		}
	case RouteEnvoyConfigType:
		for _, cfg := range globalConfigs.Configs {
			if cfg.GetTypeUrl() == "type.googleapis.com/envoy.admin.v3.RoutesConfigDump" {
				return cfg, nil
			}
		}
	case AllEnvoyConfigType:
		return globalConfigs, nil
	default:
		return nil, fmt.Errorf("unknown resourceType %s", resourceType)
	}

	return nil, fmt.Errorf("unknown resourceType %s", resourceType)
}

func newK8sClient() (client.Client, error) {
	scheme := envoygateway.GetScheme()

	cli, err := client.New(config.GetConfigOrDie(), client.Options{Scheme: scheme})
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Kubernetes client: %w", err)
	}

	return cli, nil
}

// ClosePortForwarderOnInterrupt closes the port forwarder when an interrupt signal is received
func ClosePortForwarderOnInterrupt(fw kube.PortForwarder) {
	go func() {
		signals := make(chan os.Signal, 1)
		signal.Notify(signals, os.Interrupt)
		defer signal.Stop(signals)
		<-signals
		fw.Stop()
	}()
}

// openBrowser opens the given URL in the default browser
func openBrowser(url string, writer io.Writer) {
	var err error

	fmt.Fprintf(writer, "%s\n", url)

	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		fmt.Fprintf(writer, "Unsupported platform %q; open %s in your browser.\n", runtime.GOOS, url)
	}

	if err != nil {
		fmt.Fprintf(writer, "Failed to open browser; open %s in your browser.\n", url)
	}
}
