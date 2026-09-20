// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0

package gatewayapi

import (
	"fmt"
	"io"
	"os"
	"strings"
	"testing"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/gatewayapi/resource"
	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/logging"
	"github.com/stretchr/testify/require"
)

// Research-only: assert intended listener routing, without modifying production code.
func TestResearch9696(t *testing.T) {
	for _, kind := range []string{"HTTP", "GRPC", "TCP", "TLS", "UDP"} {
		for _, merge := range []bool{false, true} {
			for _, explicit := range []bool{false, true} {
				mode := "omitted"
				if explicit {
					mode = "explicit"
				}
				t.Run(fmt.Sprintf("%s/merge=%t/%s", kind, merge, mode), func(t *testing.T) {
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
					resources := &resource.Resources{}
					mustUnmarshal(t, []byte(input), resources)
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
					for _, listener := range got.Gateways[0].Status.Listeners {
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
					x := got.XdsIR["default/gateway-1"]
					require.NotNil(t, x)
					destinations := map[string]*ir.RouteDestination{}
					for _, listener := range x.HTTP {
						require.Len(t, listener.Routes, 1)
						destinations[listener.Name] = listener.Routes[0].Destination
					}
					for _, listener := range x.TCP {
						require.Len(t, listener.Routes, 1)
						destinations[listener.Name] = listener.Routes[0].Destination
					}
					for _, listener := range x.UDP {
						require.NotNil(t, listener.Route)
						destinations[listener.Name] = listener.Route.Destination
					}
					require.Len(t, destinations, 2)
					for _, listener := range []string{"listener-1", "listener-2"} {
						d := destinations["default/gateway-1/"+listener]
						require.NotNil(t, d)
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
						require.Len(t, settings, 1)
						require.Len(t, settings[0].Endpoints, 1)
						want := "10.0.0.7"
						if listener == "listener-1" {
							want = "10.96.0.7"
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
