// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package gatewayapi

import (
	"fmt"
	"io"
	"testing"

	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/gatewayapi/resource"
	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/logging"
)

type backendTrafficPolicyBenchmarkCase struct {
	routes, policies, listeners, matches int
	gatewayPolicy, merge, selector       bool
}

func BenchmarkBackendTrafficPolicyRoutes(b *testing.B) {
	for _, tc := range []struct {
		name string
		backendTrafficPolicyBenchmarkCase
	}{
		{"small", backendTrafficPolicyBenchmarkCase{routes: 1, policies: 1, listeners: 1, matches: 1}},
		{"medium", backendTrafficPolicyBenchmarkCase{routes: 100, policies: 100, listeners: 1, matches: 1}},
		{"dense", backendTrafficPolicyBenchmarkCase{routes: 1000, policies: 1000, listeners: 2, matches: 4}},
		{"merged", backendTrafficPolicyBenchmarkCase{routes: 1000, policies: 1000, listeners: 2, matches: 4, gatewayPolicy: true, merge: true}},
		{"selector", backendTrafficPolicyBenchmarkCase{routes: 1000, policies: 1, listeners: 2, matches: 4, selector: true}},
		{"one-policy", backendTrafficPolicyBenchmarkCase{routes: 1000, policies: 1, listeners: 2, matches: 4}},
		{"two-policies", backendTrafficPolicyBenchmarkCase{routes: 1000, policies: 2, listeners: 2, matches: 4}},
		{"ten-policies", backendTrafficPolicyBenchmarkCase{routes: 1000, policies: 10, listeners: 2, matches: 4}},
		{"thirty-two-policies", backendTrafficPolicyBenchmarkCase{routes: 1000, policies: 32, listeners: 2, matches: 4}},
		{"sixty-four-policies", backendTrafficPolicyBenchmarkCase{routes: 1000, policies: 64, listeners: 2, matches: 4}},
		{"127-policies", backendTrafficPolicyBenchmarkCase{routes: 1000, policies: 127, listeners: 2, matches: 4}},
		{"128-policies", backendTrafficPolicyBenchmarkCase{routes: 1000, policies: 128, listeners: 2, matches: 4}},
		{"129-policies", backendTrafficPolicyBenchmarkCase{routes: 1000, policies: 129, listeners: 2, matches: 4}},
		{"single-match-32", backendTrafficPolicyBenchmarkCase{routes: 1000, policies: 32, listeners: 1, matches: 1}},
		{"single-match-64", backendTrafficPolicyBenchmarkCase{routes: 1000, policies: 64, listeners: 1, matches: 1}},
		{"single-match-127", backendTrafficPolicyBenchmarkCase{routes: 1000, policies: 127, listeners: 1, matches: 1}},
		{"single-match-128", backendTrafficPolicyBenchmarkCase{routes: 1000, policies: 128, listeners: 1, matches: 1}},
		{"single-match-129", backendTrafficPolicyBenchmarkCase{routes: 1000, policies: 129, listeners: 1, matches: 1}},
		{"no-policies", backendTrafficPolicyBenchmarkCase{routes: 1000, listeners: 2, matches: 4}},
		{"gateway-only", backendTrafficPolicyBenchmarkCase{routes: 1000, listeners: 2, matches: 4, gatewayPolicy: true}},
	} {
		b.Run(tc.name, func(b *testing.B) {
			fixture := newBackendTrafficPolicyBenchmark(tc.backendTrafficPolicyBenchmarkCase)
			require.Len(b, fixture.run(), len(fixture.resources.BackendTrafficPolicies))
			var configured int
			for _, x := range fixture.xdsIR {
				for _, listener := range x.HTTP {
					for _, route := range listener.Routes {
						if route.Traffic != nil && route.Traffic.LoadBalancer != nil {
							configured++
						}
					}
				}
			}
			targets := tc.policies
			if tc.gatewayPolicy || tc.selector {
				targets = tc.routes
			}
			require.Equal(b, targets*tc.listeners*tc.matches, configured)

			b.ReportAllocs()
			for b.Loop() {
				fixture.run()
			}
		})
	}
}

type backendTrafficPolicyBenchmark struct {
	translator *Translator
	resources  *resource.Resources
	gateways   []*GatewayContext
	routes     []RouteContext
	xdsIR      resource.XdsIRMap
}

func newBackendTrafficPolicyBenchmark(tc backendTrafficPolicyBenchmarkCase) *backendTrafficPolicyBenchmark {
	f := &backendTrafficPolicyBenchmark{
		translator: &Translator{
			TranslatorContext:     &TranslatorContext{},
			GatewayControllerName: egv1a1.GatewayControllerName,
			Logger:                logging.DefaultLogger(io.Discard, egv1a1.LogLevelError),
		},
		resources: resource.NewResources(),
		xdsIR:     make(resource.XdsIRMap),
	}
	gateway := &GatewayContext{Gateway: &gwapiv1.Gateway{
		ObjectMeta: metav1.ObjectMeta{Namespace: "default", Name: "gateway"},
	}}
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
		if tc.selector {
			route.SetGroupVersionKind(gwapiv1.SchemeGroupVersion.WithKind(resource.KindHTTPRoute))
		}
		GetRouteParentContext(route, parent, egv1a1.GatewayControllerName).SetListeners(gateway.listeners...)
		f.routes = append(f.routes, route)
		for _, listener := range x.HTTP {
			for match := range tc.matches {
				listener.Routes = append(listener.Routes, &ir.HTTPRoute{
					Name: irRouteName(route, 0, match), Metadata: buildResourceMetadata(route, route.Spec.Rules[0].Name),
				})
			}
		}
		if i < tc.policies {
			policy := backendTrafficPolicyBenchmarkPolicy(route.GetNamespace(), route.GetName(), resource.KindHTTPRoute)
			if tc.merge {
				policy.Spec.MergeType = new(egv1a1.StrategicMerge)
			}
			f.resources.BackendTrafficPolicies = append(f.resources.BackendTrafficPolicies, policy)
		}
	}
	if tc.selector {
		policy := f.resources.BackendTrafficPolicies[0]
		policy.Spec.TargetRefs = nil
		policy.Spec.TargetSelectors = []egv1a1.TargetSelector{{Kind: resource.KindHTTPRoute}}
	}
	if tc.gatewayPolicy {
		f.resources.BackendTrafficPolicies = append(f.resources.BackendTrafficPolicies,
			backendTrafficPolicyBenchmarkPolicy(gateway.Namespace, gateway.Name, resource.KindGateway))
	}
	return f
}

func backendTrafficPolicyBenchmarkPolicy(namespace, name, kind string) *egv1a1.BackendTrafficPolicy {
	policy := &egv1a1.BackendTrafficPolicy{ObjectMeta: metav1.ObjectMeta{Namespace: namespace, Name: name}}
	policy.Spec.TargetRefs = []gwapiv1.LocalPolicyTargetReferenceWithSectionName{{
		LocalPolicyTargetReference: gwapiv1.LocalPolicyTargetReference{
			Group: gwapiv1.GroupName, Kind: gwapiv1.Kind(kind), Name: gwapiv1.ObjectName(name),
		},
	}}
	policy.Spec.LoadBalancer = &egv1a1.LoadBalancer{Type: egv1a1.RoundRobinLoadBalancerType}
	return policy
}

func (f *backendTrafficPolicyBenchmark) run() []*egv1a1.BackendTrafficPolicy {
	// Reset the policy pass's mutations so every iteration does the same work.
	// This reset and the full pass, including index construction, are timed.
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
