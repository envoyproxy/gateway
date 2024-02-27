// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package egctl

import (
	"fmt"

	"github.com/spf13/cobra"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"

	"github.com/envoyproxy/gateway/internal/cmd/options"
)

func newConfigCommand() *cobra.Command {
	cfgCommand := &cobra.Command{
		Use:     "config",
		Aliases: []string{"c"},
		Short:   "Retrieve proxy configuration.",
		Long:    "Retrieve information about proxy configuration from envoy proxy and gateway.",
	}

	cfgCommand.AddCommand(proxyCommand())
	cfgCommand.AddCommand(ratelimitCommand())

	flags := cfgCommand.Flags()
	options.AddKubeConfigFlags(flags)

	cfgCommand.PersistentFlags().StringVarP(&output, "output", "o", "json", "One of 'yaml' or 'json'")
	cfgCommand.PersistentFlags().StringVarP(&podNamespace, "namespace", "n", "envoy-gateway-system", "Namespace where envoy proxy pod are installed.")
	cfgCommand.PersistentFlags().StringArrayVarP(&labelSelectors, "labels", "l", nil, "Labels to select the envoy proxy pod.")
	cfgCommand.PersistentFlags().BoolVarP(&allNamespaces, "all-namespaces", "A", false, "List all envoy proxy pods from all namespaces.")

	return cfgCommand
}

func ratelimitCommand() *cobra.Command {
	return ratelimitConfigCommand()
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
	c.AddCommand(endpointConfigCmd())
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

  # Retrieve summary about all configuration for a pod matching label selectors
  egctl config envoy-proxy all --labels gateway.envoyproxy.io/owning-gateway-name=eg -l gateway.envoyproxy.io/owning-gateway-namespace=default

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
	configDump, err := retrieveConfigDump(args, true, AllEnvoyConfigType)
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
