// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package egctl

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sync"

	adminv3 "github.com/envoyproxy/go-control-plane/envoy/admin/v3"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/reflect/protoreflect"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/yaml"

	"github.com/envoyproxy/gateway/internal/cmd/options"
	"github.com/envoyproxy/gateway/internal/infrastructure/kubernetes/proxy"
	kube "github.com/envoyproxy/gateway/internal/kubernetes"
	"github.com/envoyproxy/gateway/internal/utils"
)

var (
	output         string
	podName        string
	podNamespace   string
	labelSelectors []string
	allNamespaces  bool
)

const (
	adminPort          = 19000   // TODO: make this configurable until EG support
	rateLimitDebugPort = 6070    // TODO: make this configurable until EG support
	containerName      = "envoy" // TODO: make this configurable until EG support
)

type aggregatedConfigDump map[string]map[string]protoreflect.ProtoMessage

func retrieveConfigDump(args []string, includeEds bool, configType envoyConfigType) (aggregatedConfigDump, error) {
	if !allNamespaces {
		if len(labelSelectors) == 0 {
			if len(args) != 0 && args[0] != "" {
				podName = args[0]
			}
		}

		if podNamespace == "" {
			return nil, fmt.Errorf("pod namespace is required")
		}
	}

	cli, err := getCLIClient()
	if err != nil {
		return nil, err
	}

	pods, err := fetchRunningEnvoyPods(cli, types.NamespacedName{Namespace: podNamespace, Name: podName}, labelSelectors, allNamespaces)
	if err != nil {
		return nil, err
	}

	podConfigDumps := make(aggregatedConfigDump, 0)
	// Initialize the map with namespaces
	for _, pod := range pods {
		if _, ok := podConfigDumps[pod.Namespace]; !ok {
			podConfigDumps[pod.Namespace] = make(map[string]protoreflect.ProtoMessage)
		}
	}

	var errs error
	var wg sync.WaitGroup
	wg.Add(len(pods))
	for _, pod := range pods {
		pod := pod
		go func() {
			fw, err := portForwarder(cli, pod, adminPort)
			if err != nil {
				errs = errors.Join(errs, err)
				return
			}

			if err := fw.Start(); err != nil {
				errs = errors.Join(errs, err)
				return
			}
			defer fw.Stop()
			defer wg.Done()

			configDump, err := extractConfigDump(fw, includeEds, configType)
			if err != nil {
				errs = errors.Join(errs, err)
				return
			}

			podConfigDumps[pod.Namespace][pod.Name] = configDump
		}()
	}

	wg.Wait()
	if errs != nil {
		return nil, errs
	}

	return podConfigDumps, nil
}

// fetchRunningEnvoyPods gets the Pods, either based on the NamespacedName or the labelSelectors.
// It further filters out only those Pods that are in "Running" state.
// labelSelectors, if provided, take precedence over the pod NamespacedName.
func fetchRunningEnvoyPods(c kube.CLIClient, nn types.NamespacedName, labelSelectors []string, allNamespaces bool) ([]types.NamespacedName, error) {
	var pods []corev1.Pod

	switch {
	case allNamespaces:
		namespaces, err := c.Kube().CoreV1().Namespaces().List(context.Background(), v1.ListOptions{})
		if err != nil {
			return nil, err
		}
		for _, i := range namespaces.Items {
			podList, err := c.PodsForSelector(i.Name, proxy.EnvoyAppLabelSelector()...)
			if err != nil {
				return nil, fmt.Errorf("list pods failed in ns %s: %w", i.Name, err)
			}

			if len(podList.Items) == 0 {
				continue
			}

			pods = append(pods, podList.Items...)
		}
	case len(labelSelectors) > 0:
		podList, err := c.PodsForSelector(nn.Namespace, labelSelectors...)
		if err != nil {
			return nil, fmt.Errorf("get pod %s fail: %w", nn, err)
		}

		if len(podList.Items) == 0 {
			return nil, fmt.Errorf("no Pods found for label selectors %+v", labelSelectors)
		}

		pods = podList.Items
	case nn.Name != "":
		pod, err := c.Pod(nn)
		if err != nil {
			return nil, fmt.Errorf("get pod %s fail: %w", nn, err)
		}

		pods = []corev1.Pod{*pod}

	case nn.Name == "":
		podList, err := c.PodsForSelector(nn.Namespace, proxy.EnvoyAppLabelSelector()...)
		if err != nil {
			return nil, fmt.Errorf("get pod %s fail: %w", nn, err)
		}

		if len(podList.Items) == 0 {
			return nil, fmt.Errorf("no Pods found for label selectors %+v", proxy.EnvoyAppLabelSelector())
		}

		pods = podList.Items
	}

	podsNamespacedNames := []types.NamespacedName{}
	for _, pod := range pods {
		pod := pod
		podNsName := utils.NamespacedName(&pod)
		if pod.Status.Phase != "Running" {
			return podsNamespacedNames, fmt.Errorf("pod %s is not running", podNsName)
		}

		podsNamespacedNames = append(podsNamespacedNames, podNsName)
	}

	return podsNamespacedNames, nil
}

// portForwarder returns a port forwarder instance for a single Pod.
func portForwarder(cli kube.CLIClient, nn types.NamespacedName, port int) (kube.PortForwarder, error) {
	fw, err := kube.NewLocalPortForwarder(cli, nn, 0, port)
	if err != nil {
		return nil, err
	}

	return fw, nil
}

// getCLIClient returns a new kubernetes CLI Client.
func getCLIClient() (kube.CLIClient, error) {
	c, err := kube.NewCLIClient(options.DefaultConfigFlags.ToRawKubeConfigLoader())
	if err != nil {
		return nil, fmt.Errorf("build CLI client fail: %w", err)
	}

	return c, nil
}

func marshalEnvoyProxyConfig(configDump aggregatedConfigDump, output string) ([]byte, error) {
	configDumpMap := make(map[string]map[string]interface{})
	for ns, nsConfigs := range configDump {
		configDumpMap[ns] = make(map[string]interface{})
		for pod, podConfigs := range nsConfigs {
			var newConfig interface{}
			if err := json.Unmarshal([]byte(protojson.MarshalOptions{Multiline: false}.Format(podConfigs)), &newConfig); err != nil {
				return nil, err
			}
			configDumpMap[ns][pod] = newConfig
		}
	}

	out, err := json.MarshalIndent(configDumpMap, "", "  ")
	if output == "yaml" {
		return yaml.JSONToYAML(out)
	}

	return out, err
}

func extractConfigDump(fw kube.PortForwarder, includeEds bool, configType envoyConfigType) (protoreflect.ProtoMessage, error) {
	out, err := configDumpRequest(fw.Address(), includeEds)
	if err != nil {
		return nil, err
	}

	configDumpResponse := &adminv3.ConfigDump{}
	if err := protojson.Unmarshal(out, configDumpResponse); err != nil {
		return nil, err
	}

	return findXDSResourceFromConfigDump(configType, configDumpResponse)
}

func configDumpRequest(address string, includeEds bool) ([]byte, error) {
	url := fmt.Sprintf("http://%s/config_dump", address)
	if includeEds {
		url = fmt.Sprintf("%s?include_eds", url)
	}
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
