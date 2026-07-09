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

	ep := func(mt *egv1a1.MergeType, label string) *egv1a1.EnvoyProxy {
		d := &egv1a1.BackendTrafficPolicyDefaults{MergeSettings: egv1a1.MergeSettings{MergeType: mt}}
		if label != "" {
			d.MergeExcludeLabel = new(label)
		}
		return &egv1a1.EnvoyProxy{Spec: egv1a1.EnvoyProxySpec{PolicyDefaults: &egv1a1.PolicyDefaults{BackendTrafficPolicy: d}}}
	}
	btp := func(ns string, labels map[string]string, mt *egv1a1.MergeType) *egv1a1.BackendTrafficPolicy {
		return &egv1a1.BackendTrafficPolicy{
			ObjectMeta: metav1.ObjectMeta{Namespace: ns, Labels: labels},
			Spec:       egv1a1.BackendTrafficPolicySpec{MergeType: mt},
		}
	}

	tr := &Translator{ControllerNamespace: "eg"}

	tests := []struct {
		name string
		pol  *egv1a1.BackendTrafficPolicy
		ep   *egv1a1.EnvoyProxy
		want *egv1a1.MergeType
	}{
		{"explicit value wins over default", btp("app", nil, &jsonMerge), ep(&strategic, ""), &jsonMerge},
		{"default applied when unset", btp("app", nil, nil), ep(&strategic, ""), &strategic},
		{"no envoyproxy stays nil", btp("app", nil, nil), nil, nil},
		{"no default in envoyproxy stays nil", btp("app", nil, nil), ep(nil, ""), nil},
		{"control-plane namespace excluded", btp("eg", nil, nil), ep(&strategic, ""), nil},
		{"exclude label opts out", btp("app", map[string]string{"skip": "x"}, nil), ep(&strategic, "skip"), nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tr.effectiveMergeType(tt.pol, tt.ep)
			if tt.want == nil {
				assert.Nil(t, got)
				return
			}
			assert.NotNil(t, got)
			assert.Equal(t, *tt.want, *got)
		})
	}
}

// TestEffectiveMergeType_AdditionalBranches covers the effectiveMergeType branches that
// TestEffectiveMergeType does not: an EnvoyProxy without a BackendTrafficPolicy section, and an
// exclude label that is configured but not present on the policy (so the default still applies).
func TestEffectiveMergeType_AdditionalBranches(t *testing.T) {
	strategic := egv1a1.StrategicMerge

	tr := &Translator{ControllerNamespace: "eg"}

	tests := []struct {
		name string
		pol  *egv1a1.BackendTrafficPolicy
		ep   *egv1a1.EnvoyProxy
		want *egv1a1.MergeType
	}{
		{
			// ep != nil but ep.Spec.PolicyDefaults == nil -> nil.
			name: "envoyproxy without policyDefaults stays nil",
			pol: &egv1a1.BackendTrafficPolicy{
				ObjectMeta: metav1.ObjectMeta{Namespace: "app"},
			},
			ep:   &egv1a1.EnvoyProxy{Spec: egv1a1.EnvoyProxySpec{}},
			want: nil,
		},
		{
			// exclude label configured but absent on the policy -> falls through to the default.
			name: "exclude label configured but not on policy applies default",
			pol: &egv1a1.BackendTrafficPolicy{
				ObjectMeta: metav1.ObjectMeta{Namespace: "app", Labels: map[string]string{"other": "x"}},
			},
			ep: &egv1a1.EnvoyProxy{Spec: egv1a1.EnvoyProxySpec{
				PolicyDefaults: &egv1a1.PolicyDefaults{
					BackendTrafficPolicy: &egv1a1.BackendTrafficPolicyDefaults{MergeSettings: egv1a1.MergeSettings{
						MergeType:         &strategic,
						MergeExcludeLabel: new("skip"),
					}},
				},
			}},
			want: &strategic,
		},
		{
			// nil labels map with an exclude label configured -> default still applies.
			name: "nil labels with exclude label configured applies default",
			pol: &egv1a1.BackendTrafficPolicy{
				ObjectMeta: metav1.ObjectMeta{Namespace: "app"},
			},
			ep: &egv1a1.EnvoyProxy{Spec: egv1a1.EnvoyProxySpec{
				PolicyDefaults: &egv1a1.PolicyDefaults{
					BackendTrafficPolicy: &egv1a1.BackendTrafficPolicyDefaults{MergeSettings: egv1a1.MergeSettings{
						MergeType:         &strategic,
						MergeExcludeLabel: new("skip"),
					}},
				},
			}},
			want: &strategic,
		},
		{
			// Defense in depth: a non-merge value (e.g. Replace, which can slip through the
			// unvalidated EnvoyGateway default spec) is ignored instead of producing a "merged"
			// status while actually replacing.
			name: "replace default is ignored",
			pol: &egv1a1.BackendTrafficPolicy{
				ObjectMeta: metav1.ObjectMeta{Namespace: "app"},
			},
			ep: &egv1a1.EnvoyProxy{Spec: egv1a1.EnvoyProxySpec{
				PolicyDefaults: &egv1a1.PolicyDefaults{
					BackendTrafficPolicy: &egv1a1.BackendTrafficPolicyDefaults{MergeSettings: egv1a1.MergeSettings{
						MergeType: new(egv1a1.Replace),
					}},
				},
			}},
			want: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tr.effectiveMergeType(tt.pol, tt.ep)
			if tt.want == nil {
				assert.Nil(t, got)
				return
			}
			assert.NotNil(t, got)
			assert.Equal(t, *tt.want, *got)
		})
	}
}

