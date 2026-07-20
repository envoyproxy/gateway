// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package gatewayapi

import (
	"testing"

	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/ir"
)

func TestEffectiveMergeType(t *testing.T) {
	strategic := egv1a1.StrategicMerge
	jsonMerge := egv1a1.JSONMerge
	replace := egv1a1.Replace

	child := func(mt *egv1a1.MergeType) *egv1a1.BackendTrafficPolicy {
		return &egv1a1.BackendTrafficPolicy{
			ObjectMeta: metav1.ObjectMeta{Namespace: "app", Name: "child"},
			Spec:       egv1a1.BackendTrafficPolicySpec{MergeType: mt},
		}
	}

	tests := []struct {
		name      string
		pol       *egv1a1.BackendTrafficPolicy
		defaultMT *egv1a1.MergeType
		want      *egv1a1.MergeType
	}{
		{"explicit value wins over default", child(&jsonMerge), &strategic, &jsonMerge},
		{"explicit replace opts out of default", child(&replace), &strategic, nil},
		{"default applied when unset", child(nil), &strategic, &strategic},
		{"no default stays nil", child(nil), nil, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := effectiveMergeType(tt.pol, tt.defaultMT)
			if tt.want == nil {
				assert.Nil(t, got)
				return
			}
			assert.NotNil(t, got)
			assert.Equal(t, *tt.want, *got)
		})
	}
}

// TestResolveDefaultChildMergeType covers resolveDefaultChildMergeType, which returns the
// defaultChildMergeType declared by the nearest ancestor policy (closest first). This decouples
// the default-merge intent (which, after CEL restricts defaultChildMergeType to Gateway targets,
// is only declared on the Gateway-level policy) from the merge target (the closest parent config).
func TestResolveDefaultChildMergeType(t *testing.T) {
	strategic := egv1a1.StrategicMerge
	jsonMerge := egv1a1.JSONMerge
	replace := egv1a1.Replace

	withDefault := func(mt *egv1a1.MergeType) *egv1a1.BackendTrafficPolicy {
		return &egv1a1.BackendTrafficPolicy{Spec: egv1a1.BackendTrafficPolicySpec{DefaultChildMergeType: mt}}
	}

	tests := []struct {
		name      string
		ancestors []*egv1a1.BackendTrafficPolicy
		want      *egv1a1.MergeType
	}{
		{"no ancestors", nil, nil},
		{"nil ancestor", []*egv1a1.BackendTrafficPolicy{nil}, nil},
		{"single ancestor with default", []*egv1a1.BackendTrafficPolicy{withDefault(&strategic)}, &strategic},
		{"single ancestor without default", []*egv1a1.BackendTrafficPolicy{withDefault(nil)}, nil},
		// Reading B: the closest parent (listener) declares no default, so resolution falls through
		// to the gateway-level policy that does.
		{"closest without default falls through", []*egv1a1.BackendTrafficPolicy{withDefault(nil), withDefault(&strategic)}, &strategic},
		{"closest declarer wins", []*egv1a1.BackendTrafficPolicy{withDefault(&jsonMerge), withDefault(&strategic)}, &jsonMerge},
		{"invalid nearest default is ignored", []*egv1a1.BackendTrafficPolicy{withDefault(&replace), withDefault(&strategic)}, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := resolveDefaultChildMergeType(tt.ancestors...)
			if tt.want == nil {
				assert.Nil(t, got)
				return
			}
			assert.NotNil(t, got)
			assert.Equal(t, *tt.want, *got)
		})
	}
}

