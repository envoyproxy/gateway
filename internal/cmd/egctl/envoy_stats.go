// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package egctl

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"

	kube "github.com/envoyproxy/gateway/internal/kubernetes"
	"github.com/spf13/cobra"
	"github.com/tetratelabs/multierror"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/yaml"
)

const (
	summaryOutput          = "short"
	prometheusOutput       = "prom"
	prometheusMergedOutput = "prom-merged"

	defaultProxyAdminPort = 15000
)

var (
	statsType, outputFormat string
)

func newEnvoyStatsConfigCmd() *cobra.Command {
	var podName, podNamespace string

	statsConfigCmd := &cobra.Command{
		Use:   "envoy-stats [<type>/]<name>[.<namespace>]",
		Short: "Retrieves Envoy metrics in the specified pod",
		Long:  `Retrieve Envoy emitted metrics for the specified pod.`,
		Example: `  # Retrieve Envoy emitted metrics for the specified pod.
  egctl experimental envoy-stats <pod-name[.namespace]>

  # Retrieve Envoy server metrics in prometheus format
  egctl experimental envoy-stats <pod-name[.namespace]> --output prom

  # Retrieve Envoy server metrics in prometheus format with merged application metrics
  egctl experimental envoy-stats <pod-name[.namespace]> --output prom-merged

  # Retrieve Envoy cluster metrics
  egctl experimental envoy-stats <pod-name[.namespace]> --type clusters
`,
		Aliases: []string{"es"},
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
			if statsType == "" || statsType == "server" {
				wg.Add(len(pods))
				for _, pod := range pods {
					go func(pod types.NamespacedName) {
						stats[pod.Namespace+"/"+pod.Name], err = setupEnvoyServerStatsConfig(kubeClient, pod.Name, pod.Namespace, outputFormat)
						multierror.Append(errs, err)
						wg.Done()
					}(pod)
				}
				wg.Wait()
			} else if statsType == "cluster" || statsType == "clusters" {
				wg.Add(len(pods))
				for _, pod := range pods {
					go func(pod types.NamespacedName) {
						stats[pod.Namespace+"/"+pod.Name], err = setupEnvoyClusterStatsConfig(kubeClient, pod.Name, pod.Namespace, outputFormat)
						multierror.Append(errs, err)
						wg.Done()
					}(pod)
				}
				wg.Wait()
			} else {
				return fmt.Errorf("unknown stats type %s", statsType)
			}

			if errs != nil {
				return errs
			}

			switch outputFormat {
			// convert the json output to yaml
			case yamlOutput:
				var out []byte
				statsBytes, err := json.Marshal(stats)
				if err != nil {
					return err
				}
				if out, err = yaml.JSONToYAML([]byte(statsBytes)); err != nil {
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
	statsConfigCmd.PersistentFlags().StringVarP(&outputFormat, "output", "o", summaryOutput, "Output format: one of json|yaml|prom")
	statsConfigCmd.PersistentFlags().StringVarP(&statsType, "type", "t", "server", "Where to grab the stats: one of server|clusters")
	statsConfigCmd.PersistentFlags().StringArrayVarP(&labelSelectors, "labels", "l", nil, "Labels to select the envoy proxy pod.")
	statsConfigCmd.PersistentFlags().StringVarP(&podNamespace, "namespace", "n", "envoy-gateway-system", "Namespace where envoy proxy pod are installed.")

	return statsConfigCmd
}

func setupEnvoyServerStatsConfig(kubeClient kube.CLIClient, podName, podNamespace string, outputFormat string) (string, error) {
	path := "stats"
	port := 19000
	if outputFormat == jsonOutput || outputFormat == yamlOutput {
		// for yaml output we will convert the json to yaml when printed
		path += "?format=json"
	} else {
		path += "/prometheus"
	}

	fw, err := portForwarder(kubeClient, types.NamespacedName{Namespace: podNamespace, Name: podName}, port)
	err = fw.Start()
	if err != nil {
		return "", fmt.Errorf("failed to start port forwarding for pod %s/%s: %v", podNamespace, podName, err)
	}
	defer fw.Stop()

	result, err := statsRequest(fw.Address(), path)
	if err != nil {
		return "", fmt.Errorf("failed to get stats on envoy for pod %s/%s: %v", podNamespace, podName, err)
	}
	return string(result), nil
}

func setupEnvoyClusterStatsConfig(kubeClient kube.CLIClient, podName, podNamespace string, outputFormat string) (string, error) {
	path := "clusters"
	port := 19000
	if outputFormat == jsonOutput || outputFormat == yamlOutput {
		// for yaml output we will convert the json to yaml when printed
		path += "?format=json"
	}
	fw, err := portForwarder(kubeClient, types.NamespacedName{Namespace: podNamespace, Name: podName}, port)
	err = fw.Start()
	if err != nil {
		return "", fmt.Errorf("failed to start port forwarding for pod %s/%s: %v", podNamespace, podName, err)
	}
	defer fw.Stop()

	result, err := statsRequest(fw.Address(), path)
	if err != nil {
		return "", fmt.Errorf("failed to get stats on envoy for pod %s/%s: %v", podNamespace, podName, err)
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
