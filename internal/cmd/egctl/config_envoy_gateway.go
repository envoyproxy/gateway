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
	"net/url"
	"strings"
	"sync"

	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
	"sigs.k8s.io/yaml"

	kube "github.com/envoyproxy/gateway/internal/kubernetes"
	"github.com/envoyproxy/gateway/internal/utils"
)

const (
	envoyGatewayDefaultLabelSelector = "control-plane=envoy-gateway"
	envoyGatewayConfigDumpEndpoint   = "/api/config_dump"
)

type aggregatedEnvoyGatewayConfigDump map[string]map[string]interface{}

type envoyGatewayConfigType string

const (
	AllEnvoyGatewayConfigType                  envoyGatewayConfigType = "all"
	GatewayClassEnvoyGatewayConfigType         envoyGatewayConfigType = "gatewayclass"
	GatewayEnvoyGatewayConfigType              envoyGatewayConfigType = "gateway"
	HTTPRouteEnvoyGatewayConfigType            envoyGatewayConfigType = "httproute"
	GRPCRouteEnvoyGatewayConfigType            envoyGatewayConfigType = "grpcroute"
	TLSRouteEnvoyGatewayConfigType             envoyGatewayConfigType = "tlsroute"
	TCPRouteEnvoyGatewayConfigType             envoyGatewayConfigType = "tcproute"
	UDPRouteEnvoyGatewayConfigType             envoyGatewayConfigType = "udproute"
	ClientTrafficPolicyEnvoyGatewayConfigType  envoyGatewayConfigType = "clienttrafficpolicy"
	BackendTrafficPolicyEnvoyGatewayConfigType envoyGatewayConfigType = "backendtrafficpolicy"
	BackendTLSPolicyEnvoyGatewayConfigType     envoyGatewayConfigType = "backendtlspolicy"
	SecurityPolicyEnvoyGatewayConfigType       envoyGatewayConfigType = "securitypolicy"
	EnvoyPatchPolicyEnvoyGatewayConfigType     envoyGatewayConfigType = "envoypatchpolicy"
	EnvoyExtensionPolicyEnvoyGatewayConfigType envoyGatewayConfigType = "envoyextensionpolicy"
	ServiceEnvoyGatewayConfigType              envoyGatewayConfigType = "service"
	SecretEnvoyGatewayConfigType               envoyGatewayConfigType = "secret"
	ConfigMapEnvoyGatewayConfigType            envoyGatewayConfigType = "configmap"
	NamespaceEnvoyGatewayConfigType            envoyGatewayConfigType = "namespace"
	EndpointSliceEnvoyGatewayConfigType        envoyGatewayConfigType = "endpointslice"
	ReferenceGrantEnvoyGatewayConfigType       envoyGatewayConfigType = "referencegrant"
	HTTPRouteFilterEnvoyGatewayConfigType      envoyGatewayConfigType = "httproutefilter"
	EnvoyProxyEnvoyGatewayConfigType           envoyGatewayConfigType = "envoyproxy"
	BackendEnvoyGatewayConfigType              envoyGatewayConfigType = "backend"
	ServiceImportEnvoyGatewayConfigType        envoyGatewayConfigType = "serviceimport"
)

func gatewayCommand() *cobra.Command {
	c := &cobra.Command{
		Use:     "envoy-gateway",
		Aliases: []string{"gateway"},
		Long:    "Retrieve information from envoy gateway.",
	}

	c.AddCommand(allEnvoyGatewayConfigCmd())
	for _, spec := range typedEnvoyGatewayConfigSpecs() {
		c.AddCommand(typedEnvoyGatewayConfigCmd(&spec))
	}

	return c
}

