// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package egctl

import (
	"fmt"
	"io"
	"net/http"

	"github.com/spf13/cobra"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/reflect/protoreflect"
	"k8s.io/apimachinery/pkg/types"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
	"sigs.k8s.io/yaml"

	adminv3 "github.com/envoyproxy/go-control-plane/envoy/admin/v3"

	"github.com/envoyproxy/gateway/internal/cmd/options"
	kube "github.com/envoyproxy/gateway/internal/kubernetes"
)

var (
	output       string
	podName      string
	podNamespace string
)

const (
	adminPort     = 19000   // TODO: make this configurable until EG support
	containerName = "envoy" // TODO: make this configurable until EG support
)

func NewConfigCommand() *cobra.Command {
	cfgCommand := &cobra.Command{
		Use:     "config",
		Aliases: []string{"c"},
		Short:   "Retrieve proxy configuration.",
		Long:    "Retrieve information about proxy configuration from envoy proxy and gateway.",
	}

	cfgCommand.AddCommand(proxyCommand())

	flags := cfgCommand.Flags()
	options.AddKubeConfigFlags(flags)

	cfgCommand.PersistentFlags().StringVarP(&output, "output", "o", "json", "One of 'yaml' or 'json'")
	cfgCommand.PersistentFlags().StringVarP(&podNamespace, "namespace", "n", "envoy-gateway-system", "Namespace where envoy proxy pod are installed.")

	return cfgCommand
}

func proxyCommand() *cobra.Command {
	c := &cobra.Command{
		Use:     "envoy-proxy",
		Aliases: []string{"proxy"},
		Long:    "Retrieve information from envoy proxy.",
	}

	c.AddCommand(allConfigCmd())
	c.AddCommand(bootstrapConfigCmd())
	c.AddCommand(clusterConfigCmd())
	c.AddCommand(listenerConfigCmd())
	c.AddCommand(routeConfigCmd())

	return c
}

func allConfigCmd() *cobra.Command {

	configCmd := &cobra.Command{
		Use:   "all <pod-name>",
		Short: "Retrieves all Envoy xDS resources from the specified pod",
		Long:  `Retrieves information about all Envoy xDS resources from the Envoy instance in the specified pod.`,
		Example: `  # Retrieve summary about all configuration for a given pod from Envoy.
  egctl config envoy-proxy all <pod-name> -n <pod-namespace>

  # Retrieve full configuration dump as YAML
  egctl config envoy-proxy all <pod-name> -n <pod-namespace> -o yaml

  # Retrieve full configuration dump with short syntax
  egctl c proxy all <pod-name> -n <pod-namespace>
`,
		Run: func(c *cobra.Command, args []string) {
			cmdutil.CheckErr(runAllConfig(c, args))
		},
	}

	return configCmd
}

func runAllConfig(c *cobra.Command, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("pod name is required")
	}

	podName = args[0]

	if podName == "" {
		return fmt.Errorf("pod name is required")
	}

	if podNamespace == "" {
		return fmt.Errorf("pod namespace is required")
	}

	fw, err := portForwarder(types.NamespacedName{
		Namespace: podNamespace,
		Name:      podName,
	})
	if err != nil {
		return err
	}
	if err := fw.Start(); err != nil {
		return err
	}
	defer fw.Stop()

	configDump, err := extractConfigDump(fw)
	if err != nil {
		return err
	}

	out, err := marshalEnvoyProxyConfig(configDump, output)
	if err != nil {
		return err
	}

	_, err = fmt.Fprintln(c.OutOrStdout(), string(out))
	return err
}

func bootstrapConfigCmd() *cobra.Command {

	configCmd := &cobra.Command{
		Use:   "bootstrap <pod-name>",
		Short: "Retrieves bootstrap Envoy xDS resources from the specified pod",
		Long:  `Retrieves information about bootstrap Envoy xDS resources from the Envoy instance in the specified pod.`,
		Example: `  # Retrieve summary about bootstrap configuration for a given pod from Envoy.
  egctl config envoy-proxy bootstrap <pod-name> -n <pod-namespace>

  # Retrieve full configuration dump as YAML
  egctl config envoy-proxy bootstrap <pod-name> -n <pod-namespace> -o yaml

  # Retrieve full configuration dump with short syntax
  egctl c proxy bootstrap <pod-name> -n <pod-namespace>
`,
		Run: func(c *cobra.Command, args []string) {
			cmdutil.CheckErr(runBootstrapConfig(c, args))
		},
	}

	return configCmd
}

