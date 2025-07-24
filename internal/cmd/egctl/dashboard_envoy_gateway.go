// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package egctl

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/types"

	kube "github.com/envoyproxy/gateway/internal/kubernetes"
)

const (
	// envoyGatewayAdminPort is the port for the Envoy Gateway admin server
	envoyGatewayAdminPort = 19000
)

func newEnvoyGatewayDashboardCmd() *cobra.Command {
	var namespace string
	var listenPort int

	dashboardCmd := &cobra.Command{
		Use:   "envoy-gateway",
		Short: "Open Envoy Gateway admin console in your browser",
		Long:  `Open Envoy Gateway admin console in your browser.`,
		Example: `  # Open Envoy Gateway admin console in your browser.
  egctl x dashboard envoy-gateway

  # Open Envoy Gateway admin console with a specific namespace
  egctl x dashboard envoy-gateway -n custom-namespace

  # Open Envoy Gateway admin console with a specific port
  egctl x dashboard envoy-gateway --port 8080
`,
		Aliases: []string{"eg"},
		RunE: func(c *cobra.Command, args []string) error {
			if listenPort > 65535 || listenPort < 0 {
				return fmt.Errorf("invalid port number range")
			}

			kubeClient, err := getCLIClient()
			if err != nil {
				return err
			}

			// Find the Envoy Gateway deployment pod
			labelSelectors := []string{"control-plane=envoy-gateway"}
			pl, err := kubeClient.PodsForSelector(namespace, labelSelectors...)
			if err != nil {
				return fmt.Errorf("not able to locate Envoy Gateway pod: %w", err)
			}
			if len(pl.Items) < 1 {
				return fmt.Errorf("no Envoy Gateway pods found in namespace %s", namespace)
			}

			podName := pl.Items[0].Name
			fmt.Fprintf(c.OutOrStdout(), "Found Envoy Gateway pod: %s\n", podName)

			return forwardToEnvoyGatewayAdmin(podName, namespace, listenPort, kubeClient, c.OutOrStdout())
		},
	}

	dashboardCmd.PersistentFlags().StringVarP(&namespace, "namespace", "n", "envoy-gateway-system", "Namespace where Envoy Gateway is installed.")
	dashboardCmd.PersistentFlags().IntVarP(&listenPort, "port", "p", 0, "Local port to listen to.")

	return dashboardCmd
}

// forwardToEnvoyGatewayAdmin forwards a local port to the Envoy Gateway admin port
func forwardToEnvoyGatewayAdmin(podName, namespace string, listenPort int, client kube.CLIClient, writer io.Writer) error {
	meta := types.NamespacedName{
		Namespace: namespace,
		Name:      podName,
	}

	fmt.Fprintf(writer, "Setting up port-forward to Envoy Gateway admin server...\n")

	fw, err := kube.NewLocalPortForwarder(client, meta, listenPort, envoyGatewayAdminPort)
	if err != nil {
		return fmt.Errorf("could not build port forwarder for Envoy Gateway: %w", err)
	}

	if err = fw.Start(); err != nil {
		fw.Stop()
		return fmt.Errorf("could not start port forwarder for Envoy Gateway: %w", err)
	}

	ClosePortForwarderOnInterrupt(fw)

	url := fmt.Sprintf("http://%s", fw.Address())
	fmt.Fprintf(writer, "Envoy Gateway admin console URL: %s\n", url)
	fmt.Fprintf(writer, "Opening browser to access the Envoy Gateway admin console...\n")
	fmt.Fprintf(writer, "Press Ctrl+C to quit\n")

	openBrowser(url, writer)

	fw.WaitForStop()

	return nil
}