func allEnvoyGatewayConfigCmd() *cobra.Command {
	configCmd := &cobra.Command{
		Use:   "all <pod-name>",
		Short: "Retrieves all Envoy Gateway resources from the specified pod",
		Long:  `Retrieves information about all Envoy Gateway resources from the Envoy Gateway admin API in the specified pod.`,
		Example: `  # Retrieve all configuration for a given pod from Envoy Gateway.
  egctl config envoy-gateway all <pod-name> -n <pod-namespace>

  # Retrieve all configuration for a pod matching label selectors
  egctl config envoy-gateway all --labels control-plane=envoy-gateway

  # Retrieve configuration as YAML
  egctl config envoy-gateway all <pod-name> -n <pod-namespace> -o yaml

  # Retrieve configuration with short syntax
  egctl c gateway all <pod-name> -n <pod-namespace>
`,
		Run: func(c *cobra.Command, args []string) {
			cmdutil.CheckErr(runEnvoyGatewayConfig(c, args, AllEnvoyGatewayConfigType))
		},
	}

	return configCmd
}

type typedEnvoyGatewayConfigSpec struct {
	use        string
	command    string
	display    string
	short      string
	aliases    []string
	configType envoyGatewayConfigType
}

func typedEnvoyGatewayConfigSpecs() []typedEnvoyGatewayConfigSpec {
	return []typedEnvoyGatewayConfigSpec{
		{use: string(GatewayClassEnvoyGatewayConfigType) + " <pod-name>", command: string(GatewayClassEnvoyGatewayConfigType), display: "GatewayClass", short: "Retrieves GatewayClass resources from the specified pod", aliases: []string{"gc"}, configType: GatewayClassEnvoyGatewayConfigType},
		{use: string(GatewayEnvoyGatewayConfigType) + " <pod-name>", command: string(GatewayEnvoyGatewayConfigType), display: "Gateway", short: "Retrieves Gateway resources from the specified pod", aliases: []string{"gw"}, configType: GatewayEnvoyGatewayConfigType},
		{use: string(HTTPRouteEnvoyGatewayConfigType) + " <pod-name>", command: string(HTTPRouteEnvoyGatewayConfigType), display: "HTTPRoute", short: "Retrieves HTTPRoute resources from the specified pod", aliases: []string{"hr"}, configType: HTTPRouteEnvoyGatewayConfigType},
		{use: string(GRPCRouteEnvoyGatewayConfigType) + " <pod-name>", command: string(GRPCRouteEnvoyGatewayConfigType), display: "GRPCRoute", short: "Retrieves GRPCRoute resources from the specified pod", aliases: []string{"gr"}, configType: GRPCRouteEnvoyGatewayConfigType},
		{use: string(TLSRouteEnvoyGatewayConfigType) + " <pod-name>", command: string(TLSRouteEnvoyGatewayConfigType), display: "TLSRoute", short: "Retrieves TLSRoute resources from the specified pod", aliases: []string{"tr"}, configType: TLSRouteEnvoyGatewayConfigType},
		{use: string(TCPRouteEnvoyGatewayConfigType) + " <pod-name>", command: string(TCPRouteEnvoyGatewayConfigType), display: "TCPRoute", short: "Retrieves TCPRoute resources from the specified pod", aliases: []string{"tcr"}, configType: TCPRouteEnvoyGatewayConfigType},
		{use: string(UDPRouteEnvoyGatewayConfigType) + " <pod-name>", command: string(UDPRouteEnvoyGatewayConfigType), display: "UDPRoute", short: "Retrieves UDPRoute resources from the specified pod", aliases: []string{"ur"}, configType: UDPRouteEnvoyGatewayConfigType},
		{use: string(ClientTrafficPolicyEnvoyGatewayConfigType) + " <pod-name>", command: string(ClientTrafficPolicyEnvoyGatewayConfigType), display: "ClientTrafficPolicy", short: "Retrieves ClientTrafficPolicy resources from the specified pod", aliases: []string{"ctp"}, configType: ClientTrafficPolicyEnvoyGatewayConfigType},
		{use: string(BackendTrafficPolicyEnvoyGatewayConfigType) + " <pod-name>", command: string(BackendTrafficPolicyEnvoyGatewayConfigType), display: "BackendTrafficPolicy", short: "Retrieves BackendTrafficPolicy resources from the specified pod", aliases: []string{"btp"}, configType: BackendTrafficPolicyEnvoyGatewayConfigType},
		{use: string(BackendTLSPolicyEnvoyGatewayConfigType) + " <pod-name>", command: string(BackendTLSPolicyEnvoyGatewayConfigType), display: "BackendTLSPolicy", short: "Retrieves BackendTLSPolicy resources from the specified pod", aliases: []string{"btlsp"}, configType: BackendTLSPolicyEnvoyGatewayConfigType},
		{use: string(SecurityPolicyEnvoyGatewayConfigType) + " <pod-name>", command: string(SecurityPolicyEnvoyGatewayConfigType), display: "SecurityPolicy", short: "Retrieves SecurityPolicy resources from the specified pod", aliases: []string{"sp"}, configType: SecurityPolicyEnvoyGatewayConfigType},
		{use: string(EnvoyPatchPolicyEnvoyGatewayConfigType) + " <pod-name>", command: string(EnvoyPatchPolicyEnvoyGatewayConfigType), display: "EnvoyPatchPolicy", short: "Retrieves EnvoyPatchPolicy resources from the specified pod", aliases: []string{"epp"}, configType: EnvoyPatchPolicyEnvoyGatewayConfigType},
		{use: string(EnvoyExtensionPolicyEnvoyGatewayConfigType) + " <pod-name>", command: string(EnvoyExtensionPolicyEnvoyGatewayConfigType), display: "EnvoyExtensionPolicy", short: "Retrieves EnvoyExtensionPolicy resources from the specified pod", aliases: []string{"eep"}, configType: EnvoyExtensionPolicyEnvoyGatewayConfigType},
		{use: string(ServiceEnvoyGatewayConfigType) + " <pod-name>", command: string(ServiceEnvoyGatewayConfigType), display: "Service", short: "Retrieves Service resources from the specified pod", aliases: []string{"svc"}, configType: ServiceEnvoyGatewayConfigType},
		{use: string(SecretEnvoyGatewayConfigType) + " <pod-name>", command: string(SecretEnvoyGatewayConfigType), display: "Secret", short: "Retrieves Secret resources from the specified pod", aliases: []string{"sec"}, configType: SecretEnvoyGatewayConfigType},
		{use: string(ConfigMapEnvoyGatewayConfigType) + " <pod-name>", command: string(ConfigMapEnvoyGatewayConfigType), display: "ConfigMap", short: "Retrieves ConfigMap resources from the specified pod", aliases: []string{"cm"}, configType: ConfigMapEnvoyGatewayConfigType},
		{use: string(NamespaceEnvoyGatewayConfigType) + " <pod-name>", command: string(NamespaceEnvoyGatewayConfigType), display: "Namespace", short: "Retrieves Namespace resources from the specified pod", aliases: []string{"ns"}, configType: NamespaceEnvoyGatewayConfigType},
		{use: string(EndpointSliceEnvoyGatewayConfigType) + " <pod-name>", command: string(EndpointSliceEnvoyGatewayConfigType), display: "EndpointSlice", short: "Retrieves EndpointSlice resources from the specified pod", aliases: []string{"eps"}, configType: EndpointSliceEnvoyGatewayConfigType},
		{use: string(ReferenceGrantEnvoyGatewayConfigType) + " <pod-name>", command: string(ReferenceGrantEnvoyGatewayConfigType), display: "ReferenceGrant", short: "Retrieves ReferenceGrant resources from the specified pod", aliases: []string{"rg"}, configType: ReferenceGrantEnvoyGatewayConfigType},
		{use: string(HTTPRouteFilterEnvoyGatewayConfigType) + " <pod-name>", command: string(HTTPRouteFilterEnvoyGatewayConfigType), display: "HTTPRouteFilter", short: "Retrieves HTTPRouteFilter resources from the specified pod", aliases: []string{"hrf"}, configType: HTTPRouteFilterEnvoyGatewayConfigType},
		{use: string(EnvoyProxyEnvoyGatewayConfigType) + " <pod-name>", command: string(EnvoyProxyEnvoyGatewayConfigType), display: "EnvoyProxy", short: "Retrieves EnvoyProxy resources from the specified pod", aliases: []string{"ep"}, configType: EnvoyProxyEnvoyGatewayConfigType},
		{use: string(BackendEnvoyGatewayConfigType) + " <pod-name>", command: string(BackendEnvoyGatewayConfigType), display: "Backend", short: "Retrieves Backend resources from the specified pod", aliases: []string{"be"}, configType: BackendEnvoyGatewayConfigType},
		{use: string(ServiceImportEnvoyGatewayConfigType) + " <pod-name>", command: string(ServiceImportEnvoyGatewayConfigType), display: "ServiceImport", short: "Retrieves ServiceImport resources from the specified pod", aliases: []string{"si"}, configType: ServiceImportEnvoyGatewayConfigType},
	}
}