func runBootstrapConfig(c *cobra.Command, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("pod name is required")
	}

	podName = args[0]

	if podName == "" {
		return fmt.Errorf("pod name is required")
	}

	if podNamespace == "" {
		return fmt.Errorf("pod namespace is required")
	}

	fw, err := portForwarder(types.NamespacedName{
		Namespace: podNamespace,
		Name:      podName,
	})
	if err != nil {
		return err
	}
	if err := fw.Start(); err != nil {
		return err
	}
	defer fw.Stop()

	configDump, err := extractConfigDump(fw)
	if err != nil {
		return err
	}

	bootstrap, err := findXDSResourceFromConfigDump(BootstrapEnvoyConfigType, configDump)
	if err != nil {
		return err
	}

	out, err := marshalEnvoyProxyConfig(bootstrap, output)
	if err != nil {
		return err
	}

	_, err = fmt.Fprintln(c.OutOrStdout(), string(out))
	return err
}

func clusterConfigCmd() *cobra.Command {

	configCmd := &cobra.Command{
		Use:   "cluster <pod-name>",
		Short: "Retrieves cluster Envoy xDS resources from the specified pod",
		Long:  `Retrieves information about cluster Envoy xDS resources from the Envoy instance in the specified pod.`,
		Example: `  # Retrieve summary about cluster configuration for a given pod from Envoy.
  egctl config envoy-proxy cluster <pod-name> -n <pod-namespace>

  # Retrieve full configuration dump as YAML
  egctl config envoy-proxy cluster <pod-name> -n <pod-namespace> -o yaml

  # Retrieve full configuration dump with short syntax
  egctl c proxy cluster <pod-name> -n <pod-namespace>
`,
		Run: func(c *cobra.Command, args []string) {
			cmdutil.CheckErr(runClusterConfig(c, args))
		},
	}

	return configCmd
}

func runClusterConfig(c *cobra.Command, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("pod name is required")
	}

	podName = args[0]

	if podName == "" {
		return fmt.Errorf("pod name is required")
	}

	if podNamespace == "" {
		return fmt.Errorf("pod namespace is required")
	}

	fw, err := portForwarder(types.NamespacedName{
		Namespace: podNamespace,
		Name:      podName,
	})
	if err != nil {
		return err
	}
	if err := fw.Start(); err != nil {
		return err
	}
	defer fw.Stop()

	configDump, err := extractConfigDump(fw)
	if err != nil {
		return err
	}

	cluster, err := findXDSResourceFromConfigDump(ClusterEnvoyConfigType, configDump)
	if err != nil {
		return err
	}

	out, err := marshalEnvoyProxyConfig(cluster, output)
	if err != nil {
		return err
	}

	_, err = fmt.Fprintln(c.OutOrStdout(), string(out))
	return err
}

func listenerConfigCmd() *cobra.Command {

	configCmd := &cobra.Command{
		Use:   "listener <pod-name>",
		Short: "Retrieves listener Envoy xDS resources from the specified pod",
		Long:  `Retrieves information about listener Envoy xDS resources from the Envoy instance in the specified pod.`,
		Example: `  # Retrieve summary about listener configuration for a given pod from Envoy.
  egctl config envoy-proxy listener <pod-name> -n <pod-namespace>

  # Retrieve full configuration dump as YAML
  egctl config envoy-proxy listener <pod-name> -n <pod-namespace> -o yaml

  # Retrieve full configuration dump with short syntax
  egctl c proxy listener <pod-name> -n <pod-namespace>
`,
		Run: func(c *cobra.Command, args []string) {
			cmdutil.CheckErr(runListenerConfig(c, args))
		},
	}

	return configCmd
}