// TestAnyParentPolicyMergeDefault covers anyParentPolicyMergeDefault, which reports whether any
// gateway- or listener-level parent policy of the route's parent gateways sets
// defaultChildMergeType.
func TestAnyParentPolicyMergeDefault(t *testing.T) {
	strategic := egv1a1.StrategicMerge
	gwNN := types.NamespacedName{Namespace: "envoy-gateway", Name: "gw"}

	parentCtx := func(listenerName string) *RouteParentContext {
		return &RouteParentContext{
			listeners: []*ListenerContext{{
				Listener: &gwapiv1.Listener{Name: gwapiv1.SectionName(listenerName)},
				gateway: &GatewayContext{Gateway: &gwapiv1.Gateway{
					ObjectMeta: metav1.ObjectMeta{Namespace: gwNN.Namespace, Name: gwNN.Name},
				}},
			}},
		}
	}
	parentPolicy := func(defaultMT *egv1a1.MergeType) *egv1a1.BackendTrafficPolicy {
		return &egv1a1.BackendTrafficPolicy{
			Spec: egv1a1.BackendTrafficPolicySpec{DefaultChildMergeType: defaultMT},
		}
	}

	tests := []struct {
		name    string
		parents []*RouteParentContext
		policy  map[NamespacedNameWithSection]*egv1a1.BackendTrafficPolicy
		want    bool
	}{
		{"no parents", nil, nil, false},
		{"no parent policies", []*RouteParentContext{parentCtx("http")}, nil, false},
		{
			"gateway-level policy without default",
			[]*RouteParentContext{parentCtx("http")},
			map[NamespacedNameWithSection]*egv1a1.BackendTrafficPolicy{
				{NamespacedName: gwNN}: parentPolicy(nil),
			},
			false,
		},
		{
			"gateway-level policy with default",
			[]*RouteParentContext{parentCtx("http")},
			map[NamespacedNameWithSection]*egv1a1.BackendTrafficPolicy{
				{NamespacedName: gwNN}: parentPolicy(&strategic),
			},
			true,
		},
		{
			"listener-level policy with default",
			[]*RouteParentContext{parentCtx("http")},
			map[NamespacedNameWithSection]*egv1a1.BackendTrafficPolicy{
				{NamespacedName: gwNN, SectionName: "http"}: parentPolicy(&strategic),
			},
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, anyParentPolicyMergeDefault(tt.parents, tt.policy))
		})
	}
}

// TestApplyTrafficFeatureToRoute_MergeGatewayScoping covers the MergeGateways scoping in
// applyTrafficFeatureToRoute for the TCP and UDP listener loops: a listener belonging to a
// different Gateway must be skipped, so a merged policy does not bleed across Gateways that
// share one IR. The HTTP path is covered by the merged-gateways golden fixture.
func TestApplyTrafficFeatureToRoute_MergeGatewayScoping(t *testing.T) {
	tr := &Translator{}
	gwNN := &types.NamespacedName{Namespace: "envoy-gateway", Name: "gw"}
	policy := &egv1a1.BackendTrafficPolicy{}
	target := policyTargetReferenceWithSectionName{}

	t.Run("tcp listener of another gateway is skipped", func(t *testing.T) {
		route := &TCPRouteContext{TCPRoute: &gwapiv1.TCPRoute{
			ObjectMeta: metav1.ObjectMeta{Namespace: "default", Name: "tcproute-1"},
		}}
		sibling := &ir.TCPRoute{Destination: &ir.RouteDestination{Name: irRoutePrefix(route) + "rule/0"}}
		x := &ir.Xds{TCP: []*ir.TCPListener{{
			CoreListenerDetails: ir.CoreListenerDetails{Name: "envoy-gateway/other-gw/tcp"},
			Routes:              []*ir.TCPRoute{sibling},
		}}}
		tr.applyTrafficFeatureToRoute(route, &ir.TrafficFeatures{CircuitBreaker: &ir.CircuitBreaker{}},
			nil, policy, target, x, gwNN, nil)
		assert.Nil(t, sibling.CircuitBreaker, "route on a sibling Gateway's listener must be skipped")
	})

	t.Run("udp listener of another gateway is skipped", func(t *testing.T) {
		route := &UDPRouteContext{UDPRoute: &gwapiv1.UDPRoute{
			ObjectMeta: metav1.ObjectMeta{Namespace: "default", Name: "udproute-1"},
		}}
		sibling := &ir.UDPRoute{Destination: &ir.RouteDestination{Name: irRoutePrefix(route) + "rule/0"}}
		x := &ir.Xds{UDP: []*ir.UDPListener{{
			CoreListenerDetails: ir.CoreListenerDetails{Name: "envoy-gateway/other-gw/udp"},
			Route:               sibling,
		}}}
		tr.applyTrafficFeatureToRoute(route, &ir.TrafficFeatures{LoadBalancer: &ir.LoadBalancer{}},
			nil, policy, target, x, gwNN, nil)
		assert.Nil(t, sibling.LoadBalancer, "route on a sibling Gateway's listener must be skipped")
	})
}
