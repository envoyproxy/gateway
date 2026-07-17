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
	parent := func(defaultMT *egv1a1.MergeType) *egv1a1.BackendTrafficPolicy {
		return &egv1a1.BackendTrafficPolicy{
			ObjectMeta: metav1.ObjectMeta{Namespace: "app", Name: "parent"},
			Spec:       egv1a1.BackendTrafficPolicySpec{DefaultChildMergeType: defaultMT},
		}
	}

	tests := []struct {
		name   string
		pol    *egv1a1.BackendTrafficPolicy
		parent *egv1a1.BackendTrafficPolicy
		want   *egv1a1.MergeType
	}{
		{"explicit value wins over parent default", child(&jsonMerge), parent(&strategic), &jsonMerge},
		{"explicit replace opts out of parent default", child(&replace), parent(&strategic), nil},
		{"parent default applied when unset", child(nil), parent(&strategic), &strategic},
		{"no parent policy stays nil", child(nil), nil, nil},
		{"parent without default stays nil", child(nil), parent(nil), nil},
		{"invalid parent default is ignored", child(nil), parent(&replace), nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := effectiveMergeType(tt.pol, tt.parent)
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