func typedEnvoyGatewayConfigCmd(spec *typedEnvoyGatewayConfigSpec) *cobra.Command {
	shortAlias := spec.command
	if len(spec.aliases) > 0 {
		shortAlias = spec.aliases[0]
	}
	configCmd := &cobra.Command{
		Use:   spec.use,
		Short: spec.short,
		Long:  fmt.Sprintf("Retrieves information about %s Envoy Gateway resources from the Envoy Gateway admin API in the specified pod.", spec.display),
		Example: fmt.Sprintf(`  # Retrieve %s configuration for a given pod from Envoy Gateway.
  egctl config envoy-gateway %s <pod-name> -n <pod-namespace>

  # Retrieve %s configuration for a pod matching label selectors
  egctl config envoy-gateway %s --labels control-plane=envoy-gateway

  # Retrieve configuration as YAML
  egctl config envoy-gateway %s <pod-name> -n <pod-namespace> -o yaml

  # Retrieve configuration with short syntax
  egctl c gateway %s <pod-name> -n <pod-namespace>
`, spec.display, spec.command, spec.display, spec.command, spec.command, shortAlias),
		Aliases: spec.aliases,
		Run: func(c *cobra.Command, args []string) {
			cmdutil.CheckErr(runEnvoyGatewayConfig(c, args, spec.configType))
		},
	}

	return configCmd
}

