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

func routeConfigCmd() *cobra.Command {
	configCmd := &cobra.Command{
		Use:     "route <pod-name>",
		Aliases: []string{"r"},
		Short:   "Retrieves route Envoy xDS resources from the specified pod",
		Long:    `Retrieves information about route Envoy xDS resources from the Envoy instance in the specified pod.`,
		Example: `  # Retrieve summary about route configuration for a given pod from Envoy.
  egctl config envoy-proxy route <pod-name> -n <pod-namespace>

  # Retrieve full configuration dump as YAML
  egctl config envoy-proxy route <pod-name> -n <pod-namespace> -o yaml

  # Retrieve full configuration dump with short syntax
  egctl c proxy r <pod-name> -n <pod-namespace>
`,
		Run: func(c *cobra.Command, args []string) {
			cmdutil.CheckErr(runRouteConfig(c, args))
		},
	}

	return configCmd
}

func runRouteConfig(c *cobra.Command, args []string) error {
	configDump, err := retrieveConfigDump(args, false)
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
