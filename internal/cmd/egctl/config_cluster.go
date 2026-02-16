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

func clusterConfigCmd() *cobra.Command {
	configCmd := &cobra.Command{
		Use:     "cluster <pod-name>",
		Short:   "Retrieves cluster Envoy xDS resources from the specified pod",
		Aliases: []string{"c"},
		Long:    `Retrieves information about cluster Envoy xDS resources from the Envoy instance in the specified pod.`,
		Example: `  # Retrieve summary about cluster configuration for a given pod from Envoy.
  egctl config envoy-proxy cluster <pod-name> -n <pod-namespace>

  # Retrieve summary about cluster configuration for a pod matching label selectors
  egctl config envoy-proxy cluster --labels gateway.envoyproxy.io/owning-gateway-name=eg -l gateway.envoyproxy.io/owning-gateway-namespace=default

  # Retrieve full configuration dump as YAML
  egctl config envoy-proxy cluster <pod-name> -n <pod-namespace> -o yaml

  # Retrieve full configuration dump with short syntax
  egctl c proxy c <pod-name> -n <pod-namespace>
`,
		Run: func(c *cobra.Command, args []string) {
			cmdutil.CheckErr(runClusterConfig(c, args))
		},
	}

	return configCmd
}

func runClusterConfig(c *cobra.Command, args []string) error {
	configDump, err := retrieveConfigDump(args, false, ClusterEnvoyConfigType)
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