func runEnvoyGatewayConfig(c *cobra.Command, args []string, configType envoyGatewayConfigType) error {
	configDump, err := retrieveEnvoyGatewayConfigDump(args, configType)
	if err != nil {
		return err
	}

	out, err := marshalEnvoyGatewayConfig(configDump, output)
	if err != nil {
		return err
	}

	_, err = fmt.Fprintln(c.OutOrStdout(), string(out))
	return err
}

func retrieveEnvoyGatewayConfigDump(args []string, configType envoyGatewayConfigType) (aggregatedEnvoyGatewayConfigDump, error) {
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

	pods, err := fetchRunningEnvoyGatewayPods(cli, types.NamespacedName{Namespace: podNamespace, Name: podName}, labelSelectors, allNamespaces)
	if err != nil {
		return nil, err
	}

	podConfigDumps := make(aggregatedEnvoyGatewayConfigDump)
	mu := sync.Mutex{}
	errMu := sync.Mutex{}

	for _, pod := range pods {
		if _, ok := podConfigDumps[pod.Namespace]; !ok {
			podConfigDumps[pod.Namespace] = make(map[string]interface{})
		}
	}

	var errs error
	var wg sync.WaitGroup
	wg.Add(len(pods))
	for _, pod := range pods {
		go func() {
			defer wg.Done()

			fw, err := portForwarder(cli, pod, adminPort)
			if err != nil {
				errMu.Lock()
				errs = errors.Join(errs, err)
				errMu.Unlock()
				return
			}

			if err := fw.Start(); err != nil {
				errMu.Lock()
				errs = errors.Join(errs, err)
				errMu.Unlock()
				return
			}
			defer fw.Stop()

			configDump, err := extractEnvoyGatewayConfigDump(fw, configType)
			if err != nil {
				errMu.Lock()
				errs = errors.Join(errs, err)
				errMu.Unlock()
				return
			}

			mu.Lock()
			podConfigDumps[pod.Namespace][pod.Name] = configDump
			mu.Unlock()
		}()
	}

	wg.Wait()
	if errs != nil {
		return nil, errs
	}

	return podConfigDumps, nil
}

