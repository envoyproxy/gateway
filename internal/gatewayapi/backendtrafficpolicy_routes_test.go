// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package gatewayapi

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/gatewayapi/resource"
	"github.com/envoyproxy/gateway/internal/ir"
)

func TestBackendTrafficPolicyRouteSelection(t *testing.T) {
	for _, tc := range []struct {
		otherPolicies int
		section       string
	}{{0, ""}, {0, "second"}, {256, ""}, {256, "second"}} {
		t.Run(fmt.Sprintf("other-policies=%d/section=%s", tc.otherPolicies, tc.section), func(t *testing.T) {
			f := newBackendTrafficPolicyBenchmark(backendTrafficPolicyBenchmarkCase{
				routes: tc.otherPolicies, policies: tc.otherPolicies, listeners: 2, matches: 1,
			})
			parent := gwapiv1.ParentReference{Name: "gateway", Namespace: new(gwapiv1.Namespace("default"))}
			for _, kind := range []string{resource.KindHTTPRoute, resource.KindGRPCRoute} {
				for _, identity := range []struct{ namespace, name string }{
					{"default", "route"},
					{"default", "route-extra"},
					{"other", "route"},
				} {
					var route RouteContext
					common := gwapiv1.CommonRouteSpec{ParentRefs: []gwapiv1.ParentReference{parent}}
					// Typed objects may omit TypeMeta, as they do in informer caches.
					if kind == resource.KindHTTPRoute {
						route = &HTTPRouteContext{HTTPRoute: &gwapiv1.HTTPRoute{
							ObjectMeta: metav1.ObjectMeta{Namespace: identity.namespace, Name: identity.name},
							Spec: gwapiv1.HTTPRouteSpec{CommonRouteSpec: common, Rules: []gwapiv1.HTTPRouteRule{
								{Name: new(gwapiv1.SectionName("first"))}, {Name: new(gwapiv1.SectionName("second"))},
							}},
						}}
					} else {
						route = &GRPCRouteContext{GRPCRoute: &gwapiv1.GRPCRoute{
							ObjectMeta: metav1.ObjectMeta{Namespace: identity.namespace, Name: identity.name},
							Spec: gwapiv1.GRPCRouteSpec{CommonRouteSpec: common, Rules: []gwapiv1.GRPCRouteRule{
								{Name: new(gwapiv1.SectionName("first"))}, {Name: new(gwapiv1.SectionName("second"))},
							}},
						}}
					}
					GetRouteParentContext(route, parent, egv1a1.GatewayControllerName).SetListeners(f.gateways[0].listeners...)
					f.routes = append(f.routes, route)
					for _, x := range f.xdsIR {
						for _, listener := range x.HTTP {
							for rule, name := range []gwapiv1.SectionName{"first", "second"} {
								for match := range 2 {
									listener.Routes = append(listener.Routes, &ir.HTTPRoute{
										Name: irRouteName(route, rule, match), Metadata: buildResourceMetadata(route, &name),
									})
								}
							}
						}
					}
				}
				policy := backendTrafficPolicyBenchmarkPolicy("default", "route", kind)
				policy.Name = strings.ToLower(kind)
				if tc.section != "" {
					policy.Spec.TargetRefs[0].SectionName = new(gwapiv1.SectionName(tc.section))
				}
				if kind == resource.KindGRPCRoute {
					policy.Spec.LoadBalancer.Type = egv1a1.RandomLoadBalancerType
				}
				f.resources.BackendTrafficPolicies = append(f.resources.BackendTrafficPolicies, policy)
			}

			for range 2 {
				require.Len(t, f.run(), tc.otherPolicies+2)
				for _, x := range f.xdsIR {
					for _, listener := range x.HTTP {
						for i, route := range listener.Routes {
							if i < tc.otherPolicies {
								continue
							}
							m := route.Metadata
							if m.Namespace == "default" && m.Name == "route" && (tc.section == "" || m.SectionName == tc.section) {
								require.NotNil(t, route.Traffic, route.Name)
								require.NotNil(t, route.Traffic.LoadBalancer, route.Name)
								if m.Kind == resource.KindHTTPRoute {
									require.NotNil(t, route.Traffic.LoadBalancer.RoundRobin, route.Name)
								} else {
									require.NotNil(t, route.Traffic.LoadBalancer.Random, route.Name)
								}
							} else {
								require.Nil(t, route.Traffic, route.Name)
							}
							// The next translation has new IR objects, not the previous pass's pointers.
							listener.Routes[i] = route.DeepCopy()
						}
					}
				}
			}
		})
	}
}
