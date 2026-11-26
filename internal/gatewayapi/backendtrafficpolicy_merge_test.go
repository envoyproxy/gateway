// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package gatewayapi

import (
	"fmt"
	"io"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/gatewayapi/resource"
	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/logging"
)

func TestBackendTrafficPolicyMergeAttachments(t *testing.T) {
	for _, mergeType := range []egv1a1.MergeType{egv1a1.StrategicMerge, egv1a1.JSONMerge} {
		for _, attachments := range []struct{ targets, listeners int }{{1, 1}, {2, 1}, {1, 2}, {2, 2}} {
			for _, distinctParents := range []bool{false, true} {
				t.Run(fmt.Sprintf("%s/targets=%d/listeners=%d/distinct-parents=%t", mergeType, attachments.targets, attachments.listeners, distinctParents), func(t *testing.T) {
					f := newPolicyMergeFixture(policyMergeCase{
						routes: 2 * attachments.targets, policies: 2, targets: attachments.targets, listeners: attachments.listeners, mergeType: mergeType, distinctParents: distinctParents,
					})
					for pass := range 2 {
						// Change the same policy objects between passes to detect stale results.
						for i, policy := range f.resources.BackendTrafficPolicies {
							if i < 2 {
								policy.Spec.Timeout.HTTP.RequestTimeout = new(gwapiv1.Duration(fmt.Sprintf("%ds", 5+i+pass)))
							} else {
								policy.Spec.Timeout.HTTP.ConnectionIdleTimeout = new(gwapiv1.Duration(fmt.Sprintf("%ds", 10+i-2+pass)))
							}
						}
						before := f.resources.DeepCopy()
						require.Len(t, f.run(), len(f.resources.BackendTrafficPolicies))
						expected := make(map[string]time.Duration)
						for i, route := range f.routes {
							expected[route.GetName()] = time.Duration(5+i/attachments.targets+pass) * time.Second
						}
						for _, x := range f.xdsIR {
							for listenerIndex, listener := range x.HTTP {
								parentIndex := 0
								if distinctParents {
									parentIndex = listenerIndex
								}
								for _, route := range listener.Routes {
									require.NotNil(t, route.Traffic)
									require.NotNil(t, route.Traffic.Timeout)
									require.NotNil(t, route.Traffic.Timeout.HTTP)
									require.Equal(t, ir.MetaV1DurationPtr(expected[route.Metadata.Name]), route.Traffic.Timeout.HTTP.RequestTimeout)
									require.Equal(t, ir.MetaV1DurationPtr(time.Duration(10+parentIndex+pass)*time.Second), route.Traffic.Timeout.HTTP.ConnectionIdleTimeout)
								}
							}
							// Each attachment owns its translated settings, even when its merge is reused.
							if len(x.HTTP) > 1 {
								x.HTTP[0].Routes[0].Traffic.Timeout.HTTP.RequestTimeout.Duration = time.Hour
								require.NotEqual(t, time.Hour, x.HTTP[1].Routes[0].Traffic.Timeout.HTTP.RequestTimeout.Duration)
							}
						}
						for i, policy := range f.resources.BackendTrafficPolicies {
							require.Equal(t, before.BackendTrafficPolicies[i].Spec, policy.Spec)
						}
					}
				})
			}
		}
	}
}

