// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package egctl

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"runtime"

	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/types"

	kube "github.com/envoyproxy/gateway/internal/kubernetes"
)

func newEnvoyDashboardCmd() *cobra.Command {
	var podName, podNamespace string
	var listenPort int

	dashboardCmd := &cobra.Command{
		Use:   "envoy-proxy <name> -n <namespace>",
		Short: "Retrieve Envoy admin dashboard for the specified pod",
		Long:  `Retrieve Envoy admin dashboard for the specified pod.`,
		Example: `  # Retrieve Envoy admin dashboard for the specified pod.
  egctl experimental dashboard envoy-proxy <pod-name> -n <namespace>

  # short syntax
  egctl experimental d envoy-proxy <pod-name> -n <namespace>
`,
		Aliases: []string{"d"},
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 && len(labelSelectors) == 0 {
				cmd.Println(cmd.UsageString())
				return fmt.Errorf("dashboard requires pod name or label selector")
			}
			if len(args) > 0 && len(labelSelectors) > 0 {
				cmd.Println(cmd.UsageString())
				return fmt.Errorf("name cannot be provided when a selector is specified")
			}
			return nil
		},
		RunE: func(c *cobra.Command, args []string) error {
			if listenPort > 65535 || listenPort < 0 {
				return fmt.Errorf("invalid port number range")
			}

			kubeClient, err := getCLIClient()
			if err != nil {
				return err
			}
			if len(args) != 0 {
				podName = args[0]
			}
			if len(labelSelectors) > 0 {
				pl, err := kubeClient.PodsForSelector(podNamespace, labelSelectors...)
				if err != nil {
					return fmt.Errorf("not able to locate pod with selector %s: %w", labelSelectors, err)
				}
				if len(pl.Items) < 1 {
					return errors.New("no pods found")
				}
				podName = pl.Items[0].Name
				podNamespace = pl.Items[0].Namespace
			}

			return portForward(podName, podNamespace, "http://%s", listenPort, kubeClient, c.OutOrStdout())
		},
	}
	dashboardCmd.PersistentFlags().StringArrayVarP(&labelSelectors, "labels", "l", nil, "Labels to select the envoy proxy pod.")
	dashboardCmd.PersistentFlags().StringVarP(&podNamespace, "namespace", "n", "envoy-gateway-system", "Namespace where envoy proxy pod are installed.")
	dashboardCmd.PersistentFlags().IntVarP(&listenPort, "port", "p", 0, "Local port to listen to.")

	return dashboardCmd
}

// portForward first tries to forward localhost:remotePort to podName:remotePort, falls back to dynamic local port
func portForward(podName, namespace, urlFormat string, listenPort int, client kube.CLIClient, writer io.Writer) error {
	var fw kube.PortForwarder
	meta := types.NamespacedName{
		Namespace: namespace,
		Name:      podName,
	}
	fw, err := kube.NewLocalPortForwarder(client, meta, listenPort, adminPort)
	if err != nil {
		return fmt.Errorf("could not build port forwarder for envoy proxy: %w", err)
	}

	if err = fw.Start(); err != nil {
		fw.Stop()
		return fmt.Errorf("could not start port forwarder for envoy proxy: %w", err)
	}

	ClosePortForwarderOnInterrupt(fw)

	openBrowser(fmt.Sprintf(urlFormat, fw.Address()), writer)

	fw.WaitForStop()

	return nil
}

func ClosePortForwarderOnInterrupt(fw kube.PortForwarder) {
	go func() {
		signals := make(chan os.Signal, 1)
		signal.Notify(signals, os.Interrupt)
		defer signal.Stop(signals)
		<-signals
		fw.Stop()
	}()
}

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