func runListenerConfig(c *cobra.Command, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("pod name is required")
	}

	podName = args[0]

	if podName == "" {
		return fmt.Errorf("pod name is required")
	}

	if podNamespace == "" {
		return fmt.Errorf("pod namespace is required")
	}

	fw, err := portForwarder(types.NamespacedName{
		Namespace: podNamespace,
		Name:      podName,
	})
	if err != nil {
		return err
	}
	if err := fw.Start(); err != nil {
		return err
	}
	defer fw.Stop()

	configDump, err := extractConfigDump(fw)
	if err != nil {
		return err
	}

	listener, err := findXDSResourceFromConfigDump(ListenerEnvoyConfigType, configDump)
	if err != nil {
		return err
	}

	out, err := marshalEnvoyProxyConfig(listener, output)
	if err != nil {
		return err
	}

	_, err = fmt.Fprintln(c.OutOrStdout(), string(out))
	return err
}

func routeConfigCmd() *cobra.Command {

	configCmd := &cobra.Command{
		Use:   "route <pod-name>",
		Short: "Retrieves route Envoy xDS resources from the specified pod",
		Long:  `Retrieves information about route Envoy xDS resources from the Envoy instance in the specified pod.`,
		Example: `  # Retrieve summary about route configuration for a given pod from Envoy.
  egctl config envoy-proxy route <pod-name> -n <pod-namespace>

  # Retrieve full configuration dump as YAML
  egctl config envoy-proxy route <pod-name> -n <pod-namespace> -o yaml

  # Retrieve full configuration dump with short syntax
  egctl c proxy route <pod-name> -n <pod-namespace>
`,
		Run: func(c *cobra.Command, args []string) {
			cmdutil.CheckErr(runRouteConfig(c, args))
		},
	}

	return configCmd
}

func runRouteConfig(c *cobra.Command, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("pod name is required")
	}

	podName = args[0]

	if podName == "" {
		return fmt.Errorf("pod name is required")
	}

	if podNamespace == "" {
		return fmt.Errorf("pod namespace is required")
	}

	fw, err := portForwarder(types.NamespacedName{
		Namespace: podNamespace,
		Name:      podName,
	})
	if err != nil {
		return err
	}
	if err := fw.Start(); err != nil {
		return err
	}
	defer fw.Stop()

	configDump, err := extractConfigDump(fw)
	if err != nil {
		return err
	}

	route, err := findXDSResourceFromConfigDump(RouteEnvoyConfigType, configDump)
	if err != nil {
		return err
	}

	out, err := marshalEnvoyProxyConfig(route, output)
	if err != nil {
		return err
	}

	_, err = fmt.Fprintln(c.OutOrStdout(), string(out))
	return err
}

func portForwarder(nn types.NamespacedName) (kube.PortForwarder, error) {
	c, err := kube.NewCLIClient(options.DefaultConfigFlags.ToRawKubeConfigLoader())
	if err != nil {
		return nil, fmt.Errorf("build CLI client fail: %w", err)
	}

	pod, err := c.Pod(nn)
	if err != nil {
		return nil, fmt.Errorf("get pod %s fail: %w", nn, err)
	}
	if pod.Status.Phase != "Running" {
		return nil, fmt.Errorf("pod %s is not running", nn)
	}

	fw, err := kube.NewLocalPortForwarder(c, nn, 0, int(adminPort))
	if err != nil {
		return nil, err
	}

	return fw, nil
}

func marshalEnvoyProxyConfig(configDump protoreflect.ProtoMessage, output string) ([]byte, error) {
	out, err := protojson.Marshal(configDump)
	if err != nil {
		return nil, err
	}

	if output == "yaml" {
		out, err = yaml.JSONToYAML(out)
		if err != nil {
			return nil, err
		}
	}

	return out, nil
}

func extractConfigDump(fw kube.PortForwarder) (*adminv3.ConfigDump, error) {
	out, err := configDumpRequest(fw.Address())
	if err != nil {
		return nil, err
	}

	configDump := &adminv3.ConfigDump{}
	if err := protojson.Unmarshal(out, configDump); err != nil {
		return nil, err
	}

	return configDump, nil
}

func configDumpRequest(address string) ([]byte, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("http://%s/config_dump", address), nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	return io.ReadAll(resp.Body)
}
