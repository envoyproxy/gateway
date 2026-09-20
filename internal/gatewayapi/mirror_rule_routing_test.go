// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0

package gatewayapi

import (
	"fmt"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/gatewayapi/resource"
	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/logging"
)

// Mirror routing must use the original rule identity, including non-first GRPC rules.
func TestMirrorRuleRouting(t *testing.T) {
	input, err := os.ReadFile("testdata/backendtrafficpolicy-listener-routing-omitted.in.yaml")
	require.NoError(t, err)
	for _, kind := range []string{"HTTP", "GRPC"} {
		for _, explicit := range []bool{false, true} {
			for _, merge := range []bool{false, true} {
				for _, scenario := range []string{"rule", "second-rule", "listener", "unnamed", "cross-namespace", "without-grant"} {
					t.Run(fmt.Sprintf("%s/explicit=%t/merge=%t/%s", kind, explicit, merge, scenario), func(t *testing.T) {
						text := string(input) + "      filters:\n      - type: RequestMirror\n        requestMirror:\n          backendRef: {name: backend, port: 8080}\n"
						override := scenario != "listener" && scenario != "unnamed"
						if override {
							policy := "- apiVersion: gateway.envoyproxy.io/v1alpha1\n  kind: BackendTrafficPolicy\n  metadata: {name: rule-routing, namespace: default}\n  spec:\n    targetRefs:\n    - {group: gateway.networking.k8s.io, kind: HTTPRoute, name: route-1, sectionName: rule-1}\n    routingType: Endpoint\n"
							text = strings.Replace(text, "httpRoutes:", policy+"httpRoutes:", 1)
						}
						if explicit {
							text = strings.Replace(text, "    - {name: gateway-1}\n    rules:", "    - {name: gateway-1, sectionName: listener-1}\n    - {name: gateway-1, sectionName: listener-2}\n    rules:", 1)
						}
						if kind == "GRPC" {
							text = strings.ReplaceAll(text, "HTTPRoute", "GRPCRoute")
							text = strings.ReplaceAll(text, "httpRoutes:", "grpcRoutes:")
						}
						resources := &resource.Resources{}
						mustUnmarshal(t, []byte(text), resources)
						base, err := os.ReadFile("testdata/base/base.yaml")
						require.NoError(t, err)
						baseResources := &resource.Resources{}
						mustUnmarshal(t, base, baseResources)
						resources.Secrets = baseResources.Secrets
						ruleIdx := 0
						if scenario == "second-rule" {
							ruleIdx = 1
						}
						if kind == "HTTP" {
							r := resources.HTTPRoutes[0]
							if scenario == "unnamed" {
								r.Spec.Rules[0].Name = nil
							}
							if ruleIdx == 1 {
								first := *r.Spec.Rules[0].DeepCopy()
								first.Name = ptr.To(gwapiv1.SectionName("other-rule"))
								first.Filters = nil
								first.Matches = []gwapiv1.HTTPRouteMatch{{Path: &gwapiv1.HTTPPathMatch{Type: ptr.To(gwapiv1.PathMatchExact), Value: ptr.To("/other")}}}
								r.Spec.Rules = append([]gwapiv1.HTTPRouteRule{first}, r.Spec.Rules...)
							}
						} else {
							r := resources.GRPCRoutes[0]
							if scenario == "unnamed" {
								r.Spec.Rules[0].Name = nil
							}
							if ruleIdx == 1 {
								first := *r.Spec.Rules[0].DeepCopy()
								first.Name = ptr.To(gwapiv1.SectionName("other-rule"))
								first.Filters = nil
								first.Matches = []gwapiv1.GRPCRouteMatch{{Method: &gwapiv1.GRPCMethodMatch{Service: ptr.To("other.Service")}}}
								r.Spec.Rules = append([]gwapiv1.GRPCRouteRule{first}, r.Spec.Rules...)
							}
						}
						crossNamespace := scenario == "cross-namespace" || scenario == "without-grant"
						if crossNamespace {
							svc := resources.Services[0].DeepCopy()
							svc.Namespace = "mirror-ns"
							svc.Spec.ClusterIP = "10.96.0.8"
							resources.Services = append(resources.Services, svc)
							eps := resources.EndpointSlices[0].DeepCopy()
							eps.Namespace = "mirror-ns"
							eps.Endpoints[0].Addresses = []string{"10.0.0.8"}
							resources.EndpointSlices = append(resources.EndpointSlices, eps)
							if kind == "HTTP" {
								resources.HTTPRoutes[0].Spec.Rules[0].Filters[0].RequestMirror.BackendRef.Namespace = ptr.To(gwapiv1.Namespace("mirror-ns"))
							} else {
								resources.GRPCRoutes[0].Spec.Rules[0].Filters[0].RequestMirror.BackendRef.Namespace = ptr.To(gwapiv1.Namespace("mirror-ns"))
							}
							if scenario == "cross-namespace" {
								grants := &resource.Resources{}
								mustUnmarshal(t, []byte(fmt.Sprintf("referenceGrants:\n- metadata: {name: mirror, namespace: mirror-ns}\n  spec:\n    from:\n    - {group: gateway.networking.k8s.io, kind: %sRoute, namespace: default}\n    to:\n    - {group: '', kind: Service, name: backend}\n", kind)), grants)
								resources.ReferenceGrants = grants.ReferenceGrants
							}
						}
						translator := &Translator{GatewayControllerName: egv1a1.GatewayControllerName, GatewayClassName: "envoy-gateway-class", ControllerNamespace: "envoy-gateway-system", Logger: logging.DefaultLogger(io.Discard, egv1a1.LogLevelInfo)}
						if merge {
							translator.MergeBackends = &MergeBackendsConfig{}
						}
						got, err := translator.Translate(t.Context(), resources)
						require.NoError(t, err)
						require.Len(t, got.BackendTrafficPolicies, len(resources.BackendTrafficPolicies))
						for _, policy := range got.BackendTrafficPolicies {
							require.NotEmpty(t, policy.Status.Ancestors)
							for _, ancestor := range policy.Status.Ancestors {
								requireMirrorCondition(t, ancestor.Conditions, "Accepted", "True", "Accepted")
							}
						}
						var parents []gwapiv1.RouteParentStatus
						if kind == "HTTP" {
							parents = got.HTTPRoutes[0].Status.Parents
						} else {
							parents = got.GRPCRoutes[0].Status.Parents
						}
						parentCount := 1
						if explicit {
							parentCount = 2
						}
						require.Len(t, parents, parentCount)
						for _, parent := range parents {
							requireMirrorCondition(t, parent.Conditions, "Accepted", "True", "Accepted")
							if scenario == "without-grant" {
								requireMirrorCondition(t, parent.Conditions, "ResolvedRefs", "False", "RefNotPermitted")
							} else {
								requireMirrorCondition(t, parent.Conditions, "ResolvedRefs", "True", "ResolvedRefs")
							}
						}
						x := got.XdsIR["default/gateway-1"]
						require.NotNil(t, x)
						require.Len(t, x.HTTP, 2)
						for _, listener := range x.HTTP {
							if scenario == "without-grant" {
								for _, route := range listener.Routes {
									require.Empty(t, route.Mirrors)
								}
								continue
							}
							var selected *ir.HTTPRoute
							prefix := fmt.Sprintf("%sroute/default/route-1/rule/%d", strings.ToLower(kind), ruleIdx)
							for _, route := range listener.Routes {
								if strings.HasPrefix(route.Name, prefix+"/") {
									require.Nil(t, selected)
									selected = route
								}
							}
							require.NotNil(t, selected)
							mainName, mirrorName, host := prefix, prefix+"-mirror-0", "10.0.0.7"
							if !override && listener.Name == "default/gateway-1/listener-1" {
								mainName += "/listener/" + listener.Name
								mirrorName += "/listener/" + listener.Name
								host = "10.96.0.7"
							}
							require.Equal(t, mainName, selected.Destination.Name)
							if merge && !override && listener.Name == "default/gateway-1/listener-2" {
								require.Len(t, selected.Destination.BackendClusterRefs, 1)
								require.Empty(t, selected.Destination.Settings)
							} else {
								require.Empty(t, selected.Destination.BackendClusterRefs)
							}
							settings := append([]*ir.DestinationSetting{}, selected.Destination.Settings...)
							for _, ref := range selected.Destination.BackendClusterRefs {
								for _, cluster := range x.BackendClusters {
									if cluster.Name == ref.Name {
										settings = append(settings, cluster.Setting)
									}
								}
							}
							require.Len(t, settings, 1)
							require.Len(t, settings[0].Endpoints, 1)
							require.Equal(t, host, settings[0].Endpoints[0].Host)
							require.Len(t, selected.Mirrors, 1)
							mirror := selected.Mirrors[0].Destination
							require.Equal(t, mirrorName, mirror.Name)
							require.Empty(t, mirror.BackendClusterRefs)
							require.Len(t, mirror.Settings, 1)
							require.Equal(t, mirrorName+"/backend/-1", mirror.Settings[0].Name)
							require.Len(t, mirror.Settings[0].Endpoints, 1)
							if crossNamespace {
								host = "10.0.0.8"
							}
							require.Equal(t, host, mirror.Settings[0].Endpoints[0].Host)
							t.Logf("listener=%s main=%s mirror=%s hosts=%s/%s shared=%d", listener.Name, mainName, mirrorName, settings[0].Endpoints[0].Host, host, len(selected.Destination.BackendClusterRefs))
						}
					})
				}
			}
		}
	}
}

func requireMirrorCondition(t *testing.T, conditions []metav1.Condition, kind, status, reason string) {
	t.Helper()
	for _, c := range conditions {
		if c.Type == kind {
			require.Equal(t, status, string(c.Status))
			require.Equal(t, reason, c.Reason)
			return
		}
	}
	t.Fatalf("missing %s condition", kind)
}
