// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package egctl

import (
	"fmt"

	"github.com/spf13/cobra"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
)

func gatewayAllConfigCmd() *cobra.Command {
	configCmd := &cobra.Command{
		Use:   "all",
		Short: "Retrieves complete configuration dump from Envoy Gateway",
		Long:  `Retrieves complete configuration dump including all provider resources from the Envoy Gateway admin console.`,
		Example: `  # Retrieve complete configuration dump from Envoy Gateway
  egctl config envoy-gateway all -n envoy-gateway-system

  # Retrieve complete configuration dump as YAML
  egctl config envoy-gateway all -n envoy-gateway-system -o yaml

  # Retrieve complete configuration dump with short syntax
  egctl c gateway all -n envoy-gateway-system
`,
		Run: func(c *cobra.Command, args []string) {
			cmdutil.CheckErr(runGatewayConfig(c, args, "all"))
		},
	}

	return configCmd
}

func gatewaySummaryConfigCmd() *cobra.Command {
	configCmd := &cobra.Command{
		Use:   "summary",
		Short: "Retrieves summary configuration dump from Envoy Gateway",
		Long:  `Retrieves summary configuration dump from the Envoy Gateway admin console.`,
		Example: `  # Retrieve summary configuration dump from Envoy Gateway
  egctl config envoy-gateway summary -n envoy-gateway-system

  # Retrieve summary configuration dump as YAML
  egctl config envoy-gateway summary -n envoy-gateway-system -o yaml

  # Retrieve summary configuration dump with short syntax
  egctl c gateway summary -n envoy-gateway-system
`,
		Run: func(c *cobra.Command, args []string) {
			cmdutil.CheckErr(runGatewayConfig(c, args, "summary"))
		},
	}

	return configCmd
}

func gatewayInfoCmd() *cobra.Command {
	configCmd := &cobra.Command{
		Use:   "info",
		Short: "Retrieves system information from Envoy Gateway",
		Long:  `Retrieves basic system information from the Envoy Gateway admin console.`,
		Example: `  # Retrieve system information from Envoy Gateway
  egctl config envoy-gateway info -n envoy-gateway-system

  # Retrieve system information as YAML
  egctl config envoy-gateway info -n envoy-gateway-system -o yaml
`,
		Run: func(c *cobra.Command, args []string) {
			cmdutil.CheckErr(runGatewayConfig(c, args, "info"))
		},
	}

	return configCmd
}

func gatewayServerInfoCmd() *cobra.Command {
	configCmd := &cobra.Command{
		Use:   "server-info",
		Short: "Retrieves server status information from Envoy Gateway",
		Long:  `Retrieves server status and component information from the Envoy Gateway admin console.`,
		Example: `  # Retrieve server status information from Envoy Gateway
  egctl config envoy-gateway server-info -n envoy-gateway-system

  # Retrieve server status information as YAML
  egctl config envoy-gateway server-info -n envoy-gateway-system -o yaml
`,
		Run: func(c *cobra.Command, args []string) {
			cmdutil.CheckErr(runGatewayConfig(c, args, "server-info"))
		},
	}

	return configCmd
}

func runGatewayConfig(c *cobra.Command, args []string, configType string) error {
	configDump, err := retrieveGatewayConfigDump(args, configType)
	if err != nil {
		return err
	}

	out, err := marshalGatewayConfig(configDump, output)
	if err != nil {
		return err
	}

	_, err = fmt.Fprintln(c.OutOrStdout(), string(out))
	return err
}
