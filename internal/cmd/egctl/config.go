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
	"k8s.io/apimachinery/pkg/types"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
	"sigs.k8s.io/yaml"

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
	cfgCommand.PersistentFlags().StringVarP(&podNamespace, "namespace", "n", "envoy-gateway", "Namespace where envoy proxy pod are installed.")

	return cfgCommand
}

func proxyCommand() *cobra.Command {
	c := &cobra.Command{
		Use:     "envoy-proxy",
		Aliases: []string{"proxy"},
		Long:    "Retrieve information from envoy proxy.",
	}

	c.AddCommand(allConfigCmd())

	return c
}

func allConfigCmd() *cobra.Command {

	allConfigCmd := &cobra.Command{
		Use:   "all <pod-name>",
		Short: "Retrieves all configuration for the Envoy in the specified pod",
		Long:  `Retrieve information about all configuration for the Envoy instance in the specified pod.`,
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

	return allConfigCmd
}

func runAllConfig(c *cobra.Command, args []string) error {
	podName = args[0]

	if podName == "" {
		return fmt.Errorf("pod name is required")
	}

	if podNamespace == "" {
		return fmt.Errorf("pod namespace is required")
	}

	out, err := extractConfigDump(types.NamespacedName{
		Namespace: podNamespace,
		Name:      podName,
	})
	if err != nil {
		return err
	}

	if output == "yaml" {
		out, err = yaml.JSONToYAML(out)
		if err != nil {
			return err
		}
	}
	_, err = fmt.Fprintln(c.OutOrStdout(), string(out))
	return err
}

func extractConfigDump(nn types.NamespacedName) ([]byte, error) {
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

	if err := fw.Start(); err != nil {
		return nil, err
	}
	defer fw.Stop()

	return configDumpRequest(fw.Address())
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
