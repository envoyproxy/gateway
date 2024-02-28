// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package egctl

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sync"

	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/yaml"

	"github.com/envoyproxy/gateway/internal/kubernetes"
)

const (
	prometheusOutput = "prom"
)

var (
	statsType, outputFormat string
)

func newEnvoyStatsCmd() *cobra.Command {
	var podName, podNamespace string

	statsConfigCmd := &cobra.Command{
		Use:   "envoy-proxy <name> -n <namespace>",
		Short: "Retrieves Envoy metrics in the specified pod",
		Long:  `Retrieve Envoy emitted metrics for the specified pod.`,
		Example: `  # Retrieve Envoy emitted metrics for the specified pod.
  egctl experimental stats <pod-name> -n <namespace>

  # Retrieve Envoy server metrics in prometheus format
  egctl experimental stats envoy-proxy <pod-name> -n <namespace> --output prom

  # Retrieve Envoy cluster metrics
  egctl experimental stats  envoy-proxy <pod-name> -n <namespace> --type clusters
`,
		Aliases: []string{"ep"},
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 && len(labelSelectors) == 0 {
				cmd.Println(cmd.UsageString())
				return fmt.Errorf("stats requires pod name or label selector")
			}
			return nil
		},
		RunE: func(c *cobra.Command, args []string) error {
			stats := map[string]string{}
			kubeClient, err := getCLIClient()
			if err != nil {
				return err
			}
			if len(args) != 0 {
				podName = args[0]
			}
			pods, err := fetchRunningEnvoyPods(kubeClient, types.NamespacedName{Namespace: podNamespace, Name: podName}, labelSelectors, allNamespaces)
			if err != nil {
				return err
			}
			var errs error
			var wg sync.WaitGroup
			switch statsType {
			case "", "server":
				wg.Add(len(pods))
				for _, pod := range pods {
					go func(pod types.NamespacedName) {
						stats[pod.Namespace+"/"+pod.Name], err = setupEnvoyServerStatsConfig(kubeClient, pod.Name, pod.Namespace, outputFormat)
						if err != nil {
							errs = errors.Join(errs, err)
						}
						wg.Done()
					}(pod)
				}

				wg.Wait()
			case "cluster", "clusters":
				wg.Add(len(pods))
				for _, pod := range pods {
					go func(pod types.NamespacedName) {
						stats[pod.Namespace+"/"+pod.Name], err = setupEnvoyClusterStatsConfig(kubeClient, pod.Name, pod.Namespace, outputFormat)
						if err != nil {
							errs = errors.Join(errs, err)
						}
						wg.Done()
					}(pod)
				}
				wg.Wait()
			default:
				return fmt.Errorf("unknown stats type %s", statsType)
			}

			if errs != nil {
				return errs
			}

			switch outputFormat {
			case jsonOutput:
				statsBytes, err := json.Marshal(stats)
				if err != nil {
					return err
				}
				_, _ = fmt.Fprint(c.OutOrStdout(), string(statsBytes))
			// convert the json output to yaml
			case yamlOutput:
				var out []byte
				statsBytes, err := json.Marshal(stats)
				if err != nil {
					return err
				}
				if out, err = yaml.JSONToYAML(statsBytes); err != nil {
					return err
				}
				_, _ = fmt.Fprint(c.OutOrStdout(), string(out))
			default:
				for namespacedName, stat := range stats {
					_, _ = fmt.Fprint(c.OutOrStdout(), namespacedName+":\n")
					_, _ = fmt.Fprint(c.OutOrStdout(), stat)
				}
			}

			return nil
		},
	}
	statsConfigCmd.PersistentFlags().StringVarP(&outputFormat, "output", "o", prometheusOutput, "Output format: one of json|yaml|prom")
	statsConfigCmd.PersistentFlags().StringVarP(&statsType, "type", "t", "server", "Where to grab the stats: one of server|clusters")
	statsConfigCmd.PersistentFlags().StringArrayVarP(&labelSelectors, "labels", "l", nil, "Labels to select the envoy proxy pod.")
	statsConfigCmd.PersistentFlags().StringVarP(&podNamespace, "namespace", "n", "envoy-gateway-system", "Namespace where envoy proxy pod are installed.")

	return statsConfigCmd
}

func setupEnvoyServerStatsConfig(kubeClient kubernetes.CLIClient, podName, podNamespace string, outputFormat string) (string, error) {
	path := "stats"
	if outputFormat == jsonOutput || outputFormat == yamlOutput {
		// for yaml output we will convert the json to yaml when printed
		path += "?format=json"
	} else {
		path += "/prometheus"
	}

	fw, err := portForwarder(kubeClient, types.NamespacedName{Namespace: podNamespace, Name: podName}, adminPort)
	if err != nil {
		return "", fmt.Errorf("failed to initialize pod-forwarding for %s/%s: %w", podNamespace, podName, err)
	}
	err = fw.Start()
	if err != nil {
		return "", fmt.Errorf("failed to start port forwarding for pod %s/%s: %w", podNamespace, podName, err)
	}
	defer fw.Stop()

	result, err := statsRequest(fw.Address(), path)
	if err != nil {
		return "", fmt.Errorf("failed to get stats on envoy for pod %s/%s: %w", podNamespace, podName, err)
	}
	return string(result), nil
}

func setupEnvoyClusterStatsConfig(kubeClient kubernetes.CLIClient, podName, podNamespace string, outputFormat string) (string, error) {
	path := "clusters"
	if outputFormat == jsonOutput || outputFormat == yamlOutput {
		// for yaml output we will convert the json to yaml when printed
		path += "?format=json"
	}
	fw, err := portForwarder(kubeClient, types.NamespacedName{Namespace: podNamespace, Name: podName}, adminPort)
	if err != nil {
		return "", fmt.Errorf("failed to initialize pod-forwarding for %s/%s: %w", podNamespace, podName, err)
	}
	err = fw.Start()
	if err != nil {
		return "", fmt.Errorf("failed to start port forwarding for pod %s/%s: %w", podNamespace, podName, err)
	}
	defer fw.Stop()

	result, err := statsRequest(fw.Address(), path)
	if err != nil {
		return "", fmt.Errorf("failed to get stats on envoy for pod %s/%s: %w", podNamespace, podName, err)
	}
	return string(result), nil
}

func statsRequest(address string, path string) ([]byte, error) {
	url := fmt.Sprintf("http://%s/%s", address, path)
	req, err := http.NewRequest("GET", url, nil)
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
