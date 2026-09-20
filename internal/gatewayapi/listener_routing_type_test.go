// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0

package gatewayapi

import (
	"fmt"
	"io"
	"os"
	"slices"
	"strings"
	"testing"

	clusterv3 "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	endpointv3 "github.com/envoyproxy/go-control-plane/envoy/config/endpoint/v3"
	resourcev3 "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	"github.com/stretchr/testify/require"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/gatewayapi/resource"
	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/logging"
	xdstranslator "github.com/envoyproxy/gateway/internal/xds/translator"
)

// Exercise listener routing through both translators, including cluster de-duplication.
func TestListenerRoutingType(t *testing.T) {
	for _, kind := range []string{"HTTP", "GRPC", "TCP", "TLS", "UDP"} {
		for _, merge := range []bool{false, true} {
			for _, explicit := range []bool{false, true} {
				for _, scenario := range []string{"gateway", "reverse", "listener-set", "headless", "serviceimport", "weighted", "mirror", "route", "rule", "replace", "tls-pool"} {
					if (scenario == "mirror" || scenario == "weighted") && kind != "HTTP" && kind != "GRPC" {
						continue
					}
					if scenario == "tls-pool" && kind != "TLS" {
						continue
					}
					mode := "omitted"
					if explicit {
						mode = "explicit"
					}
					t.Run(fmt.Sprintf("%s/merge=%t/%s/%s", kind, merge, mode, scenario), func(t *testing.T) {
						protocol, transport, collection, version := kind, "TCP", strings.ToLower(kind)+"Routes", "v1"
						if kind == "GRPC" {
							protocol = "HTTP"
						}
						if kind == "UDP" {
							transport = "UDP"
						}
						if kind == "UDP" || kind == "TCP" {
							version = "v1alpha2"
						}
						tlsConfig, hostnames := "", ""
						if kind == "TLS" {
							tlsConfig = "\n      tls: {mode: Passthrough}"
							hostnames = "\n    hostnames: [example.test]"
						}
						parents := "    - {name: gateway-1}"
						if explicit {
							parents = "    - {name: gateway-1, sectionName: listener-1}\n    - {name: gateway-1, sectionName: listener-2}"
						}
						input := fmt.Sprintf(`namespaces:
- metadata: {name: default}
gateways:
- apiVersion: gateway.networking.k8s.io/v1
  kind: Gateway
  metadata: {name: gateway-1, namespace: default}
  spec:
    gatewayClassName: envoy-gateway-class
    listeners:
    - name: listener-1
      protocol: %s
      port: 9081%s
    - name: listener-2
      protocol: %s
      port: 9082%s
services:
- apiVersion: v1
  kind: Service
  metadata: {name: backend, namespace: default}
  spec:
    clusterIP: 10.96.0.7
    ports:
    - {name: backend, port: 8080, targetPort: 8080, protocol: %s}
endpointSlices:
- apiVersion: discovery.k8s.io/v1
  kind: EndpointSlice
  metadata:
    name: backend-1
    namespace: default
    labels: {kubernetes.io/service-name: backend}
  addressType: IPv4
  ports:
  - {name: backend, port: 8080, protocol: %s}
  endpoints:
  - addresses: [10.0.0.7]
    conditions: {ready: true}
backendTrafficPolicies:
- apiVersion: gateway.envoyproxy.io/v1alpha1
  kind: BackendTrafficPolicy
  metadata: {name: gateway-routing, namespace: default}
  spec:
    targetRefs:
    - {group: gateway.networking.k8s.io, kind: Gateway, name: gateway-1}
    routingType: Endpoint
- apiVersion: gateway.envoyproxy.io/v1alpha1
  kind: BackendTrafficPolicy
  metadata: {name: listener-routing, namespace: default}
  spec:
    targetRefs:
    - {group: gateway.networking.k8s.io, kind: Gateway, name: gateway-1, sectionName: listener-1}
    routingType: Service
%s:
- apiVersion: gateway.networking.k8s.io/%s
  kind: %sRoute
  metadata: {name: route-1, namespace: default}
  spec:
    parentRefs:
%s%s
    rules:
    - name: rule-1
      backendRefs:
      - {name: backend, port: 8080}
`, protocol, tlsConfig, protocol, tlsConfig, transport, transport, collection, version, kind, parents, hostnames)
						listenerPrefix := "default/gateway-1/"
						if scenario == "listener-set" {
							start := strings.Index(input, "gateways:")
							end := strings.Index(input, "services:")
							gateway := input[start:end]
							ls := strings.Replace(gateway, "gateways:", "listenerSets:", 1)
							ls = strings.Replace(ls, "kind: Gateway", "kind: ListenerSet", 1)
							ls = strings.Replace(ls, "name: gateway-1, namespace:", "name: extra, namespace:", 1)
							ls = strings.Replace(ls, "gatewayClassName: envoy-gateway-class", "parentRef: {name: gateway-1}", 1)
							gateway = strings.Replace(gateway, "gatewayClassName: envoy-gateway-class", "gatewayClassName: envoy-gateway-class\n    allowedListeners: {namespaces: {from: All}}", 1)
							gateway = strings.ReplaceAll(gateway, "908", "909")
							input = input[:start] + gateway + ls + input[end:]
							input = strings.ReplaceAll(input, "{name: gateway-1}", "{name: extra, group: gateway.networking.k8s.io, kind: ListenerSet}")
							// Restore the ListenerSet's own Gateway parent.
							input = strings.Replace(input, "parentRef: {name: extra, group: gateway.networking.k8s.io, kind: ListenerSet}", "parentRef: {name: gateway-1}", 1)
							input = strings.ReplaceAll(input, "{name: gateway-1, sectionName:", "{name: extra, group: gateway.networking.k8s.io, kind: ListenerSet, sectionName:")
							input = strings.Replace(input, "kind: Gateway, name: gateway-1, sectionName: listener-1", "kind: ListenerSet, name: extra, sectionName: listener-1", 1)
							listenerPrefix += "default/extra/"
							collision := "- apiVersion: gateway.envoyproxy.io/v1alpha1\n  kind: BackendTrafficPolicy\n  metadata: {name: unrelated-gateway-listener, namespace: default}\n  spec:\n    targetRefs:\n    - {group: gateway.networking.k8s.io, kind: Gateway, name: gateway-1, sectionName: listener-2}\n    routingType: Service\n"
							input = strings.Replace(input, collection+":", collision+collection+":", 1)
						}
						if scenario == "weighted" {
							input = strings.Replace(input, "      - {name: backend, port: 8080}", "      - name: backend\n        port: 8080\n        weight: 2\n        filters:\n        - type: RequestHeaderModifier\n          requestHeaderModifier: {add: [{name: x-probe, value: listener}]}\n      - {name: backend, port: 8080, weight: 3}", 1)
						}
						if scenario == "serviceimport" {
							input = strings.Replace(input, "services:\n- apiVersion: v1\n  kind: Service", "serviceImports:\n- apiVersion: multicluster.x-k8s.io/v1alpha1\n  kind: ServiceImport", 1)
							input = strings.Replace(input, "clusterIP: 10.96.0.7", "ips: [10.96.0.7]\n    type: ClusterSetIP", 1)
							input = strings.ReplaceAll(input, ", targetPort: 8080", "")
							input = strings.ReplaceAll(input, "kubernetes.io/service-name", "multicluster.kubernetes.io/service-name")
							input = strings.ReplaceAll(input, "{name: backend, port: 8080}", "{name: backend, port: 8080, kind: ServiceImport, group: multicluster.x-k8s.io}")
						}
						if scenario == "mirror" {
							input += "      filters:\n      - type: RequestMirror\n        requestMirror:\n          backendRef: {name: backend, port: 8080}\n"
						}
						if scenario == "tls-pool" {
							input += "    - name: rule-2\n      backendRefs:\n      - {name: backend, port: 8080}\n"
						}
						if scenario == "route" || scenario == "rule" || scenario == "replace" || scenario == "tls-pool" {
							policy := fmt.Sprintf("- apiVersion: gateway.envoyproxy.io/v1alpha1\n  kind: BackendTrafficPolicy\n  metadata: {name: route-routing, namespace: default}\n  spec:\n    targetRefs:\n    - {group: gateway.networking.k8s.io, kind: %sRoute, name: route-1", kind)
							if scenario == "rule" {
								policy += ", sectionName: rule-1"
							}
							if scenario == "tls-pool" {
								policy += ", sectionName: rule-2"
							}
							policy += "}\n"
							if scenario != "replace" {
								policy += "    routingType: Endpoint\n"
							}
							if scenario == "replace" {
								policy += "    loadBalancer: {type: RoundRobin}\n"
							}
							input = strings.Replace(input, collection+":", policy+collection+":", 1)
						}
						resources := &resource.Resources{}
						mustUnmarshal(t, []byte(input), resources)
						if scenario == "headless" {
							resources.Services[0].Spec.ClusterIP = "None"
						}
						if scenario == "reverse" {
							slices.Reverse(resources.Gateways[0].Spec.Listeners)
						}
						base, err := os.ReadFile("testdata/base/base.yaml")
						require.NoError(t, err)
						baseResources := &resource.Resources{}
						mustUnmarshal(t, base, baseResources)
						resources.Secrets = append(resources.Secrets, baseResources.Secrets...)
						translator := &Translator{
							GatewayControllerName: egv1a1.GatewayControllerName,
							GatewayClassName:      "envoy-gateway-class",
							ControllerNamespace:   "envoy-gateway-system",
							Logger:                logging.DefaultLogger(io.Discard, egv1a1.LogLevelInfo),
						}
						if merge {
							translator.MergeBackends = &MergeBackendsConfig{}
						}
						got, err := translator.Translate(t.Context(), resources)
						require.NoError(t, err)
						require.Len(t, got.Gateways, 1)
						require.Len(t, got.Gateways[0].Status.Listeners, 2)
						listeners := got.Gateways[0].Status.Listeners
						if scenario == "listener-set" {
							require.Len(t, got.ListenerSets, 1)
							require.Len(t, got.ListenerSets[0].Status.Listeners, 2)
							for _, l := range listeners {
								require.Zero(t, l.AttachedRoutes)
							}
							listeners = nil
							for _, l := range got.ListenerSets[0].Status.Listeners {
								require.EqualValues(t, 1, l.AttachedRoutes)
							}
						}
						for _, listener := range listeners {
							require.EqualValues(t, 1, listener.AttachedRoutes, "attachment for %s", listener.Name)
						}
						for _, policy := range got.BackendTrafficPolicies {
							require.NotEmpty(t, policy.Status.Ancestors)
							accepted := false
							for _, ancestor := range policy.Status.Ancestors {
								for _, c := range ancestor.Conditions {
									if c.Type == "Accepted" && c.Status == "True" {
										accepted = true
									}
								}
							}
							require.True(t, accepted, "policy %s must be accepted", policy.Name)
						}
						var parentStatuses []gwapiv1.RouteParentStatus
						switch kind {
						case "HTTP":
							parentStatuses = got.HTTPRoutes[0].Status.Parents
						case "GRPC":
							parentStatuses = got.GRPCRoutes[0].Status.Parents
						case "TCP":
							parentStatuses = got.TCPRoutes[0].Status.Parents
						case "TLS":
							parentStatuses = got.TLSRoutes[0].Status.Parents
						case "UDP":
							parentStatuses = got.UDPRoutes[0].Status.Parents
						}
						parentCount := 1
						if explicit {
							parentCount = 2
						}
						require.Len(t, parentStatuses, parentCount)
						for _, parent := range parentStatuses {
							seen := map[string]bool{}
							for _, condition := range parent.Conditions {
								require.False(t, seen[condition.Type], "duplicate parent condition")
								seen[condition.Type] = true
								if condition.Type == "Accepted" || condition.Type == "ResolvedRefs" {
									require.Equal(t, "True", string(condition.Status))
								}
							}
						}
						x := got.XdsIR["default/gateway-1"]
						require.NotNil(t, x)
						destinations := map[string]*ir.RouteDestination{}
						for _, listener := range x.HTTP {
							if scenario == "listener-set" && !strings.HasPrefix(listener.Name, listenerPrefix) {
								continue
							}
							require.Len(t, listener.Routes, 1)
							destinations[listener.Name] = listener.Routes[0].Destination
						}
						for _, listener := range x.TCP {
							if scenario == "listener-set" && !strings.HasPrefix(listener.Name, listenerPrefix) {
								continue
							}
							require.Len(t, listener.Routes, 1)
							destinations[listener.Name] = listener.Routes[0].Destination
						}
						for _, listener := range x.UDP {
							if scenario == "listener-set" && !strings.HasPrefix(listener.Name, listenerPrefix) {
								continue
							}
							require.NotNil(t, listener.Route)
							destinations[listener.Name] = listener.Route.Destination
						}
						require.Len(t, destinations, 2)
						xt := &xdstranslator.Translator{Logger: translator.Logger}
						table, err := xt.Translate(t.Context(), x)
						require.NoError(t, err)
						hostsByCluster := map[string][]string{}
						clusterNames := map[string]bool{}
						for _, resource := range table.XdsResources[resourcev3.ClusterType] {
							cluster := resource.(*clusterv3.Cluster)
							require.False(t, clusterNames[cluster.Name], "duplicate cluster %s", cluster.Name)
							clusterNames[cluster.Name] = true
						}
						for _, resource := range table.XdsResources[resourcev3.EndpointType] {
							endpoints := resource.(*endpointv3.ClusterLoadAssignment)
							for _, locality := range endpoints.Endpoints {
								for _, ep := range locality.LbEndpoints {
									hostsByCluster[endpoints.ClusterName] = append(hostsByCluster[endpoints.ClusterName], ep.GetEndpoint().GetAddress().GetSocketAddress().Address)
								}
							}
						}
						for _, listener := range []string{"listener-1", "listener-2"} {
							d := destinations[listenerPrefix+listener]
							require.NotNil(t, d)
							if kind == "TLS" || kind == "TCP" || kind == "UDP" {
								require.True(t, len(d.Settings) == 0 || len(d.BackendClusterRefs) == 0,
									"a single-cluster route must not mix inline and merged backends")
							}
							settings := append([]*ir.DestinationSetting{}, d.Settings...)
							for _, ref := range d.BackendClusterRefs {
								found := false
								for _, cluster := range x.BackendClusters {
									if cluster.Name == ref.Name {
										settings = append(settings, cluster.Setting)
										found = true
									}
								}
								require.True(t, found, "shared cluster reference %s", ref.Name)
							}
							expectedSettings := 1
							if scenario == "tls-pool" || scenario == "weighted" {
								expectedSettings = 2
							}
							require.Len(t, settings, expectedSettings)
							if scenario == "weighted" {
								var weights []uint32
								for _, setting := range d.Settings {
									weights = append(weights, *setting.Weight)
								}
								for _, ref := range d.BackendClusterRefs {
									weights = append(weights, *ref.Weight)
								}
								require.ElementsMatch(t, []uint32{2, 3}, weights)
							}
							require.Len(t, settings[0].Endpoints, 1)
							want := "10.0.0.7"
							if listener == "listener-1" && scenario != "headless" && scenario != "route" && scenario != "rule" && scenario != "replace" {
								want = "10.96.0.7"
							}
							expectedHosts := []string{want}
							if scenario == "tls-pool" {
								expectedHosts = append(expectedHosts, "10.0.0.7")
							}
							if scenario == "weighted" {
								expectedHosts = append(expectedHosts, want)
							}
							var clusterHosts []string
							if len(d.Settings) > 0 {
								if scenario == "weighted" {
									for _, setting := range d.Settings {
										require.True(t, clusterNames[setting.Name], "missing weighted xDS cluster %s", setting.Name)
										clusterHosts = append(clusterHosts, hostsByCluster[setting.Name]...)
									}
								} else {
									require.True(t, clusterNames[d.Name], "missing inline xDS cluster %s", d.Name)
									clusterHosts = append(clusterHosts, hostsByCluster[d.Name]...)
								}
							}
							for _, ref := range d.BackendClusterRefs {
								require.True(t, clusterNames[ref.Name])
								clusterHosts = append(clusterHosts, hostsByCluster[ref.Name]...)
							}
							require.ElementsMatch(t, expectedHosts, clusterHosts, "xDS endpoints for %s", listener)
							if scenario == "mirror" {
								for _, l := range x.HTTP {
									if l.Name != listenerPrefix+listener {
										continue
									}
									require.Len(t, l.Routes[0].Mirrors, 1)
									mirror := l.Routes[0].Mirrors[0].Destination
									require.Equal(t, want, mirror.Settings[0].Endpoints[0].Host)
									require.Equal(t, []string{want}, hostsByCluster[mirror.Name])
								}
							}
							actual := settings[0].Endpoints[0].Host
							t.Logf("%s host=%s expected=%s mergedRefs=%d", listener, actual, want, len(d.BackendClusterRefs))
							if actual != want {
								t.Errorf("listener RoutingType ignored: %s got %s, want %s", listener, actual, want)
							}
						}
					})
				}
			}
		}
	}
}