func BenchmarkBackendTrafficPolicyMerge(b *testing.B) {
	for _, tc := range []struct {
		name string
		policyMergeCase
	}{
		{"small-unique", policyMergeCase{routes: 1, policies: 1, targets: 1, listeners: 1, mergeType: egv1a1.StrategicMerge}},
		{"unique-pairs", policyMergeCase{routes: 1000, policies: 1000, targets: 1, listeners: 1, mergeType: egv1a1.StrategicMerge}},
		{"two-targets", policyMergeCase{routes: 1000, policies: 500, targets: 2, listeners: 1, mergeType: egv1a1.StrategicMerge}},
		{"two-listeners", policyMergeCase{routes: 1000, policies: 1000, targets: 1, listeners: 2, mergeType: egv1a1.StrategicMerge}},
		{"two-targets-two-listeners", policyMergeCase{routes: 1000, policies: 500, targets: 2, listeners: 2, mergeType: egv1a1.StrategicMerge}},
		{"alternating-parents", policyMergeCase{routes: 1000, policies: 500, targets: 2, listeners: 2, mergeType: egv1a1.StrategicMerge, distinctParents: true}},
		{"two-policies", policyMergeCase{routes: 1000, policies: 2, targets: 1, listeners: 2, mergeType: egv1a1.StrategicMerge}},
		{"ten-policies", policyMergeCase{routes: 1000, policies: 10, targets: 1, listeners: 2, mergeType: egv1a1.StrategicMerge}},
		{"unmerged", policyMergeCase{routes: 1000, policies: 1000, targets: 1, listeners: 2}},
	} {
		b.Run(tc.name, func(b *testing.B) {
			f := newPolicyMergeFixture(tc.policyMergeCase)
			require.Len(b, f.run(), len(f.resources.BackendTrafficPolicies))
			for _, x := range f.xdsIR {
				for _, listener := range x.HTTP {
					require.Len(b, listener.Routes, tc.routes*4)
					for _, route := range listener.Routes {
						require.NotNil(b, route.Traffic.Timeout)
					}
				}
			}
			b.ReportAllocs()
			for b.Loop() {
				f.run()
			}
		})
	}
}

func BenchmarkBackendTrafficPolicyMergeSparse(b *testing.B) {
	for _, policies := range []int{2, 10} {
		b.Run(fmt.Sprintf("policies=%d", policies), func(b *testing.B) {
			f := newPolicyMergeFixture(policyMergeCase{routes: 1000, policies: policies, targets: 1, listeners: 2})
			f.resources.BackendTrafficPolicies = f.resources.BackendTrafficPolicies[:policies]
			require.Len(b, f.run(), policies)
			configured := 0
			for _, x := range f.xdsIR {
				for _, listener := range x.HTTP {
					for _, route := range listener.Routes {
						if route.Traffic != nil && route.Traffic.Timeout != nil {
							configured++
						}
					}
				}
			}
			require.Equal(b, policies*2*4, configured)
			b.ReportAllocs()
			for b.Loop() {
				f.run()
			}
		})
	}
}

type policyMergeCase struct {
	routes, policies, targets, listeners int
	mergeType                            egv1a1.MergeType
	distinctParents                      bool
}

type policyMergeFixture struct {
	translator *Translator
	resources  *resource.Resources
	gateways   []*GatewayContext
	routes     []RouteContext
	xdsIR      resource.XdsIRMap
}