func fetchRunningEnvoyGatewayPods(c kube.CLIClient, nn types.NamespacedName, labelSelectors []string, allNamespaces bool) ([]types.NamespacedName, error) {
	var pods []corev1.Pod
	defaultSelector := []string{envoyGatewayDefaultLabelSelector}

	switch {
	case allNamespaces:
		namespaces, err := c.Kube().CoreV1().Namespaces().List(context.Background(), metav1.ListOptions{})
		if err != nil {
			return nil, err
		}
		selectors := defaultSelector
		if len(labelSelectors) > 0 {
			selectors = labelSelectors
		}
		for i := range namespaces.Items {
			podList, err := c.PodsForSelector(namespaces.Items[i].Name, selectors...)
			if err != nil {
				return nil, fmt.Errorf("list pods failed in ns %s: %w", namespaces.Items[i].Name, err)
			}
			if len(podList.Items) == 0 {
				continue
			}
			pods = append(pods, podList.Items...)
		}
		if len(pods) == 0 {
			return nil, fmt.Errorf("no Pods found for label selectors %+v", selectors)
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
	default:
		podList, err := c.PodsForSelector(nn.Namespace, defaultSelector...)
		if err != nil {
			return nil, fmt.Errorf("get pod %s fail: %w", nn, err)
		}
		if len(podList.Items) == 0 {
			return nil, fmt.Errorf("no Pods found for label selectors %+v", defaultSelector)
		}
		pods = podList.Items
	}

	podsNamespacedNames := make([]types.NamespacedName, 0, len(pods))
	for i := range pods {
		podNsName := utils.NamespacedName(&pods[i])
		if pods[i].Status.Phase != corev1.PodRunning {
			return podsNamespacedNames, fmt.Errorf("pod %s is not running", podNsName)
		}
		podsNamespacedNames = append(podsNamespacedNames, podNsName)
	}

	return podsNamespacedNames, nil
}

func extractEnvoyGatewayConfigDump(fw kube.PortForwarder, configType envoyGatewayConfigType) (interface{}, error) {
	out, err := envoyGatewayConfigDumpRequest(fw.Address(), configType)
	if err != nil {
		return nil, err
	}

	response := struct {
		Resources []json.RawMessage `json:"resources"`
	}{}
	if err := json.Unmarshal(out, &response); err != nil {
		return nil, err
	}

	items := make([]interface{}, 0, len(response.Resources))
	for _, raw := range response.Resources {
		var item interface{}
		if err := json.Unmarshal(raw, &item); err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	return items, nil
}

func marshalEnvoyGatewayConfig(configDump aggregatedEnvoyGatewayConfigDump, output string) ([]byte, error) {
	out, err := json.MarshalIndent(configDump, "", "  ")
	if err != nil {
		return nil, err
	}

	if output == "yaml" {
		return yaml.JSONToYAML(out)
	}

	return out, nil
}

func envoyGatewayConfigDumpRequest(address string, configType envoyGatewayConfigType) ([]byte, error) {
	query := url.Values{}
	query.Set("resource", string(configType))
	url := fmt.Sprintf("http://%s%s?%s", address, envoyGatewayConfigDumpEndpoint, query.Encode())

	ctx, cancel := context.WithTimeout(context.Background(), configRequestTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
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

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		bodyText := strings.TrimSpace(string(body))
		if len(bodyText) > 200 {
			bodyText = bodyText[:200] + "..."
		}
		return nil, fmt.Errorf("request to %s failed with status %d: %s", envoyGatewayConfigDumpEndpoint, resp.StatusCode, bodyText)
	}

	return body, nil
}