// TestAnyGatewayMergeDefault covers anyGatewayMergeDefault, which reports whether any of a route's
// parent gateways supplies a default mergeType for the policy via its EnvoyProxy.
func TestAnyGatewayMergeDefault(t *testing.T) {
	strategic := egv1a1.StrategicMerge

	epWithDefault := &egv1a1.EnvoyProxy{Spec: egv1a1.EnvoyProxySpec{
		PolicyDefaults: &egv1a1.PolicyDefaults{BackendTrafficPolicy: &egv1a1.BackendTrafficPolicyDefaults{MergeSettings: egv1a1.MergeSettings{MergeType: &strategic}}},
	}}
	epNoDefault := &egv1a1.EnvoyProxy{Spec: egv1a1.EnvoyProxySpec{}}

	// parentWith builds a parent context whose listeners reference gateways carrying the given
	// EnvoyProxies (one listener per EnvoyProxy).
	parentWith := func(eps ...*egv1a1.EnvoyProxy) *RouteParentContext {
		p := &RouteParentContext{}
		for _, ep := range eps {
			p.listeners = append(p.listeners, &ListenerContext{
				gateway: &GatewayContext{envoyProxy: ep},
			})
		}
		return p
	}

	policy := &egv1a1.BackendTrafficPolicy{
		ObjectMeta: metav1.ObjectMeta{Namespace: "app"},
	}

	tr := &Translator{ControllerNamespace: "eg"}

	tests := []struct {
		name    string
		parents []*RouteParentContext
		want    bool
	}{
		{"no parents", nil, false},
		{"parent with no listeners", []*RouteParentContext{parentWith()}, false},
		{"single parent supplies default", []*RouteParentContext{parentWith(epWithDefault)}, true},
		{"single parent no default", []*RouteParentContext{parentWith(epNoDefault)}, false},
		{"mixed listeners on one parent", []*RouteParentContext{parentWith(epNoDefault, epWithDefault)}, true},
		{
			"multiple parents one supplies default",
			[]*RouteParentContext{parentWith(epNoDefault), parentWith(epWithDefault)},
			true,
		},
		{
			"multiple parents none supply default",
			[]*RouteParentContext{parentWith(epNoDefault), parentWith(nil)},
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tr.anyGatewayMergeDefault(policy, tt.parents))
		})
	}
}

// TestApplyTrafficFeatureToRoute_MergeGatewayScoping covers the MergeGateways scoping in
// applyTrafficFeatureToRoute for the TCP and UDP listener loops: a listener belonging to a
// different Gateway must be skipped, so a defaulted merged policy does not bleed across Gateways
// that share one IR. The HTTP path is covered by the merged-gateways golden fixture.
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