func newPolicyMergeFixture(tc policyMergeCase) *policyMergeFixture {
	f := &policyMergeFixture{
		translator: &Translator{
			TranslatorContext: &TranslatorContext{}, GatewayControllerName: egv1a1.GatewayControllerName,
			Logger: logging.DefaultLogger(io.Discard, egv1a1.LogLevelError),
		},
		resources: resource.NewResources(), xdsIR: make(resource.XdsIRMap),
	}
	gateway := &GatewayContext{Gateway: &gwapiv1.Gateway{ObjectMeta: metav1.ObjectMeta{Namespace: "default", Name: "gateway"}}}
	for i := range tc.listeners {
		gateway.Spec.Listeners = append(gateway.Spec.Listeners, gwapiv1.Listener{
			Name: gwapiv1.SectionName(fmt.Sprintf("http-%d", i)), Protocol: gwapiv1.HTTPProtocolType, Port: gwapiv1.PortNumber(8080 + i),
		})
	}
	gateway.ResetListeners()
	f.gateways = []*GatewayContext{gateway}
	x := &ir.Xds{}
	f.xdsIR[f.translator.getIRKey(gateway.Gateway)] = x
	for _, listener := range gateway.listeners {
		x.HTTP = append(x.HTTP, &ir.HTTPListener{CoreListenerDetails: ir.CoreListenerDetails{Name: irListenerName(listener)}})
	}
	parent := gwapiv1.ParentReference{Name: gwapiv1.ObjectName(gateway.Name)}
	for i := range tc.routes {
		route := &HTTPRouteContext{HTTPRoute: &gwapiv1.HTTPRoute{
			ObjectMeta: metav1.ObjectMeta{Namespace: "default", Name: fmt.Sprintf("route-%04d", i)},
			Spec: gwapiv1.HTTPRouteSpec{
				CommonRouteSpec: gwapiv1.CommonRouteSpec{ParentRefs: []gwapiv1.ParentReference{parent}},
				Rules:           []gwapiv1.HTTPRouteRule{{Name: new(gwapiv1.SectionName("rule"))}},
			},
		}}
		GetRouteParentContext(route, parent, egv1a1.GatewayControllerName).SetListeners(gateway.listeners...)
		f.routes = append(f.routes, route)
		for _, listener := range x.HTTP {
			for match := range 4 {
				listener.Routes = append(listener.Routes, &ir.HTTPRoute{
					Name: irRouteName(route, 0, match), Metadata: buildResourceMetadata(route, route.Spec.Rules[0].Name),
				})
			}
		}
	}
	for i := range tc.policies {
		policy := &egv1a1.BackendTrafficPolicy{ObjectMeta: metav1.ObjectMeta{Namespace: "default", Name: fmt.Sprintf("policy-%04d", i)}}
		for j := range tc.targets {
			policy.Spec.TargetRefs = append(policy.Spec.TargetRefs, policyMergeTarget(resource.KindHTTPRoute, f.routes[i*tc.targets+j].GetName()))
		}
		if tc.mergeType != "" {
			policy.Spec.MergeType = new(tc.mergeType)
		}
		policy.Spec.Timeout = &egv1a1.Timeout{HTTP: &egv1a1.HTTPTimeout{RequestTimeout: new(gwapiv1.Duration("5s"))}}
		f.resources.BackendTrafficPolicies = append(f.resources.BackendTrafficPolicies, policy)
	}
	parents := 1
	if tc.distinctParents {
		parents = tc.listeners
	}
	for i := range parents {
		policy := &egv1a1.BackendTrafficPolicy{ObjectMeta: metav1.ObjectMeta{Namespace: "default", Name: fmt.Sprintf("parent-%d", i)}}
		policy.Spec.TargetRefs = []gwapiv1.LocalPolicyTargetReferenceWithSectionName{policyMergeTarget(resource.KindGateway, gateway.Name)}
		if tc.distinctParents {
			policy.Spec.TargetRefs[0].SectionName = new(gateway.Spec.Listeners[i].Name)
		}
		policy.Spec.Timeout = &egv1a1.Timeout{HTTP: &egv1a1.HTTPTimeout{
			RequestTimeout: new(gwapiv1.Duration("10s")), ConnectionIdleTimeout: new(gwapiv1.Duration("10s")),
		}}
		f.resources.BackendTrafficPolicies = append(f.resources.BackendTrafficPolicies, policy)
	}
	return f
}

func policyMergeTarget(kind, name string) gwapiv1.LocalPolicyTargetReferenceWithSectionName {
	return gwapiv1.LocalPolicyTargetReferenceWithSectionName{LocalPolicyTargetReference: gwapiv1.LocalPolicyTargetReference{
		Group: gwapiv1.GroupName, Kind: gwapiv1.Kind(kind), Name: gwapiv1.ObjectName(name),
	}}
}

func (f *policyMergeFixture) run() []*egv1a1.BackendTrafficPolicy {
	// Include resetting pass-local mutations in every timed iteration.
	for _, x := range f.xdsIR {
		for _, listener := range x.HTTP {
			for _, route := range listener.Routes {
				route.Traffic = nil
				route.Metadata.Policies = nil
			}
		}
	}
	for _, policy := range f.resources.BackendTrafficPolicies {
		policy.Status = gwapiv1.PolicyStatus{}
	}
	return f.translator.ProcessBackendTrafficPolicies(f.resources, f.gateways, f.routes, f.xdsIR)
}
